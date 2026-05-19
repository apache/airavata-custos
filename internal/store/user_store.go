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

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

const userColumns = `id, organization_id, first_name, last_name, middle_name, email, status`

type mysqlUserStore struct {
	db *sqlx.DB
}

// NewUserStore returns a MySQL-backed UserStore.
func NewUserStore(db *sqlx.DB) UserStore {
	return &mysqlUserStore{db: db}
}

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
		`INSERT INTO users (id, organization_id, first_name, last_name, middle_name, email, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.OrganizationID, u.FirstName, u.LastName, u.MiddleName, u.Email, u.Status)
	return err
}

func (s *mysqlUserStore) Update(ctx context.Context, tx *sql.Tx, u *models.User) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE users SET organization_id = ?, first_name = ?, last_name = ?, middle_name = ?, email = ?
		 WHERE id = ?`,
		u.OrganizationID, u.FirstName, u.LastName, u.MiddleName, u.Email, u.ID)
	return err
}

func (s *mysqlUserStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	return err
}
