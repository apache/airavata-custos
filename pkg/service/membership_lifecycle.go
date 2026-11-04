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

// Core alone decides when a membership grants access: it and its allocation are ACTIVE, the scheduler has the
// allocation's account, and the member's account on the allocation's cluster is provisioned; ready_at records it.
// Every ready_at write goes through syncMembershipTx, which publishes ready or deactivate when the answer changes.
// Writes to a membership's allocation or account resync it through syncMembershipsTx, in the same transaction.
// A membership create share-locks its user and allocation, and syncMembershipsTx locks them for update, so the two
// serialise and neither misses the other's write.

package service

import (
	"context"
	"database/sql"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// syncMembershipTx updates the membership's ready_at and publishes ready or deactivate when it changes.
func (s *Service) syncMembershipTx(ctx context.Context, tx *sql.Tx, membershipID string) error {
	m, alloc, ready, err := s.eligibilityTx(ctx, tx, membershipID)
	if err != nil || m == nil || ready == (m.ReadyAt != nil) {
		return err
	}
	if !ready {
		return s.deactivateMembershipTx(ctx, tx, m)
	}
	return s.readyMembershipTx(ctx, tx, m, alloc)
}

// readyMembershipTx sets ready_at and publishes ready.
func (s *Service) readyMembershipTx(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationMembership, a *models.ComputeAllocation) error {
	now := nowUTC()
	m.ReadyAt = &now
	if err := s.memberships.SetReady(ctx, tx, m.ID, m.ReadyAt); err != nil {
		return err
	}
	if err := s.eventBus.Publish(ctx, tx, events.ComputeAllocationMembershipReadyEvent, m); err != nil {
		return err
	}
	return s.requestStorageTx(ctx, tx, m, a)
}

// syncMembershipsTx syncs the memberships matching f after a write to their allocation or account.
func (s *Service) syncMembershipsTx(ctx context.Context, tx *sql.Tx, f store.MembershipFilter) error {
	if f.UserID != "" {
		if _, err := s.users.FindByIDTx(ctx, tx, f.UserID, store.UpdateLock); err != nil {
			return err
		}
	}
	if f.AllocationID != "" {
		if _, err := s.allocs.FindByIDTx(ctx, tx, f.AllocationID, store.UpdateLock); err != nil {
			return err
		}
	}
	ms, err := s.memberships.FindTx(ctx, tx, f)
	if err != nil {
		return err
	}
	for _, m := range ms {
		if err := s.syncMembershipTx(ctx, tx, m.ID); err != nil {
			return err
		}
	}
	return nil
}

// lockMembershipPairTx share-locks the user and allocation a membership is about to reference.
func (s *Service) lockMembershipPairTx(ctx context.Context, tx *sql.Tx, allocationID, userID string) error {
	if _, err := s.users.FindByIDTx(ctx, tx, userID, store.ShareLock); err != nil {
		return err
	}
	_, err := s.allocs.FindByIDTx(ctx, tx, allocationID, store.ShareLock)
	return err
}

// eligibilityTx locks the membership and reports whether it grants access; nil when it or its allocation is gone.
func (s *Service) eligibilityTx(ctx context.Context, tx *sql.Tx, membershipID string) (*models.ComputeAllocationMembership, *models.ComputeAllocation, bool, error) {
	m, err := s.memberships.FindByIDTx(ctx, tx, membershipID)
	if err != nil || m == nil {
		return nil, nil, false, err
	}
	alloc, err := s.allocs.FindByIDTx(ctx, tx, m.ComputeAllocationID, store.NoLock)
	if err != nil || alloc == nil {
		return nil, nil, false, err
	}
	account, err := s.clusterUsers.FindByPairTx(ctx, tx, alloc.ComputeClusterID, m.UserID)
	if err != nil {
		return nil, nil, false, err
	}
	return m, alloc, m.MembershipStatus == models.ACTIVE && alloc.Status == models.ACTIVE && alloc.SchedulerProvisionedAt != nil &&
		account != nil && account.ProvisionedAt != nil, nil
}

// deactivateMembershipTx clears ready_at and publishes deactivate with m as stored.
func (s *Service) deactivateMembershipTx(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationMembership) error {
	m.ReadyAt = nil
	if err := s.memberships.SetReady(ctx, tx, m.ID, nil); err != nil {
		return err
	}
	return s.eventBus.Publish(ctx, tx, events.ComputeAllocationMembershipDeactivateEvent, m)
}
