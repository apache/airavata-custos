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
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
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
	base, err := usernameBase(firstName, lastName, email, posix.Prefix())
	if err != nil {
		return "", err
	}
	name, err := firstFreeUsername(base, func(candidate string) (bool, error) {
		existing, err := s.clusterUsers.FindByClusterAndLocalUsername(ctx, clusterID, candidate)
		if err != nil {
			return false, fmt.Errorf("check username: %w", err)
		}
		return existing == nil, nil
	})
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", fmt.Errorf("username allocator exhausted for base %q", base)
	}
	return name, nil
}

// usernameBase falls back to the email local part when both names are empty after normalizing.
func usernameBase(firstName, lastName, email, prefix string) (string, error) {
	u := &models.User{FirstName: firstName, LastName: lastName}
	if posix.Normalize(firstName) == "" && posix.Normalize(lastName) == "" {
		u.LastName, _, _ = strings.Cut(email, "@")
	}
	base, _, err := posix.BuildBase(u, prefix)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	if prefix == "" {
		base = strings.TrimPrefix(base, "-")
	}
	return base, nil
}

// firstFreeUsername tries base, base2, base3 and so on. It returns "" when every suffix is taken.
func firstFreeUsername(base string, free func(string) (bool, error)) (string, error) {
	for n := range posix.MaxCollisionSuffix {
		name := posix.Candidate(base, n)
		ok, err := free(name)
		if err != nil {
			return "", err
		}
		if ok {
			return name, nil
		}
	}
	return "", nil
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

	// Onboarding is an admin creating the user account on purpose, so it counts as the approval.
	reviewedAt := nowUTC()
	clusterUser := &models.ComputeClusterUser{
		ID:               newID(),
		ComputeClusterID: in.ComputeClusterID,
		UserID:           created.ID,
		LocalUsername:    in.Username,
		AccessLevel:      models.ClusterAccessUser,
		ApprovalStatus:   models.ClusterAccountApproved,
		ReviewedAt:       &reviewedAt,
		ReviewedBy:       &in.OnboardedBy,
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

		if err := s.eventBus.Publish(ctx, tx, events.UserCreateEvent, created); err != nil {
			return err
		}
		if needsCluster {
			if err := s.eventBus.Publish(ctx, tx, events.ComputeClusterUserCreateEvent, clusterUser); err != nil {
				return err
			}
			// Approved at creation, so provisioning starts now.
			return s.eventBus.Publish(ctx, tx, events.ComputeClusterUserApproveEvent, clusterUser)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if in.ClusterAdmin {
		slog.InfoContext(ctx, "cluster admin account created", "user_id", created.ID, "cluster_id", clusterUser.ComputeClusterID, "actor_id", in.OnboardedBy)
	}

	return created, nil
}

type OnboardingInput struct {
	Username    string
	AcceptedIDs []string
}

type UsernameCheck struct {
	Valid      bool   `json:"valid"`
	Available  bool   `json:"available"`
	Suggestion string `json:"suggestion"`
}

// CheckOnboardingUsername checks a chosen login against every cluster the user has an account and suggests a free one.
func (s *Service) CheckOnboardingUsername(ctx context.Context, userID, candidate string) (*UsernameCheck, error) {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	accounts, err := s.clusterUsers.FindByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("lookup cluster accounts: %w", err)
	}
	free := func(name string) (bool, error) { return s.usernameFree(ctx, accounts, userID, name) }

	base, err := usernameBase(user.FirstName, user.LastName, user.Email, "")
	if err != nil {
		return nil, err
	}
	check := &UsernameCheck{}
	if check.Suggestion, err = firstFreeUsername(base, free); err != nil {
		return nil, err
	}

	if candidate == "" {
		return check, nil
	}
	if check.Valid = s.ValidateUsername(candidate) == nil; check.Valid {
		if check.Available, err = free(candidate); err != nil {
			return nil, err
		}
	}
	return check, nil
}

// usernameFree reports whether no other user holds name on any cluster in accounts.
func (s *Service) usernameFree(ctx context.Context, accounts []models.ComputeClusterUser, userID, name string) (bool, error) {
	for _, acc := range accounts {
		taken, err := s.clusterUsers.FindByClusterAndLocalUsername(ctx, acc.ComputeClusterID, name)
		if err != nil {
			return false, fmt.Errorf("check username: %w", err)
		}
		if taken != nil && taken.UserID != userID {
			return false, nil
		}
	}
	return true, nil
}

// SetReservedUsernames sets the names a user cannot choose as a cluster login.
func (s *Service) SetReservedUsernames(names []string) {
	s.reservedUsernames = make(map[string]bool, len(names))
	for _, n := range names {
		s.reservedUsernames[strings.ToLower(n)] = true
	}
}

// ValidateUsername checks a username a person supplied. The login format and the configured reserved list.
func (s *Service) ValidateUsername(name string) error {
	if !posix.ValidChosen(name) {
		return fmt.Errorf("%w: username %q must be 1 to 32 characters, start with a lowercase letter, then use lowercase letters, digits, - or _", ErrInvalidInput, name)
	}
	if s.reservedUsernames[name] {
		return fmt.Errorf("%w: username %q is reserved", ErrInvalidInput, name)
	}
	return nil
}

// CompleteOnboarding binds the caller's sub to their PENDING user. The request
// must accept the current agreement. Nothing is written unless all of it succeeds.
func (s *Service) CompleteOnboarding(ctx context.Context, claims *identity.Claims, in OnboardingInput) (*models.User, error) {
	matched, err := s.matchPendingByEmail(ctx, claims)
	if errors.Is(err, identity.ErrUserAlreadyOIDCLinked) || errors.Is(err, identity.ErrUserNotPending) {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	if err != nil {
		return nil, err
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		// Recheck under the row lock. A parallel submit may have onboarded this user already.
		user, err := s.users.FindByIDTx(ctx, tx, matched.ID)
		if err != nil {
			return fmt.Errorf("lock user: %w", err)
		}
		if user == nil || user.Status != models.UserPending {
			return fmt.Errorf("%w: user is not waiting for onboarding", ErrInvalidInput)
		}
		bound, err := s.userIdentities.FindByUserAndSource(ctx, user.ID, identitySourceOIDC)
		if err != nil {
			return fmt.Errorf("lookup oidc identity: %w", err)
		}
		if bound != nil {
			return fmt.Errorf("%w: user already has a sign-in identity", ErrInvalidInput)
		}

		current, err := s.agreements.FindCurrent(ctx)
		if err != nil {
			return fmt.Errorf("lookup current agreement: %w", err)
		}
		if current != nil && !slices.Contains(in.AcceptedIDs, current.ID) {
			return fmt.Errorf("%w: agreement %q must be accepted", ErrInvalidInput, current.VersionLabel)
		}

		if user.Type == models.UserTypeClusterLocal && in.Username != "" {
			if err := s.renameClusterAccountsTx(ctx, tx, user, in.Username); err != nil {
				return err
			}
		}

		if err := s.bindSubTx(ctx, tx, user, claims); err != nil {
			return err
		}
		if current != nil {
			if _, err := s.acceptTx(ctx, tx, user.ID, current); err != nil {
				return err
			}
		}
		user.Status = models.UserActive
		return s.eventBus.Publish(ctx, tx, events.UserOnboardedEvent, *user)
	}); err != nil {
		return nil, err
	}
	return s.GetUser(ctx, matched.ID)
}

// A locked name is skipped, not refused, so onboarding still finishes.
func (s *Service) renameClusterAccountsTx(ctx context.Context, tx *sql.Tx, user *models.User, name string) error {
	accounts, err := s.clusterUsers.FindByUserTx(ctx, tx, user.ID)
	if err != nil {
		return fmt.Errorf("lookup cluster accounts: %w", err)
	}
	editable, err := s.UsernameEditable(ctx, user, accounts)
	if err != nil || !editable {
		return err
	}
	if err := s.ValidateUsername(name); err != nil {
		return err
	}
	for i := range accounts {
		acc := &accounts[i]
		if acc.LocalUsername == name {
			continue
		}
		free, err := s.usernameFree(ctx, accounts[i:i+1], user.ID, name)
		if err != nil {
			return err
		}
		if !free {
			return fmt.Errorf("%w: username %q is taken on cluster %q", ErrAlreadyExists, name, acc.ComputeClusterID)
		}
		acc.LocalUsername = name
		if err := s.clusterUsers.Update(ctx, tx, acc); err != nil {
			return fmt.Errorf("rename cluster account: %w", err)
		}
	}
	return nil
}

// UsernameEditable is false once the registry knows the user, since the name may already be on the cluster.
func (s *Service) UsernameEditable(ctx context.Context, user *models.User, accounts []models.ComputeClusterUser) (bool, error) {
	if user.Type != models.UserTypeClusterLocal || len(accounts) == 0 {
		return false, nil
	}
	registered, err := s.userIdentities.FindByUserAndSource(ctx, user.ID, identitySourceComanage)
	if err != nil {
		return false, fmt.Errorf("lookup registry identity: %w", err)
	}
	return registered == nil, nil
}

// SetupStep is one provisioning step shown after onboarding. Name is the cluster or
// allocation the step is for, and DoneAt is nil until it is done.
type SetupStep struct {
	Key    string     `json:"key"`
	Name   string     `json:"name,omitempty"`
	DoneAt *time.Time `json:"done_at,omitempty"`
}

// SetupSteps returns the provisioning steps that apply to the user and when each was done.
func (s *Service) SetupSteps(ctx context.Context, userID string) ([]SetupStep, error) {
	registered, err := s.userIdentities.FindByUserAndSource(ctx, userID, identitySourceComanage)
	if err != nil {
		return nil, fmt.Errorf("lookup registry identity: %w", err)
	}
	var registeredAt *time.Time
	if registered != nil {
		registeredAt = &registered.CreatedAt
	}
	steps := []SetupStep{{Key: "registry", DoneAt: registeredAt}}

	accounts, err := s.clusterUsers.FindByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("lookup cluster accounts: %w", err)
	}
	for _, a := range accounts {
		cluster, err := s.GetComputeCluster(ctx, a.ComputeClusterID)
		if err != nil {
			return nil, err
		}
		steps = append(steps, SetupStep{Key: "cluster_account", Name: cluster.Name, DoneAt: a.ProvisionedAt})
	}

	allocations, err := s.ListComputeAllocationsForParticipant(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, a := range allocations {
		if a.PosixGroup != nil {
			steps = append(steps, SetupStep{Key: "project_group", Name: a.Name, DoneAt: a.GroupProvisionedAt})
		}
	}
	return steps, nil
}
