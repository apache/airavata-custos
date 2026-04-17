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

	"github.com/apache/airavata-custos/allocations/domain/model"
	"github.com/jmoiron/sqlx"
)

type mariaDBClusterAccountStore struct {
	db *sqlx.DB
}

func NewClusterAccountStore(db *sqlx.DB) ClusterAccountStore {
	return &mariaDBClusterAccountStore{db: db}
}

func (s *mariaDBClusterAccountStore) FindByUsername(ctx context.Context, username string) (*model.ClusterAccount, error) {
	var a model.ClusterAccount
	err := s.db.GetContext(ctx, &a,
		`SELECT id, person_id, username, created_at, updated_at
		 FROM cluster_accounts WHERE username = ?`, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (s *mariaDBClusterAccountStore) FindByPerson(ctx context.Context, personID string) ([]model.ClusterAccount, error) {
	var results []model.ClusterAccount
	err := s.db.SelectContext(ctx, &results,
		`SELECT id, person_id, username, created_at, updated_at
		 FROM cluster_accounts WHERE person_id = ?`, personID)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *mariaDBClusterAccountStore) Save(ctx context.Context, tx *sql.Tx, a *model.ClusterAccount) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO cluster_accounts (id, person_id, username, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		a.ID, a.PersonID, a.Username, a.CreatedAt, a.UpdatedAt)
	return err
}

func (s *mariaDBClusterAccountStore) UpdatePersonID(ctx context.Context, tx *sql.Tx, accountID, newPersonID string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE cluster_accounts SET person_id = ? WHERE id = ?`, newPersonID, accountID)
	return err
}
