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
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

// UserRoleDetail is the joined view of one held role: the role row, the
// privileges it carries, and when it was granted to the user.
type UserRoleDetail struct {
	Role       models.Role
	Privileges []models.PrivilegeKey
	GrantedAt  time.Time
}

const userRoleColumns = "user_id, role_id, granted_by, granted_at, reason"

type pgUserRoleStore struct {
	db *sqlx.DB
}

func NewUserRoleStore(db *sqlx.DB) UserRoleStore {
	return &pgUserRoleStore{db: db}
}

func (s *pgUserRoleStore) FindForUpdate(ctx context.Context, tx *sql.Tx, userID, roleID string) (*models.UserRole, error) {
	row := tx.QueryRowContext(ctx,
		`SELECT `+userRoleColumns+` FROM user_roles WHERE user_id = $1 AND role_id = $2 FOR UPDATE`,
		userID, roleID)
	var r models.UserRole
	if err := row.Scan(&r.UserID, &r.RoleID, &r.GrantedBy, &r.GrantedAt, &r.Reason); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *pgUserRoleStore) ListByUser(ctx context.Context, userID string) ([]models.UserRole, error) {
	var rows []models.UserRole
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+userRoleColumns+` FROM user_roles WHERE user_id = $1 ORDER BY granted_at`,
		userID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgUserRoleStore) ListByRole(ctx context.Context, roleID string) ([]models.UserRole, error) {
	var rows []models.UserRole
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+userRoleColumns+` FROM user_roles WHERE role_id = $1 ORDER BY granted_at`,
		roleID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgUserRoleStore) Create(ctx context.Context, tx *sql.Tx, r *models.UserRole) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO user_roles (`+userRoleColumns+`) VALUES ($1, $2, $3, $4, $5)`,
		r.UserID, r.RoleID, r.GrantedBy, r.GrantedAt, r.Reason)
	return err
}

func (s *pgUserRoleStore) Delete(ctx context.Context, tx *sql.Tx, userID, roleID string) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`,
		userID, roleID)
	return err
}

// ListDetailedByUser returns the user's roles with their privileges and
// grant dates in one query.
func (s *pgUserRoleStore) ListDetailedByUser(ctx context.Context, userID string) ([]UserRoleDetail, error) {
	var rows []struct {
		models.Role
		UserGrantedAt time.Time      `db:"user_granted_at"`
		Privilege     sql.NullString `db:"privilege"`
	}
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+Qualify("r", roleColumns)+`,
		        ur.granted_at AS user_granted_at, rp.privilege
		 FROM user_roles ur
		 JOIN roles r ON r.id = ur.role_id
		 LEFT JOIN role_privileges rp ON rp.role_id = r.id
		 WHERE ur.user_id = $1
		 ORDER BY r.name, rp.privilege`, userID)
	if err != nil {
		return nil, err
	}
	out := make([]UserRoleDetail, 0)
	index := make(map[string]int)
	for _, row := range rows {
		i, seen := index[row.Role.ID]
		if !seen {
			out = append(out, UserRoleDetail{
				Role:       row.Role,
				Privileges: []models.PrivilegeKey{},
				GrantedAt:  row.UserGrantedAt,
			})
			i = len(out) - 1
			index[row.Role.ID] = i
		}
		if row.Privilege.Valid {
			out[i].Privileges = append(out[i].Privileges, models.PrivilegeKey(row.Privilege.String))
		}
	}
	return out, nil
}
