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
	"slices"
	"strings"
	"time"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/posix"
)

const userAgreementAuditAccepted = "USER_AGREEMENT_ACCEPTED"

type OnboardingInput struct {
	Username    string
	AcceptedIDs []string
}

type UsernameCheck struct {
	Valid      bool   `json:"valid"`
	Available  bool   `json:"available"`
	Suggestion string `json:"suggestion"`
}

// PendingUserAgreements returns the current agreement if the user has not accepted it.
func (s *Service) PendingUserAgreements(ctx context.Context, userID string) ([]models.UserAgreement, error) {
	return s.agreements.ListPendingForUser(ctx, userID)
}

// AcceptUserAgreement records the acceptance once. A repeat returns the first record.
func (s *Service) AcceptUserAgreement(ctx context.Context, userID, agreementID string) (*models.UserAgreementAcceptance, error) {
	a, err := s.GetUserAgreement(ctx, agreementID)
	if err != nil {
		return nil, err
	}
	if a.Status != models.UserAgreementPublished {
		return nil, fmt.Errorf("%w: only a published agreement can be accepted", ErrInvalidInput)
	}
	var acc *models.UserAgreementAcceptance
	if err := s.inTx(ctx, func(tx *sql.Tx) (err error) {
		acc, err = s.acceptTx(ctx, tx, userID, a)
		return err
	}); err != nil {
		return nil, err
	}
	return acc, nil
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

func (s *Service) acceptTx(ctx context.Context, tx *sql.Tx, userID string, a *models.UserAgreement) (*models.UserAgreementAcceptance, error) {
	acc := &models.UserAgreementAcceptance{
		ID:          newID(),
		UserID:      userID,
		AgreementID: a.ID,
		AcceptedAt:  nowUTC(),
		BodyHash:    a.BodyHash,
	}
	created, err := s.acceptances.Create(ctx, tx, acc)
	if err != nil {
		return nil, fmt.Errorf("insert acceptance: %w", err)
	}
	if !created {
		return s.acceptances.FindTx(ctx, tx, userID, a.ID)
	}
	if err := s.writePrivilegeAuditTx(ctx, tx, userAgreementAuditAccepted, userID, map[string]any{
		"agreement_id":  a.ID,
		"version_label": a.VersionLabel,
		"body_hash":     a.BodyHash,
	}); err != nil {
		return nil, err
	}
	return acc, s.eventBus.Publish(ctx, tx, events.UserAgreementAcceptanceCreateEvent, acc)
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
