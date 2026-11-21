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
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

// StorageScope narrows storages to a cluster, an account or an allocation. An empty field matches any.
type StorageScope struct {
	ClusterID, UserID, AllocationID string
}

// StorageHolder is who a storage is for: an account, an allocation, both for a membership, or a connector's user.
type StorageHolder struct {
	ClusterID            string  `db:"compute_cluster_id"`
	ComputeClusterUserID *string `db:"compute_cluster_user_id"`
	AllocationID         *string `db:"allocation_id"`
	UserID               *string `db:"user_id"`
	Username             string  `db:"username"`
	PosixGroup           string  `db:"posix_group"`
	Name                 string  `db:"-"`
}

const (
	allocationHolderColumns = `a.compute_cluster_id, a.id AS allocation_id, a.posix_group`
	// allocationGivesStorage holds for an active allocation a with a provisioned group.
	allocationGivesStorage = `a.status = 'ACTIVE' AND a.posix_group IS NOT NULL AND a.group_provisioned_at IS NOT NULL`
	// userHasAccess holds for a user u whose provisioned accounts reach the cluster.
	userHasAccess = `u.status = 'ACTIVE'`
)

// storageInScope matches storage s to the scope's cluster $1, user $2 and allocation $3.
const storageInScope = `($1 = '' OR s.compute_cluster_id = $1) AND ($3 = '' OR s.allocation_id = $3)
	AND ($2 = '' OR s.user_id = $2 OR s.compute_cluster_user_id IN (SELECT id FROM compute_cluster_users WHERE user_id = $2))`

type pgStorageStore struct {
	db *sqlx.DB
}

// NewStorageStore returns a PostgreSQL-backed StorageStore.
func NewStorageStore(db *sqlx.DB) StorageStore {
	return &pgStorageStore{db: db}
}

func (s *pgStorageStore) List(ctx context.Context) ([]models.Storage, error) {
	storages := []models.Storage{}
	err := s.db.SelectContext(ctx, &storages, `SELECT s.*, cu.local_username AS username, a.name AS allocation_name, a.start_time, a.end_time
		FROM storages s
		LEFT JOIN compute_cluster_users cu ON cu.id = s.compute_cluster_user_id
		LEFT JOIN compute_allocations a ON a.id = s.allocation_id
		ORDER BY s.status <> 'pending_deletion', s.path`)
	return storages, err
}

