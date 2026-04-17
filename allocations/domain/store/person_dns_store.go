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

	"github.com/apache/airavata-custos/allocations/domain/model"
	"github.com/jmoiron/sqlx"
)

type mariaDBPersonDNStore struct {
	db *sqlx.DB
}

func NewPersonDNStore(db *sqlx.DB) PersonDNStore {
	return &mariaDBPersonDNStore{db: db}
}

func (s *mariaDBPersonDNStore) ExistsByPersonAndDN(ctx context.Context, personID, dn string) (bool, error) {
	var count int
	err := s.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM person_dns WHERE person_id = ? AND dn = ?`, personID, dn)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *mariaDBPersonDNStore) Save(ctx context.Context, tx *sql.Tx, d *model.PersonDN) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO person_dns (person_id, dn) VALUES (?, ?)`,
		d.PersonID, d.DN)
	return err
}

func (s *mariaDBPersonDNStore) DeleteByPersonID(ctx context.Context, tx *sql.Tx, personID string) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM person_dns WHERE person_id = ?`, personID)
	return err
}

func (s *mariaDBPersonDNStore) DeleteByPersonIDNotIn(ctx context.Context, tx *sql.Tx, personID string, dnsToKeep []string) error {
	if len(dnsToKeep) == 0 {
		return s.DeleteByPersonID(ctx, tx, personID)
	}
	query, args, err := sqlx.In(
		`DELETE FROM person_dns WHERE person_id = ? AND dn NOT IN (?)`, personID, dnsToKeep)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, query, args...)
	return err
}

func (s *mariaDBPersonDNStore) FindByPersonID(ctx context.Context, personID string) ([]model.PersonDN, error) {
	var results []model.PersonDN
	err := s.db.SelectContext(ctx, &results,
		`SELECT id, person_id, dn FROM person_dns WHERE person_id = ?`, personID)
	if err != nil {
		return nil, err
	}
	return results, nil
}
