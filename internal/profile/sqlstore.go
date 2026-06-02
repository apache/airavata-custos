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

package profile

import (
	"database/sql"
	"errors"
	"time"

	"github.com/apache/airavata-custos/pkg/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// SQLStore is a lightweight SQL-backed implementation of the profile store.
// This scaffold persists the stable profile and links external identities while
// leaving richer attribute history and authorization inputs for follow-up work.
type SQLStore struct {
	db *sqlx.DB
}

func NewSQLStore(db *sqlx.DB) *SQLStore {
	return &SQLStore{db: db}
}

func (s *SQLStore) GetByID(id string) (*models.Profile, error) {
	p := &models.Profile{}
	err := s.db.Get(p, `SELECT id, display_name, created_at, updated_at FROM profiles WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var emails []string
	err = s.db.Select(&emails, `SELECT email FROM profile_emails WHERE profile_id = ?`, id)
	if err != nil {
		return nil, err
	}
	p.Emails = emails

	return p, nil
}

func (s *SQLStore) GetByEmail(email string) (*models.Profile, error) {
	var id string
	err := s.db.Get(&id, `SELECT profile_id FROM profile_emails WHERE email = ?`, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return s.GetByID(id)
}

func (s *SQLStore) Create(p *models.Profile) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err = tx.Exec(`INSERT INTO profiles (id, display_name, created_at, updated_at) VALUES (?, ?, ?, ?)`, p.ID, p.DisplayName, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, e := range p.Emails {
		_, err = tx.Exec(`INSERT INTO profile_emails (profile_id, email) VALUES (?, ?)`, p.ID, e)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLStore) Update(p *models.Profile) error {
	p.UpdatedAt = time.Now().UTC()
	_, err := s.db.Exec(`UPDATE profiles SET display_name = ?, updated_at = ? WHERE id = ?`, p.DisplayName, p.UpdatedAt, p.ID)
	return err
}

func (s *SQLStore) LinkExternalID(profileID, provider, externalID string) error {
	_, err := s.db.Exec(`INSERT INTO external_identities (provider, external_id, profile_id, linked_at) VALUES (?, ?, ?, ?)`, provider, externalID, profileID, time.Now().UTC())
	return err
}
