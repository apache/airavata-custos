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

type mysqlUserMergeStore struct {
	db *sqlx.DB
}

// NewUserMergeStore returns a MySQL-backed UserMergeStore.
func NewUserMergeStore(db *sqlx.DB) UserMergeStore {
	return &mysqlUserMergeStore{db: db}
}

const userMergeColumns = `id, retiring_user_id, surviving_user_id, COALESCE(reason, '') AS reason, merged_at`

func (s *mysqlUserMergeStore) Record(ctx context.Context, tx *sql.Tx, retiringUserID, survivingUserID, reason string) error {
	var reasonArg any
	if reason == "" {
		reasonArg = nil
	} else {
		reasonArg = reason
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO user_merges (retiring_user_id, surviving_user_id, reason)
		 VALUES (?, ?, ?)`,
		retiringUserID, survivingUserID, reasonArg)
	return err
}

func (s *mysqlUserMergeStore) FindByRetiringUser(ctx context.Context, retiringUserID string) (*models.UserMerge, error) {
	var m models.UserMerge
	err := s.db.GetContext(ctx, &m,
		`SELECT `+userMergeColumns+` FROM user_merges WHERE retiring_user_id = ?`, retiringUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (s *mysqlUserMergeStore) FindBySurvivingUser(ctx context.Context, survivingUserID string) ([]models.UserMerge, error) {
	var out []models.UserMerge
	err := s.db.SelectContext(ctx, &out,
		`SELECT `+userMergeColumns+` FROM user_merges WHERE surviving_user_id = ? ORDER BY merged_at ASC`,
		survivingUserID)
	if err != nil {
		return nil, err
	}
	return out, nil
}
