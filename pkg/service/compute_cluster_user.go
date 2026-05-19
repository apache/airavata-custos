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

// CreateComputeClusterUser persists a new compute-cluster user mapping. If
// the ID is empty, a UUID is generated. The (possibly populated) record is
// returned.
func (s *Service) CreateComputeClusterUser(ctx context.Context, cu *models.ComputeClusterUser) (*models.ComputeClusterUser, error) {
	if cu == nil {
		return nil, fmt.Errorf("%w: compute cluster user is nil", ErrInvalidInput)
	}
	if cu.ComputeClusterID == "" {
		return nil, fmt.Errorf("%w: compute_cluster_id is required", ErrInvalidInput)
	}
	if cu.UserID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if cu.LocalUsername == "" {
		return nil, fmt.Errorf("%w: local_username is required", ErrInvalidInput)
	}
	if cu.ID == "" {
		cu.ID = newID()
	}

	if cluster, err := s.clusters.FindByID(ctx, cu.ComputeClusterID); err != nil {
		return nil, fmt.Errorf("lookup compute cluster: %w", err)
	} else if cluster == nil {
		return nil, fmt.Errorf("%w: compute cluster %q not found", ErrInvalidInput, cu.ComputeClusterID)
	}
	if user, err := s.users.FindByID(ctx, cu.UserID); err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	} else if user == nil {
		return nil, fmt.Errorf("%w: user %q not found", ErrInvalidInput, cu.UserID)
	}

	if existing, err := s.clusterUsers.FindByPair(ctx, cu.ComputeClusterID, cu.UserID); err != nil {
		return nil, fmt.Errorf("lookup compute cluster user by pair: %w", err)
	} else if existing != nil {
		return nil, fmt.Errorf("%w: user %q is already mapped on cluster %q",
			ErrAlreadyExists, cu.UserID, cu.ComputeClusterID)
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.clusterUsers.Create(ctx, tx, cu)
	}); err != nil {
		return nil, fmt.Errorf("create compute cluster user: %w", err)
	}

	s.eventBus.Publish(events.ComputeClusterUserCreateEvent, cu)
	return cu, nil
}

// GetComputeClusterUser retrieves a compute-cluster user by its ID.
func (s *Service) GetComputeClusterUser(ctx context.Context, id string) (*models.ComputeClusterUser, error) {
	c, err := s.clusterUsers.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get compute cluster user: %w", err)
	}
	if c == nil {
		return nil, ErrNotFound
	}
	return c, nil
}

// GetComputeClusterUserByPair retrieves the compute-cluster user mapping for
// the given (compute_cluster_id, user_id) pair.
func (s *Service) GetComputeClusterUserByPair(ctx context.Context, clusterID, userID string) (*models.ComputeClusterUser, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("%w: compute_cluster_id is required", ErrInvalidInput)
	}
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	c, err := s.clusterUsers.FindByPair(ctx, clusterID, userID)
	if err != nil {
		return nil, fmt.Errorf("get compute cluster user by pair: %w", err)
	}
	if c == nil {
		return nil, ErrNotFound
	}
	return c, nil
}

// ListComputeClusterUsersByCluster returns every user mapping for the given
// compute cluster, ordered by local username.
func (s *Service) ListComputeClusterUsersByCluster(ctx context.Context, clusterID string) ([]models.ComputeClusterUser, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("%w: compute_cluster_id is required", ErrInvalidInput)
	}
	users, err := s.clusterUsers.FindByCluster(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("list compute cluster users by cluster: %w", err)
	}
	return users, nil
}

// ListComputeClusterUsersByUser returns every cluster mapping held by the
// given Custos user.
func (s *Service) ListComputeClusterUsersByUser(ctx context.Context, userID string) ([]models.ComputeClusterUser, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	users, err := s.clusterUsers.FindByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list compute cluster users by user: %w", err)
	}
	return users, nil
}

// UpdateComputeClusterUser persists changes to an existing compute-cluster
// user mapping.
func (s *Service) UpdateComputeClusterUser(ctx context.Context, cu *models.ComputeClusterUser) error {
	if cu == nil || cu.ID == "" {
		return fmt.Errorf("%w: compute cluster user id is required", ErrInvalidInput)
	}
	if cu.ComputeClusterID == "" {
		return fmt.Errorf("%w: compute_cluster_id is required", ErrInvalidInput)
	}
	if cu.UserID == "" {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if cu.LocalUsername == "" {
		return fmt.Errorf("%w: local_username is required", ErrInvalidInput)
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.clusterUsers.Update(ctx, tx, cu)
	}); err != nil {
		return fmt.Errorf("update compute cluster user: %w", err)
	}

	s.eventBus.Publish(events.ComputeClusterUserUpdateEvent, cu)
	return nil
}

// DeleteComputeClusterUser removes a compute-cluster user mapping by ID.
func (s *Service) DeleteComputeClusterUser(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: compute cluster user id is required", ErrInvalidInput)
	}
	cu, err := s.clusterUsers.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup compute cluster user: %w", err)
	}
	if cu == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.clusterUsers.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete compute cluster user: %w", err)
	}

	s.eventBus.Publish(events.ComputeClusterUserDeleteEvent, cu)
	return nil
}
