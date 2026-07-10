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

type mysqlUserStore struct {
	db *sqlx.DB
}

// NewUserStore returns a MySQL-backed UserStore.
func NewUserStore(db *sqlx.DB) UserStore {
	return &mysqlUserStore{db: db}
}

const userColumns = `id, organization_id, first_name, last_name, middle_name, email, status, type`

func (s *mysqlUserStore) FindByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := s.db.GetContext(ctx, &u,
		`SELECT `+userColumns+` FROM users WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (s *mysqlUserStore) List(ctx context.Context, limit, offset int) ([]models.User, int, error) {
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
		`SELECT `+userColumns+` FROM users ORDER BY email LIMIT ? OFFSET ?`, limit, offset); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (s *mysqlUserStore) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := s.db.GetContext(ctx, &u,
		`SELECT `+userColumns+` FROM users WHERE email = ?`, email)
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
func (s *mysqlUserStore) GetUserByOIDCSub(ctx context.Context, oidcSub string) (*models.User, error) {
	if oidcSub == "" {
		return nil, nil
	}
	var u models.User
	err := s.db.GetContext(ctx, &u,
		`SELECT `+prefixed("u", userColumns)+`
		 FROM users u
		 JOIN user_identities ui ON ui.user_id = u.id
		 WHERE ui.oidc_sub = ?`, oidcSub)
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

func (s *mysqlUserStore) FindByOrganization(ctx context.Context, organizationID string) ([]models.User, error) {
	var users []models.User
	err := s.db.SelectContext(ctx, &users,
		`SELECT `+userColumns+` FROM users WHERE organization_id = ?`, organizationID)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *mysqlUserStore) Create(ctx context.Context, tx *sql.Tx, u *models.User) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.OrganizationID, u.FirstName, u.LastName, u.MiddleName, u.Email, u.Status, u.Type)
	return err
}

func (s *mysqlUserStore) Update(ctx context.Context, tx *sql.Tx, u *models.User) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE users SET organization_id = ?, first_name = ?, last_name = ?, middle_name = ?, email = ?, status = ?, type = ?
		 WHERE id = ?`,
		u.OrganizationID, u.FirstName, u.LastName, u.MiddleName, u.Email, u.Status, u.Type, u.ID)
	return err
}

func (s *mysqlUserStore) UpdateStatus(ctx context.Context, tx *sql.Tx, id string, status models.UserStatus) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE users SET status = ? WHERE id = ?`,
		status, id)
	return err
}

func (s *mysqlUserStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}
