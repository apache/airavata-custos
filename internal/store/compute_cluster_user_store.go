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

// ComputeClusterUserWithUser is a cluster user joined with the fields the
// review list shows.
type ComputeClusterUserWithUser struct {
	models.ComputeClusterUser
	RequestedAt    time.Time `db:"created_at"`
	FirstName      string    `db:"first_name"`
	LastName       string    `db:"last_name"`
	Email          string    `db:"user_email"`
	ClusterName    string    `db:"cluster_name"`
	ReviewerName   *string   `db:"reviewer_name"`
	AllocationID   *string   `db:"allocation_id"`
	AllocationName *string   `db:"allocation_name"`
	ProjectID      *string   `db:"project_id"`
}

type pgComputeClusterUserStore struct {
	db *sqlx.DB
}

// NewComputeClusterUserStore returns a PostgreSQL-backed ComputeClusterUserStore.
func NewComputeClusterUserStore(db *sqlx.DB) ComputeClusterUserStore {
	return &pgComputeClusterUserStore{db: db}
}

const computeClusterUserColumns = `id, compute_cluster_id, user_id, local_username, access_level, provisioned_at, approval_status, reviewed_at, reviewed_by, review_note`

func (s *pgComputeClusterUserStore) FindByID(ctx context.Context, id string) (*models.ComputeClusterUser, error) {
	var c models.ComputeClusterUser
	err := s.db.GetContext(ctx, &c,
		`SELECT `+computeClusterUserColumns+`
           FROM compute_cluster_users WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *pgComputeClusterUserStore) FindByIDTx(ctx context.Context, tx *sql.Tx, id string) (*models.ComputeClusterUser, error) {
	var c models.ComputeClusterUser
	err := (&sqlx.Tx{Tx: tx, Mapper: s.db.Mapper}).GetContext(ctx, &c,
		`SELECT `+computeClusterUserColumns+` FROM compute_cluster_users WHERE id = $1 FOR UPDATE`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *pgComputeClusterUserStore) FindByPair(ctx context.Context, clusterID, userID string) (*models.ComputeClusterUser, error) {
	var c models.ComputeClusterUser
	err := s.db.GetContext(ctx, &c,
		`SELECT `+computeClusterUserColumns+`
           FROM compute_cluster_users
          WHERE compute_cluster_id = $1 AND user_id = $2`, clusterID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *pgComputeClusterUserStore) FindByClusterAndLocalUsername(ctx context.Context, clusterID, localUsername string) (*models.ComputeClusterUser, error) {
	var c models.ComputeClusterUser
	err := s.db.GetContext(ctx, &c,
		`SELECT `+computeClusterUserColumns+`
           FROM compute_cluster_users
          WHERE compute_cluster_id = $1 AND local_username = $2`, clusterID, localUsername)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *pgComputeClusterUserStore) FindByCluster(ctx context.Context, clusterID string) ([]models.ComputeClusterUser, error) {
	var users []models.ComputeClusterUser
	err := s.db.SelectContext(ctx, &users,
		`SELECT `+computeClusterUserColumns+`
           FROM compute_cluster_users
          WHERE compute_cluster_id = $1
          ORDER BY local_username`, clusterID)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *pgComputeClusterUserStore) FindByUser(ctx context.Context, userID string) ([]models.ComputeClusterUser, error) {
	var users []models.ComputeClusterUser
	err := s.db.SelectContext(ctx, &users,
		`SELECT `+computeClusterUserColumns+`
           FROM compute_cluster_users
          WHERE user_id = $1
          ORDER BY compute_cluster_id`, userID)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *pgComputeClusterUserStore) FindByUserTx(ctx context.Context, tx *sql.Tx, userID string) ([]models.ComputeClusterUser, error) {
	var users []models.ComputeClusterUser
	err := (&sqlx.Tx{Tx: tx, Mapper: s.db.Mapper}).SelectContext(ctx, &users,
		`SELECT `+computeClusterUserColumns+` FROM compute_cluster_users WHERE user_id = $1 ORDER BY compute_cluster_id FOR UPDATE`, userID)
	return users, err
}

func (s *pgComputeClusterUserStore) Create(ctx context.Context, tx *sql.Tx, c *models.ComputeClusterUser) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_cluster_users (id, compute_cluster_id, user_id, local_username, access_level, approval_status, reviewed_at, reviewed_by)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID, c.ComputeClusterID, c.UserID, c.LocalUsername, c.AccessLevel, c.ApprovalStatus, c.ReviewedAt, c.ReviewedBy)
	if err != nil {
		return err
	}
	return claimLocalUsername(ctx, tx, c)
}

// claimLocalUsername runs after the write, which waits for a concurrent delete of this username to commit, so its
// retirement is visible. A user taking back their own held username frees it.
func claimLocalUsername(ctx context.Context, tx *sql.Tx, c *models.ComputeClusterUser) error {
	var held bool
	err := tx.QueryRowContext(ctx, `WITH d AS (DELETE FROM retired_local_usernames WHERE compute_cluster_id = $1 AND local_username = $2 RETURNING user_id)
		SELECT EXISTS (SELECT 1 FROM d WHERE user_id <> $3)`, c.ComputeClusterID, c.LocalUsername, c.UserID).Scan(&held)
	if err == nil && held {
		return ErrUsernameHeld
	}
	return err
}

// ErrUsernameHeld reports a username held for the user who left it.
var ErrUsernameHeld = errors.New("local username is held for its former user")

// clusterAccountListFrom links each account to the user's newest allocation on
// the same cluster.
// TODO The request does not store its allocation, so this is inferred.
const clusterAccountListFrom = `
	  FROM compute_cluster_users cu
	  JOIN users u            ON u.id = cu.user_id
	  JOIN compute_clusters c ON c.id = cu.compute_cluster_id
	  LEFT JOIN users r       ON r.id = cu.reviewed_by
	  LEFT JOIN LATERAL (
	       SELECT ca.id, ca.name, ca.project_id, p.title
	         FROM compute_allocation_memberships m
	         JOIN compute_allocations ca ON ca.id = m.compute_allocation_id
	         JOIN projects p             ON p.id = ca.project_id
	        WHERE m.user_id = cu.user_id AND ca.compute_cluster_id = cu.compute_cluster_id
	        ORDER BY m.created_at DESC
	        LIMIT 1) a ON true
	 WHERE ($1 = '' OR cu.approval_status = $1)
	   AND ($2 = '' OR CONCAT_WS(' ', u.first_name, u.last_name, u.email, cu.local_username, a.name, a.title) ILIKE '%' || $2 || '%')`

func (s *pgComputeClusterUserStore) ListByApprovalStatus(ctx context.Context, status models.ClusterAccountApproval, query string, limit, offset int) ([]ComputeClusterUserWithUser, int, error) {
	// Pending is worked one allocation at a time, so its rows come grouped.
	order := `cu.created_at DESC, cu.id`
	if status == models.ClusterAccountPending {
		order = `a.id IS NULL, a.title, a.name, a.id, cu.created_at, cu.id`
	}
	var rows []ComputeClusterUserWithUser
	err := s.db.SelectContext(ctx, &rows,
		`SELECT cu.id, cu.compute_cluster_id, cu.user_id, cu.local_username, cu.access_level, cu.provisioned_at,
		        cu.approval_status, cu.reviewed_at, cu.reviewed_by, cu.review_note, cu.created_at,
		        u.first_name, u.last_name, u.email AS user_email, c.name AS cluster_name,
		        NULLIF(TRIM(CONCAT_WS(' ', r.first_name, r.last_name)), '') AS reviewer_name,
		        a.id AS allocation_id, a.name AS allocation_name, a.project_id`+
			clusterAccountListFrom+`
		  ORDER BY `+order+`
		  LIMIT $3 OFFSET $4`,
		status, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	var total int
	if err := s.db.GetContext(ctx, &total, `SELECT COUNT(*)`+clusterAccountListFrom, status, query); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (s *pgComputeClusterUserStore) Review(ctx context.Context, tx *sql.Tx, id string, status models.ClusterAccountApproval, reviewedBy, note string, at time.Time) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_cluster_users
		    SET approval_status = $1, reviewed_by = $2, reviewed_at = $3, review_note = NULLIF($4, '')
		  WHERE id = $5`,
		status, reviewedBy, at, note, id)
	return err
}

