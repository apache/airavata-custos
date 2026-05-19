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

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.users.Create(ctx, tx, user)
	}); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
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

// GetUserByExternalIdentity resolves a user via their (source, external_id)
// entry. Returns ErrNotFound when either the external identity does not
// exist or the user it points to has been deleted.
func (s *Service) GetUserByExternalIdentity(ctx context.Context, source, externalID string) (*models.User, error) {
	ext, err := s.extIDs.FindBySourceAndExternalID(ctx, source, externalID)
	if err != nil {
		return nil, fmt.Errorf("lookup external identity: %w", err)
	}
	if ext == nil {
		return nil, ErrNotFound
	}
	u, err := s.users.FindByID(ctx, ext.UserID)
	if err != nil {
		return nil, fmt.Errorf("lookup user: %w", err)
	}
	if u == nil {
		return nil, ErrNotFound
	}
	return u, nil
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

// UpdateUser persists changes to an existing user.
func (s *Service) UpdateUser(ctx context.Context, user *models.User) error {
	if user == nil || user.ID == "" {
		return fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.users.Update(ctx, tx, user)
	}); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// DeleteUser removes a user by ID.
func (s *Service) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: user id is required", ErrInvalidInput)
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.users.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
