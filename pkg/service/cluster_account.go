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

// CreateClusterAccount persists a new cluster account for a user on a cluster.
// If account.ID is empty, a new UUID is generated. The referenced user and
// cluster must exist; (compute_cluster_id, username) is unique.
func (s *Service) CreateClusterAccount(ctx context.Context, account *models.ClusterAccount) (*models.ClusterAccount, error) {
	if account == nil {
		return nil, fmt.Errorf("%w: cluster account is nil", ErrInvalidInput)
	}
	if account.UserID == "" {
		return nil, fmt.Errorf("%w: cluster account user_id is required", ErrInvalidInput)
	}
	if account.ComputeClusterID == "" {
		return nil, fmt.Errorf("%w: cluster account compute_cluster_id is required", ErrInvalidInput)
	}
	if account.Username == "" {
		return nil, fmt.Errorf("%w: cluster account username is required", ErrInvalidInput)
	}

	if user, err := s.users.FindByID(ctx, account.UserID); err != nil {
		return nil, fmt.Errorf("verify user: %w", err)
	} else if user == nil {
		return nil, fmt.Errorf("%w: user %q does not exist", ErrInvalidInput, account.UserID)
	}
	if cluster, err := s.clusters.FindByID(ctx, account.ComputeClusterID); err != nil {
		return nil, fmt.Errorf("verify compute cluster: %w", err)
	} else if cluster == nil {
		return nil, fmt.Errorf("%w: compute cluster %q does not exist", ErrInvalidInput, account.ComputeClusterID)
	}

	if existing, err := s.clusterAccounts.FindByClusterAndUsername(ctx, account.ComputeClusterID, account.Username); err != nil {
		return nil, fmt.Errorf("lookup cluster account: %w", err)
	} else if existing != nil {
		return nil, fmt.Errorf("%w: cluster account %q on cluster %q", ErrAlreadyExists, account.Username, account.ComputeClusterID)
	}

	if account.ID == "" {
		account.ID = newID()
	}
	if account.Status == "" {
		account.Status = models.ACTIVE
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.clusterAccounts.Create(ctx, tx, account)
	}); err != nil {
		return nil, fmt.Errorf("create cluster account: %w", err)
	}

	s.eventBus.Publish(events.ClusterAccountCreateEvent, account)
	return account, nil
}

// GetClusterAccount retrieves a cluster account by ID.
func (s *Service) GetClusterAccount(ctx context.Context, id string) (*models.ClusterAccount, error) {
	a, err := s.clusterAccounts.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get cluster account: %w", err)
	}
	if a == nil {
		return nil, ErrNotFound
	}
	return a, nil
}

// GetClusterAccountByClusterAndUsername resolves a cluster account by its
// natural key.
func (s *Service) GetClusterAccountByClusterAndUsername(ctx context.Context, clusterID, username string) (*models.ClusterAccount, error) {
	a, err := s.clusterAccounts.FindByClusterAndUsername(ctx, clusterID, username)
	if err != nil {
		return nil, fmt.Errorf("get cluster account by cluster/username: %w", err)
	}
	if a == nil {
		return nil, ErrNotFound
	}
	return a, nil
}

// ListClusterAccountsForUser returns every cluster account belonging to a user.
func (s *Service) ListClusterAccountsForUser(ctx context.Context, userID string) ([]models.ClusterAccount, error) {
	out, err := s.clusterAccounts.FindByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list cluster accounts by user: %w", err)
	}
	return out, nil
}

// ListClusterAccountsForCluster returns every cluster account on a cluster.
func (s *Service) ListClusterAccountsForCluster(ctx context.Context, clusterID string) ([]models.ClusterAccount, error) {
	out, err := s.clusterAccounts.FindByCluster(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("list cluster accounts by cluster: %w", err)
	}
	return out, nil
}

// UpdateClusterAccountStatus flips only the lifecycle status.
func (s *Service) UpdateClusterAccountStatus(ctx context.Context, id string, status models.AllocationStatus) (*models.ClusterAccount, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: cluster account id is required", ErrInvalidInput)
	}
	switch status {
	case models.ACTIVE, models.INACTIVE, models.DELETED:
	default:
		return nil, fmt.Errorf("%w: invalid cluster account status %q", ErrInvalidInput, status)
	}

	a, err := s.clusterAccounts.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("lookup cluster account: %w", err)
	}
	if a == nil {
		return nil, ErrNotFound
	}
	if a.Status == status {
		return a, nil
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.clusterAccounts.UpdateStatus(ctx, tx, id, status)
	}); err != nil {
		return nil, fmt.Errorf("update cluster account status: %w", err)
	}
	a.Status = status

	s.eventBus.Publish(events.ClusterAccountUpdateEvent, a)
	return a, nil
}

// DeleteClusterAccount removes a cluster account by ID.
func (s *Service) DeleteClusterAccount(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: cluster account id is required", ErrInvalidInput)
	}
	a, err := s.clusterAccounts.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup cluster account: %w", err)
	}
	if a == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.clusterAccounts.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete cluster account: %w", err)
	}

	s.eventBus.Publish(events.ClusterAccountDeleteEvent, a)
	return nil
}
