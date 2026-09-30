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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

const (
	clusterUserAuditAdminGranted = "CLUSTER_ADMIN_GRANTED"
	clusterUserAuditApproved     = "CLUSTER_ACCOUNT_APPROVED"
	clusterUserAuditDenied       = "CLUSTER_ACCOUNT_DENIED"
)

func (s *Service) writeClusterUserAuditTx(ctx context.Context, tx *sql.Tx, eventType, entityID string, details map[string]any) error {
	payload, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("marshal audit details: %w", err)
	}
	return s.auditEvents.Create(ctx, tx, &models.AuditEvent{
		ID:         newID(),
		EventType:  eventType,
		EventTime:  nowUTC(),
		EntityID:   entityID,
		EntityType: "compute_cluster_user",
		Details:    string(payload),
	})
}

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
	if cu.AccessLevel == "" {
		cu.AccessLevel = models.ClusterAccessUser
	}
	if cu.AccessLevel != models.ClusterAccessUser && cu.AccessLevel != models.ClusterAccessAdmin {
		return nil, fmt.Errorf("%w: unknown access level %q for user %s on cluster %s, must be %s or %s",
			ErrInvalidInput, cu.AccessLevel, cu.UserID, cu.ComputeClusterID, models.ClusterAccessUser, models.ClusterAccessAdmin)
	}
	// A new row waits for review. The admin routes pass it as approved, with the admin who added it as the reviewer.
	if cu.ApprovalStatus == "" {
		cu.ApprovalStatus = models.ClusterAccountPending
	}
	switch cu.ApprovalStatus {
	case models.ClusterAccountPending:
	case models.ClusterAccountApproved:
		if cu.ReviewedBy == nil || *cu.ReviewedBy == "" {
			return nil, fmt.Errorf("%w: an approved cluster user needs a reviewer", ErrInvalidInput)
		}
		if cu.ReviewedAt == nil {
			now := nowUTC()
			cu.ReviewedAt = &now
		}
	default:
		return nil, fmt.Errorf("%w: a new cluster user can be %s or %s, not %s",
			ErrInvalidInput, models.ClusterAccountPending, models.ClusterAccountApproved, cu.ApprovalStatus)
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
		if err := s.clusterUsers.Create(ctx, tx, cu); err != nil {
			return err
		}
		if err := s.eventBus.Publish(ctx, tx, events.ComputeClusterUserCreateEvent, cu); err != nil {
			return err
		}
		// Approved at creation, so provisioning starts now.
		if cu.ApprovalStatus == models.ClusterAccountApproved {
			return s.eventBus.Publish(ctx, tx, events.ComputeClusterUserApproveEvent, cu)
		}
		return nil
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
	return cu, nil
}

// ListComputeClusterUsersByApproval returns cluster users in the given
// approval status with their user and cluster, newest first. An empty status means any status.
func (s *Service) ListComputeClusterUsersByApproval(ctx context.Context, status models.ClusterAccountApproval, limit, offset int) ([]store.ComputeClusterUserWithUser, int, error) {
	switch status {
	case "", models.ClusterAccountPending, models.ClusterAccountApproved, models.ClusterAccountDenied:
	default:
		return nil, 0, fmt.Errorf("%w: unknown approval status %q", ErrInvalidInput, status)
	}
	rows, total, err := s.clusterUsers.ListByApprovalStatus(ctx, status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list compute cluster users by approval: %w", err)
	}
	return rows, total, nil
}

