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

type mysqlUserDNStore struct {
	db *sqlx.DB
}

// NewUserDNStore returns a MySQL-backed UserDNStore.
func NewUserDNStore(db *sqlx.DB) UserDNStore {
	return &mysqlUserDNStore{db: db}
}

const userDNColumns = `id, user_id, dn, created_at`

func (s *mysqlUserDNStore) FindByID(ctx context.Context, id string) (*models.UserDN, error) {
	var d models.UserDN
	err := s.db.GetContext(ctx, &d,
		`SELECT `+userDNColumns+` FROM user_dns WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (s *mysqlUserDNStore) FindByDN(ctx context.Context, dn string) (*models.UserDN, error) {
	var d models.UserDN
	err := s.db.GetContext(ctx, &d,
		`SELECT `+userDNColumns+` FROM user_dns WHERE dn = ?`, dn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (s *mysqlUserDNStore) FindByUser(ctx context.Context, userID string) ([]models.UserDN, error) {
	var out []models.UserDN
	err := s.db.SelectContext(ctx, &out,
		`SELECT `+userDNColumns+` FROM user_dns WHERE user_id = ? ORDER BY created_at ASC`,
		userID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *mysqlUserDNStore) Create(ctx context.Context, tx *sql.Tx, d *models.UserDN) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO user_dns (id, user_id, dn) VALUES (?, ?, ?)`,
		d.ID, d.UserID, d.DN)
	return err
}

func (s *mysqlUserDNStore) ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	// Drop fromUserID's DNs already held by the survivor, then move the rest.
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM user_dns
		 WHERE user_id = ?
		   AND dn IN (SELECT dn FROM (SELECT dn FROM user_dns WHERE user_id = ?) AS s)`,
		fromUserID, toUserID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE user_dns SET user_id = ? WHERE user_id = ?`,
		toUserID, fromUserID)
	return err
}

func (s *mysqlUserDNStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM user_dns WHERE id = ?`, id)
	return err
}
