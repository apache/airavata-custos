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

package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// MergeUsers consolidates the retiring user into the surviving user. All
// identity-forward state moves to the survivor; historical truth stays in
// place. The retiring user is flipped to status=MERGED. All work happens in a
// single transaction.
//
// Moved to survivor (duplicates on the retiring user are dropped first):
//   - user_identities
//   - compute_cluster_users (on a shared cluster the provisioned one wins; if both are, both stay)
//   - projects.project_pi_id
//   - compute_allocation_memberships
//
// A dropped cluster account publishes compute_cluster_user::delete, so it is
// also removed outside Custos.
//
// Left in place (who actually did the thing):
//   - compute_allocation_change_requests (requester / approver)
//   - compute_allocation_usages
//
// Not idempotent — a retiring user already in status=MERGED is rejected with
// ErrAlreadyExists. Callers that need replay safety should fetch the retiring
// user by ID and skip the call when its status is already MERGED.
func (s *Service) MergeUsers(ctx context.Context, survivingID, retiringID string) (survivor *models.User, err error) {
	err = s.inTx(ctx, func(tx *sql.Tx) (err error) {
		if _, err = s.MergeUsersTx(ctx, tx, survivingID, retiringID); err == nil {
			survivor, err = s.users.FindByIDTx(ctx, tx, survivingID, store.NoLock)
		}
		return
	})
	return
}

func (s *Service) MergeUsersTx(ctx context.Context, tx *sql.Tx, survivingID, retiringID string) ([]models.ComputeClusterUser, error) {
	if survivingID == "" || retiringID == "" {
		return nil, fmt.Errorf("%w: surviving and retiring user IDs are required", ErrInvalidInput)
	}
	if survivingID == retiringID {
		return nil, fmt.Errorf("%w: cannot merge a user with itself", ErrInvalidInput)
	}

	survivor, err := s.users.FindByIDTx(ctx, tx, survivingID, store.NoKeyUpdateLock)
	if err != nil {
		return nil, fmt.Errorf("lookup surviving user: %w", err)
	}
	if survivor == nil {
		return nil, fmt.Errorf("%w: surviving user %q does not exist", ErrInvalidInput, survivingID)
	}
	if !canBeMerged(survivor.Status) {
		return nil, fmt.Errorf("%w: surviving user %q must be ACTIVE or PENDING (got %s)",
			ErrInvalidInput, survivingID, survivor.Status)
	}
	retiring, err := s.users.FindByIDTx(ctx, tx, retiringID, store.NoKeyUpdateLock)
	if err != nil {
		return nil, fmt.Errorf("lookup retiring user: %w", err)
	}
	if retiring == nil {
		return nil, fmt.Errorf("%w: retiring user %q does not exist", ErrInvalidInput, retiringID)
	}
	if retiring.Status == models.UserMerged {
		return nil, fmt.Errorf("%w: retiring user %q is already merged", ErrAlreadyExists, retiringID)
	}
	if !canBeMerged(retiring.Status) {
		return nil, fmt.Errorf("%w: retiring user %q must be ACTIVE or PENDING (got %s)",
			ErrInvalidInput, retiringID, retiring.Status)
	}

	retiringAccounts, err := s.clusterUsers.FindByUserTx(ctx, tx, retiringID)
	if err != nil {
		return nil, fmt.Errorf("list retiring cluster users: %w", err)
	}
	survivorAccounts, err := s.clusterUsers.FindByUserTx(ctx, tx, survivingID)
	if err != nil {
		return nil, fmt.Errorf("list surviving cluster users: %w", err)
	}
	survivorClusters := make(map[string]models.ComputeClusterUser, len(survivorAccounts))
	for _, cu := range survivorAccounts {
		survivorClusters[cu.ComputeClusterID] = cu
	}
	var dropped []models.ComputeClusterUser
	for _, cu := range retiringAccounts {
		theirs, clash := survivorClusters[cu.ComputeClusterID]
		switch {
		case clash && cu.ProvisionedAt == nil:
			dropped = append(dropped, cu)
		case clash && theirs.ProvisionedAt == nil:
			dropped = append(dropped, theirs)
		}
	}

	if err := func() error {
		if err := s.userIdentities.ReassignUser(ctx, tx, retiringID, survivingID); err != nil {
			return fmt.Errorf("reassign user identities: %w", err)
		}
		for _, cu := range dropped {
			if err := s.clusterUsers.Delete(ctx, tx, cu.ID); err != nil {
				return fmt.Errorf("delete compute cluster user: %w", err)
			}
			if err := s.eventBus.Publish(ctx, tx, events.ComputeClusterUserDeleteEvent, cu); err != nil {
				return err
			}
		}
		if err := s.clusterUsers.ReassignUser(ctx, tx, retiringID, survivingID); err != nil {
			return fmt.Errorf("reassign compute cluster users: %w", err)
		}
		if err := s.projs.ReassignPI(ctx, tx, retiringID, survivingID); err != nil {
			return fmt.Errorf("reassign project PI: %w", err)
		}
		if err := s.memberships.ReassignUser(ctx, tx, retiringID, survivingID); err != nil {
			return fmt.Errorf("reassign memberships: %w", err)
		}
		if err := s.projMemberships.ReassignUser(ctx, tx, retiringID, survivingID); err != nil {
			return fmt.Errorf("reassign project memberships: %w", err)
		}
		if err := s.users.UpdateStatus(ctx, tx, retiringID, models.UserMerged); err != nil {
			return err
		}
		retiring.Status = models.UserMerged
		if err := s.eventBus.Publish(ctx, tx, events.UserUpdateEvent, retiring); err != nil {
			return err
		}
		return s.eventBus.Publish(ctx, tx, events.UserUpdateEvent, survivor)
	}(); err != nil {
		return nil, fmt.Errorf("merge users: %w", err)
	}
	s.InvalidateAllIdentities()
	return dropped, nil
}

// canBeMerged returns true for the user statuses eligible for merging:
// ACTIVE (already linked) or PENDING (provisioned upstream, not yet linked).
// INACTIVE, SUSPENDED, and MERGED are intentionally excluded.
func canBeMerged(s models.UserStatus) bool {
	return s == models.UserActive || s == models.UserPending
}
