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

// AddUserDN binds a DN to a user. If d.ID is empty, a new UUID is generated.
// The referenced user must already exist; (user_id, dn) is unique.
func (s *Service) AddUserDN(ctx context.Context, d *models.UserDN) (*models.UserDN, error) {
	if d == nil {
		return nil, fmt.Errorf("%w: user dn is nil", ErrInvalidInput)
	}
	if d.UserID == "" {
		return nil, fmt.Errorf("%w: user dn user_id is required", ErrInvalidInput)
	}
	if d.DN == "" {
		return nil, fmt.Errorf("%w: user dn dn is required", ErrInvalidInput)
	}

	if user, err := s.users.FindByID(ctx, d.UserID); err != nil {
		return nil, fmt.Errorf("verify user: %w", err)
	} else if user == nil {
		return nil, fmt.Errorf("%w: user %q does not exist", ErrInvalidInput, d.UserID)
	}

	if existing, err := s.userDNs.FindByDN(ctx, d.DN); err != nil {
		return nil, fmt.Errorf("lookup user dn: %w", err)
	} else if existing != nil {
		return nil, fmt.Errorf("%w: dn %q", ErrAlreadyExists, d.DN)
	}

	if d.ID == "" {
		d.ID = newID()
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.userDNs.Create(ctx, tx, d)
	}); err != nil {
		return nil, fmt.Errorf("add user dn: %w", err)
	}

	s.eventBus.Publish(events.UserDNCreateEvent, d)
	return d, nil
}

// GetUserDN retrieves a DN binding by ID. Returns ErrNotFound when no row matches.
func (s *Service) GetUserDN(ctx context.Context, id string) (*models.UserDN, error) {
	d, err := s.userDNs.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user dn: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}
	return d, nil
}

// GetUserDNByDN performs a reverse lookup from DN to binding.
func (s *Service) GetUserDNByDN(ctx context.Context, dn string) (*models.UserDN, error) {
	if dn == "" {
		return nil, fmt.Errorf("%w: dn is required", ErrInvalidInput)
	}
	d, err := s.userDNs.FindByDN(ctx, dn)
	if err != nil {
		return nil, fmt.Errorf("get user dn by dn: %w", err)
	}
	if d == nil {
		return nil, ErrNotFound
	}
	return d, nil
}

// ListUserDNs returns every DN bound to the given user.
func (s *Service) ListUserDNs(ctx context.Context, userID string) ([]models.UserDN, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	out, err := s.userDNs.FindByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user dns: %w", err)
	}
	return out, nil
}

// RemoveUserDN removes a DN binding by ID.
func (s *Service) RemoveUserDN(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: user dn id is required", ErrInvalidInput)
	}
	d, err := s.userDNs.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup user dn: %w", err)
	}
	if d == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.userDNs.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("remove user dn: %w", err)
	}

	s.eventBus.Publish(events.UserDNDeleteEvent, d)
	return nil
}
