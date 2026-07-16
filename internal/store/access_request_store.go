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

// COALESCE folds the nullable columns into Go zero values on read; writes use
// nullIfEmpty so empty strings never trip the users FKs.
const accessRequestColumns = `id, oidc_sub, email, name, institution, event_code,
	COALESCE(reason, '') AS reason, status,
	COALESCE(approver_id, '') AS approver_id,
	COALESCE(deny_reason, '') AS deny_reason,
	expires_at,
	COALESCE(created_user_id, '') AS created_user_id,
	timestamp`

type mysqlAccessRequestStore struct {
	db *sqlx.DB
}

// NewAccessRequestStore returns a MySQL-backed AccessRequestStore.
func NewAccessRequestStore(db *sqlx.DB) AccessRequestStore {
	return &mysqlAccessRequestStore{db: db}
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (s *mysqlAccessRequestStore) FindByID(ctx context.Context, id string) (*models.AccessRequest, error) {
	var r models.AccessRequest
	err := s.db.GetContext(ctx, &r,
		`SELECT `+accessRequestColumns+` FROM access_requests WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *mysqlAccessRequestStore) FindLatestBySub(ctx context.Context, oidcSub string) (*models.AccessRequest, error) {
	var r models.AccessRequest
	err := s.db.GetContext(ctx, &r,
		`SELECT `+accessRequestColumns+`
		 FROM access_requests
		 WHERE oidc_sub = ?
		 ORDER BY timestamp DESC
		 LIMIT 1`, oidcSub)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *mysqlAccessRequestStore) HasPendingBySub(ctx context.Context, oidcSub string) (bool, error) {
	var n int
	err := s.db.GetContext(ctx, &n,
		`SELECT COUNT(*) FROM access_requests WHERE oidc_sub = ? AND status = ?`,
		oidcSub, string(models.AccessRequestPending))
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *mysqlAccessRequestStore) List(ctx context.Context, f AccessRequestListFilter) ([]models.AccessRequest, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	query := `SELECT ` + accessRequestColumns + ` FROM access_requests WHERE 1=1`
	args := []any{}
	if f.Status != "" {
		query += ` AND status = ?`
		args = append(args, f.Status)
	}
	if f.EventCode != "" {
		query += ` AND event_code = ?`
		args = append(args, f.EventCode)
	}
	query += ` ORDER BY timestamp DESC LIMIT ?`
	args = append(args, limit)
	var rows []models.AccessRequest
	err := s.db.SelectContext(ctx, &rows, query, args...)
	return rows, err
}

func (s *mysqlAccessRequestStore) Create(ctx context.Context, tx *sql.Tx, r *models.AccessRequest) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO access_requests
		     (id, oidc_sub, email, name, institution, event_code, reason, status, approver_id, deny_reason, expires_at, created_user_id, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.OIDCSub, r.Email, r.Name, r.Institution, r.EventCode,
		nullIfEmpty(r.Reason), string(r.Status), nullIfEmpty(r.ApproverID),
		nullIfEmpty(r.DenyReason), r.ExpiresAt, nullIfEmpty(r.CreatedUserID), r.Timestamp)
	return err
}

func (s *mysqlAccessRequestStore) Update(ctx context.Context, tx *sql.Tx, r *models.AccessRequest) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE access_requests
		    SET status          = ?,
		        approver_id     = ?,
		        deny_reason     = ?,
		        expires_at      = ?,
		        created_user_id = ?
		  WHERE id = ?`,
		string(r.Status), nullIfEmpty(r.ApproverID), nullIfEmpty(r.DenyReason),
		r.ExpiresAt, nullIfEmpty(r.CreatedUserID), r.ID)
	return err
}
