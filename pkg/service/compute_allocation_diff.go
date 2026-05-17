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

// CreateComputeAllocationDiff records a new diff against a compute allocation.
// The referenced allocation must exist. If Timestamp is zero the server's
// current UTC time is used. Diff records are intended to be append-only and
// have no update operation.
func (s *Service) CreateComputeAllocationDiff(ctx context.Context, diff *models.ComputeAllocationDiff) (*models.ComputeAllocationDiff, error) {
	if diff == nil {
		return nil, fmt.Errorf("%w: compute allocation diff is nil", ErrInvalidInput)
	}
	if diff.ComputeAllocationID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_id is required", ErrInvalidInput)
	}
	if diff.DiffType == "" {
		return nil, fmt.Errorf("%w: diff_type is required", ErrInvalidInput)
	}
	if diff.Status == "" {
		return nil, fmt.Errorf("%w: status is required", ErrInvalidInput)
	}

	if alloc, err := s.allocs.FindByID(ctx, diff.ComputeAllocationID); err != nil {
		return nil, fmt.Errorf("lookup compute allocation: %w", err)
	} else if alloc == nil {
		return nil, fmt.Errorf("%w: compute allocation %q not found", ErrInvalidInput, diff.ComputeAllocationID)
	}

	if diff.ID == "" {
		diff.ID = newID()
	}
	if diff.Timestamp.IsZero() {
		diff.Timestamp = nowUTC()
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.allocDiffs.Create(ctx, tx, diff)
	}); err != nil {
		return nil, fmt.Errorf("create compute allocation diff: %w", err)
	}
	return diff, nil
}

// GetComputeAllocationDiff retrieves a diff by its ID. Returns ErrNotFound
// when no diff matches.
func (s *Service) GetComputeAllocationDiff(ctx context.Context, id string) (*models.ComputeAllocationDiff, error) {
	d, err := s.allocDiffs.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get compute allocation diff: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}
	return d, nil
}

// ListDiffsForAllocation returns every diff ever recorded against the given
// compute allocation, ordered chronologically by timestamp.
func (s *Service) ListDiffsForAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationDiff, error) {
	if allocationID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_id is required", ErrInvalidInput)
	}
	diffs, err := s.allocDiffs.FindByAllocation(ctx, allocationID)
	if err != nil {
		return nil, fmt.Errorf("list diffs for allocation: %w", err)
	}
	return diffs, nil
}

// GetLatestDiffForAllocation returns the most recent diff for the given
// allocation. Returns ErrNotFound when the allocation has no diffs.
func (s *Service) GetLatestDiffForAllocation(ctx context.Context, allocationID string) (*models.ComputeAllocationDiff, error) {
	if allocationID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_id is required", ErrInvalidInput)
	}
	d, err := s.allocDiffs.FindLatestByAllocation(ctx, allocationID)
	if err != nil {
		return nil, fmt.Errorf("get latest diff for allocation: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}
	return d, nil
}

// DeleteComputeAllocationDiff removes a diff record by ID. This is intended
// for administrative cleanup of erroneous entries; the diff log is otherwise
// append-only.
func (s *Service) DeleteComputeAllocationDiff(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: compute allocation diff id is required", ErrInvalidInput)
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.allocDiffs.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete compute allocation diff: %w", err)
	}
	return nil
}
