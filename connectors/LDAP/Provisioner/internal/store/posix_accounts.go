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

// Package store keeps the POSIX attributes the connector assigns, one row
// per compute_cluster_users row.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

const (
	StatusActive   = "ACTIVE"
	StatusInactive = "INACTIVE"
)

// ErrUsernameTaken means another row already holds the username.
var ErrUsernameTaken = errors.New("username is held by another ldap_posix_accounts row")

// numberLock serializes number assignment across Custos instances, since
// each assignment reads MAX(uid_number).
const numberLock = `SELECT pg_advisory_xact_lock(hashtext('ldap_posix_accounts.uid_number'))`

type Account struct {
	ComputeClusterUserID string    `db:"compute_cluster_user_id"`
	UserID               string    `db:"user_id"`
	UIDNumber            int64     `db:"uid_number"`
	GIDNumber            int64     `db:"gid_number"`
	Username             string    `db:"username"`
	HomeDirectory        string    `db:"home_directory"`
	LoginShell           string    `db:"login_shell"`
	EntryDN              string    `db:"entry_dn"`
	Status               string    `db:"status"`
	CreatedAt            time.Time `db:"created_at"`
	UpdatedAt            time.Time `db:"updated_at"`
}

type Accounts struct {
	db *sqlx.DB
}

func NewAccounts(db *sqlx.DB) *Accounts {
	return &Accounts{db: db}
}

// Get returns the row for a cluster account, or nil when there is none.
func (s *Accounts) Get(ctx context.Context, computeClusterUserID string) (*Account, error) {
	return s.getBy(ctx, "compute_cluster_user_id", computeClusterUserID)
}

// GetByUsername returns the row holding username, or nil when there is none.
func (s *Accounts) GetByUsername(ctx context.Context, username string) (*Account, error) {
	return s.getBy(ctx, "username", username)
}

func (s *Accounts) getBy(ctx context.Context, column, value string) (*Account, error) {
	var a Account
	err := s.db.GetContext(ctx, &a, `SELECT * FROM ldap_posix_accounts WHERE `+column+` = $1`, value)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get ldap_posix_accounts by %s: %w", column, err)
	}
	return &a, nil
}

// Create inserts a row numbered one above every assigned uid_number, and at
// least floor. Inactive rows count too, so a number is never reused.
func (s *Accounts) Create(ctx context.Context, a Account, floor int64) (*Account, error) {
	var created Account
	err := s.withNumberLock(ctx, func(tx *sqlx.Tx) error {
		return tx.GetContext(ctx, &created, `INSERT INTO ldap_posix_accounts
			(compute_cluster_user_id, user_id, uid_number, gid_number, username, home_directory, login_shell, entry_dn)
			SELECT $1, $2, n, n, $3, $4, $5, $6
			FROM (SELECT GREATEST(COALESCE(MAX(uid_number) + 1, $7), $7) AS n FROM ldap_posix_accounts) next
			RETURNING *`,
			a.ComputeClusterUserID, a.UserID, a.Username, a.HomeDirectory, a.LoginShell, a.EntryDN, floor)
	})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "uq_ldap_posix_accounts_username" {
		return nil, ErrUsernameTaken
	}
	if err != nil {
		return nil, fmt.Errorf("insert ldap_posix_accounts: %w", err)
	}
	return &created, nil
}

// Renumber moves the row to a fresh uid_number and gid_number, for when the
// directory already holds the one it had.
func (s *Accounts) Renumber(ctx context.Context, computeClusterUserID string, floor int64) (*Account, error) {
	var updated Account
	err := s.withNumberLock(ctx, func(tx *sqlx.Tx) error {
		return tx.GetContext(ctx, &updated, `UPDATE ldap_posix_accounts
			SET uid_number = next.n, gid_number = next.n, updated_at = CURRENT_TIMESTAMP(6)
			FROM (SELECT GREATEST(MAX(uid_number) + 1, $2) AS n FROM ldap_posix_accounts) next
			WHERE compute_cluster_user_id = $1
			RETURNING ldap_posix_accounts.*`,
			computeClusterUserID, floor)
	})
	if err != nil {
		return nil, fmt.Errorf("renumber ldap_posix_accounts: %w", err)
	}
	return &updated, nil
}

// Reclaim moves an inactive row, numbers included, to a new cluster account
// and reactivates it. It returns nil when the row is no longer inactive.
func (s *Accounts) Reclaim(ctx context.Context, fromID, toID string) (*Account, error) {
	var a Account
	err := s.db.GetContext(ctx, &a, `UPDATE ldap_posix_accounts
		SET compute_cluster_user_id = $2, status = 'ACTIVE', updated_at = CURRENT_TIMESTAMP(6)
		WHERE compute_cluster_user_id = $1 AND status = 'INACTIVE'
		RETURNING *`,
		fromID, toID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reclaim ldap_posix_accounts: %w", err)
	}
	return &a, nil
}

func (s *Accounts) SetStatus(ctx context.Context, computeClusterUserID, status string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE ldap_posix_accounts
		SET status = $2, updated_at = CURRENT_TIMESTAMP(6)
		WHERE compute_cluster_user_id = $1`,
		computeClusterUserID, status)
	if err != nil {
		return fmt.Errorf("set ldap_posix_accounts status: %w", err)
	}
	return nil
}

func (s *Accounts) withNumberLock(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, numberLock); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
