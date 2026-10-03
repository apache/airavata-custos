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
	FirstName   string `db:"first_name"`
	LastName    string `db:"last_name"`
	Email       string `db:"user_email"`
	ClusterName string `db:"cluster_name"`
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

func (s *pgComputeClusterUserStore) Create(ctx context.Context, tx *sql.Tx, c *models.ComputeClusterUser) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_cluster_users (id, compute_cluster_id, user_id, local_username, access_level, approval_status, reviewed_at, reviewed_by)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID, c.ComputeClusterID, c.UserID, c.LocalUsername, c.AccessLevel, c.ApprovalStatus, c.ReviewedAt, c.ReviewedBy)
	return err
}

func (s *pgComputeClusterUserStore) ListByApprovalStatus(ctx context.Context, status models.ClusterAccountApproval, limit, offset int) ([]ComputeClusterUserWithUser, int, error) {
	var rows []ComputeClusterUserWithUser
	err := s.db.SelectContext(ctx, &rows,
		`SELECT cu.id, cu.compute_cluster_id, cu.user_id, cu.local_username, cu.access_level, cu.provisioned_at,
		        cu.approval_status, cu.reviewed_at, cu.reviewed_by, cu.review_note,
		        u.first_name, u.last_name, u.email AS user_email, c.name AS cluster_name
		   FROM compute_cluster_users cu
		   JOIN users u            ON u.id = cu.user_id
		   JOIN compute_clusters c ON c.id = cu.compute_cluster_id
		  WHERE $1 = '' OR cu.approval_status = $1
		  ORDER BY cu.created_at DESC
		  LIMIT $2 OFFSET $3`,
		status, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	var total int
	if err := s.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM compute_cluster_users WHERE $1 = '' OR approval_status = $1`, status); err != nil {
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
	return err
}

func (s *pgComputeClusterUserStore) MarkProvisioned(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `UPDATE compute_cluster_users SET provisioned_at = NOW() WHERE id = $1`, id)
	return err
}

func (s *pgComputeClusterUserStore) ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
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
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_cluster_users WHERE id = $1`, id)
	return err
}
