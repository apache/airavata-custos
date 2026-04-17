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

	"github.com/apache/airavata-custos/allocations/access-amie/model"
	"github.com/jmoiron/sqlx"
)

type ProjectStore interface {
	FindByID(ctx context.Context, id string) (*model.Project, error)
	Save(ctx context.Context, tx *sql.Tx, p *model.Project) error
	Update(ctx context.Context, tx *sql.Tx, p *model.Project) error
}

type mariaDBProjectStore struct {
	db *sqlx.DB
}

func NewProjectStore(db *sqlx.DB) ProjectStore {
	return &mariaDBProjectStore{db: db}
}

func (s *mariaDBProjectStore) FindByID(ctx context.Context, id string) (*model.Project, error) {
	var p model.Project
	err := s.db.GetContext(ctx, &p,
		`SELECT id, grant_number, is_active, created_at, updated_at
		 FROM projects WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *mariaDBProjectStore) Save(ctx context.Context, tx *sql.Tx, p *model.Project) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO projects (id, grant_number, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		p.ID, p.GrantNumber, p.IsActive, p.CreatedAt, p.UpdatedAt)
	return err
}

func (s *mariaDBProjectStore) Update(ctx context.Context, tx *sql.Tx, p *model.Project) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE projects SET grant_number = ?, is_active = ?, updated_at = ?
		 WHERE id = ?`,
		p.GrantNumber, p.IsActive, p.UpdatedAt, p.ID)
	return err
}
