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

type mariaDBPersonGlobalIDStore struct {
	db *sqlx.DB
}

func NewPersonGlobalIDStore(db *sqlx.DB) PersonGlobalIDStore {
	return &mariaDBPersonGlobalIDStore{db: db}
}

func (s *mariaDBPersonGlobalIDStore) FindPersonByGlobalID(ctx context.Context, globalID string) (*model.Person, error) {
	var p model.Person
	err := s.db.GetContext(ctx, &p,
		`SELECT p.id, p.access_global_id, p.first_name, p.last_name, p.email,
            p.organization, p.org_code, p.nsf_status_code, p.is_active, p.created_at, p.updated_at
     FROM persons p
     JOIN person_global_ids g ON p.id = g.person_id
     WHERE g.global_id = ?`, globalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *mariaDBPersonGlobalIDStore) Save(ctx context.Context, tx *sql.Tx, g *model.PersonGlobalID) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO person_global_ids (person_id, global_id) VALUES (?, ?)`,
		g.PersonID, g.GlobalID)
	return err
}

func (s *mariaDBPersonGlobalIDStore) UpdatePersonID(ctx context.Context, tx *sql.Tx, oldPersonID, newPersonID string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE person_global_ids SET person_id = ? WHERE person_id = ?`,
		newPersonID, oldPersonID)
	return err
}