func (s *pgStorageStore) Lock(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('storages'))`)
	return err
}

func (s *pgStorageStore) AccountHolders(ctx context.Context, tx *sql.Tx, scope StorageScope) (holders []StorageHolder, err error) {
	err = (&sqlx.Tx{Tx: tx, Mapper: s.db.Mapper}).SelectContext(ctx, &holders, `SELECT compute_cluster_id, id AS compute_cluster_user_id, local_username AS username FROM compute_cluster_users
		WHERE provisioned_at IS NOT NULL AND ($1 = '' OR compute_cluster_id = $1) AND ($2 = '' OR user_id = $2)
		AND EXISTS (SELECT 1 FROM users u WHERE u.id = compute_cluster_users.user_id AND `+userHasAccess+`)`, scope.ClusterID, scope.UserID)
	return
}

func (s *pgStorageStore) AllocationHolders(ctx context.Context, tx *sql.Tx, scope StorageScope) (holders []StorageHolder, err error) {
	err = (&sqlx.Tx{Tx: tx, Mapper: s.db.Mapper}).SelectContext(ctx, &holders, `SELECT `+allocationHolderColumns+` FROM compute_allocations a
		WHERE `+allocationGivesStorage+` AND ($1 = '' OR a.compute_cluster_id = $1) AND ($2 = '' OR a.id = $2)`, scope.ClusterID, scope.AllocationID)
	return
}

func (s *pgStorageStore) MemberHolders(ctx context.Context, tx *sql.Tx, scope StorageScope) (holders []StorageHolder, err error) {
	err = (&sqlx.Tx{Tx: tx, Mapper: s.db.Mapper}).SelectContext(ctx, &holders, `SELECT cu.id AS compute_cluster_user_id, cu.local_username AS username, `+allocationHolderColumns+`
		FROM compute_allocation_memberships m
		JOIN compute_allocations a ON a.id = m.compute_allocation_id AND `+allocationGivesStorage+`
		JOIN compute_cluster_users cu ON cu.compute_cluster_id = a.compute_cluster_id AND cu.user_id = m.user_id AND cu.provisioned_at IS NOT NULL
		JOIN users u ON u.id = m.user_id AND `+userHasAccess+`
		WHERE m.membership_status = 'ACTIVE' AND ($1 = '' OR a.compute_cluster_id = $1) AND ($2 = '' OR m.user_id = $2) AND ($3 = '' OR a.id = $3)`,
		scope.ClusterID, scope.UserID, scope.AllocationID)
	return
}

func (s *pgStorageStore) Users(ctx context.Context, tx *sql.Tx, userID string) (users []models.User, err error) {
	err = (&sqlx.Tx{Tx: tx, Mapper: s.db.Mapper}).SelectContext(ctx, &users, `SELECT `+userColumns+` FROM users WHERE $1 = '' OR id = $1 ORDER BY created_at, id`, userID)
	return
}

func (s *pgStorageStore) Retire(ctx context.Context, tx *sql.Tx, scope StorageScope, claimedClusterIDs []string, keptTypes []models.StorageType, desired []models.Storage) error {
	var desiredClusterIDs, desiredPaths []string
	for _, st := range desired {
		desiredClusterIDs, desiredPaths = append(desiredClusterIDs, st.ComputeClusterID), append(desiredPaths, st.Path)
	}
	_, err := tx.ExecContext(ctx, `UPDATE storages s SET status = 'pending_deletion'
		WHERE s.status = 'active' AND s.compute_cluster_id = ANY($4) AND `+storageInScope+` AND s.type <> ALL(COALESCE($7::text[], '{}'))
			AND (s.compute_cluster_id, s.path) NOT IN (SELECT * FROM unnest($5::text[], $6::text[]))`,
		scope.ClusterID, scope.UserID, scope.AllocationID, claimedClusterIDs, desiredClusterIDs, desiredPaths, keptTypes)
	return err
}

func (s *pgStorageStore) Upsert(ctx context.Context, tx *sql.Tx, desired []models.Storage) (changed []models.Storage, err error) {
	if len(desired) == 0 {
		return nil, nil
	}
	rows, err := json.Marshal(desired)
	if err != nil {
		return nil, err
	}
	// A storage another user holds stays theirs.
	err = (&sqlx.Tx{Tx: tx, Mapper: s.db.Mapper}).SelectContext(ctx, &changed, `INSERT INTO storages AS s (id, compute_cluster_id,
		compute_cluster_user_id, allocation_id, user_id, type, path, owner, "group", mode, quota_bytes, quota_inodes)
	SELECT id, compute_cluster_id, compute_cluster_user_id, allocation_id, user_id, type, path, owner, "group", mode, quota_bytes,
		quota_inodes FROM json_populate_recordset(NULL::storages, $1::json)
	ON CONFLICT (compute_cluster_id, path) DO UPDATE SET compute_cluster_user_id = EXCLUDED.compute_cluster_user_id,
		allocation_id = EXCLUDED.allocation_id, user_id = EXCLUDED.user_id, type = EXCLUDED.type, owner = EXCLUDED.owner, "group" = EXCLUDED."group", mode = EXCLUDED.mode,
		quota_bytes = CASE WHEN s.quota_until > now() THEN s.quota_bytes ELSE EXCLUDED.quota_bytes END,
		quota_until = CASE WHEN s.quota_until > now() THEN s.quota_until END,
		quota_inodes = EXCLUDED.quota_inodes, status = 'active'
	WHERE (s.user_id IS NULL OR s.user_id = EXCLUDED.user_id) AND (s.status <> 'active' OR s.quota_until <= now()
		OR s.quota_until IS NULL AND s.quota_bytes <> EXCLUDED.quota_bytes
		OR (s.compute_cluster_user_id, s.allocation_id, s.user_id, s.type, s.owner, s."group", s.mode, s.quota_inodes)
			IS DISTINCT FROM (EXCLUDED.compute_cluster_user_id, EXCLUDED.allocation_id, EXCLUDED.user_id, EXCLUDED.type, EXCLUDED.owner,
				EXCLUDED."group", EXCLUDED.mode, EXCLUDED.quota_inodes))
	RETURNING *`, string(rows))
	slices.SortFunc(changed, func(a, b models.Storage) int { return strings.Compare(a.Path, b.Path) }) // a parent first
	return changed, err
}

func (s *pgStorageStore) SetQuota(ctx context.Context, tx *sql.Tx, id string, bytes int64, until time.Time) (*models.Storage, int64, error) {
	var row struct {
		models.Storage
		PreviousQuotaBytes int64 `db:"previous_quota_bytes"`
	}
	err := (&sqlx.Tx{Tx: tx, Mapper: s.db.Mapper}).GetContext(ctx, &row, `UPDATE storages s SET quota_bytes = $2, quota_until = $3
		FROM (SELECT id, quota_bytes FROM storages WHERE id = $1 FOR UPDATE) old
		WHERE s.id = old.id AND old.quota_bytes <> 0 RETURNING s.*, old.quota_bytes AS previous_quota_bytes`, id, bytes, until)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, nil
	}
	return &row.Storage, row.PreviousQuotaBytes, err
}

func (s *pgStorageStore) Delete(ctx context.Context, tx *sql.Tx, id string) (*models.Storage, error) {
	var deleted models.Storage
	err := (&sqlx.Tx{Tx: tx, Mapper: s.db.Mapper}).GetContext(ctx, &deleted, `WITH subtree AS (SELECT s.id, s.status FROM storages s JOIN storages root
			ON s.compute_cluster_id = root.compute_cluster_id AND (s.id = root.id OR starts_with(s.path, root.path || '/')) WHERE root.id = $1)
		, deleted AS (DELETE FROM storages WHERE id IN (SELECT id FROM subtree)
			AND NOT EXISTS (SELECT 1 FROM subtree WHERE status = 'active') RETURNING *)
		SELECT * FROM deleted WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &deleted, err
}
