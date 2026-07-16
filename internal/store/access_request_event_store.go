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

const accessRequestEventColumns = "id, access_request_id, event_type, COALESCE(description, '') AS description, timestamp"

type mysqlAccessRequestEventStore struct {
	db *sqlx.DB
}

// NewAccessRequestEventStore returns a MySQL-backed AccessRequestEventStore.
func NewAccessRequestEventStore(db *sqlx.DB) AccessRequestEventStore {
	return &mysqlAccessRequestEventStore{db: db}
}

func (s *mysqlAccessRequestEventStore) FindByID(ctx context.Context, id string) (*models.AccessRequestEvent, error) {
	var e models.AccessRequestEvent
	err := s.db.GetContext(ctx, &e,
		`SELECT `+accessRequestEventColumns+` FROM access_request_events WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (s *mysqlAccessRequestEventStore) FindByRequest(ctx context.Context, accessRequestID string) ([]models.AccessRequestEvent, error) {
	var rows []models.AccessRequestEvent
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+accessRequestEventColumns+`
		 FROM access_request_events
		 WHERE access_request_id = ?
		 ORDER BY timestamp`, accessRequestID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlAccessRequestEventStore) Create(ctx context.Context, tx *sql.Tx, e *models.AccessRequestEvent) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO access_request_events
		     (id, access_request_id, event_type, description, timestamp)
		 VALUES (?, ?, ?, ?, ?)`,
		e.ID, e.AccessRequestID, e.EventType, e.Description, e.Timestamp)
	return err
}
