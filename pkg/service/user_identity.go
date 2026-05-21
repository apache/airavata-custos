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

// CreateUserIdentity persists a new user identity. If e.ID is empty, a new
// UUID is generated. The referenced user must already exist and the (source, external_id) pair is unique.
func (s *Service) CreateUserIdentity(ctx context.Context, e *models.UserIdentity) (*models.UserIdentity, error) {
	if e == nil {
		return nil, fmt.Errorf("%w: user identity is nil", ErrInvalidInput)
	}
	if e.UserID == "" {
		return nil, fmt.Errorf("%w: user identity user_id is required", ErrInvalidInput)
	}
	if e.Source == "" {
		return nil, fmt.Errorf("%w: user identity source is required", ErrInvalidInput)
	}
	if e.ExternalID == "" {
		return nil, fmt.Errorf("%w: user identity external_id is required", ErrInvalidInput)
	}

	if user, err := s.users.FindByID(ctx, e.UserID); err != nil {
		return nil, fmt.Errorf("verify user: %w", err)
	} else if user == nil {
		return nil, fmt.Errorf("%w: user %q does not exist", ErrInvalidInput, e.UserID)
	}

	if existing, err := s.userIdentities.FindBySourceAndExternalID(ctx, e.Source, e.ExternalID); err != nil {
		return nil, fmt.Errorf("lookup user identity: %w", err)
	} else if existing != nil {
		return nil, fmt.Errorf("%w: user identity for source %q, external_id %q", ErrAlreadyExists, e.Source, e.ExternalID)
	}

	if e.ID == "" {
		e.ID = newID()
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.userIdentities.Create(ctx, tx, e)
	}); err != nil {
		return nil, fmt.Errorf("create user identity: %w", err)
	}

	s.eventBus.Publish(events.UserIdentityCreateEvent, e)
	return e, nil
}

// GetUserIdentity retrieves a user identity by ID. Returns ErrNotFound when no row matches.
func (s *Service) GetUserIdentity(ctx context.Context, id string) (*models.UserIdentity, error) {
	e, err := s.userIdentities.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user identity: %w", err)
	}
	if e == nil {
		return nil, ErrNotFound
	}
	return e, nil
}

// GetUserIdentityBySourceAndExternalID retrieves the unique user identity for
// the given (source, external_id) pair.
func (s *Service) GetUserIdentityBySourceAndExternalID(ctx context.Context, source, externalID string) (*models.UserIdentity, error) {
	if source == "" {
		return nil, fmt.Errorf("%w: source is required", ErrInvalidInput)
	}
	if externalID == "" {
		return nil, fmt.Errorf("%w: external_id is required", ErrInvalidInput)
	}
	e, err := s.userIdentities.FindBySourceAndExternalID(ctx, source, externalID)
	if err != nil {
		return nil, fmt.Errorf("get user identity by source/external_id: %w", err)
	}
	if e == nil {
		return nil, ErrNotFound
	}
	return e, nil
}

// GetUserIdentityByOIDCSub retrieves the first user identity matching the given OIDC subject.
func (s *Service) GetUserIdentityByOIDCSub(ctx context.Context, oidcSub string) (*models.UserIdentity, error) {
	if oidcSub == "" {
		return nil, fmt.Errorf("%w: oidc_sub is required", ErrInvalidInput)
	}
	e, err := s.userIdentities.FindByOIDCSub(ctx, oidcSub)
	if err != nil {
		return nil, fmt.Errorf("get user identity by oidc_sub: %w", err)
	}
	if e == nil {
		return nil, ErrNotFound
	}
	return e, nil
}

// ListUserIdentitiesForUser returns every user identity belonging to the given user.
func (s *Service) ListUserIdentitiesForUser(ctx context.Context, userID string) ([]models.UserIdentity, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	out, err := s.userIdentities.FindByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user identities by user: %w", err)
	}
	return out, nil
}

// UpdateUserIdentity persists changes to an existing user identity. Fields
// left blank/zero on the supplied record fall back to the stored value.
func (s *Service) UpdateUserIdentity(ctx context.Context, e *models.UserIdentity) error {
	if e == nil || e.ID == "" {
		return fmt.Errorf("%w: user identity id is required", ErrInvalidInput)
	}
	existing, err := s.userIdentities.FindByID(ctx, e.ID)
	if err != nil {
		return fmt.Errorf("lookup user identity: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}
	if e.UserID == "" {
		e.UserID = existing.UserID
	}
	if e.Source == "" {
		e.Source = existing.Source
	}
	if e.ExternalID == "" {
		e.ExternalID = existing.ExternalID
	}
	if e.Email == "" {
		e.Email = existing.Email
	}
	if e.OIDCSub == "" {
		e.OIDCSub = existing.OIDCSub
	}
	if e.Metadata == "" {
		e.Metadata = existing.Metadata
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.userIdentities.Update(ctx, tx, e)
	}); err != nil {
		return fmt.Errorf("update user identity: %w", err)
	}

	s.eventBus.Publish(events.UserIdentityUpdateEvent, e)
	return nil
}

// DeleteUserIdentity removes a user identity by ID.
func (s *Service) DeleteUserIdentity(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: user identity id is required", ErrInvalidInput)
	}
	e, err := s.userIdentities.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup user identity: %w", err)
	}
	if e == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.userIdentities.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete user identity: %w", err)
	}

	s.eventBus.Publish(events.UserIdentityDeleteEvent, e)
	return nil
}
