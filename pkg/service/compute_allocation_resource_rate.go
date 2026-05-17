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
	"time"

	"github.com/apache/airavata-custos/pkg/models"
)

// CreateComputeAllocationResourceRate persists a new rate for a compute
// allocation resource. The referenced resource must exist, start_time must be
// strictly before end_time, and rate must be non-negative.
func (s *Service) CreateComputeAllocationResourceRate(ctx context.Context, rate *models.ComputeAllocationResourceRate) (*models.ComputeAllocationResourceRate, error) {
	if rate == nil {
		return nil, fmt.Errorf("%w: compute allocation resource rate is nil", ErrInvalidInput)
	}
	if rate.ComputeAllocationResourceID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_resource_id is required", ErrInvalidInput)
	}
	if rate.Rate < 0 {
		return nil, fmt.Errorf("%w: rate must be non-negative", ErrInvalidInput)
	}
	if rate.StartTime.IsZero() || rate.EndTime.IsZero() {
		return nil, fmt.Errorf("%w: start_time and end_time are required", ErrInvalidInput)
	}
	if !rate.StartTime.Before(rate.EndTime) {
		return nil, fmt.Errorf("%w: start_time must be before end_time", ErrInvalidInput)
	}

	if res, err := s.resources.FindByID(ctx, rate.ComputeAllocationResourceID); err != nil {
		return nil, fmt.Errorf("lookup compute allocation resource: %w", err)
	} else if res == nil {
		return nil, fmt.Errorf("%w: compute allocation resource %q not found", ErrInvalidInput, rate.ComputeAllocationResourceID)
	}

	if rate.ID == "" {
		rate.ID = newID()
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.resourceRates.Create(ctx, tx, rate)
	}); err != nil {
		return nil, fmt.Errorf("create compute allocation resource rate: %w", err)
	}
	return rate, nil
}

// GetComputeAllocationResourceRate retrieves a rate by its ID. Returns
// ErrNotFound when no rate matches.
func (s *Service) GetComputeAllocationResourceRate(ctx context.Context, id string) (*models.ComputeAllocationResourceRate, error) {
	r, err := s.resourceRates.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get compute allocation resource rate: %w", err)
	}
	if r == nil {
		return nil, ErrNotFound
	}
	return r, nil
}

// ListRatesForResource returns every rate ever defined for the given resource,
// ordered chronologically by start_time.
func (s *Service) ListRatesForResource(ctx context.Context, resourceID string) ([]models.ComputeAllocationResourceRate, error) {
	if resourceID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_resource_id is required", ErrInvalidInput)
	}
	rates, err := s.resourceRates.FindByResource(ctx, resourceID)
	if err != nil {
		return nil, fmt.Errorf("list rates for resource: %w", err)
	}
	return rates, nil
}

// GetEffectiveRateForResource returns the rate effective for the given resource
// at the supplied instant. Returns ErrNotFound when no rate applies.
func (s *Service) GetEffectiveRateForResource(ctx context.Context, resourceID string, at time.Time) (*models.ComputeAllocationResourceRate, error) {
	if resourceID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_resource_id is required", ErrInvalidInput)
	}
	if at.IsZero() {
		at = nowUTC()
	}
	r, err := s.resourceRates.FindEffective(ctx, resourceID, at)
	if err != nil {
		return nil, fmt.Errorf("get effective rate for resource: %w", err)
	}
	if r == nil {
		return nil, ErrNotFound
	}
	return r, nil
}

// UpdateComputeAllocationResourceRate persists changes to an existing rate.
// The referenced resource and ID are immutable from this call's perspective —
// only rate, start_time, and end_time are updated.
func (s *Service) UpdateComputeAllocationResourceRate(ctx context.Context, rate *models.ComputeAllocationResourceRate) error {
	if rate == nil || rate.ID == "" {
		return fmt.Errorf("%w: compute allocation resource rate id is required", ErrInvalidInput)
	}
	if rate.Rate < 0 {
		return fmt.Errorf("%w: rate must be non-negative", ErrInvalidInput)
	}
	if !rate.StartTime.IsZero() && !rate.EndTime.IsZero() && !rate.StartTime.Before(rate.EndTime) {
		return fmt.Errorf("%w: start_time must be before end_time", ErrInvalidInput)
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.resourceRates.Update(ctx, tx, rate)
	}); err != nil {
		return fmt.Errorf("update compute allocation resource rate: %w", err)
	}
	return nil
}

// DeleteComputeAllocationResourceRate removes a rate by ID.
func (s *Service) DeleteComputeAllocationResourceRate(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: compute allocation resource rate id is required", ErrInvalidInput)
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.resourceRates.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete compute allocation resource rate: %w", err)
	}
	return nil
}
