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

type mysqlClusterAccountStore struct {
	db *sqlx.DB
}

// NewClusterAccountStore returns a MySQL-backed ClusterAccountStore.
func NewClusterAccountStore(db *sqlx.DB) ClusterAccountStore {
	return &mysqlClusterAccountStore{db: db}
}

const clusterAccountColumns = `id, user_id, username, status`

func (s *mysqlClusterAccountStore) FindByID(ctx context.Context, id string) (*models.ClusterAccount, error) {
	var a models.ClusterAccount
	err := s.db.GetContext(ctx, &a,
		`SELECT `+clusterAccountColumns+` FROM cluster_accounts WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (s *mysqlClusterAccountStore) FindByUsername(ctx context.Context, username string) (*models.ClusterAccount, error) {
	var a models.ClusterAccount
	err := s.db.GetContext(ctx, &a,
		`SELECT `+clusterAccountColumns+` FROM cluster_accounts WHERE username = ?`, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (s *mysqlClusterAccountStore) FindByUser(ctx context.Context, userID string) ([]models.ClusterAccount, error) {
	var out []models.ClusterAccount
	err := s.db.SelectContext(ctx, &out,
		`SELECT `+clusterAccountColumns+` FROM cluster_accounts WHERE user_id = ? ORDER BY created_at ASC`,
		userID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *mysqlClusterAccountStore) Create(ctx context.Context, tx *sql.Tx, a *models.ClusterAccount) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO cluster_accounts (id, user_id, username, status)
		 VALUES (?, ?, ?, ?)`,
		a.ID, a.UserID, a.Username, a.Status)
	return err
}

func (s *mysqlClusterAccountStore) UpdateStatus(ctx context.Context, tx *sql.Tx, id string, status models.AllocationStatus) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE cluster_accounts SET status = ? WHERE id = ?`,
		status, id)
	return err
}

func (s *mysqlClusterAccountStore) ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	// Drop fromUserID's accounts whose username the survivor already owns,
	// then move the rest. Username is globally unique on the deployment.
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM cluster_accounts
		 WHERE user_id = ?
		   AND username IN (SELECT username FROM (SELECT username FROM cluster_accounts WHERE user_id = ?) AS s)`,
		fromUserID, toUserID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE cluster_accounts SET user_id = ? WHERE user_id = ?`,
		toUserID, fromUserID)
	return err
}

func (s *mysqlClusterAccountStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM cluster_accounts WHERE id = ?`, id)
	return err
}
