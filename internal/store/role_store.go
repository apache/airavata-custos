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

type pgRoleStore struct {
	db *sqlx.DB
}

func NewRoleStore(db *sqlx.DB) RoleStore {
	return &pgRoleStore{db: db}
}

func (s *pgRoleStore) FindByID(ctx context.Context, id string) (*models.Role, error) {
	var r models.Role
	err := s.db.GetContext(ctx, &r,
		`SELECT `+roleColumns+` FROM roles WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *pgRoleStore) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var r models.Role
	err := s.db.GetContext(ctx, &r,
		`SELECT `+roleColumns+` FROM roles WHERE name = $1`, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *pgRoleStore) List(ctx context.Context) ([]models.Role, error) {
	var rows []models.Role
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+roleColumns+` FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgRoleStore) Create(ctx context.Context, tx *sql.Tx, r *models.Role) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO roles (id, name, description, is_system) VALUES ($1, $2, $3, $4)`,
		r.ID, r.Name, r.Description, r.IsSystem)
	return err
}

func (s *pgRoleStore) Update(ctx context.Context, tx *sql.Tx, r *models.Role) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE roles SET name = $1, description = $2 WHERE id = $3`,
		r.Name, r.Description, r.ID)
	return err
}

func (s *pgRoleStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM roles WHERE id = $1`, id)
	return err
}

func (s *pgRoleStore) ListPrivileges(ctx context.Context, roleID string) ([]models.PrivilegeKey, error) {
	var keys []models.PrivilegeKey
	err := s.db.SelectContext(ctx, &keys,
		`SELECT privilege FROM role_privileges WHERE role_id = $1 ORDER BY privilege`, roleID)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *pgRoleStore) AddPrivilege(ctx context.Context, tx *sql.Tx, roleID string, privilege models.PrivilegeKey) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO role_privileges (role_id, privilege) VALUES ($1, $2)`,
		roleID, privilege)
	return err
}

func (s *pgRoleStore) RemovePrivilege(ctx context.Context, tx *sql.Tx, roleID string, privilege models.PrivilegeKey) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM role_privileges WHERE role_id = $1 AND privilege = $2`,
		roleID, privilege)
	return err
}

func (s *pgRoleStore) HasPrivilege(ctx context.Context, tx *sql.Tx, roleID string, privilege models.PrivilegeKey) (bool, error) {
	var n int
	err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM role_privileges WHERE role_id = $1 AND privilege = $2`,
		roleID, privilege).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// CountRolesGrantingPrivilege returns the number of roles carrying the
// given key. Used by the last-meta-holder guard.
func (s *pgRoleStore) CountRolesGrantingPrivilege(ctx context.Context, tx *sql.Tx, privilege models.PrivilegeKey) (int, error) {
	var n int
	err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM role_privileges WHERE privilege = $1`,
		privilege).Scan(&n)
	if err != nil {
		return 0, err
	}
	return n, nil
}
