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

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
)

type UserDNStore interface {
	// Add inserts a DN row. Uses INSERT IGNORE so a duplicate DN is a silent
	// no-op rather than an error.
	Add(ctx context.Context, tx *sql.Tx, d *model.UserDN) error
	// ListByUser returns every DN bound to userID, oldest first.
	ListByUser(ctx context.Context, userID string) ([]model.UserDN, error)
	// DeleteByID removes a single DN row.
	DeleteByID(ctx context.Context, tx *sql.Tx, id string) error
	// ReassignUser moves every DN owned by fromUserID over to toUserID,
	// dropping duplicates that would collide with toUserID's existing rows.
	ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error
}

type mariaDBUserDNStore struct {
	db *sqlx.DB
}

func NewUserDNStore(db *sqlx.DB) UserDNStore {
	return &mariaDBUserDNStore{db: db}
}

func (s *mariaDBUserDNStore) Add(ctx context.Context, tx *sql.Tx, d *model.UserDN) error {
	if d.ID == "" {
		d.ID = uuid.NewString()
	}
	_, err := tx.ExecContext(ctx,
		`INSERT IGNORE INTO amie_user_dns (id, user_id, dn) VALUES (?, ?, ?)`,
		d.ID, d.UserID, d.DN)
	return err
}

func (s *mariaDBUserDNStore) ListByUser(ctx context.Context, userID string) ([]model.UserDN, error) {
	var out []model.UserDN
	err := s.db.SelectContext(ctx, &out,
		`SELECT id, user_id, dn, created_at
           FROM amie_user_dns
          WHERE user_id = ?
          ORDER BY created_at ASC`, userID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *mariaDBUserDNStore) DeleteByID(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM amie_user_dns WHERE id = ?`, id)
	return err
}

func (s *mariaDBUserDNStore) ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	// Drop fromUserID's rows whose DN is already held by toUserID; UNIQUE(dn)
	// would otherwise reject the UPDATE.
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM amie_user_dns
		  WHERE user_id = ?
		    AND dn IN (SELECT dn FROM (SELECT dn FROM amie_user_dns WHERE user_id = ?) AS s)`,
		fromUserID, toUserID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE amie_user_dns SET user_id = ? WHERE user_id = ?`,
		toUserID, fromUserID)
	return err
}
