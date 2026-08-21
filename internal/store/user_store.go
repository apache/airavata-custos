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

package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

type pgUserStore struct {
	db *sqlx.DB
}

// NewUserStore returns a PostgreSQL-backed UserStore.
func NewUserStore(db *sqlx.DB) UserStore {
	return &pgUserStore{db: db}
}

const userColumns = `id, organization_id, first_name, last_name, middle_name, email, status, type`

func (s *pgUserStore) FindByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := s.db.GetContext(ctx, &u,
		`SELECT `+userColumns+` FROM users WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (s *pgUserStore) List(ctx context.Context, limit, offset int) ([]models.User, int, error) {
	var total int
	if err := s.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM users`); err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	var rows []models.User
	if err := s.db.SelectContext(ctx, &rows,
		`SELECT `+userColumns+` FROM users ORDER BY email LIMIT $1 OFFSET $2`, limit, offset); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (s *pgUserStore) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := s.db.GetContext(ctx, &u,
		`SELECT `+userColumns+` FROM users WHERE email = $1`, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// GetUserByOIDCSub returns the user owning the user_identities row whose
// oidc_sub matches. Returns nil when the OIDC subject is empty or no row
// links it to a Custos user.
func (s *pgUserStore) GetUserByOIDCSub(ctx context.Context, oidcSub string) (*models.User, error) {
	if oidcSub == "" {
		return nil, nil
	}
	var u models.User
	err := s.db.GetContext(ctx, &u,
		`SELECT `+prefixed("u", userColumns)+`
		 FROM users u
		 JOIN user_identities ui ON ui.user_id = u.id
		 WHERE ui.oidc_sub = $1`, oidcSub)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// prefixed returns the comma-separated column list with each bare column
// prefixed by alias. Used to disambiguate joined queries.
func prefixed(alias, columns string) string {
	parts := strings.Split(columns, ", ")
	for i, p := range parts {
		parts[i] = alias + "." + p
	}
	return strings.Join(parts, ", ")
}

func (s *pgUserStore) FindByOrganization(ctx context.Context, organizationID string) ([]models.User, error) {
	var users []models.User
	err := s.db.SelectContext(ctx, &users,
		`SELECT `+userColumns+` FROM users WHERE organization_id = $1`, organizationID)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *pgUserStore) Create(ctx context.Context, tx *sql.Tx, u *models.User) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		u.ID, u.OrganizationID, u.FirstName, u.LastName, u.MiddleName, u.Email, u.Status, u.Type)
	return err
}

func (s *pgUserStore) Update(ctx context.Context, tx *sql.Tx, u *models.User) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE users SET organization_id = $1, first_name = $2, last_name = $3, middle_name = $4, email = $5, status = $6, type = $7
		 WHERE id = $8`,
		u.OrganizationID, u.FirstName, u.LastName, u.MiddleName, u.Email, u.Status, u.Type, u.ID)
	return err
}

func (s *pgUserStore) UpdateStatus(ctx context.Context, tx *sql.Tx, id string, status models.UserStatus) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE users SET status = $1 WHERE id = $2`,
		status, id)
	return err
}

func (s *pgUserStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}