func (s *pgComputeClusterUserStore) Update(ctx context.Context, tx *sql.Tx, c *models.ComputeClusterUser) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_cluster_users
            SET compute_cluster_id = $1,
                user_id            = $2,
                local_username     = $3
          WHERE id = $4`,
		c.ComputeClusterID, c.UserID, c.LocalUsername, c.ID)
	if err != nil {
		return err
	}
	return claimLocalUsername(ctx, tx, c)
}

func (s *pgComputeClusterUserStore) MarkProvisioned(ctx context.Context, tx *sql.Tx, id string, at time.Time) (bool, error) {
	res, err := tx.ExecContext(ctx, `UPDATE compute_cluster_users SET provisioned_at = $1 WHERE id = $2 AND provisioned_at IS NULL`, at, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}

func (s *pgComputeClusterUserStore) ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	if _, err := tx.ExecContext(ctx,
		retireLocalUsernames+`
		 WHERE user_id = $1
		   AND compute_cluster_id IN (SELECT compute_cluster_id FROM compute_cluster_users WHERE user_id = $2)
		ON CONFLICT DO NOTHING`,
		fromUserID, toUserID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE retired_local_usernames SET user_id = $2 WHERE user_id = $1`, fromUserID, toUserID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM compute_cluster_users
		 WHERE user_id = $1
		   AND compute_cluster_id IN (
		       SELECT compute_cluster_id FROM (
		           SELECT compute_cluster_id FROM compute_cluster_users WHERE user_id = $2
		       ) AS s
		   )`,
		fromUserID, toUserID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_cluster_users SET user_id = $1 WHERE user_id = $2`,
		toUserID, fromUserID)
	return err
}

