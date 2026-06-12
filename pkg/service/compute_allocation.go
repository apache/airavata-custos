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

// CreateComputeAllocation persists a new compute allocation. The referenced
// project and compute cluster must already exist. If alloc.ID is empty a new
// UUID is generated; if Status is empty it defaults to ACTIVE.
func (s *Service) CreateComputeAllocation(ctx context.Context, alloc *models.ComputeAllocation) (*models.ComputeAllocation, error) {
	if alloc == nil {
		return nil, fmt.Errorf("%w: compute allocation is nil", ErrInvalidInput)
	}
	if alloc.Name == "" {
		return nil, fmt.Errorf("%w: compute allocation name is required", ErrInvalidInput)
	}
	if alloc.ProjectID == "" {
		return nil, fmt.Errorf("%w: project_id is required", ErrInvalidInput)
	}
	if alloc.ComputeClusterID == "" {
		return nil, fmt.Errorf("%w: compute_cluster_id is required", ErrInvalidInput)
	}
	if alloc.ID == "" {
		alloc.ID = newID()
	}
	if alloc.Status == "" {
		alloc.Status = models.ACTIVE
	}

	if proj, err := s.projs.FindByID(ctx, alloc.ProjectID); err != nil {
		return nil, fmt.Errorf("lookup project: %w", err)
	} else if proj == nil {
		return nil, fmt.Errorf("%w: project %q not found", ErrInvalidInput, alloc.ProjectID)
	}

	if cluster, err := s.clusters.FindByID(ctx, alloc.ComputeClusterID); err != nil {
		return nil, fmt.Errorf("lookup compute cluster: %w", err)
	} else if cluster == nil {
		return nil, fmt.Errorf("%w: compute cluster %q not found", ErrInvalidInput, alloc.ComputeClusterID)
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.allocs.Create(ctx, tx, alloc)
	}); err != nil {
		return nil, fmt.Errorf("create compute allocation: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationCreateEvent, alloc)
	return alloc, nil
}

// GetComputeAllocation retrieves a compute allocation by its ID. Returns
// ErrNotFound when no allocation matches.
func (s *Service) GetComputeAllocation(ctx context.Context, id string) (*models.ComputeAllocation, error) {
	a, err := s.allocs.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get compute allocation: %w", err)
	}
	if a == nil {
		return nil, ErrNotFound
	}
	return a, nil
}

// ListComputeAllocationsByProject returns every compute allocation attached to a project.
func (s *Service) ListComputeAllocationsByProject(ctx context.Context, projectID string) ([]models.ComputeAllocation, error) {
	allocs, err := s.allocs.FindByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list compute allocations by project: %w", err)
	}
	return allocs, nil
}

// ListComputeAllocationsByCluster returns every compute allocation attached to a cluster.
func (s *Service) ListComputeAllocationsByCluster(ctx context.Context, clusterID string) ([]models.ComputeAllocation, error) {
	allocs, err := s.allocs.FindByCluster(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("list compute allocations by cluster: %w", err)
	}
	return allocs, nil
}

// UpdateComputeAllocation persists changes to an existing compute allocation.
func (s *Service) UpdateComputeAllocation(ctx context.Context, alloc *models.ComputeAllocation) error {
	if alloc == nil || alloc.ID == "" {
		return fmt.Errorf("%w: compute allocation id is required", ErrInvalidInput)
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.allocs.Update(ctx, tx, alloc)
	}); err != nil {
		return fmt.Errorf("update compute allocation: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationUpdateEvent, alloc)
	return nil
}

// DeleteComputeAllocation removes a compute allocation by ID.
func (s *Service) DeleteComputeAllocation(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: compute allocation id is required", ErrInvalidInput)
	}
	alloc, err := s.allocs.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup compute allocation: %w", err)
	}
	if alloc == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.allocs.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete compute allocation: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationDeleteEvent, alloc)
	return nil
}
