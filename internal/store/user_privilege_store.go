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

const userPrivilegeColumns = "id, user_id, privilege, granted_by, granted_at, reason"

type mysqlUserPrivilegeStore struct {
	db *sqlx.DB
}

// NewUserPrivilegeStore returns a MySQL-backed UserPrivilegeStore.
func NewUserPrivilegeStore(db *sqlx.DB) UserPrivilegeStore {
	return &mysqlUserPrivilegeStore{db: db}
}

func (s *mysqlUserPrivilegeStore) Find(ctx context.Context, userID string, privilege models.PrivilegeKey) (*models.UserPrivilege, error) {
	var r models.UserPrivilege
	err := s.db.GetContext(ctx, &r,
		`SELECT `+userPrivilegeColumns+`
		 FROM user_privileges
		 WHERE user_id = ? AND privilege = ?`,
		userID, privilege)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *mysqlUserPrivilegeStore) FindForUpdate(ctx context.Context, tx *sql.Tx, userID string, privilege models.PrivilegeKey) (*models.UserPrivilege, error) {
	row := tx.QueryRowContext(ctx,
		`SELECT `+userPrivilegeColumns+`
		 FROM user_privileges
		 WHERE user_id = ? AND privilege = ?
		 FOR UPDATE`,
		userID, privilege)
	var r models.UserPrivilege
	if err := row.Scan(
		&r.ID, &r.UserID, &r.Privilege,
		&r.GrantedBy, &r.GrantedAt, &r.Reason,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *mysqlUserPrivilegeStore) ListByUser(ctx context.Context, userID string) ([]models.UserPrivilege, error) {
	var rows []models.UserPrivilege
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+userPrivilegeColumns+`
		 FROM user_privileges
		 WHERE user_id = ?
		 ORDER BY granted_at`, userID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlUserPrivilegeStore) ListByPrivilege(ctx context.Context, privilege models.PrivilegeKey) ([]models.UserPrivilege, error) {
	var rows []models.UserPrivilege
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+userPrivilegeColumns+`
		 FROM user_privileges
		 WHERE privilege = ?
		 ORDER BY granted_at`, privilege)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlUserPrivilegeStore) CountByPrivilege(ctx context.Context, tx *sql.Tx, privilege models.PrivilegeKey) (int, error) {
	var n int
	err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_privileges WHERE privilege = ?`, privilege).Scan(&n)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (s *mysqlUserPrivilegeStore) Create(ctx context.Context, tx *sql.Tx, r *models.UserPrivilege) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO user_privileges
		  (id, user_id, privilege, granted_by, granted_at, reason)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		r.ID, r.UserID, r.Privilege, r.GrantedBy, r.GrantedAt, r.Reason)
	return err
}

func (s *mysqlUserPrivilegeStore) Delete(ctx context.Context, tx *sql.Tx, userID string, privilege models.PrivilegeKey) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM user_privileges WHERE user_id = ? AND privilege = ?`,
		userID, privilege)
	return err
}
