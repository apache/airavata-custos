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

type mysqlComputeClusterUserStore struct {
	db *sqlx.DB
}

// NewComputeClusterUserStore returns a MySQL-backed ComputeClusterUserStore.
func NewComputeClusterUserStore(db *sqlx.DB) ComputeClusterUserStore {
	return &mysqlComputeClusterUserStore{db: db}
}

const computeClusterUserColumns = `id, compute_cluster_id, user_id, local_username`

func (s *mysqlComputeClusterUserStore) FindByID(ctx context.Context, id string) (*models.ComputeClusterUser, error) {
	var c models.ComputeClusterUser
	err := s.db.GetContext(ctx, &c,
		`SELECT `+computeClusterUserColumns+`
           FROM compute_cluster_users WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *mysqlComputeClusterUserStore) FindByPair(ctx context.Context, clusterID, userID string) (*models.ComputeClusterUser, error) {
	var c models.ComputeClusterUser
	err := s.db.GetContext(ctx, &c,
		`SELECT `+computeClusterUserColumns+`
           FROM compute_cluster_users
          WHERE compute_cluster_id = ? AND user_id = ?`, clusterID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *mysqlComputeClusterUserStore) FindByClusterAndLocalUsername(ctx context.Context, clusterID, localUsername string) (*models.ComputeClusterUser, error) {
	var c models.ComputeClusterUser
	err := s.db.GetContext(ctx, &c,
		`SELECT `+computeClusterUserColumns+`
           FROM compute_cluster_users
          WHERE compute_cluster_id = ? AND local_username = ?`, clusterID, localUsername)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *mysqlComputeClusterUserStore) FindByCluster(ctx context.Context, clusterID string) ([]models.ComputeClusterUser, error) {
	var users []models.ComputeClusterUser
	err := s.db.SelectContext(ctx, &users,
		`SELECT `+computeClusterUserColumns+`
           FROM compute_cluster_users
          WHERE compute_cluster_id = ?
          ORDER BY local_username`, clusterID)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *mysqlComputeClusterUserStore) FindByUser(ctx context.Context, userID string) ([]models.ComputeClusterUser, error) {
	var users []models.ComputeClusterUser
	err := s.db.SelectContext(ctx, &users,
		`SELECT `+computeClusterUserColumns+`
           FROM compute_cluster_users
          WHERE user_id = ?
          ORDER BY compute_cluster_id`, userID)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *mysqlComputeClusterUserStore) Create(ctx context.Context, tx *sql.Tx, c *models.ComputeClusterUser) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_cluster_users (id, compute_cluster_id, user_id, local_username)
         VALUES (?, ?, ?, ?)`,
		c.ID, c.ComputeClusterID, c.UserID, c.LocalUsername)
	return err
}

func (s *mysqlComputeClusterUserStore) Update(ctx context.Context, tx *sql.Tx, c *models.ComputeClusterUser) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_cluster_users
            SET compute_cluster_id = ?,
                user_id            = ?,
                local_username     = ?
          WHERE id = ?`,
		c.ComputeClusterID, c.UserID, c.LocalUsername, c.ID)
	return err
}

func (s *mysqlComputeClusterUserStore) ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM compute_cluster_users
		 WHERE user_id = ?
		   AND compute_cluster_id IN (
		       SELECT compute_cluster_id FROM (
		           SELECT compute_cluster_id FROM compute_cluster_users WHERE user_id = ?
		       ) AS s
		   )`,
		fromUserID, toUserID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_cluster_users SET user_id = ? WHERE user_id = ?`,
		toUserID, fromUserID)
	return err
}

func (s *mysqlComputeClusterUserStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_cluster_users WHERE id = ?`, id)
	return err
}
