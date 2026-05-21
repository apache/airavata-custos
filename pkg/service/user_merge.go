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
// place. The retiring user is flipped to status=MERGED and a row is written
// to user_merges with the surviving user and the given reason. All work
// happens in a single transaction.
//
// Moved to survivor (duplicates on the retiring user are dropped first):
//   - user_identities
//   - user_dns
//   - compute_cluster_users
//   - projects.project_pi_id
//   - compute_allocation_memberships
//
// Left in place (who actually did the thing):
//   - compute_allocation_change_requests (requester / approver)
//   - compute_allocation_usages
func (s *Service) MergeUsers(ctx context.Context, survivingID, retiringID, reason string) (*models.User, error) {
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
	if survivor.Status == models.UserMerged {
		return nil, fmt.Errorf("%w: surviving user %q is itself merged", ErrInvalidInput, survivingID)
	}
	retiring, err := s.users.FindByID(ctx, retiringID)
	if err != nil {
		return nil, fmt.Errorf("lookup retiring user: %w", err)
	}
	if retiring == nil {
		return nil, fmt.Errorf("%w: retiring user %q does not exist", ErrInvalidInput, retiringID)
	}

	// Idempotency: re-running the same merge is a no-op; merging the same
	// retiring user into a different survivor is rejected.
	if retiring.Status == models.UserMerged {
		prior, err := s.userMerges.FindByRetiringUser(ctx, retiringID)
		if err != nil {
			return nil, fmt.Errorf("lookup prior merge: %w", err)
		}
		if prior != nil {
			if prior.SurvivingUserID == survivingID {
				return survivor, nil
			}
			return nil, fmt.Errorf("%w: user %q already merged into %q",
				ErrAlreadyExists, retiringID, prior.SurvivingUserID)
		}
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.userIdentities.ReassignUser(ctx, tx, retiringID, survivingID); err != nil {
			return fmt.Errorf("reassign user identities: %w", err)
		}
		if err := s.userDNs.ReassignUser(ctx, tx, retiringID, survivingID); err != nil {
			return fmt.Errorf("reassign user dns: %w", err)
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
		if err := s.users.UpdateStatus(ctx, tx, retiringID, models.UserMerged); err != nil {
			return fmt.Errorf("mark retiring user merged: %w", err)
		}
		return s.userMerges.Record(ctx, tx, retiringID, survivingID, reason)
	}); err != nil {
		return nil, fmt.Errorf("merge users: %w", err)
	}

	retiring.Status = models.UserMerged
	s.eventBus.Publish(events.UserUpdateEvent, retiring)
	s.eventBus.Publish(events.UserUpdateEvent, survivor)
	return survivor, nil
}

// GetUserMergeByRetiringUser returns the merge record for a retiring user, or
// ErrNotFound if the user has not been merged.
func (s *Service) GetUserMergeByRetiringUser(ctx context.Context, retiringUserID string) (*models.UserMerge, error) {
	if retiringUserID == "" {
		return nil, fmt.Errorf("%w: retiring_user_id is required", ErrInvalidInput)
	}
	m, err := s.userMerges.FindByRetiringUser(ctx, retiringUserID)
	if err != nil {
		return nil, fmt.Errorf("get user merge: %w", err)
	}
	if m == nil {
		return nil, ErrNotFound
	}
	return m, nil
}

// ListUserMergesBySurvivingUser returns every merge record absorbed by the
// given surviving user, oldest first.
func (s *Service) ListUserMergesBySurvivingUser(ctx context.Context, survivingUserID string) ([]models.UserMerge, error) {
	if survivingUserID == "" {
		return nil, fmt.Errorf("%w: surviving_user_id is required", ErrInvalidInput)
	}
	out, err := s.userMerges.FindBySurvivingUser(ctx, survivingUserID)
	if err != nil {
		return nil, fmt.Errorf("list user merges by surviving user: %w", err)
	}
	return out, nil
}
