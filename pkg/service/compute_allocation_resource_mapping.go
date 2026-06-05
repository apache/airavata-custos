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

// AttachResourceToAllocation links a compute allocation resource to a compute
// allocation, recording the resource amount and wall-clock time granted to
// the allocation. Both entities must already exist. The link is idempotent —
// if the same (allocation, resource) pair is already linked, ErrAlreadyExists
// is returned.
func (s *Service) AttachResourceToAllocation(ctx context.Context, allocationID, resourceID string, resourceAmount, resourceTime int64) (*models.ComputeAllocationResourceMapping, error) {
	if allocationID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_id is required", ErrInvalidInput)
	}
	if resourceID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_resource_id is required", ErrInvalidInput)
	}
	if resourceAmount < 0 {
		return nil, fmt.Errorf("%w: resource_amount must be non-negative", ErrInvalidInput)
	}
	if resourceTime < 0 {
		return nil, fmt.Errorf("%w: resource_time must be non-negative", ErrInvalidInput)
	}

	if alloc, err := s.allocs.FindByID(ctx, allocationID); err != nil {
		return nil, fmt.Errorf("lookup compute allocation: %w", err)
	} else if alloc == nil {
		return nil, fmt.Errorf("%w: compute allocation %q not found", ErrInvalidInput, allocationID)
	}

	if res, err := s.resources.FindByID(ctx, resourceID); err != nil {
		return nil, fmt.Errorf("lookup compute allocation resource: %w", err)
	} else if res == nil {
		return nil, fmt.Errorf("%w: compute allocation resource %q not found", ErrInvalidInput, resourceID)
	}

	if existing, err := s.resourceMappings.FindByPair(ctx, allocationID, resourceID); err != nil {
		return nil, fmt.Errorf("lookup mapping: %w", err)
	} else if existing != nil {
		return nil, fmt.Errorf("%w: resource %q already attached to allocation %q", ErrAlreadyExists, resourceID, allocationID)
	}

	mapping := &models.ComputeAllocationResourceMapping{
		ID:                          newID(),
		ComputeAllocationID:         allocationID,
		ComputeAllocationResourceID: resourceID,
		ResourceAmount:              resourceAmount,
		ResourceTime:                resourceTime,
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.resourceMappings.Create(ctx, tx, mapping)
	}); err != nil {
		return nil, fmt.Errorf("attach resource to allocation: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationResourceMappingCreateEvent, mapping)
	return mapping, nil
}

// UpdateAllocationResourceMapping updates the resource_amount and
// resource_time recorded against an existing (allocation, resource) mapping.
func (s *Service) UpdateAllocationResourceMapping(ctx context.Context, allocationID, resourceID string, resourceAmount, resourceTime int64) (*models.ComputeAllocationResourceMapping, error) {
	if allocationID == "" || resourceID == "" {
		return nil, fmt.Errorf("%w: allocation and resource ids are required", ErrInvalidInput)
	}
	if resourceAmount < 0 {
		return nil, fmt.Errorf("%w: resource_amount must be non-negative", ErrInvalidInput)
	}
	if resourceTime < 0 {
		return nil, fmt.Errorf("%w: resource_time must be non-negative", ErrInvalidInput)
	}

	existing, err := s.resourceMappings.FindByPair(ctx, allocationID, resourceID)
	if err != nil {
		return nil, fmt.Errorf("lookup mapping: %w", err)
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	existing.ResourceAmount = resourceAmount
	existing.ResourceTime = resourceTime
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.resourceMappings.Update(ctx, tx, existing)
	}); err != nil {
		return nil, fmt.Errorf("update allocation resource mapping: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationResourceMappingUpdateEvent, existing)
	return existing, nil
}

// DetachResourceFromAllocation removes the link between a compute allocation
// and a compute allocation resource. Returns ErrNotFound when no such mapping
// exists.
func (s *Service) DetachResourceFromAllocation(ctx context.Context, allocationID, resourceID string) error {
	if allocationID == "" || resourceID == "" {
		return fmt.Errorf("%w: allocation and resource ids are required", ErrInvalidInput)
	}

	existing, err := s.resourceMappings.FindByPair(ctx, allocationID, resourceID)
	if err != nil {
		return fmt.Errorf("lookup mapping: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.resourceMappings.DeleteByPair(ctx, tx, allocationID, resourceID)
	}); err != nil {
		return fmt.Errorf("detach resource from allocation: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationResourceMappingDeleteEvent, existing)
	return nil
}

// ListResourcesForAllocation returns every compute allocation resource
// attached to the given compute allocation.
func (s *Service) ListResourcesForAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationResource, error) {
	if allocationID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_id is required", ErrInvalidInput)
	}
	resources, err := s.resourceMappings.FindResourcesByAllocation(ctx, allocationID)
	if err != nil {
		return nil, fmt.Errorf("list resources for allocation: %w", err)
	}
	return resources, nil
}

// ListAllocationsForResource returns every compute allocation that has the
// given compute allocation resource attached.
func (s *Service) ListAllocationsForResource(ctx context.Context, resourceID string) ([]models.ComputeAllocation, error) {
	if resourceID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_resource_id is required", ErrInvalidInput)
	}
	allocs, err := s.resourceMappings.FindAllocationsByResource(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("list allocations for resource: %w", err)
	}
	return allocs, nil
}
