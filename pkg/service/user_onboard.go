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
	"log/slog"
	"strconv"
	"strings"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/posix"
)

// OnboardUserInput is what an admin fills in to create a user. A researcher
// has both admin flags false. The user type, role, and cluster account come with the flags.
type OnboardUserInput struct {
	Email            string
	FirstName        string
	LastName         string
	OrganizationID   string // defaults to the system org
	Username         string // cluster login name, generated when empty
	ComputeClusterID string // taken from the allocation
	AllocationID     string // researchers only
	PortalAdmin      bool
	ClusterAdmin     bool
	OnboardedBy      string
}

// generateClusterUsername builds a base name with the same rule ingestion uses.
func (s *Service) generateClusterUsername(ctx context.Context, clusterID, firstName, lastName, email string) (string, error) {
	u := &models.User{FirstName: firstName, LastName: lastName}
	if posix.Normalize(firstName) == "" && posix.Normalize(lastName) == "" {
		u.LastName, _, _ = strings.Cut(email, "@")
	}
	base, _, err := posix.BuildBase(u, posix.Prefix())
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	for n := 0; n < posix.MaxCollisionSuffix; n++ {
		candidate := base
		if n > 0 {
			candidate = base + strconv.Itoa(n+1)
		}
		existing, err := s.clusterUsers.FindByClusterAndLocalUsername(ctx, clusterID, candidate)
		if err != nil {
			return "", fmt.Errorf("check username: %w", err)
		}
		if existing == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("username allocator exhausted for base %q", base)
}

// OnboardUser creates the same rows ingestion does, so the existing `cluster user create event` drives provisioning.
func (s *Service) OnboardUser(ctx context.Context, in OnboardUserInput) (*models.User, error) {
	// A cluster admin still gets a cluster account, a portal-only admin does not.
	admin := in.PortalAdmin || in.ClusterAdmin
	needsCluster := !admin || in.ClusterAdmin

	if admin && in.AllocationID != "" {
		return nil, fmt.Errorf("%w: an allocation can only be attached to a researcher", ErrInvalidInput)
	}
	if in.OnboardedBy == "" {
		return nil, fmt.Errorf("%w: onboarded_by is required", ErrInvalidInput)
	}

	// Only the users with `roles:manage` can onboard admin users.
	if admin {
		has, err := s.HasPrivilege(ctx, in.OnboardedBy, models.RolesManage)
		if err != nil {
			return nil, err
		}
		if !has {
			return nil, fmt.Errorf("%w: caller does not hold %s", ErrInvalidInput, models.RolesManage)
		}
	}

	var alloc *models.ComputeAllocation
	if in.AllocationID != "" {
		a, err := s.GetComputeAllocation(ctx, in.AllocationID)
		if err != nil {
			return nil, fmt.Errorf("lookup allocation: %w", err)
		}
		if a.Status != models.ACTIVE {
			return nil, fmt.Errorf("%w: allocation %q is %s", ErrInvalidInput, a.ID, a.Status)
		}
		alloc = a
		// The cluster account and the membership have to land on the same
		// cluster, or the association mapper never finds the account.
		if in.ComputeClusterID != "" && in.ComputeClusterID != a.ComputeClusterID {
			return nil, fmt.Errorf("%w: allocation %q is on cluster %q", ErrInvalidInput, a.ID, a.ComputeClusterID)
		}
		in.ComputeClusterID = a.ComputeClusterID
		// TODO should the allocation's source (stored in the project#Origination) be limited to "SYSTEM" level?
	}
	// TODO let the admin pick the cluster when more than one is registered.
	if needsCluster && in.ComputeClusterID == "" {
		clusters, err := s.ListComputeClusters(ctx)
		if err != nil {
			return nil, fmt.Errorf("list compute clusters: %w", err)
		}
		// TODO update the logic and support cluster selection
		if len(clusters) != 1 {
			return nil, fmt.Errorf("%w: pick a cluster, %d are registered", ErrInvalidInput, len(clusters))
		}
		in.ComputeClusterID = clusters[0].ID
	}

	// Resolved before the transaction so a taken name fails with a clear message instead of a constraint violation.
	if needsCluster {
		if in.Username == "" {
			generated, err := s.generateClusterUsername(ctx, in.ComputeClusterID, in.FirstName, in.LastName, in.Email)
			if err != nil {
				return nil, err
			}
			in.Username = generated
		} else {
			if len(in.Username) > posix.MaxLoginLen {
				return nil, fmt.Errorf("%w: username %q is longer than %d characters", ErrInvalidInput, in.Username, posix.MaxLoginLen)
			}
			taken, err := s.clusterUsers.FindByClusterAndLocalUsername(ctx, in.ComputeClusterID, in.Username)
			if err != nil {
				return nil, fmt.Errorf("check username: %w", err)
			}
			if taken != nil {
				return nil, fmt.Errorf("%w: username %q is taken on cluster %q", ErrAlreadyExists, in.Username, in.ComputeClusterID)
			}
		}
	}

	userType := models.UserTypeClusterLocal
	if !needsCluster {
		userType = models.UserTypeSystem
	}
	created := &models.User{
		OrganizationID: in.OrganizationID,
		Email:          in.Email,
		FirstName:      in.FirstName,
		LastName:       in.LastName,
		Type:           userType,
	}
	if created.OrganizationID == "" {
		// Created ahead of the main transaction so the user check below can see it.
		// If the rest fails, the empty system org left behind, which is harmless.
		if err := s.inTx(ctx, func(tx *sql.Tx) error {
			_, err := s.ensureSystemOrgTx(ctx, tx)
			return err
		}); err != nil {
			return nil, err
		}
		created.OrganizationID = bootstrapSystemOrgID
	}
	if err := s.checkNewUser(ctx, created); err != nil {
		return nil, err
	}

	if needsCluster {
		if cluster, err := s.clusters.FindByID(ctx, in.ComputeClusterID); err != nil {
			return nil, fmt.Errorf("lookup compute cluster: %w", err)
		} else if cluster == nil {
			return nil, fmt.Errorf("%w: compute cluster %q not found", ErrInvalidInput, in.ComputeClusterID)
		}
	}

	clusterUser := &models.ComputeClusterUser{
		ID:               newID(),
		ComputeClusterID: in.ComputeClusterID,
		UserID:           created.ID,
		LocalUsername:    in.Username,
		AccessLevel:      models.ClusterAccessUser,
	}
	if in.ClusterAdmin {
		clusterUser.AccessLevel = models.ClusterAccessAdmin
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.users.Create(ctx, tx, created); err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		if in.PortalAdmin {
			role, err := s.ensureAdminRoleTx(ctx, tx)
			if err != nil {
				return err
			}
			if err := s.userRoles.Create(ctx, tx, &models.UserRole{
				UserID:    created.ID,
				RoleID:    role.ID,
				GrantedBy: stringPtrOrNil(in.OnboardedBy),
				GrantedAt: nowUTC(),
				Reason:    stringPtrOrNil("user onboarding"),
			}); err != nil {
				return fmt.Errorf("grant admin role: %w", err)
			}

			if err := s.writeRoleAuditTx(ctx, tx, userRoleAuditGranted, created.ID, map[string]any{
				"actor_id":  in.OnboardedBy,
				"role_id":   role.ID,
				"role_name": role.Name,
				"reason":    "user onboarding",
			}); err != nil {
				return err
			}
		}

		if needsCluster {
			if err := s.clusterUsers.Create(ctx, tx, clusterUser); err != nil {
				return fmt.Errorf("create cluster user: %w", err)
			}
			if in.ClusterAdmin {
				if err := s.writeClusterUserAuditTx(ctx, tx, clusterUserAuditAdminGranted, clusterUser.ID, map[string]any{
					"actor_id":       in.OnboardedBy,
					"user_id":        created.ID,
					"cluster_id":     clusterUser.ComputeClusterID,
					"local_username": clusterUser.LocalUsername,
					"reason":         "user onboarding",
				}); err != nil {
					return err
				}
			}
		}

		if alloc != nil {
			if err := s.memberships.Create(ctx, tx, &models.ComputeAllocationMembership{
				ID:                  newID(),
				ComputeAllocationID: alloc.ID,
				UserID:              created.ID,
				StartTime:           alloc.StartTime,
				EndTime:             alloc.EndTime,
				MembershipStatus:    models.ACTIVE,
			}); err != nil {
				return fmt.Errorf("create allocation membership: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Published after commit so a subscriber never provisions a rolled-back row.
	s.eventBus.Publish(ctx, events.UserCreateEvent, created)
	if needsCluster {
		s.eventBus.Publish(ctx, events.ComputeClusterUserCreateEvent, clusterUser)
	}
	if in.ClusterAdmin {
		slog.Info("cluster admin account created", "user_id", created.ID, "cluster_id", clusterUser.ComputeClusterID, "actor_id", in.OnboardedBy)
	}

	return created, nil
}
