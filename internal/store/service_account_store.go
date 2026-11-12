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

const serviceAccountColumns = `id, key, name, description, status, secret_hash, secret_hint, created_by, created_at, updated_at`

type pgServiceAccountStore struct {
	db *sqlx.DB
}

func NewServiceAccountStore(db *sqlx.DB) ServiceAccountStore {
	return &pgServiceAccountStore{db: db}
}

func (s *pgServiceAccountStore) privileges(ctx context.Context, accountID string) ([]models.PrivilegeKey, error) {
	privileges := []models.PrivilegeKey{}
	err := s.db.SelectContext(ctx, &privileges, `SELECT privilege FROM service_account_privileges WHERE service_account_id = $1 ORDER BY privilege`, accountID)
	return privileges, err
}

func (s *pgServiceAccountStore) find(ctx context.Context, q sqlx.QueryerContext, where string, arg any) (*models.ServiceAccount, error) {
	var a models.ServiceAccount
	err := sqlx.GetContext(ctx, q, &a, `SELECT `+serviceAccountColumns+` FROM service_accounts WHERE `+where, arg)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if a.Privileges, err = s.privileges(ctx, a.ID); err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *pgServiceAccountStore) FindByID(ctx context.Context, id string) (*models.ServiceAccount, error) {
	return s.find(ctx, s.db, `id = $1`, id)
}

func (s *pgServiceAccountStore) FindByKey(ctx context.Context, key string) (*models.ServiceAccount, error) {
	return s.find(ctx, s.db, `key = $1`, key)
}

func (s *pgServiceAccountStore) FindByIDTx(ctx context.Context, tx *sql.Tx, id string) (*models.ServiceAccount, error) {
	return s.find(ctx, &sqlx.Tx{Tx: tx, Mapper: s.db.Mapper}, `id = $1 FOR UPDATE`, id)
}

func (s *pgServiceAccountStore) List(ctx context.Context) ([]models.ServiceAccount, error) {
	accounts := []models.ServiceAccount{}
	if err := s.db.SelectContext(ctx, &accounts, `SELECT `+serviceAccountColumns+` FROM service_accounts ORDER BY name`); err != nil {
		return nil, err
	}
	for i := range accounts {
		privileges, err := s.privileges(ctx, accounts[i].ID)
		if err != nil {
			return nil, err
		}
		accounts[i].Privileges = privileges
	}
	return accounts, nil
}

func (s *pgServiceAccountStore) Create(ctx context.Context, tx *sql.Tx, a *models.ServiceAccount) error {
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO service_accounts (id, key, name, description, status, secret_hash, secret_hint, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		a.ID, a.Key, a.Name, a.Description, a.Status, a.SecretHash, a.SecretHint, a.CreatedBy); err != nil {
		return err
	}
	for _, privilege := range a.Privileges {
		if _, err := tx.ExecContext(ctx, `INSERT INTO service_account_privileges (service_account_id, privilege) VALUES ($1, $2)`, a.ID, privilege); err != nil {
			return err
		}
	}
	return nil
}

func (s *pgServiceAccountStore) Update(ctx context.Context, tx *sql.Tx, a *models.ServiceAccount) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE service_accounts SET name = $1, description = $2, status = $3, secret_hash = $4, secret_hint = $5, updated_at = CURRENT_TIMESTAMP(6)
		 WHERE id = $6`,
		a.Name, a.Description, a.Status, a.SecretHash, a.SecretHint, a.ID)
	return err
}
