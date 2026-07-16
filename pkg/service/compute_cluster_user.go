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
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/posix"
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
		switch {
		case isLocalUsernameDuplicate(err):
			return nil, fmt.Errorf("%w: %s", ErrAlreadyExists, cu.LocalUsername)
		case isPairDuplicate(err):
			return nil, fmt.Errorf("%w: user %q is already mapped on cluster %q",
				ErrAlreadyExists, cu.UserID, cu.ComputeClusterID)
		default:
			return nil, fmt.Errorf("create compute cluster user: %w", err)
		}
	}

	s.eventBus.Publish(ctx, events.ComputeClusterUserCreateEvent, cu)
	return cu, nil
}

// AllocateComputeClusterUser builds a POSIX username for the user and creates
// the cluster mapping, retrying with a numeric suffix on username collisions.
func (s *Service) AllocateComputeClusterUser(ctx context.Context, user *models.User, clusterID string) (*models.ComputeClusterUser, error) {
	if user == nil || user.ID == "" {
		return nil, fmt.Errorf("%w: user is required", ErrInvalidInput)
	}

	base, truncated, err := posix.BuildBase(user, posix.Prefix())
	if err != nil {
		_, _ = s.CreateAuditEvent(ctx, &models.AuditEvent{
			EventType:  "PosixUsernameUnbuildable",
			EntityID:   user.ID,
			EntityType: "user",
			Details:    err.Error(),
		})
		return nil, err
	}
	if truncated {
		_, _ = s.CreateAuditEvent(ctx, &models.AuditEvent{
			EventType:  "PosixUsernameTruncated",
			EntityID:   user.ID,
			EntityType: "user",
			Details:    base,
		})
	}

	for n := 0; n < posix.MaxCollisionSuffix; n++ {
		candidate := base
		if n > 0 {
			candidate = base + strconv.Itoa(n+1)
		}
		ccu, err := s.CreateComputeClusterUser(ctx, &models.ComputeClusterUser{
			ComputeClusterID: clusterID,
			UserID:           user.ID,
			LocalUsername:    candidate,
		})
		if err == nil {
			return ccu, nil
		}
		if errors.Is(err, ErrAlreadyExists) {
			continue
		}
		return nil, err
	}

	_, _ = s.CreateAuditEvent(ctx, &models.AuditEvent{
		EventType:  "PosixUsernameAllocatorExhausted",
		EntityID:   user.ID,
		EntityType: "user",
		Details:    base,
	})
	return nil, fmt.Errorf("posix username allocator exhausted for base %q", base)
}

func isLocalUsernameDuplicate(err error) bool {
	return err != nil && strings.Contains(err.Error(), "uq_compute_cluster_users_local_username")
}

func isPairDuplicate(err error) bool {
	return err != nil && strings.Contains(err.Error(), "uq_compute_cluster_users_pair")
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

// GetComputeClusterUserByClusterAndLocalUsername retrieves the compute-cluster
// user mapping for the given (compute_cluster_id, local_username) pair.
func (s *Service) GetComputeClusterUserByClusterAndLocalUsername(ctx context.Context, clusterID, localUsername string) (*models.ComputeClusterUser, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("%w: compute_cluster_id is required", ErrInvalidInput)
	}
	if localUsername == "" {
		return nil, fmt.Errorf("%w: local_username is required", ErrInvalidInput)
	}
	c, err := s.clusterUsers.FindByClusterAndLocalUsername(ctx, clusterID, localUsername)
	if err != nil {
		return nil, fmt.Errorf("get compute cluster user by local username and cluster: %w", err)
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
// user mapping. Fields left blank/zero on the supplied record fall back to
// the stored value.
func (s *Service) UpdateComputeClusterUser(ctx context.Context, cu *models.ComputeClusterUser) error {
	if cu == nil || cu.ID == "" {
		return fmt.Errorf("%w: compute cluster user id is required", ErrInvalidInput)
	}
	existing, err := s.clusterUsers.FindByID(ctx, cu.ID)
	if err != nil {
		return fmt.Errorf("lookup compute cluster user: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}
	if cu.ComputeClusterID == "" {
		cu.ComputeClusterID = existing.ComputeClusterID
	}
	if cu.UserID == "" {
		cu.UserID = existing.UserID
	}
	if cu.LocalUsername == "" {
		cu.LocalUsername = existing.LocalUsername
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.clusterUsers.Update(ctx, tx, cu)
	}); err != nil {
		return fmt.Errorf("update compute cluster user: %w", err)
	}

	s.eventBus.Publish(ctx, events.ComputeClusterUserUpdateEvent, cu)
	return nil
}

// MarkComputeClusterUserProvisioned stamps provisioned_at on the mapping,
// signaling that the account exists in the registry.
func (s *Service) MarkComputeClusterUserProvisioned(ctx context.Context, id string) error {
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
		return s.clusterUsers.MarkProvisioned(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("mark compute cluster user provisioned: %w", err)
	}
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

	s.eventBus.Publish(ctx, events.ComputeClusterUserDeleteEvent, cu)
	return nil
}
