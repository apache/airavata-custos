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

const roleColumns = "id, name, description, is_system, created_at"

type mysqlRoleStore struct {
	db *sqlx.DB
}

func NewRoleStore(db *sqlx.DB) RoleStore {
	return &mysqlRoleStore{db: db}
}

func (s *mysqlRoleStore) FindByID(ctx context.Context, id string) (*models.Role, error) {
	var r models.Role
	err := s.db.GetContext(ctx, &r,
		`SELECT `+roleColumns+` FROM roles WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *mysqlRoleStore) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var r models.Role
	err := s.db.GetContext(ctx, &r,
		`SELECT `+roleColumns+` FROM roles WHERE name = ?`, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *mysqlRoleStore) List(ctx context.Context) ([]models.Role, error) {
	var rows []models.Role
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+roleColumns+` FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlRoleStore) Create(ctx context.Context, tx *sql.Tx, r *models.Role) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO roles (id, name, description, is_system) VALUES (?, ?, ?, ?)`,
		r.ID, r.Name, r.Description, r.IsSystem)
	return err
}

func (s *mysqlRoleStore) Update(ctx context.Context, tx *sql.Tx, r *models.Role) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE roles SET name = ?, description = ? WHERE id = ?`,
		r.Name, r.Description, r.ID)
	return err
}

func (s *mysqlRoleStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM roles WHERE id = ?`, id)
	return err
}

func (s *mysqlRoleStore) ListPrivileges(ctx context.Context, roleID string) ([]models.PrivilegeKey, error) {
	var keys []models.PrivilegeKey
	err := s.db.SelectContext(ctx, &keys,
		`SELECT privilege FROM role_privileges WHERE role_id = ? ORDER BY privilege`, roleID)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *mysqlRoleStore) AddPrivilege(ctx context.Context, tx *sql.Tx, roleID string, privilege models.PrivilegeKey) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO role_privileges (role_id, privilege) VALUES (?, ?)`,
		roleID, privilege)
	return err
}

func (s *mysqlRoleStore) RemovePrivilege(ctx context.Context, tx *sql.Tx, roleID string, privilege models.PrivilegeKey) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM role_privileges WHERE role_id = ? AND privilege = ?`,
		roleID, privilege)
	return err
}

func (s *mysqlRoleStore) HasPrivilege(ctx context.Context, tx *sql.Tx, roleID string, privilege models.PrivilegeKey) (bool, error) {
	var n int
	err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM role_privileges WHERE role_id = ? AND privilege = ?`,
		roleID, privilege).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// CountRolesGrantingPrivilege returns the number of roles carrying the
// given key. Used by the last-meta-holder guard.
func (s *mysqlRoleStore) CountRolesGrantingPrivilege(ctx context.Context, tx *sql.Tx, privilege models.PrivilegeKey) (int, error) {
	var n int
	err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM role_privileges WHERE privilege = ?`,
		privilege).Scan(&n)
	if err != nil {
		return 0, err
	}
	return n, nil
}
