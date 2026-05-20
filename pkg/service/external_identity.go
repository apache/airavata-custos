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

// CreateExternalIdentity persists a new external identity. If e.ID is empty, a
// new UUID is generated. The referenced user must already exist and the
// (source, external_id) pair is unique.
func (s *Service) CreateExternalIdentity(ctx context.Context, e *models.ExternalIdentity) (*models.ExternalIdentity, error) {
	if e == nil {
		return nil, fmt.Errorf("%w: external identity is nil", ErrInvalidInput)
	}
	if e.UserID == "" {
		return nil, fmt.Errorf("%w: external identity user_id is required", ErrInvalidInput)
	}
	if e.Source == "" {
		return nil, fmt.Errorf("%w: external identity source is required", ErrInvalidInput)
	}
	if e.ExternalID == "" {
		return nil, fmt.Errorf("%w: external identity external_id is required", ErrInvalidInput)
	}

	if user, err := s.users.FindByID(ctx, e.UserID); err != nil {
		return nil, fmt.Errorf("verify user: %w", err)
	} else if user == nil {
		return nil, fmt.Errorf("%w: user %q does not exist", ErrInvalidInput, e.UserID)
	}

	if existing, err := s.extIDs.FindBySourceAndExternalID(ctx, e.Source, e.ExternalID); err != nil {
		return nil, fmt.Errorf("lookup external identity: %w", err)
	} else if existing != nil {
		return nil, fmt.Errorf("%w: external identity for source %q, external_id %q", ErrAlreadyExists, e.Source, e.ExternalID)
	}

	if e.ID == "" {
		e.ID = newID()
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.extIDs.Create(ctx, tx, e)
	}); err != nil {
		return nil, fmt.Errorf("create external identity: %w", err)
	}

	s.eventBus.Publish(events.ExternalIdentityCreateEvent, e)
	return e, nil
}

// GetExternalIdentity retrieves an external identity by ID. Returns
// ErrNotFound when no row matches.
func (s *Service) GetExternalIdentity(ctx context.Context, id string) (*models.ExternalIdentity, error) {
	e, err := s.extIDs.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get external identity: %w", err)
	}
	if e == nil {
		return nil, ErrNotFound
	}
	return e, nil
}

// GetExternalIdentityBySourceAndExternalID retrieves the unique external
// identity for the given (source, external_id) pair.
func (s *Service) GetExternalIdentityBySourceAndExternalID(ctx context.Context, source, externalID string) (*models.ExternalIdentity, error) {
	if source == "" {
		return nil, fmt.Errorf("%w: source is required", ErrInvalidInput)
	}
	if externalID == "" {
		return nil, fmt.Errorf("%w: external_id is required", ErrInvalidInput)
	}
	e, err := s.extIDs.FindBySourceAndExternalID(ctx, source, externalID)
	if err != nil {
		return nil, fmt.Errorf("get external identity by source/external_id: %w", err)
	}
	if e == nil {
		return nil, ErrNotFound
	}
	return e, nil
}

// GetExternalIdentityByOIDCSub retrieves the first external identity matching
// the given OIDC subject.
func (s *Service) GetExternalIdentityByOIDCSub(ctx context.Context, oidcSub string) (*models.ExternalIdentity, error) {
	if oidcSub == "" {
		return nil, fmt.Errorf("%w: oidc_sub is required", ErrInvalidInput)
	}
	e, err := s.extIDs.FindByOIDCSub(ctx, oidcSub)
	if err != nil {
		return nil, fmt.Errorf("get external identity by oidc_sub: %w", err)
	}
	if e == nil {
		return nil, ErrNotFound
	}
	return e, nil
}

// ListExternalIdentitiesForUser returns every external identity belonging to
// the given user.
func (s *Service) ListExternalIdentitiesForUser(ctx context.Context, userID string) ([]models.ExternalIdentity, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	out, err := s.extIDs.FindByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list external identities by user: %w", err)
	}
	return out, nil
}

// UpdateExternalIdentity persists changes to an existing external identity.
// Fields left blank/zero on the supplied record fall back to the stored value.
func (s *Service) UpdateExternalIdentity(ctx context.Context, e *models.ExternalIdentity) error {
	if e == nil || e.ID == "" {
		return fmt.Errorf("%w: external identity id is required", ErrInvalidInput)
	}
	existing, err := s.extIDs.FindByID(ctx, e.ID)
	if err != nil {
		return fmt.Errorf("lookup external identity: %w", err)
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
	if e.OIDCSub == "" {
		e.OIDCSub = existing.OIDCSub
	}
	if e.Metadata == "" {
		e.Metadata = existing.Metadata
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.extIDs.Update(ctx, tx, e)
	}); err != nil {
		return fmt.Errorf("update external identity: %w", err)
	}

	s.eventBus.Publish(events.ExternalIdentityUpdateEvent, e)
	return nil
}

// DeleteExternalIdentity removes an external identity by ID.
func (s *Service) DeleteExternalIdentity(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: external identity id is required", ErrInvalidInput)
	}
	e, err := s.extIDs.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup external identity: %w", err)
	}
	if e == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.extIDs.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete external identity: %w", err)
	}

	s.eventBus.Publish(events.ExternalIdentityDeleteEvent, e)
	return nil
}