func (s *pgComputeClusterUserStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	if _, err := tx.ExecContext(ctx, retireLocalUsernames+` WHERE id = $1 ON CONFLICT DO NOTHING`, id); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_cluster_users WHERE id = $1`, id)
	return err
}

// retireLocalUsernames holds the usernames of the mappings about to be deleted for their users.
const retireLocalUsernames = `INSERT INTO retired_local_usernames (compute_cluster_id, local_username, user_id)
	SELECT compute_cluster_id, local_username, user_id FROM compute_cluster_users`

func (s *pgComputeClusterUserStore) Retired(ctx context.Context, clusterID, userID, localUsername string) (string, bool, error) {
	var r struct {
		Own   string `db:"own"`
		Taken bool   `db:"taken"`
	}
	err := s.db.GetContext(ctx, &r,
		`SELECT COALESCE(MAX(local_username) FILTER (WHERE user_id = $2), '') AS own,
		        COALESCE(BOOL_OR(local_username = $3 AND user_id <> $2), false) AS taken
		   FROM retired_local_usernames WHERE compute_cluster_id = $1`,
		clusterID, userID, localUsername)
	return r.Own, r.Taken, err
}

func (s *pgComputeClusterUserStore) ListUsernames(ctx context.Context, clusterID string) ([]models.ClusterUsername, error) {
	rows := []models.ClusterUsername{}
	err := s.db.SelectContext(ctx, &rows,
		`SELECT n.local_username, n.user_id, COALESCE(u.email, '') AS email, n.since, n.held
		   FROM (SELECT local_username, user_id, created_at AS since, false AS held FROM compute_cluster_users WHERE compute_cluster_id = $1
		         UNION ALL
		         SELECT local_username, user_id, retired_at, true FROM retired_local_usernames WHERE compute_cluster_id = $1) n
		   LEFT JOIN users u ON u.id = n.user_id
		  ORDER BY n.held DESC, n.local_username`, clusterID)
	return rows, err
}

func (s *pgComputeClusterUserStore) Release(ctx context.Context, tx *sql.Tx, clusterID, localUsername string) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM retired_local_usernames WHERE compute_cluster_id = $1 AND local_username = $2`, clusterID, localUsername)
	return err
}
