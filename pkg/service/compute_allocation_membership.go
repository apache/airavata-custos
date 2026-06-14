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

// CreateComputeAllocationMembership grants a user a sub-allocation against a
// compute allocation. Both the referenced allocation and user must exist, and
// the (allocation, user) pair must not already have a membership.
// MembershipStatus defaults to ACTIVE when unset.
func (s *Service) CreateComputeAllocationMembership(ctx context.Context, m *models.ComputeAllocationMembership) (*models.ComputeAllocationMembership, error) {
	if m == nil {
		return nil, fmt.Errorf("%w: compute allocation membership is nil", ErrInvalidInput)
	}
	if m.ComputeAllocationID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_id is required", ErrInvalidInput)
	}
	if m.UserID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}

	if alloc, err := s.allocs.FindByID(ctx, m.ComputeAllocationID); err != nil {
		return nil, fmt.Errorf("lookup compute allocation: %w", err)
	} else if alloc == nil {
		return nil, fmt.Errorf("%w: compute allocation %q not found", ErrInvalidInput, m.ComputeAllocationID)
	}
	if u, err := s.users.FindByID(ctx, m.UserID); err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	} else if u == nil {
		return nil, fmt.Errorf("%w: user %q not found", ErrInvalidInput, m.UserID)
	}

	if existing, err := s.memberships.FindByPair(ctx, m.ComputeAllocationID, m.UserID); err != nil {
		return nil, fmt.Errorf("lookup compute allocation membership: %w", err)
	} else if existing != nil {
		return nil, fmt.Errorf("%w: user %q already has a membership on allocation %q", ErrAlreadyExists, m.UserID, m.ComputeAllocationID)
	}

	if m.ID == "" {
		m.ID = newID()
	}
	if m.MembershipStatus == "" {
		m.MembershipStatus = models.ACTIVE
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.memberships.Create(ctx, tx, m)
	}); err != nil {
		return nil, fmt.Errorf("create compute allocation membership: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationMembershipCreateEvent, m)
	return m, nil
}

// GetComputeAllocationMembership retrieves a membership by its ID. Returns
// ErrNotFound when no membership matches.
func (s *Service) GetComputeAllocationMembership(ctx context.Context, id string) (*models.ComputeAllocationMembership, error) {
	m, err := s.memberships.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get compute allocation membership: %w", err)
	}
	if m == nil {
		return nil, ErrNotFound
	}
	return m, nil
}

// ListMembersForAllocation returns every membership recorded against the
// given allocation, ordered by start_time ascending.
func (s *Service) ListMembersForAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationMembership, error) {
	if allocationID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_id is required", ErrInvalidInput)
	}
	rows, err := s.memberships.FindByAllocation(ctx, allocationID)
	if err != nil {
		return nil, fmt.Errorf("list members for allocation: %w", err)
	}
	return rows, nil
}

// ListAllocationsForUser returns every membership held by the given user,
// ordered by start_time ascending.
func (s *Service) ListAllocationsForUser(ctx context.Context, userID string) ([]models.ComputeAllocationMembership, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	rows, err := s.memberships.FindByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list allocations for user: %w", err)
	}
	return rows, nil
}

// UpdateComputeAllocationMembership replaces mutable fields of an existing
// membership. Fields left blank/zero on the supplied record fall back to the
// stored value (so callers can perform partial updates).
func (s *Service) UpdateComputeAllocationMembership(ctx context.Context, m *models.ComputeAllocationMembership) (*models.ComputeAllocationMembership, error) {
	if m == nil || m.ID == "" {
		return nil, fmt.Errorf("%w: compute allocation membership id is required", ErrInvalidInput)
	}
	existing, err := s.memberships.FindByID(ctx, m.ID)
	if err != nil {
		return nil, fmt.Errorf("lookup compute allocation membership: %w", err)
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if m.ComputeAllocationID == "" {
		m.ComputeAllocationID = existing.ComputeAllocationID
	}
	if m.UserID == "" {
		m.UserID = existing.UserID
	}
	if m.StartTime.IsZero() {
		m.StartTime = existing.StartTime
	}
	if m.EndTime.IsZero() {
		m.EndTime = existing.EndTime
	}
	if m.MembershipStatus == "" {
		m.MembershipStatus = existing.MembershipStatus
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.memberships.Update(ctx, tx, m)
	}); err != nil {
		return nil, fmt.Errorf("update compute allocation membership: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationMembershipUpdateEvent, m)
	return m, nil
}

// UpdateMembershipStatus sets the lifecycle status (ACTIVE, INACTIVE, etc.)
// of the membership identified by the given ID. Other fields are preserved.
func (s *Service) UpdateMembershipStatus(ctx context.Context, id string, status models.AllocationStatus) (*models.ComputeAllocationMembership, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: compute allocation membership id is required", ErrInvalidInput)
	}
	if status == "" {
		return nil, fmt.Errorf("%w: membership_status is required", ErrInvalidInput)
	}
	existing, err := s.memberships.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("lookup compute allocation membership: %w", err)
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	existing.MembershipStatus = status
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.memberships.Update(ctx, tx, existing)
	}); err != nil {
		return nil, fmt.Errorf("update compute allocation membership status: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationMembershipUpdateEvent, existing)
	return existing, nil
}

// DeleteComputeAllocationMembership removes a membership by ID.
func (s *Service) DeleteComputeAllocationMembership(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: compute allocation membership id is required", ErrInvalidInput)
	}
	existing, err := s.memberships.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup compute allocation membership: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.memberships.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete compute allocation membership: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationMembershipDeleteEvent, existing)
	return nil
}
