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

// StorageStore keeps each cluster's storages and the storage policies they are made from.
type StorageStore interface {
	// List pages through every cluster's storages, or one cluster's, pending ones first.
	List(ctx context.Context, clusterID string, limit, offset int) ([]models.Storage, int, error)
	// Ensure makes the storages the policies give the owners, so a returning owner's is active again, and returns
	// those it made or changed, a parent before what is under it. An administrator's quota stands until it lapses.
	Ensure(ctx context.Context, tx *sql.Tx, o StorageOwners) ([]models.Storage, error)
	// SetQuota replaces the quota of a storage that has one of its own, and returns the storage and its old quota; nil without one.
	SetQuota(ctx context.Context, tx *sql.Tx, id string, bytes int64, until time.Time) (*models.Storage, int64, error)
	// Delete drops a pending storage and those under it once all of them are pending, and returns them, deepest first.
	Delete(ctx context.Context, tx *sql.Tx, id string) ([]models.Storage, error)
	// RecordUsage sets the use of the cluster's storages, matched by position to paths.
	RecordUsage(ctx context.Context, clusterID string, paths []string, used []int64) error
	// MarkForDeletion marks the active storages of an account's username or of an allocation.
	MarkForDeletion(ctx context.Context, tx *sql.Tx, clusterID, username, allocationID string) error

	// ListPolicies returns the cluster's policies ordered by pattern, so a parent precedes what is under it.
	ListPolicies(ctx context.Context, clusterID string) ([]models.StoragePolicy, error)
	CreatePolicy(ctx context.Context, p *models.StoragePolicy) error
	// UpdatePolicy replaces what the policy of p's id and type says and reads the rest into p; false without one.
	UpdatePolicy(ctx context.Context, p *models.StoragePolicy) (bool, error)
	// DeletePolicy reports whether the policy existed.
	DeletePolicy(ctx context.Context, tx *sql.Tx, id string) (bool, error)
}

// under matches storage s when it is storage p or beneath it.
const under = `s.compute_cluster_id = p.compute_cluster_id AND (s.id = p.id OR starts_with(s.path, p.path || '/'))`

// StorageOwners narrows Ensure to a cluster, an account or an allocation. An empty field matches any.
type StorageOwners struct {
	ClusterID, UserID, AllocationID string
}

// lockStorageAccounts and lockStorageAllocations lock the account named, then the allocation named or those of the
// account, though not their keys. An account, a group and a membership provisioned at the same moment then queue and
// the last sees the others, while memberships being inserted, which hold a key share on the allocation, do not
// deadlock. A scope naming neither locks nothing.
const lockStorageAccounts = `SELECT 1 FROM compute_cluster_users cu
	WHERE cu.user_id = $2 AND ($1 = '' OR cu.compute_cluster_id = $1)
		AND ($3 = '' OR cu.compute_cluster_id = (SELECT compute_cluster_id FROM compute_allocations WHERE id = $3))
	ORDER BY cu.id FOR NO KEY UPDATE`

const lockStorageAllocations = `SELECT 1 FROM compute_allocations a
	WHERE ($1 = '' OR a.compute_cluster_id = $1) AND ($3 = '' OR a.id = $3)
		AND ($3 <> '' OR EXISTS (SELECT 1 FROM compute_allocation_memberships m WHERE m.compute_allocation_id = a.id AND m.user_id = $2))
	ORDER BY a.id FOR NO KEY UPDATE OF a`

// filled is a policy's template with its placeholders replaced by the owners' names.
func filled(template string) string {
	return `replace(replace(` + template + `, '{user}', COALESCE(cu.local_username, '')), '{allocation}', COALESCE(a.posix_group, ''))`
}

