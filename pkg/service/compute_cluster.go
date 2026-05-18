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

// CreateComputeCluster persists a new compute cluster. If cluster.ID is empty
// a new UUID is generated. The (possibly populated) cluster is returned.
func (s *Service) CreateComputeCluster(ctx context.Context, cluster *models.ComputeCluster) (*models.ComputeCluster, error) {
	if cluster == nil {
		return nil, fmt.Errorf("%w: compute cluster is nil", ErrInvalidInput)
	}
	if cluster.Name == "" {
		return nil, fmt.Errorf("%w: compute cluster name is required", ErrInvalidInput)
	}
	if cluster.ID == "" {
		cluster.ID = newID()
	}

	if existing, err := s.clusters.FindByName(ctx, cluster.Name); err != nil {
		return nil, fmt.Errorf("lookup compute cluster by name: %w", err)
	} else if existing != nil {
		return nil, fmt.Errorf("%w: compute cluster with name %q", ErrAlreadyExists, cluster.Name)
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.clusters.Create(ctx, tx, cluster)
	}); err != nil {
		return nil, fmt.Errorf("create compute cluster: %w", err)
	}

	s.eventBus.Publish(events.ComputeAllocationCreateEvent, cluster)
	return cluster, nil
}

// GetComputeCluster retrieves a compute cluster by its ID. Returns
// ErrNotFound when no cluster matches.
func (s *Service) GetComputeCluster(ctx context.Context, id string) (*models.ComputeCluster, error) {
	c, err := s.clusters.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get compute cluster: %w", err)
	}
	if c == nil {
		return nil, ErrNotFound
	}
	return c, nil
}

// GetComputeClusterByName retrieves a compute cluster by its name.
func (s *Service) GetComputeClusterByName(ctx context.Context, name string) (*models.ComputeCluster, error) {
	c, err := s.clusters.FindByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get compute cluster by name: %w", err)
	}
	if c == nil {
		return nil, ErrNotFound
	}
	return c, nil
}

// ListComputeClusters returns every compute cluster, ordered by name.
func (s *Service) ListComputeClusters(ctx context.Context) ([]models.ComputeCluster, error) {
	clusters, err := s.clusters.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list compute clusters: %w", err)
	}
	return clusters, nil
}

// UpdateComputeCluster persists changes to an existing compute cluster.
func (s *Service) UpdateComputeCluster(ctx context.Context, cluster *models.ComputeCluster) error {
	if cluster == nil || cluster.ID == "" {
		return fmt.Errorf("%w: compute cluster id is required", ErrInvalidInput)
	}
	if cluster.Name == "" {
		return fmt.Errorf("%w: compute cluster name is required", ErrInvalidInput)
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.clusters.Update(ctx, tx, cluster)
	}); err != nil {
		return fmt.Errorf("update compute cluster: %w", err)
	}

	s.eventBus.Publish(events.ComputeAllocationUpdateEvent, cluster)
	return nil
}

// DeleteComputeCluster removes a compute cluster by ID.
func (s *Service) DeleteComputeCluster(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: compute cluster id is required", ErrInvalidInput)
	}
	cluster, err := s.clusters.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup compute cluster: %w", err)
	}
	if cluster == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.clusters.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete compute cluster: %w", err)
	}

	s.eventBus.Publish(events.ComputeAllocationDeleteEvent, cluster)
	return nil
}
