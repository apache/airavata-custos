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

	"github.com/apache/airavata-custos/allocations/domain/model"
	"github.com/jmoiron/sqlx"
)

type mariaDBPersonStore struct {
	db *sqlx.DB
}

func NewPersonStore(db *sqlx.DB) PersonStore {
	return &mariaDBPersonStore{db: db}
}

func (s *mariaDBPersonStore) FindByID(ctx context.Context, id string) (*model.Person, error) {
	var p model.Person
	err := s.db.GetContext(ctx, &p,
		`SELECT id, access_global_id, first_name, last_name, email, organization, org_code, nsf_status_code, created_at, updated_at
		 FROM persons WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *mariaDBPersonStore) FindByAccessGlobalID(ctx context.Context, globalID string) (*model.Person, error) {
	var p model.Person
	err := s.db.GetContext(ctx, &p,
		`SELECT id, access_global_id, first_name, last_name, email, organization, org_code, nsf_status_code, created_at, updated_at
		 FROM persons WHERE access_global_id = ?`, globalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *mariaDBPersonStore) Save(ctx context.Context, tx *sql.Tx, p *model.Person) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO persons (id, access_global_id, first_name, last_name, email, organization, org_code, nsf_status_code, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.AccessGlobalID, p.FirstName, p.LastName, p.Email,
		p.Organization, p.OrgCode, p.NsfStatusCode,
		p.CreatedAt, p.UpdatedAt)
	return err
}

func (s *mariaDBPersonStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM persons WHERE id = ?`, id)
	return err
}
