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

// CreateUser persists a new user. If user.ID is empty a new UUID is generated.
// The referenced organization must already exist.
func (s *Service) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	if user == nil {
		return nil, fmt.Errorf("%w: user is nil", ErrInvalidInput)
	}
	if user.Email == "" {
		return nil, fmt.Errorf("%w: user email is required", ErrInvalidInput)
	}
	if user.OrganizationID == "" {
		return nil, fmt.Errorf("%w: user organization_id is required", ErrInvalidInput)
	}

	if org, err := s.orgs.FindByID(ctx, user.OrganizationID); err != nil {
		return nil, fmt.Errorf("verify organization: %w", err)
	} else if org == nil {
		return nil, fmt.Errorf("%w: organization %q does not exist", ErrInvalidInput, user.OrganizationID)
	}

	if existing, err := s.users.FindByEmail(ctx, user.Email); err != nil {
		return nil, fmt.Errorf("lookup user by email: %w", err)
	} else if existing != nil {
		return nil, fmt.Errorf("%w: user with email %q", ErrAlreadyExists, user.Email)
	}

	if user.ID == "" {
		user.ID = newID()
	}
	if user.Status == "" {
		user.Status = models.UserPending
	}
	if user.Type == "" {
		user.Type = models.UserTypeClusterLocal
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.users.Create(ctx, tx, user)
	}); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	s.eventBus.Publish(ctx, events.UserCreateEvent, user)
	return user, nil
}

// GetUser retrieves a user by ID. Returns ErrNotFound when no user matches.
func (s *Service) GetUser(ctx context.Context, id string) (*models.User, error) {
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}

// GetUserByUserIdentity returns the user owning the user identity uniquely
// identified by (source, externalID). Returns ErrNotFound when no such binding exists.
func (s *Service) GetUserByUserIdentity(ctx context.Context, source, externalID string) (*models.User, error) {
	if source == "" {
		return nil, fmt.Errorf("%w: source is required", ErrInvalidInput)
	}
	if externalID == "" {
		return nil, fmt.Errorf("%w: external_id is required", ErrInvalidInput)
	}
	ident, err := s.userIdentities.FindBySourceAndExternalID(ctx, source, externalID)
	if err != nil {
		return nil, fmt.Errorf("lookup user identity: %w", err)
	}
	if ident == nil {
		return nil, ErrNotFound
	}
	return s.GetUser(ctx, ident.UserID)
}

// GetUserByEmail retrieves a user by email.
func (s *Service) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
}

// ListUsersByOrganization returns all users belonging to an organization.
func (s *Service) ListUsersByOrganization(ctx context.Context, organizationID string) ([]models.User, error) {
	users, err := s.users.FindByOrganization(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list users by organization: %w", err)
	}
	return users, nil
}

// UpdateUser persists changes to an existing user. Fields left blank/zero on
// the supplied record fall back to the stored value.
func (s *Service) UpdateUser(ctx context.Context, user *models.User) error {
	if user == nil || user.ID == "" {
		return fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}
	existing, err := s.users.FindByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("lookup user: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}

	if user.Email != "" && user.Email != existing.Email {
		if other, err := s.users.FindByEmail(ctx, user.Email); err != nil {
			return fmt.Errorf("lookup user by email: %w", err)
		} else if other != nil && other.ID != user.ID {
			return fmt.Errorf("%w: user with email %q", ErrAlreadyExists, user.Email)
		}
	}

	if user.OrganizationID == "" {
		user.OrganizationID = existing.OrganizationID
	}
	if user.FirstName == "" {
		user.FirstName = existing.FirstName
	}
	if user.LastName == "" {
		user.LastName = existing.LastName
	}
	if user.MiddleName == "" {
		user.MiddleName = existing.MiddleName
	}
	if user.Email == "" {
		user.Email = existing.Email
	}
	if user.Status == "" {
		user.Status = existing.Status
	}
	if user.Type == "" {
		user.Type = existing.Type
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.users.Update(ctx, tx, user)
	}); err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	s.eventBus.Publish(ctx, events.UserUpdateEvent, user)
	return nil
}

// UpdateUserStatus sets the lifecycle status of the user identified by id.
// Other fields are preserved.
func (s *Service) UpdateUserStatus(ctx context.Context, id string, status models.UserStatus) (*models.User, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}
	if status == "" {
		return nil, fmt.Errorf("%w: status is required", ErrInvalidInput)
	}
	existing, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.users.UpdateStatus(ctx, tx, id, status)
	}); err != nil {
		return nil, fmt.Errorf("update user status: %w", err)
	}
	existing.Status = status

	s.eventBus.Publish(ctx, events.UserUpdateEvent, existing)
	return existing, nil
}

// DeleteUser removes a user by ID.
func (s *Service) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}
	existing, err := s.users.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup user: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.users.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	s.eventBus.Publish(ctx, events.UserDeleteEvent, existing)
	return nil
}
