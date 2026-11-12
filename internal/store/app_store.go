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

const appColumns = `id, key, name, description, status, secret_hash, secret_hint, created_by, created_at, updated_at`

type pgAppStore struct {
	db *sqlx.DB
}

func NewAppStore(db *sqlx.DB) AppStore {
	return &pgAppStore{db: db}
}

func (s *pgAppStore) privileges(ctx context.Context, appID string) ([]models.PrivilegeKey, error) {
	privileges := []models.PrivilegeKey{}
	err := s.db.SelectContext(ctx, &privileges, `SELECT privilege FROM app_privileges WHERE app_id = $1 ORDER BY privilege`, appID)
	return privileges, err
}

func (s *pgAppStore) find(ctx context.Context, q sqlx.QueryerContext, where string, arg any) (*models.App, error) {
	var a models.App
	err := sqlx.GetContext(ctx, q, &a, `SELECT `+appColumns+` FROM apps WHERE `+where, arg)
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

func (s *pgAppStore) FindByID(ctx context.Context, id string) (*models.App, error) {
	return s.find(ctx, s.db, `id = $1`, id)
}

func (s *pgAppStore) FindByKey(ctx context.Context, key string) (*models.App, error) {
	return s.find(ctx, s.db, `key = $1`, key)
}

func (s *pgAppStore) FindByIDTx(ctx context.Context, tx *sql.Tx, id string) (*models.App, error) {
	return s.find(ctx, &sqlx.Tx{Tx: tx, Mapper: s.db.Mapper}, `id = $1 FOR UPDATE`, id)
}

func (s *pgAppStore) List(ctx context.Context) ([]models.App, error) {
	apps := []models.App{}
	if err := s.db.SelectContext(ctx, &apps, `SELECT `+appColumns+` FROM apps ORDER BY name`); err != nil {
		return nil, err
	}
	for i := range apps {
		privileges, err := s.privileges(ctx, apps[i].ID)
		if err != nil {
			return nil, err
		}
		apps[i].Privileges = privileges
	}
	return apps, nil
}

func (s *pgAppStore) Create(ctx context.Context, tx *sql.Tx, a *models.App) error {
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO apps (id, key, name, description, status, secret_hash, secret_hint, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		a.ID, a.Key, a.Name, a.Description, a.Status, a.SecretHash, a.SecretHint, a.CreatedBy); err != nil {
		return err
	}
	for _, privilege := range a.Privileges {
		if _, err := tx.ExecContext(ctx, `INSERT INTO app_privileges (app_id, privilege) VALUES ($1, $2)`, a.ID, privilege); err != nil {
			return err
		}
	}
	return nil
}

func (s *pgAppStore) Update(ctx context.Context, tx *sql.Tx, a *models.App) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE apps SET name = $1, description = $2, status = $3, secret_hash = $4, secret_hint = $5, updated_at = CURRENT_TIMESTAMP(6)
		 WHERE id = $6`,
		a.Name, a.Description, a.Status, a.SecretHash, a.SecretHint, a.ID)
	return err
}
