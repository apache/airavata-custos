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

// CreateOrganization persists a new organization. If org.ID is empty a new
// UUID is generated. The (possibly populated) organization is returned.
func (s *Service) CreateOrganization(ctx context.Context, org *models.Organization) (*models.Organization, error) {
	if org == nil {
		return nil, fmt.Errorf("%w: organization is nil", ErrInvalidInput)
	}
	if org.Name == "" {
		return nil, fmt.Errorf("%w: organization name is required", ErrInvalidInput)
	}
	if org.ID == "" {
		org.ID = newID()
	}

	if org.OriginatedID != "" {
		if existing, err := s.orgs.FindByOriginatedID(ctx, org.OriginatedID); err != nil {
			return nil, fmt.Errorf("lookup organization by originated id: %w", err)
		} else if existing != nil {
			return nil, fmt.Errorf("%w: organization with originated_id %q", ErrAlreadyExists, org.OriginatedID)
		}
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.orgs.Create(ctx, tx, org)
	}); err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}

	s.eventBus.Publish(ctx, events.OrganizationCreateEvent, org)
	return org, nil
}

// GetOrganization retrieves an organization by its ID. Returns ErrNotFound
// when no organization matches.
func (s *Service) GetOrganization(ctx context.Context, id string) (*models.Organization, error) {
	org, err := s.orgs.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get organization: %w", err)
	}
	if org == nil {
		return nil, ErrNotFound
	}
	return org, nil
}

// GetOrganizationByOriginatedID retrieves an organization by its external originated ID.
func (s *Service) GetOrganizationByOriginatedID(ctx context.Context, originatedID string) (*models.Organization, error) {
	org, err := s.orgs.FindByOriginatedID(ctx, originatedID)
	if err != nil {
		return nil, fmt.Errorf("get organization by originated id: %w", err)
	}
	if org == nil {
		return nil, ErrNotFound
	}
	return org, nil
}

// UpdateOrganization persists changes to an existing organization.
func (s *Service) UpdateOrganization(ctx context.Context, org *models.Organization) error {
	if org == nil || org.ID == "" {
		return fmt.Errorf("%w: organization id is required", ErrInvalidInput)
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.orgs.Update(ctx, tx, org)
	}); err != nil {
		return fmt.Errorf("update organization: %w", err)
	}

	s.eventBus.Publish(ctx, events.OrganizationUpdateEvent, org)
	return nil
}

// DeleteOrganization removes an organization by ID.
func (s *Service) DeleteOrganization(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: organization id is required", ErrInvalidInput)
	}
	org, err := s.orgs.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup organization: %w", err)
	}
	if org == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.orgs.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete organization: %w", err)
	}

	s.eventBus.Publish(ctx, events.OrganizationDeleteEvent, org)
	return nil
}
