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
// KIND, either express or implied. See the License for the
// specific language governing permissions and limitations
// under the License.

// Package store holds the LDAP Provisioner connector's persistent
// state accessors. Currently just the UID sequence — every allocation
// against LDAP goes through Allocate() so the number never regresses,
// even after LDAP entries are deleted or the process restarts.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// UIDSequence is the persistent monotonic UID allocator. One row per
// cluster; each row's next_uid is the value that will be handed out on
// the NEXT Allocate call. The row is created lazily by Seed and
// incremented atomically by Allocate under InnoDB row locking.
type UIDSequence struct {
	db *sqlx.DB
}

func NewUIDSequence(db *sqlx.DB) *UIDSequence {
	return &UIDSequence{db: db}
}

// Seed initialises the counter row for clusterID to the greater of
// initialNextUID and any existing next_uid on that row. Idempotent:
// safe to call at every startup. Never lowers an existing counter, so
// re-seeding with a stale scan cannot regress the sequence.
//
// The caller passes initialNextUID as max(LDAP scan, minUID) + 1 so a
// fresh install picks up above any entries that were provisioned
// out-of-band before this connector ran.
func (s *UIDSequence) Seed(ctx context.Context, clusterID string, initialNextUID int64) error {
	if clusterID == "" {
		return errors.New("UIDSequence.Seed: clusterID is required")
	}
	if initialNextUID <= 0 {
		return errors.New("UIDSequence.Seed: initialNextUID must be positive")
	}
	// INSERT with ON DUPLICATE KEY UPDATE performs an atomic
	// upsert-take-max: existing rows keep their value if it's already
	// higher, so re-running startup with a lower LDAP-scan hint is
	// harmless.
	const q = `
		INSERT INTO ldap_uid_sequence (cluster_id, next_uid)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE next_uid = GREATEST(next_uid, VALUES(next_uid))
	`
	if _, err := s.db.ExecContext(ctx, q, clusterID, initialNextUID); err != nil {
		return fmt.Errorf("seed ldap_uid_sequence for %s: %w", clusterID, err)
	}
	return nil
}

// Allocate atomically increments the counter for clusterID and returns
// the value that was current BEFORE the increment — i.e. the uid to
// assign to the new account. Uses the MySQL LAST_INSERT_ID(expr) trick
// so read and write happen in one statement with InnoDB row locking
// serialising concurrent callers.
//
// Errors if Seed has not been called for this cluster.
func (s *UIDSequence) Allocate(ctx context.Context, clusterID string) (int64, error) {
	if clusterID == "" {
		return 0, errors.New("UIDSequence.Allocate: clusterID is required")
	}

	// Run in a transaction so LAST_INSERT_ID reads the value this same
	// connection just wrote. Without an explicit tx some connection
	// pools could hand back a different session for the SELECT.
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin allocate tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// The atomic increment: LAST_INSERT_ID(expr) both writes expr and
	// stashes it in the session's LAST_INSERT_ID for retrieval.
	res, err := tx.ExecContext(ctx,
		`UPDATE ldap_uid_sequence
		    SET next_uid = LAST_INSERT_ID(next_uid + 1)
		  WHERE cluster_id = ?`,
		clusterID,
	)
	if err != nil {
		return 0, fmt.Errorf("increment ldap_uid_sequence: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return 0, fmt.Errorf("ldap_uid_sequence row missing for cluster %s (was Seed called?)", clusterID)
	}

	var newNext int64
	if err := tx.QueryRowxContext(ctx, `SELECT LAST_INSERT_ID()`).Scan(&newNext); err != nil {
		return 0, fmt.Errorf("read LAST_INSERT_ID: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit allocate: %w", err)
	}
	// LAST_INSERT_ID(next_uid+1) sets it to next_uid+1 which is the
	// new "next_uid" value. The value we allocate is what next_uid
	// USED to be — i.e. newNext - 1.
	return newNext - 1, nil
}

// Peek returns the current next_uid for clusterID without incrementing.
// Used by tests and diagnostics; not part of the allocation path.
func (s *UIDSequence) Peek(ctx context.Context, clusterID string) (int64, error) {
	var next int64
	err := s.db.QueryRowxContext(ctx,
		`SELECT next_uid FROM ldap_uid_sequence WHERE cluster_id = ?`,
		clusterID,
	).Scan(&next)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("peek ldap_uid_sequence: %w", err)
	}
	return next, nil
}
