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
	"slices"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// requestIdentityTx asks the identity registry to create or relink each approved account.
func (s *Service) requestIdentityTx(ctx context.Context, tx *sql.Tx, accounts ...models.ComputeClusterUser) error {
	for _, cu := range accounts {
		if cu.ApprovalStatus != models.ClusterAccountApproved {
			continue
		}
		if err := s.requestTx(ctx, tx, events.ComputeClusterUserIdentityRequestEvent, cu, func() error {
			return s.markProvisionedTx(ctx, tx, &cu)
		}); err != nil {
			return err
		}
	}
	return nil
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
	return s.requestIdentityTx(ctx, tx, *cu)
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
			return s.groupWithoutRegistryTx(ctx, tx, a)
		}); err != nil {
			return err
		}
	}
	return nil
}

// groupWithoutRegistryTx marks the group provisioned once it has a member, as the registry would on its first sync.
func (s *Service) groupWithoutRegistryTx(ctx context.Context, tx *sql.Tx, a *models.ComputeAllocation) error {
	ok, err := s.memberships.HasProvisionedMemberTx(ctx, tx, a.ID)
	if err != nil || !ok {
		return err
	}
	return s.markGroupProvisionedTx(ctx, tx, a)
}

// requestMemberGroupsTx resyncs the groups of the allocations holding the memberships matching f.
func (s *Service) requestMemberGroupsTx(ctx context.Context, tx *sql.Tx, f store.MembershipFilter) error {
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

// requestStorageTx asks for a ready member's directories once its allocation group is provisioned.
func (s *Service) requestStorageTx(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationMembership, a *models.ComputeAllocation) error {
	if m.ReadyAt == nil || a.GroupProvisionedAt == nil {
		return nil
	}
	return s.eventBus.Publish(ctx, tx, events.ComputeAllocationMembershipStorageRequestEvent, m)
}

// requestSchedulerTx asks for the allocation's scheduler account.
func (s *Service) requestSchedulerTx(ctx context.Context, tx *sql.Tx, a *models.ComputeAllocation) error {
	return s.requestTx(ctx, tx, events.ComputeAllocationSchedulerRequestEvent, a, func() error {
		return s.markSchedulerProvisionedTx(ctx, tx, a.ID)
	})
}

// requestAccountStorageTx asks for the account's own directories.
func (s *Service) requestAccountStorageTx(ctx context.Context, tx *sql.Tx, cu *models.ComputeClusterUser) error {
	return s.requestTx(ctx, tx, events.ComputeClusterUserStorageRequestEvent, cu, func() error {
		return s.markStorageProvisionedTx(ctx, tx, cu.ID)
	})
}

// readyAccountTx publishes ready the first time the account's identity and storage are both done.
func (s *Service) readyAccountTx(ctx context.Context, tx *sql.Tx, id string) error {
	firstTime, err := s.clusterUsers.MarkReady(ctx, tx, id, nowUTC())
	if err != nil || !firstTime {
		return err
	}
	cu, err := s.clusterUsers.FindByIDTx(ctx, tx, id)
	if err != nil || cu == nil {
		return err
	}
	return s.eventBus.Publish(ctx, tx, events.ComputeClusterUserReadyEvent, cu)
}
