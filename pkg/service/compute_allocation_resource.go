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

// CreateComputeAllocationResource persists a new compute allocation resource.
// If resource.ID is empty a new UUID is generated.
func (s *Service) CreateComputeAllocationResource(ctx context.Context, resource *models.ComputeAllocationResource) (*models.ComputeAllocationResource, error) {
	if resource == nil {
		return nil, fmt.Errorf("%w: compute allocation resource is nil", ErrInvalidInput)
	}
	if resource.Name == "" {
		return nil, fmt.Errorf("%w: resource name is required", ErrInvalidInput)
	}
	if resource.ResourceType == "" {
		return nil, fmt.Errorf("%w: resource_type is required", ErrInvalidInput)
	}
	if resource.ComputeClusterID == "" {
		return nil, fmt.Errorf("%w: compute_cluster_id is required", ErrInvalidInput)
	}
	if cluster, err := s.clusters.FindByID(ctx, resource.ComputeClusterID); err != nil {
		return nil, fmt.Errorf("lookup compute cluster: %w", err)
	} else if cluster == nil {
		return nil, fmt.Errorf("%w: compute cluster %q not found", ErrInvalidInput, resource.ComputeClusterID)
	}
	if resource.ID == "" {
		resource.ID = newID()
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.resources.Create(ctx, tx, resource)
	}); err != nil {
		return nil, fmt.Errorf("create compute allocation resource: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationResourceCreateEvent, resource)
	return resource, nil
}

// GetComputeAllocationResource retrieves a compute allocation resource by its ID.
// Returns ErrNotFound when no resource matches.
func (s *Service) GetComputeAllocationResource(ctx context.Context, id string) (*models.ComputeAllocationResource, error) {
	r, err := s.resources.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get compute allocation resource: %w", err)
	}
	if r == nil {
		return nil, ErrNotFound
	}
	return r, nil
}

// GetComputeAllocationResourceByNameAndCluster retrieves a compute allocation
// resource by its name within a given compute cluster. Returns ErrNotFound
// when no resource matches.
func (s *Service) GetComputeAllocationResourceByNameAndCluster(ctx context.Context, name, clusterID string) (*models.ComputeAllocationResource, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: resource name is required", ErrInvalidInput)
	}
	if clusterID == "" {
		return nil, fmt.Errorf("%w: compute_cluster_id is required", ErrInvalidInput)
	}
	r, err := s.resources.FindByNameAndCluster(ctx, name, clusterID)
	if err != nil {
		return nil, fmt.Errorf("get compute allocation resource by name and cluster: %w", err)
	}
	if r == nil {
		return nil, ErrNotFound
	}
	return r, nil
}

// ListComputeAllocationResources returns every compute allocation resource.
func (s *Service) ListComputeAllocationResources(ctx context.Context) ([]models.ComputeAllocationResource, error) {
	resources, err := s.resources.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list compute allocation resources: %w", err)
	}
	return resources, nil
}

// ListComputeAllocationResourcesByTypeAndCluster returns all resources of the
// given type on the given compute cluster.
func (s *Service) ListComputeAllocationResourcesByTypeAndCluster(ctx context.Context, resourceType, clusterID string) ([]models.ComputeAllocationResource, error) {
	if resourceType == "" {
		return nil, fmt.Errorf("%w: resource_type is required", ErrInvalidInput)
	}
	if clusterID == "" {
		return nil, fmt.Errorf("%w: compute_cluster_id is required", ErrInvalidInput)
	}
	resources, err := s.resources.FindByTypeAndCluster(ctx, resourceType, clusterID)
	if err != nil {
		return nil, fmt.Errorf("list compute allocation resources by type and cluster: %w", err)
	}
	return resources, nil
}

// UpdateComputeAllocationResource persists changes to an existing resource.
func (s *Service) UpdateComputeAllocationResource(ctx context.Context, resource *models.ComputeAllocationResource) error {
	if resource == nil || resource.ID == "" {
		return fmt.Errorf("%w: compute allocation resource id is required", ErrInvalidInput)
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.resources.Update(ctx, tx, resource)
	}); err != nil {
		return fmt.Errorf("update compute allocation resource: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationResourceUpdateEvent, resource)
	return nil
}

// DeleteComputeAllocationResource removes a compute allocation resource by ID.
func (s *Service) DeleteComputeAllocationResource(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: compute allocation resource id is required", ErrInvalidInput)
	}
	resource, err := s.resources.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup compute allocation resource: %w", err)
	}
	if resource == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.resources.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete compute allocation resource: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeAllocationResourceDeleteEvent, resource)
	return nil
}
