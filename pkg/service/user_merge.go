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
//   - compute_cluster_users
//   - projects.project_pi_id
//   - compute_allocation_memberships
//
// Left in place (who actually did the thing):
//   - compute_allocation_change_requests (requester / approver)
//   - compute_allocation_usages
//
// Not idempotent — a retiring user already in status=MERGED is rejected with
// ErrAlreadyExists. Callers that need replay safety should fetch the retiring
// user by ID and skip the call when its status is already MERGED.
func (s *Service) MergeUsers(ctx context.Context, survivingID, retiringID string) (*models.User, error) {
	if survivingID == "" || retiringID == "" {
		return nil, fmt.Errorf("%w: surviving and retiring user IDs are required", ErrInvalidInput)
	}
	if survivingID == retiringID {
		return nil, fmt.Errorf("%w: cannot merge a user with itself", ErrInvalidInput)
	}

	survivor, err := s.users.FindByID(ctx, survivingID)
	if err != nil {
		return nil, fmt.Errorf("lookup surviving user: %w", err)
	}
	if survivor == nil {
		return nil, fmt.Errorf("%w: surviving user %q does not exist", ErrInvalidInput, survivingID)
	}
	if survivor.Status != models.UserActive {
		return nil, fmt.Errorf("%w: surviving user %q must be ACTIVE (got %s)",
			ErrInvalidInput, survivingID, survivor.Status)
	}
	retiring, err := s.users.FindByID(ctx, retiringID)
	if err != nil {
		return nil, fmt.Errorf("lookup retiring user: %w", err)
	}
	if retiring == nil {
		return nil, fmt.Errorf("%w: retiring user %q does not exist", ErrInvalidInput, retiringID)
	}
	if retiring.Status == models.UserMerged {
		return nil, fmt.Errorf("%w: retiring user %q is already merged", ErrAlreadyExists, retiringID)
	}
	if retiring.Status != models.UserActive {
		return nil, fmt.Errorf("%w: retiring user %q must be ACTIVE (got %s)",
			ErrInvalidInput, retiringID, retiring.Status)
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.userIdentities.ReassignUser(ctx, tx, retiringID, survivingID); err != nil {
			return fmt.Errorf("reassign user identities: %w", err)
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
		return s.users.UpdateStatus(ctx, tx, retiringID, models.UserMerged)
	}); err != nil {
		return nil, fmt.Errorf("merge users: %w", err)
	}

	retiring.Status = models.UserMerged
	s.eventBus.Publish(events.UserUpdateEvent, retiring)
	s.eventBus.Publish(events.UserUpdateEvent, survivor)
	return survivor, nil
}
