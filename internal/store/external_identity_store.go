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

type mysqlExternalIdentityStore struct {
	db *sqlx.DB
}

// NewExternalIdentityStore returns a MySQL-backed ExternalIdentityStore.
func NewExternalIdentityStore(db *sqlx.DB) ExternalIdentityStore {
	return &mysqlExternalIdentityStore{db: db}
}

// oidc_sub and metadata are nullable; project NULL to ” for the model.
const externalIdentityColumns = `id, user_id, source, external_id, COALESCE(oidc_sub, '') AS oidc_sub, COALESCE(metadata, '') AS metadata, created_at`

// nullableString returns nil when s is empty so the column stores SQL NULL
// rather than ”. NULL is the only value MySQL UNIQUE allows to repeat.
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (s *mysqlExternalIdentityStore) FindByID(ctx context.Context, id string) (*models.ExternalIdentity, error) {
	var e models.ExternalIdentity
	err := s.db.GetContext(ctx, &e,
		`SELECT `+externalIdentityColumns+` FROM external_identities WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (s *mysqlExternalIdentityStore) FindBySourceAndExternalID(ctx context.Context, source, externalID string) (*models.ExternalIdentity, error) {
	var e models.ExternalIdentity
	err := s.db.GetContext(ctx, &e,
		`SELECT `+externalIdentityColumns+` FROM external_identities WHERE source = ? AND external_id = ?`,
		source, externalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (s *mysqlExternalIdentityStore) FindByOIDCSub(ctx context.Context, oidcSub string) (*models.ExternalIdentity, error) {
	if oidcSub == "" {
		return nil, nil
	}
	var e models.ExternalIdentity
	err := s.db.GetContext(ctx, &e,
		`SELECT `+externalIdentityColumns+` FROM external_identities WHERE oidc_sub = ?`, oidcSub)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (s *mysqlExternalIdentityStore) FindByUser(ctx context.Context, userID string) ([]models.ExternalIdentity, error) {
	var out []models.ExternalIdentity
	err := s.db.SelectContext(ctx, &out,
		`SELECT `+externalIdentityColumns+` FROM external_identities WHERE user_id = ? ORDER BY created_at ASC`,
		userID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *mysqlExternalIdentityStore) Create(ctx context.Context, tx *sql.Tx, e *models.ExternalIdentity) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO external_identities (id, user_id, source, external_id, oidc_sub, metadata)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		e.ID, e.UserID, e.Source, e.ExternalID, nullableString(e.OIDCSub), nullableString(e.Metadata))
	return err
}

func (s *mysqlExternalIdentityStore) Update(ctx context.Context, tx *sql.Tx, e *models.ExternalIdentity) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE external_identities
		    SET user_id = ?, source = ?, external_id = ?, oidc_sub = ?, metadata = ?
		  WHERE id = ?`,
		e.UserID, e.Source, e.ExternalID, nullableString(e.OIDCSub), nullableString(e.Metadata), e.ID)
	return err
}

func (s *mysqlExternalIdentityStore) ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE external_identities SET user_id = ? WHERE user_id = ?`,
		toUserID, fromUserID)
	return err
}

func (s *mysqlExternalIdentityStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM external_identities WHERE id = ?`, id)
	return err
}
