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

	"github.com/apache/airavata-custos/pkg/models"
)

// CreateComputeAllocationUsage records a new usage event against a compute
// allocation. The referenced allocation must exist. CalculatedTime defaults
// to the server's current UTC time when unset. Usage records are intended to
// be append-only — there is no update operation.
func (s *Service) CreateComputeAllocationUsage(ctx context.Context, u *models.ComputeAllocationUsage) (*models.ComputeAllocationUsage, error) {
	if u == nil {
		return nil, fmt.Errorf("%w: compute allocation usage is nil", ErrInvalidInput)
	}
	if u.ComputeAllocationID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_id is required", ErrInvalidInput)
	}
	if u.UsedRawAmount < 0 || u.UsedSUAmount < 0 {
		return nil, fmt.Errorf("%w: used_raw_amount and used_su_amount must be non-negative", ErrInvalidInput)
	}

	if alloc, err := s.allocs.FindByID(ctx, u.ComputeAllocationID); err != nil {
		return nil, fmt.Errorf("lookup compute allocation: %w", err)
	} else if alloc == nil {
		return nil, fmt.Errorf("%w: compute allocation %q not found", ErrInvalidInput, u.ComputeAllocationID)
	}
	if u.ComputeAllocationResourceID != "" {
		if r, err := s.resources.FindByID(ctx, u.ComputeAllocationResourceID); err != nil {
			return nil, fmt.Errorf("lookup compute allocation resource: %w", err)
		} else if r == nil {
			return nil, fmt.Errorf("%w: compute allocation resource %q not found", ErrInvalidInput, u.ComputeAllocationResourceID)
		}
	}

	if u.ID == "" {
		u.ID = newID()
	}
	if u.CalculatedTime.IsZero() {
		u.CalculatedTime = nowUTC()
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.usages.Create(ctx, tx, u)
	}); err != nil {
		return nil, fmt.Errorf("create compute allocation usage: %w", err)
	}
	return u, nil
}

// GetComputeAllocationUsage retrieves a usage event by its ID. Returns
// ErrNotFound when no usage matches.
func (s *Service) GetComputeAllocationUsage(ctx context.Context, id string) (*models.ComputeAllocationUsage, error) {
	u, err := s.usages.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get compute allocation usage: %w", err)
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}

// GetComputeAllocationUsageByComputeAllocationIDAndJobID retrieves a usage event
// by the given compute allocation ID and job ID. Returns ErrNotFound when no usage matches.
func (s *Service) GetComputeAllocationUsageByComputeAllocationIDAndJobID(ctx context.Context, allocationID, jobID string) (*models.ComputeAllocationUsage, error) {
	if allocationID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_id is required", ErrInvalidInput)
	}
	if jobID == "" {
		return nil, fmt.Errorf("%w: job_id is required", ErrInvalidInput)
	}
	u, err := s.usages.FindByComputeAllocationIDAndJobID(ctx, allocationID, jobID)
	if err != nil {
		return nil, fmt.Errorf("get compute allocation usage by allocation id and job id: %w", err)
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}

// ListUsagesForAllocation returns every usage event recorded against the
// given allocation, ordered by calculated_time ascending.
func (s *Service) ListUsagesForAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationUsage, error) {
	if allocationID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_id is required", ErrInvalidInput)
	}
	rows, err := s.usages.FindByAllocation(ctx, allocationID)
	if err != nil {
		return nil, fmt.Errorf("list usages for allocation: %w", err)
	}
	return rows, nil
}

// ListUsagesByUser returns every usage event attributed to the given user,
// ordered by calculated_time ascending.
func (s *Service) ListUsagesByUser(ctx context.Context, userID string) ([]models.ComputeAllocationUsage, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	rows, err := s.usages.FindByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list usages by user: %w", err)
	}
	return rows, nil
}

// GetTotalSUUsageForAllocation returns the total SUs consumed against the
// given allocation across all recorded usage events.
func (s *Service) GetTotalSUUsageForAllocation(ctx context.Context, allocationID string) (int64, error) {
	if allocationID == "" {
		return 0, fmt.Errorf("%w: compute_allocation_id is required", ErrInvalidInput)
	}
	total, err := s.usages.SumSUForAllocation(ctx, allocationID)
	if err != nil {
		return 0, fmt.Errorf("sum su usage for allocation: %w", err)
	}
	return total, nil
}

// GetTotalSUUsageForUserInAllocation returns the total SUs consumed by the
// given user against the given allocation.
func (s *Service) GetTotalSUUsageForUserInAllocation(ctx context.Context, allocationID, userID string) (int64, error) {
	if allocationID == "" {
		return 0, fmt.Errorf("%w: compute_allocation_id is required", ErrInvalidInput)
	}
	if userID == "" {
		return 0, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	total, err := s.usages.SumSUForUserInAllocation(ctx, allocationID, userID)
	if err != nil {
		return 0, fmt.Errorf("sum su usage for user in allocation: %w", err)
	}
	return total, nil
}

// DeleteComputeAllocationUsage removes a usage event by ID.
func (s *Service) DeleteComputeAllocationUsage(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: compute allocation usage id is required", ErrInvalidInput)
	}
	existing, err := s.usages.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup compute allocation usage: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.usages.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete compute allocation usage: %w", err)
	}
	return nil
}
