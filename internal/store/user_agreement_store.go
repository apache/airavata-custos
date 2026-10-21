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

const userAgreementColumns = `id, title, version_label, summary, body_md, body_hash, status, change_note, created_by, created_at, published_by, published_at, updated_at, removed_by, removed_at`

const currentUserAgreementID = `SELECT id FROM user_agreements WHERE status = 'PUBLISHED' ORDER BY published_at DESC LIMIT 1`

type pgUserAgreementStore struct {
	db *sqlx.DB
}

func NewUserAgreementStore(db *sqlx.DB) UserAgreementStore {
	return &pgUserAgreementStore{db: db}
}

func (s *pgUserAgreementStore) FindByID(ctx context.Context, id string) (*models.UserAgreement, error) {
	var a models.UserAgreement
	err := s.db.GetContext(ctx, &a, `SELECT `+userAgreementColumns+` FROM user_agreements WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *pgUserAgreementStore) FindByIDTx(ctx context.Context, tx *sql.Tx, id string) (*models.UserAgreement, error) {
	var a models.UserAgreement
	err := (&sqlx.Tx{Tx: tx, Mapper: s.db.Mapper}).GetContext(ctx, &a,
		`SELECT `+userAgreementColumns+` FROM user_agreements WHERE id = $1 FOR UPDATE`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *pgUserAgreementStore) FindCurrent(ctx context.Context) (*models.UserAgreement, error) {
	var a models.UserAgreement
	err := s.db.GetContext(ctx, &a,
		`SELECT `+userAgreementColumns+` FROM user_agreements WHERE id = (`+currentUserAgreementID+`)`)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *pgUserAgreementStore) List(ctx context.Context) ([]models.UserAgreement, error) {
	var rows []models.UserAgreement
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+userAgreementColumns+` FROM user_agreements ORDER BY created_at DESC`)
	return rows, err
}

func (s *pgUserAgreementStore) ListPendingForUser(ctx context.Context, userID string) ([]models.UserAgreement, error) {
	var rows []models.UserAgreement
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+userAgreementColumns+` FROM user_agreements ua
		  WHERE ua.id = (`+currentUserAgreementID+`)
		    AND NOT EXISTS (SELECT 1 FROM user_agreement_acceptances a WHERE a.agreement_id = ua.id AND a.user_id = $1)`, userID)
	return rows, err
}

func (s *pgUserAgreementStore) Create(ctx context.Context, tx *sql.Tx, a *models.UserAgreement) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO user_agreements (id, title, version_label, summary, body_md, body_hash, status, change_note, created_by, published_by, published_at, removed_by, removed_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		a.ID, a.Title, a.VersionLabel, a.Summary, a.BodyMD, a.BodyHash, a.Status, a.ChangeNote,
		a.CreatedBy, a.PublishedBy, a.PublishedAt, a.RemovedBy, a.RemovedAt)
	return err
}

func (s *pgUserAgreementStore) Update(ctx context.Context, tx *sql.Tx, a *models.UserAgreement) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE user_agreements
		    SET title = $1, version_label = $2, summary = $3, body_md = $4, body_hash = $5, status = $6, change_note = $7,
		        published_by = $8, published_at = $9, updated_at = CURRENT_TIMESTAMP(6), removed_by = $10, removed_at = $11
		  WHERE id = $12`,
		a.Title, a.VersionLabel, a.Summary, a.BodyMD, a.BodyHash, a.Status, a.ChangeNote,
		a.PublishedBy, a.PublishedAt, a.RemovedBy, a.RemovedAt, a.ID)
	return err
}

func (s *pgUserAgreementStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM user_agreements WHERE id = $1`, id)
	return err
}

func (s *pgUserAgreementStore) CountAcceptances(ctx context.Context, agreementID string) (int, error) {
	var n int
	err := s.db.GetContext(ctx, &n,
		`SELECT COUNT(*) FROM user_agreement_acceptances WHERE agreement_id = $1`, agreementID)
	return n, err
}

func (s *pgUserAgreementStore) CountPendingUsers(ctx context.Context, agreementID string) (int, error) {
	var n int
	err := s.db.GetContext(ctx, &n,
		`SELECT COUNT(*) FROM users u
		  WHERE `+pendingUserFilter, agreementID)
	return n, err
}
