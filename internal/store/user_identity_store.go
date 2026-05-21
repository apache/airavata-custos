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

type mysqlUserIdentityStore struct {
	db *sqlx.DB
}

// NewUserIdentityStore returns a MySQL-backed UserIdentityStore.
func NewUserIdentityStore(db *sqlx.DB) UserIdentityStore {
	return &mysqlUserIdentityStore{db: db}
}

// email, oidc_sub and metadata are nullable; project NULL to "" for the model.
const userIdentityColumns = `id, user_id, source, external_id, COALESCE(email, '') AS email, COALESCE(oidc_sub, '') AS oidc_sub, COALESCE(metadata, '') AS metadata, created_at`

// nullableString returns nil when s is empty so the column stores SQL NULL
// rather than "". NULL is the only value MySQL UNIQUE allows to repeat.
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (s *mysqlUserIdentityStore) FindByID(ctx context.Context, id string) (*models.UserIdentity, error) {
	var e models.UserIdentity
	err := s.db.GetContext(ctx, &e,
		`SELECT `+userIdentityColumns+` FROM user_identities WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (s *mysqlUserIdentityStore) FindBySourceAndExternalID(ctx context.Context, source, externalID string) (*models.UserIdentity, error) {
	var e models.UserIdentity
	err := s.db.GetContext(ctx, &e,
		`SELECT `+userIdentityColumns+` FROM user_identities WHERE source = ? AND external_id = ?`,
		source, externalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (s *mysqlUserIdentityStore) FindByOIDCSub(ctx context.Context, oidcSub string) (*models.UserIdentity, error) {
	if oidcSub == "" {
		return nil, nil
	}
	var e models.UserIdentity
	err := s.db.GetContext(ctx, &e,
		`SELECT `+userIdentityColumns+` FROM user_identities WHERE oidc_sub = ?`, oidcSub)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (s *mysqlUserIdentityStore) FindByUser(ctx context.Context, userID string) ([]models.UserIdentity, error) {
	var out []models.UserIdentity
	err := s.db.SelectContext(ctx, &out,
		`SELECT `+userIdentityColumns+` FROM user_identities WHERE user_id = ? ORDER BY created_at ASC`,
		userID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *mysqlUserIdentityStore) Create(ctx context.Context, tx *sql.Tx, e *models.UserIdentity) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO user_identities (id, user_id, source, external_id, email, oidc_sub, metadata)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.UserID, e.Source, e.ExternalID, nullableString(e.Email), nullableString(e.OIDCSub), nullableString(e.Metadata))
	return err
}

func (s *mysqlUserIdentityStore) Update(ctx context.Context, tx *sql.Tx, e *models.UserIdentity) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE user_identities
		    SET user_id = ?, source = ?, external_id = ?, email = ?, oidc_sub = ?, metadata = ?
		  WHERE id = ?`,
		e.UserID, e.Source, e.ExternalID, nullableString(e.Email), nullableString(e.OIDCSub), nullableString(e.Metadata), e.ID)
	return err
}

func (s *mysqlUserIdentityStore) ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE user_identities SET user_id = ? WHERE user_id = ?`,
		toUserID, fromUserID)
	return err
}

func (s *mysqlUserIdentityStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM user_identities WHERE id = ?`, id)
	return err
}
