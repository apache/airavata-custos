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

type mysqlOrganizationStore struct {
	db *sqlx.DB
}

// NewOrganizationStore returns a MySQL-backed OrganizationStore.
func NewOrganizationStore(db *sqlx.DB) OrganizationStore {
	return &mysqlOrganizationStore{db: db}
}

func (s *mysqlOrganizationStore) FindByID(ctx context.Context, id string) (*models.Organization, error) {
	var o models.Organization
	err := s.db.GetContext(ctx, &o,
		`SELECT id, originated_id, name FROM organizations WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (s *mysqlOrganizationStore) FindByOriginatedID(ctx context.Context, originatedID string) (*models.Organization, error) {
	var o models.Organization
	err := s.db.GetContext(ctx, &o,
		`SELECT id, originated_id, name FROM organizations WHERE originated_id = ?`, originatedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (s *mysqlOrganizationStore) List(ctx context.Context, limit, offset int) ([]models.Organization, int, error) {
	var total int
	if err := s.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM organizations`); err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	var rows []models.Organization
	if err := s.db.SelectContext(ctx, &rows,
		`SELECT id, originated_id, name FROM organizations ORDER BY name LIMIT ? OFFSET ?`, limit, offset); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (s *mysqlOrganizationStore) Create(ctx context.Context, tx *sql.Tx, o *models.Organization) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO organizations (id, originated_id, name) VALUES (?, ?, ?)`,
		o.ID, o.OriginatedID, o.Name)
	return err
}

func (s *mysqlOrganizationStore) Update(ctx context.Context, tx *sql.Tx, o *models.Organization) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE organizations SET originated_id = ?, name = ? WHERE id = ?`,
		o.OriginatedID, o.Name, o.ID)
	return err
}

func (s *mysqlOrganizationStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM organizations WHERE id = ?`, id)
	return err
}
