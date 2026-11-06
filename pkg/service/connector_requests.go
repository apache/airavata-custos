// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Core decides when a connector has work. Each request is published in the transaction of the write that causes it,
// and the connector reports completion through a core mark, which publishes the next event. Connectors handle one
// entity per request and never list core state to find more work. A request nobody subscribes to is not awaited:
// core does what its mark does itself, so a connector that is off never holds the state machine back.

package service

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// requestIdentityTx asks the identity registry to create or relink an approved account.
func (s *Service) requestIdentityTx(ctx context.Context, tx *sql.Tx, cu *models.ComputeClusterUser) error {
	if cu.ApprovalStatus != models.ClusterAccountApproved {
		return nil
	}
	return s.requestTx(ctx, tx, events.ComputeClusterUserIdentityRequestEvent, cu, func() error {
		return s.markProvisionedTx(ctx, tx, cu)
	})
}

// requestTx publishes a request to its connector, or runs done, the connector's mark, when nobody subscribes.
func (s *Service) requestTx(ctx context.Context, tx *sql.Tx, topic events.EventType, payload any, done func() error) error {
	if s.eventBus.Subscribed(topic) {
		return s.eventBus.Publish(ctx, tx, topic, payload)
	}
	return done()
}

// approveTx publishes the approval and asks for the account's identity.
func (s *Service) approveTx(ctx context.Context, tx *sql.Tx, cu *models.ComputeClusterUser) error {
	if err := s.eventBus.Publish(ctx, tx, events.ComputeClusterUserApproveEvent, cu); err != nil {
		return err
	}
	return s.requestIdentityTx(ctx, tx, cu)
}

// requestGroupsTx asks once per allocation for its group to be resynced.
func (s *Service) requestGroupsTx(ctx context.Context, tx *sql.Tx, allocationIDs ...string) error {
	for _, id := range slices.Compact(slices.Sorted(slices.Values(allocationIDs))) {
		a, err := s.allocs.FindByIDTx(ctx, tx, id, store.NoLock)
		if err != nil {
			return err
		}
		if a == nil {
			continue
		}
		if err := s.requestTx(ctx, tx, events.ComputeAllocationGroupRequestEvent, a, func() error {
			return s.markGroupProvisionedTx(ctx, tx, a)
		}); err != nil {
			return err
		}
	}
	return nil
}

// resyncMembersTx syncs the memberships matching f after a write to their account, then resyncs their allocations' groups.
func (s *Service) resyncMembersTx(ctx context.Context, tx *sql.Tx, f store.MembershipFilter) error {
	if err := s.syncMembershipsTx(ctx, tx, f); err != nil {
		return err
	}
	ms, err := s.memberships.FindTx(ctx, tx, f)
	if err != nil {
		return err
	}
	ids := make([]string, len(ms))
	for i, m := range ms {
		ids[i] = m.ComputeAllocationID
	}
	return s.requestGroupsTx(ctx, tx, ids...)
}

// SettleConnectors does each outstanding request whose topic nobody subscribes to.
// Subscriptions change only at startup, so core runs it once, after the connectors load.
func (s *Service) SettleConnectors(ctx context.Context) error {
	clusters, err := s.clusters.List(ctx)
	if err != nil {
		return err
	}
	off := func(t events.EventType) bool { return !s.eventBus.Subscribed(t) }
	var errs []error
	for _, c := range clusters {
		accounts, err := s.clusterUsers.FindByCluster(ctx, c.ID)
		if err != nil {
			return err
		}
		allocs, err := s.allocs.FindByCluster(ctx, c.ID)
		if err != nil {
			return err
		}
		for _, cu := range accounts {
			switch {
			case cu.ApprovalStatus == models.ClusterAccountApproved && cu.ProvisionedAt == nil && off(events.ComputeClusterUserIdentityRequestEvent):
				errs = append(errs, s.MarkComputeClusterUserProvisioned(ctx, cu.ID))
			case cu.ProvisionedAt != nil && cu.StorageProvisionedAt == nil && off(events.ComputeClusterUserStorageRequestEvent):
				errs = append(errs, s.MarkComputeClusterUserStorageProvisioned(ctx, cu.ID))
			}
		}
		for _, a := range allocs {
			if a.SchedulerProvisionedAt == nil && off(events.ComputeAllocationSchedulerRequestEvent) {
				errs = append(errs, s.MarkComputeAllocationSchedulerProvisioned(ctx, a.ID))
			}
			if a.GroupProvisionedAt == nil && off(events.ComputeAllocationGroupRequestEvent) {
				errs = append(errs, s.MarkComputeAllocationGroupProvisioned(ctx, a.ID))
			}
		}
	}
	return errors.Join(errs...)
}