// ensureStorages is the one rule for who gets a storage: a policy whose path names {user} takes each provisioned
// account, one naming {allocation} takes each active allocation with a provisioned group, and one naming both takes
// each active membership of the two. It returns the storages it made or changed, so running it again changes nothing.
var ensureStorages = `INSERT INTO storages AS s
	(id, compute_cluster_id, policy_id, type, path, username, allocation_id, allocation_name, start_time, end_time, owner, "group", mode, quota_bytes, quota_inodes)
	SELECT DISTINCT ON (p.compute_cluster_id, path) gen_random_uuid(), p.compute_cluster_id, p.id, p.type, ` + filled(`p.path_pattern`) + ` AS path,
		COALESCE(cu.local_username, ''), a.id, a.name, a.start_time, a.end_time, ` + filled(`p.owner`) + `, ` + filled(`p."group"`) + `, p.mode, p.quota_bytes, p.quota_inodes
	FROM storage_policies p
	LEFT JOIN compute_cluster_users cu ON p.path_pattern LIKE '%{user}%' AND cu.compute_cluster_id = p.compute_cluster_id AND cu.provisioned_at IS NOT NULL
	LEFT JOIN compute_allocations a ON p.path_pattern LIKE '%{allocation}%' AND a.compute_cluster_id = p.compute_cluster_id
		AND a.status = 'ACTIVE' AND a.posix_group IS NOT NULL AND a.group_provisioned_at IS NOT NULL
	WHERE (cu.id IS NOT NULL) = (p.path_pattern LIKE '%{user}%') AND (a.id IS NOT NULL) = (p.path_pattern LIKE '%{allocation}%')
		AND (cu.id IS NULL OR a.id IS NULL OR EXISTS (SELECT 1 FROM compute_allocation_memberships m
			WHERE m.compute_allocation_id = a.id AND m.user_id = cu.user_id AND m.membership_status = 'ACTIVE'))
		AND ($1 = '' OR p.compute_cluster_id = $1) AND ($2 = '' OR cu.user_id = $2) AND ($3 = '' OR a.id = $3)
	ORDER BY p.compute_cluster_id, path, p.id
	ON CONFLICT (compute_cluster_id, path) DO UPDATE SET policy_id = EXCLUDED.policy_id, type = EXCLUDED.type, username = EXCLUDED.username,
		allocation_id = EXCLUDED.allocation_id, allocation_name = EXCLUDED.allocation_name, start_time = EXCLUDED.start_time,
		end_time = EXCLUDED.end_time, owner = EXCLUDED.owner, "group" = EXCLUDED."group", mode = EXCLUDED.mode,
		quota_bytes = CASE WHEN s.quota_until > now() THEN s.quota_bytes ELSE EXCLUDED.quota_bytes END,
		quota_until = CASE WHEN s.quota_until > now() THEN s.quota_until END,
		quota_inodes = EXCLUDED.quota_inodes, status = 'active'
	WHERE s.status <> 'active' OR s.quota_until <= now() OR s.quota_until IS NULL AND s.quota_bytes <> EXCLUDED.quota_bytes
		OR (s.policy_id, s.type, s.username, s.allocation_id, s.allocation_name, s.start_time, s.end_time, s.owner, s."group", s.mode, s.quota_inodes)
			IS DISTINCT FROM (EXCLUDED.policy_id, EXCLUDED.type, EXCLUDED.username, EXCLUDED.allocation_id, EXCLUDED.allocation_name,
				EXCLUDED.start_time, EXCLUDED.end_time, EXCLUDED.owner, EXCLUDED."group", EXCLUDED.mode, EXCLUDED.quota_inodes)
	RETURNING s.id`

// selectStorages reads storages with whether each counts against the quota of one above it.
const selectStorages = `SELECT s.*, s.quota_bytes = 0 AND EXISTS (SELECT 1 FROM storages p WHERE p.quota_bytes <> 0 AND ` + under + `) AS inherited FROM storages s`

type pgStorageStore struct {
	db *sqlx.DB
}

// NewStorageStore returns a PostgreSQL-backed StorageStore.
func NewStorageStore(db *sqlx.DB) StorageStore {
	return &pgStorageStore{db: db}
}

