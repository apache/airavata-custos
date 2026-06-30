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

type mysqlResearcherProfileStore struct {
	db *sqlx.DB
}

// NewResearcherProfileStore returns a MySQL-backed ResearcherProfileStore.
func NewResearcherProfileStore(db *sqlx.DB) ResearcherProfileStore {
	return &mysqlResearcherProfileStore{db: db}
}

const researcherProfileColumns = `user_id, display_name, COALESCE(research_domain, '') AS research_domain, COALESCE(department, '') AS department, COALESCE(institution, '') AS institution, created_at, updated_at`

func (s *mysqlResearcherProfileStore) FindByUserID(ctx context.Context, userID string) (*models.ResearcherProfile, error) {
	var p models.ResearcherProfile
	err := s.db.GetContext(ctx, &p,
		`SELECT `+researcherProfileColumns+` FROM researcher_profiles WHERE user_id = ?`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *mysqlResearcherProfileStore) Create(ctx context.Context, tx *sql.Tx, p *models.ResearcherProfile) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO researcher_profiles (user_id, display_name, research_domain, department, institution)
	 VALUES (?, ?, ?, ?, ?)`,
		p.UserID, p.DisplayName, nullableString(p.ResearchDomain), nullableString(p.Department), nullableString(p.Institution))
	return err
}

func (s *mysqlResearcherProfileStore) Update(ctx context.Context, tx *sql.Tx, p *models.ResearcherProfile) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE researcher_profiles
		SET display_name = ?, research_domain = ?, department = ?, institution = ?
	  WHERE user_id = ?`,
		p.DisplayName, nullableString(p.ResearchDomain), nullableString(p.Department), nullableString(p.Institution), p.UserID)
	return err
}

func (s *mysqlResearcherProfileStore) Delete(ctx context.Context, tx *sql.Tx, userID string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM researcher_profiles WHERE user_id = ?`, userID)
	return err
}
