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

type pgComputeClusterUserStore struct {
	db *sqlx.DB
}

// NewComputeClusterUserStore returns a PostgreSQL-backed ComputeClusterUserStore.
func NewComputeClusterUserStore(db *sqlx.DB) ComputeClusterUserStore {
	return &pgComputeClusterUserStore{db: db}
}

const computeClusterUserColumns = `id, compute_cluster_id, user_id, local_username, provisioned_at`

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
		`INSERT INTO compute_cluster_users (id, compute_cluster_id, user_id, local_username)
         VALUES ($1, $2, $3, $4)`,
		c.ID, c.ComputeClusterID, c.UserID, c.LocalUsername)
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