func (s *pgStorageStore) List(ctx context.Context, clusterID string, limit, offset int) ([]models.Storage, int, error) {
	var total int
	if err := s.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM storages WHERE $1 = '' OR compute_cluster_id = $1`, clusterID); err != nil {
		return nil, 0, err
	}
	storages := []models.Storage{}
	err := s.db.SelectContext(ctx, &storages, selectStorages+` WHERE $1 = '' OR s.compute_cluster_id = $1
		ORDER BY s.status = 'active', s.path COLLATE "C", s.id LIMIT $2 OFFSET $3`, clusterID, limit, offset)
	return storages, total, err
}

func (s *pgStorageStore) Ensure(ctx context.Context, tx *sql.Tx, o StorageOwners) ([]models.Storage, error) {
	for _, lock := range []string{lockStorageAccounts, lockStorageAllocations} {
		if _, err := tx.ExecContext(ctx, lock, o.ClusterID, o.UserID, o.AllocationID); err != nil {
			return nil, err
		}
	}
	var ids []string
	if err := (&sqlx.Tx{Tx: tx, Mapper: s.db.Mapper}).SelectContext(ctx, &ids, ensureStorages, o.ClusterID, o.UserID, o.AllocationID); err != nil {
		return nil, err
	}
	return scan(tx.QueryContext(ctx, selectStorages+` WHERE s.id = ANY($1) ORDER BY s.path COLLATE "C"`, ids))
}

func (s *pgStorageStore) SetQuota(ctx context.Context, tx *sql.Tx, id string, bytes int64, until time.Time) (*models.Storage, int64, error) {
	var from int64
	err := tx.QueryRowContext(ctx, `UPDATE storages s SET quota_bytes = $2, quota_until = $3
		FROM (SELECT id, quota_bytes FROM storages WHERE id = $1 FOR UPDATE) old
		WHERE s.id = old.id AND old.quota_bytes <> 0 RETURNING old.quota_bytes`, id, bytes, until).Scan(&from)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	found, err := scan(tx.QueryContext(ctx, selectStorages+` WHERE s.id = $1`, id))
	if err != nil {
		return nil, 0, err
	}
	return &found[0], from, nil
}

func (s *pgStorageStore) Delete(ctx context.Context, tx *sql.Tx, id string) ([]models.Storage, error) {
	return scan(tx.QueryContext(ctx, `WITH subtree AS (SELECT s.id, s.status FROM storages s JOIN storages p ON `+under+` WHERE p.id = $1)
		, deleted AS (DELETE FROM storages WHERE id IN (SELECT id FROM subtree)
			AND NOT EXISTS (SELECT 1 FROM subtree WHERE status = 'active') RETURNING id) `+
		selectStorages+` WHERE s.id IN (SELECT id FROM deleted) ORDER BY s.path COLLATE "C" DESC`, id))
}

func (s *pgStorageStore) RecordUsage(ctx context.Context, clusterID string, paths []string, used []int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE storages s SET used_bytes = u.used FROM unnest($2::text[], $3::bigint[]) AS u(path, used)
		WHERE s.compute_cluster_id = $1 AND s.path = u.path`, clusterID, paths, used)
	return err
}

func (s *pgStorageStore) MarkForDeletion(ctx context.Context, tx *sql.Tx, clusterID, username, allocationID string) error {
	_, err := tx.ExecContext(ctx, `UPDATE storages SET status = 'pending_deletion'
		WHERE compute_cluster_id = $1 AND status = 'active' AND (username = NULLIF($2, '') OR allocation_id = $3)`, clusterID, username, allocationID)
	return err
}

func (s *pgStorageStore) ListPolicies(ctx context.Context, clusterID string) ([]models.StoragePolicy, error) {
	policies := []models.StoragePolicy{}
	err := s.db.SelectContext(ctx, &policies, `SELECT * FROM storage_policies WHERE compute_cluster_id = $1 ORDER BY path_pattern COLLATE "C"`, clusterID)
	return policies, err
}

func (s *pgStorageStore) CreatePolicy(ctx context.Context, p *models.StoragePolicy) error {
	_, err := s.db.NamedExecContext(ctx, `INSERT INTO storage_policies (id, compute_cluster_id, type, path_pattern, owner, "group", mode, quota_bytes, quota_inodes)
		VALUES (:id, :compute_cluster_id, :type, :path_pattern, :owner, :group, :mode, :quota_bytes, :quota_inodes)`, p)
	return err
}

func (s *pgStorageStore) UpdatePolicy(ctx context.Context, p *models.StoragePolicy) (bool, error) {
	rows, err := s.db.NamedQueryContext(ctx, `UPDATE storage_policies SET path_pattern = :path_pattern, owner = :owner, "group" = :group, mode = :mode,
		quota_bytes = :quota_bytes, quota_inodes = :quota_inodes WHERE id = :id AND type = :type RETURNING *`, p)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return false, rows.Err()
	}
	return true, rows.StructScan(p)
}

func (s *pgStorageStore) DeletePolicy(ctx context.Context, tx *sql.Tx, id string) (bool, error) {
	return execOnce(ctx, tx, `DELETE FROM storage_policies WHERE id = $1`, id)
}

func scan(rows *sql.Rows, err error) ([]models.Storage, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var storages []models.Storage
	return storages, sqlx.StructScan(rows, &storages)
}
