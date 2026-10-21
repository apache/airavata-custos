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

const userAgreementAcceptanceColumns = `id, user_id, agreement_id, accepted_at, body_hash`

// pendingUserFilter matches users who still need to consent to an agreement. $1 is the agreement id.
const pendingUserFilter = `u.status IN ('PENDING', 'ACTIVE')
	AND NOT EXISTS (SELECT 1 FROM user_agreement_acceptances a WHERE a.user_id = u.id AND a.agreement_id = $1)`

type pgUserAgreementAcceptanceStore struct {
	db *sqlx.DB
}

func NewUserAgreementAcceptanceStore(db *sqlx.DB) UserAgreementAcceptanceStore {
	return &pgUserAgreementAcceptanceStore{db: db}
}

func (s *pgUserAgreementAcceptanceStore) Find(ctx context.Context, userID, agreementID string) (*models.UserAgreementAcceptance, error) {
	var a models.UserAgreementAcceptance
	err := s.db.GetContext(ctx, &a,
		`SELECT `+userAgreementAcceptanceColumns+` FROM user_agreement_acceptances WHERE user_id = $1 AND agreement_id = $2`,
		userID, agreementID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *pgUserAgreementAcceptanceStore) Create(ctx context.Context, tx *sql.Tx, a *models.UserAgreementAcceptance) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO user_agreement_acceptances (id, user_id, agreement_id, body_hash)
		 VALUES ($1, $2, $3, $4)`,
		a.ID, a.UserID, a.AgreementID, a.BodyHash)
	return err
}

func (s *pgUserAgreementAcceptanceStore) ListByAgreement(ctx context.Context, agreementID string, limit, offset int) ([]models.UserAgreementAcceptance, error) {
	var rows []models.UserAgreementAcceptance
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+userAgreementAcceptanceColumns+` FROM user_agreement_acceptances
		  WHERE agreement_id = $1 ORDER BY accepted_at DESC LIMIT $2 OFFSET $3`,
		agreementID, limit, offset)
	return rows, err
}

func (s *pgUserAgreementAcceptanceStore) ListPendingUsers(ctx context.Context, agreementID string, limit, offset int) ([]models.User, error) {
	var rows []models.User
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+prefixed("u", userColumns)+` FROM users u
		  WHERE `+pendingUserFilter+` ORDER BY u.email LIMIT $2 OFFSET $3`,
		agreementID, limit, offset)
	return rows, err
}