// ApproveComputeClusterUser records the admin's approval and publishes the
// `approve event`, which is what creates the account on the cluster. A denied account can be approved later.
func (s *Service) ApproveComputeClusterUser(ctx context.Context, id, reviewerID string) (*models.ComputeClusterUser, error) {
	cu, err := s.clusterUserForReview(ctx, id, reviewerID)
	if err != nil {
		return nil, err
	}

	// Only a denied or pending cluster account can be approved
	if cu.ApprovalStatus == models.ClusterAccountApproved {
		return nil, fmt.Errorf("%w: cluster user %q is already approved", ErrAlreadyExists, id)
	}
	now := nowUTC()
	cu.ApprovalStatus = models.ClusterAccountApproved
	cu.ReviewedAt = &now
	cu.ReviewedBy = &reviewerID
	cu.ReviewNote = nil
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.clusterUsers.Review(ctx, tx, id, models.ClusterAccountApproved, reviewerID, "", now); err != nil {
			return err
		}
		if err := s.writeClusterUserAuditTx(ctx, tx, clusterUserAuditApproved, id, map[string]any{
			"actor_id":       reviewerID,
			"user_id":        cu.UserID,
			"cluster_id":     cu.ComputeClusterID,
			"local_username": cu.LocalUsername,
		}); err != nil {
			return err
		}
		return s.eventBus.Publish(ctx, tx, events.ComputeClusterUserApproveEvent, cu)
	}); err != nil {
		return nil, fmt.Errorf("approve compute cluster user: %w", err)
	}
	return cu, nil
}

// DenyComputeClusterUser records the admin's denial. No event is published,
// so the account is never created on the cluster. An approved account cannot
// be denied, since it may already exist on the cluster.
func (s *Service) DenyComputeClusterUser(ctx context.Context, id, reviewerID, note string) (*models.ComputeClusterUser, error) {
	cu, err := s.clusterUserForReview(ctx, id, reviewerID)
	if err != nil {
		return nil, err
	}
	switch cu.ApprovalStatus {
	case models.ClusterAccountApproved:
		return nil, fmt.Errorf("%w: cluster user %q is already approved, remove it instead", ErrAlreadyExists, id)
	case models.ClusterAccountDenied:
		return nil, fmt.Errorf("%w: cluster user %q is already denied", ErrAlreadyExists, id)
	}
	now := nowUTC()
	cu.ApprovalStatus = models.ClusterAccountDenied
	cu.ReviewedAt = &now
	cu.ReviewedBy = &reviewerID
	cu.ReviewNote = nil
	if note != "" {
		cu.ReviewNote = &note
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.clusterUsers.Review(ctx, tx, id, models.ClusterAccountDenied, reviewerID, note, now); err != nil {
			return err
		}
		return s.writeClusterUserAuditTx(ctx, tx, clusterUserAuditDenied, id, map[string]any{
			"actor_id":       reviewerID,
			"user_id":        cu.UserID,
			"cluster_id":     cu.ComputeClusterID,
			"local_username": cu.LocalUsername,
			"note":           note,
		})
	}); err != nil {
		return nil, fmt.Errorf("deny compute cluster user: %w", err)
	}
	return cu, nil
}

func (s *Service) clusterUserForReview(ctx context.Context, id, reviewerID string) (*models.ComputeClusterUser, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: compute cluster user id is required", ErrInvalidInput)
	}
	if reviewerID == "" {
		return nil, fmt.Errorf("%w: reviewer is required", ErrInvalidInput)
	}
	cu, err := s.clusterUsers.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("lookup compute cluster user: %w", err)
	}
	if cu == nil {
		return nil, ErrNotFound
	}
	return cu, nil
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
		if err := s.clusterUsers.Update(ctx, tx, cu); err != nil {
			return err
		}
		return s.eventBus.Publish(ctx, tx, events.ComputeClusterUserUpdateEvent, cu)
	}); err != nil {
		return fmt.Errorf("update compute cluster user: %w", err)
	}
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
		if err := s.clusterUsers.Delete(ctx, tx, id); err != nil {
			return err
		}
		return s.eventBus.Publish(ctx, tx, events.ComputeClusterUserDeleteEvent, cu)
	}); err != nil {
		return fmt.Errorf("delete compute cluster user: %w", err)
	}
	return nil
}
