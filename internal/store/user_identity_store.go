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

type pgUserIdentityStore struct {
	db *sqlx.DB
}

// NewUserIdentityStore returns a PostgreSQL-backed UserIdentityStore.
func NewUserIdentityStore(db *sqlx.DB) UserIdentityStore {
	return &pgUserIdentityStore{db: db}
}

// email, oidc_sub and metadata are nullable; project NULL to "" for the model.
const userIdentityColumns = `id, user_id, source, external_id, COALESCE(email, '') AS email, COALESCE(oidc_sub, '') AS oidc_sub, COALESCE(metadata, '') AS metadata, created_at`

// nullableString returns nil when s is empty so the column stores SQL NULL
// rather than "". NULL is the only value the UNIQUE index allows to repeat.
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (s *pgUserIdentityStore) FindByID(ctx context.Context, id string) (*models.UserIdentity, error) {
	var e models.UserIdentity
	err := s.db.GetContext(ctx, &e,
		`SELECT `+userIdentityColumns+` FROM user_identities WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (s *pgUserIdentityStore) FindBySourceAndExternalID(ctx context.Context, source, externalID string) (*models.UserIdentity, error) {
	var e models.UserIdentity
	err := s.db.GetContext(ctx, &e,
		`SELECT `+userIdentityColumns+` FROM user_identities WHERE source = $1 AND external_id = $2`,
		source, externalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (s *pgUserIdentityStore) FindByOIDCSub(ctx context.Context, oidcSub string) (*models.UserIdentity, error) {
	if oidcSub == "" {
		return nil, nil
	}
	var e models.UserIdentity
	err := s.db.GetContext(ctx, &e,
		`SELECT `+userIdentityColumns+` FROM user_identities WHERE oidc_sub = $1`, oidcSub)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (s *pgUserIdentityStore) FindByUser(ctx context.Context, userID string) ([]models.UserIdentity, error) {
	var out []models.UserIdentity
	err := s.db.SelectContext(ctx, &out,
		`SELECT `+userIdentityColumns+` FROM user_identities WHERE user_id = $1 ORDER BY created_at ASC`,
		userID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *pgUserIdentityStore) FindByUserAndSource(ctx context.Context, userID, source string) (*models.UserIdentity, error) {
	var e models.UserIdentity
	err := s.db.GetContext(ctx, &e,
		`SELECT `+userIdentityColumns+` FROM user_identities WHERE user_id = $1 AND source = $2 LIMIT 1`, userID, source)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *pgUserIdentityStore) Create(ctx context.Context, tx *sql.Tx, e *models.UserIdentity) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO user_identities (id, user_id, source, external_id, email, oidc_sub, metadata)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		e.ID, e.UserID, e.Source, e.ExternalID, nullableString(e.Email), nullableString(e.OIDCSub), nullableString(e.Metadata))
	return err
}

func (s *pgUserIdentityStore) Update(ctx context.Context, tx *sql.Tx, e *models.UserIdentity) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE user_identities
		    SET user_id = $1, source = $2, external_id = $3, email = $4, oidc_sub = $5, metadata = $6
		  WHERE id = $7`,
		e.UserID, e.Source, e.ExternalID, nullableString(e.Email), nullableString(e.OIDCSub), nullableString(e.Metadata), e.ID)
	return err
}

func (s *pgUserIdentityStore) ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE user_identities SET user_id = $1 WHERE user_id = $2`,
		toUserID, fromUserID)
	return err
}

func (s *pgUserIdentityStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM user_identities WHERE id = $1`, id)
	return err
}
