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

// CreateComputeAllocationMembershipResourceOverride records a per-resource
// override of the SU amount granted to the membership. The referenced
// membership and resource must exist, and the (membership, resource) pair
// must not already have an override.
func (s *Service) CreateComputeAllocationMembershipResourceOverride(ctx context.Context, o *models.ComputeAllocationMembershipResourceOverride) (*models.ComputeAllocationMembershipResourceOverride, error) {
	if o == nil {
		return nil, fmt.Errorf("%w: membership resource override is nil", ErrInvalidInput)
	}
	if o.ComputeAllocationMembershipID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_membership_id is required", ErrInvalidInput)
	}
	if o.ComputeAllocationResourceID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_resource_id is required", ErrInvalidInput)
	}
	if o.OverrideResourceAmount < 0 {
		return nil, fmt.Errorf("%w: override_resource_amount must be non-negative", ErrInvalidInput)
	}
	if o.OverrideResourceTime < 0 {
		return nil, fmt.Errorf("%w: override_resource_time must be non-negative", ErrInvalidInput)
	}

	if m, err := s.memberships.FindByID(ctx, o.ComputeAllocationMembershipID); err != nil {
		return nil, fmt.Errorf("lookup compute allocation membership: %w", err)
	} else if m == nil {
		return nil, fmt.Errorf("%w: compute allocation membership %q not found",
			ErrInvalidInput, o.ComputeAllocationMembershipID)
	}
	if r, err := s.resources.FindByID(ctx, o.ComputeAllocationResourceID); err != nil {
		return nil, fmt.Errorf("lookup compute allocation resource: %w", err)
	} else if r == nil {
		return nil, fmt.Errorf("%w: compute allocation resource %q not found",
			ErrInvalidInput, o.ComputeAllocationResourceID)
	}

	if existing, err := s.membershipOverrides.FindByPair(ctx, o.ComputeAllocationMembershipID, o.ComputeAllocationResourceID); err != nil {
		return nil, fmt.Errorf("lookup membership resource override: %w", err)
	} else if existing != nil {
		return nil, fmt.Errorf("%w: override already exists for membership %q resource %q",
			ErrAlreadyExists, o.ComputeAllocationMembershipID, o.ComputeAllocationResourceID)
	}

	if o.ID == "" {
		o.ID = newID()
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.membershipOverrides.Create(ctx, tx, o)
	}); err != nil {
		return nil, fmt.Errorf("create membership resource override: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationMembershipResourceOverrideCreateEvent, o)
	return o, nil
}

// GetComputeAllocationMembershipResourceOverride retrieves an override by ID.
func (s *Service) GetComputeAllocationMembershipResourceOverride(ctx context.Context, id string) (*models.ComputeAllocationMembershipResourceOverride, error) {
	o, err := s.membershipOverrides.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get membership resource override: %w", err)
	}
	if o == nil {
		return nil, ErrNotFound
	}
	return o, nil
}

// GetComputeAllocationMembershipResourceOverrideByPair retrieves the override
// for the given (membership, resource) pair.
func (s *Service) GetComputeAllocationMembershipResourceOverrideByPair(ctx context.Context, membershipID, resourceID string) (*models.ComputeAllocationMembershipResourceOverride, error) {
	if membershipID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_membership_id is required", ErrInvalidInput)
	}
	if resourceID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_resource_id is required", ErrInvalidInput)
	}
	o, err := s.membershipOverrides.FindByPair(ctx, membershipID, resourceID)
	if err != nil {
		return nil, fmt.Errorf("get membership resource override by pair: %w", err)
	}
	if o == nil {
		return nil, ErrNotFound
	}
	return o, nil
}

// ListOverridesForMembership returns every resource override recorded
// against the given membership.
func (s *Service) ListOverridesForMembership(ctx context.Context, membershipID string) ([]models.ComputeAllocationMembershipResourceOverride, error) {
	if membershipID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_membership_id is required", ErrInvalidInput)
	}
	rows, err := s.membershipOverrides.FindByMembership(ctx, membershipID)
	if err != nil {
		return nil, fmt.Errorf("list overrides for membership: %w", err)
	}
	return rows, nil
}

// ListOverridesForResource returns every membership override referencing the
// given resource.
func (s *Service) ListOverridesForResource(ctx context.Context, resourceID string) ([]models.ComputeAllocationMembershipResourceOverride, error) {
	if resourceID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_resource_id is required", ErrInvalidInput)
	}
	rows, err := s.membershipOverrides.FindByResource(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("list overrides for resource: %w", err)
	}
	return rows, nil
}

// UpdateComputeAllocationMembershipResourceOverride replaces mutable fields
// of an existing override.
func (s *Service) UpdateComputeAllocationMembershipResourceOverride(ctx context.Context, o *models.ComputeAllocationMembershipResourceOverride) (*models.ComputeAllocationMembershipResourceOverride, error) {
	if o == nil || o.ID == "" {
		return nil, fmt.Errorf("%w: membership resource override id is required", ErrInvalidInput)
	}
	if o.OverrideResourceAmount < 0 {
		return nil, fmt.Errorf("%w: override_resource_amount must be non-negative", ErrInvalidInput)
	}
	if o.OverrideResourceTime < 0 {
		return nil, fmt.Errorf("%w: override_resource_time must be non-negative", ErrInvalidInput)
	}
	existing, err := s.membershipOverrides.FindByID(ctx, o.ID)
	if err != nil {
		return nil, fmt.Errorf("lookup membership resource override: %w", err)
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if o.ComputeAllocationMembershipID == "" {
		o.ComputeAllocationMembershipID = existing.ComputeAllocationMembershipID
	}
	if o.ComputeAllocationResourceID == "" {
		o.ComputeAllocationResourceID = existing.ComputeAllocationResourceID
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.membershipOverrides.Update(ctx, tx, o)
	}); err != nil {
		return nil, fmt.Errorf("update membership resource override: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationMembershipResourceOverrideUpdateEvent, o)
	return o, nil
}

// DeleteComputeAllocationMembershipResourceOverride removes an override by ID.
func (s *Service) DeleteComputeAllocationMembershipResourceOverride(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: membership resource override id is required", ErrInvalidInput)
	}
	existing, err := s.membershipOverrides.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup membership resource override: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.membershipOverrides.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete membership resource override: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationMembershipResourceOverrideDeleteEvent, existing)
	return nil
}
