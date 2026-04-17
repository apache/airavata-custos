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
	"time"

	"github.com/apache/airavata-custos/allocations/access-amie/model"
	"github.com/jmoiron/sqlx"
)

type EventStore interface {
	FindByID(ctx context.Context, id string) (*model.ProcessingEvent, error)
	FindTop50EventsToProcess(ctx context.Context, statuses []model.ProcessingStatus, now time.Time) ([]model.EventWithPacket, error)
	Save(ctx context.Context, tx *sql.Tx, e *model.ProcessingEvent) error
	Update(ctx context.Context, tx *sql.Tx, e *model.ProcessingEvent) error
}

type mariaDBEventStore struct {
	db *sqlx.DB
}

func NewEventStore(db *sqlx.DB) EventStore {
	return &mariaDBEventStore{db: db}
}

func (s *mariaDBEventStore) FindByID(ctx context.Context, id string) (*model.ProcessingEvent, error) {
	var e model.ProcessingEvent
	err := s.db.GetContext(ctx, &e,
		`SELECT id, packet_id, type, status, attempts, payload, created_at, started_at, finished_at, last_error, next_retry_at
		 FROM amie_processing_events WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (s *mariaDBEventStore) FindTop50EventsToProcess(ctx context.Context, statuses []model.ProcessingStatus, now time.Time) ([]model.EventWithPacket, error) {
	query := `SELECT e.id, e.packet_id, e.type, e.status, e.attempts, e.payload, e.created_at, e.started_at, e.finished_at, e.last_error, e.next_retry_at,
	                  p.amie_id AS packet_amie_id, p.type AS packet_type, p.raw_json AS packet_raw_json
	           FROM amie_processing_events e
	           JOIN amie_packets p ON e.packet_id = p.id
	           WHERE e.status IN (?)
	             AND (e.next_retry_at IS NULL OR e.next_retry_at <= ?)
	           ORDER BY e.created_at ASC
	           LIMIT 50`

	// Expand the IN clause for the status slice.
	query, args, err := sqlx.In(query, statuses, now)
	if err != nil {
		return nil, err
	}
	// Rebind for the MySQL/MariaDB driver.
	query = s.db.Rebind(query)

	var results []model.EventWithPacket
	err = s.db.SelectContext(ctx, &results, query, args...)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *mariaDBEventStore) Save(ctx context.Context, tx *sql.Tx, e *model.ProcessingEvent) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO amie_processing_events (id, packet_id, type, status, attempts, payload, created_at, started_at, finished_at, last_error, next_retry_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.PacketID, e.Type, e.Status, e.Attempts, e.Payload,
		e.CreatedAt, e.StartedAt, e.FinishedAt,
		e.LastError, e.NextRetryAt)
	return err
}

func (s *mariaDBEventStore) Update(ctx context.Context, tx *sql.Tx, e *model.ProcessingEvent) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE amie_processing_events SET status = ?, attempts = ?, payload = ?, started_at = ?, finished_at = ?, last_error = ?, next_retry_at = ?
		 WHERE id = ?`,
		e.Status, e.Attempts, e.Payload,
		e.StartedAt, e.FinishedAt,
		e.LastError, e.NextRetryAt, e.ID)
	return err
}
