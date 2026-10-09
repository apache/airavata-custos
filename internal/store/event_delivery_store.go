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

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

// Delivery columns with the event table joined in.
const eventDeliveryWithEventColumns = `d.id, d.event_id, d.subscriber, d.status, d.attempts, d.next_run_at, d.last_error, d.created_at, d.finished_at,
	e.id AS "event.id", e.event_type AS "event.event_type", e.payload AS "event.payload", e.source AS "event.source", e.trace_id AS "event.trace_id", e.span_id AS "event.span_id", e.created_at AS "event.created_at"`

type pgEventDeliveryStore struct {
	db *sqlx.DB
}

// NewEventDeliveryStore returns a PostgreSQL-backed EventDeliveryStore.
func NewEventDeliveryStore(db *sqlx.DB) EventDeliveryStore {
	return &pgEventDeliveryStore{db: db}
}

func (s *pgEventDeliveryStore) SaveSubscription(ctx context.Context, subscriber, eventType string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO event_subscriptions (subscriber, event_type)
		 VALUES ($1, $2)
		 ON CONFLICT (subscriber, event_type) DO NOTHING`,
		subscriber, eventType)
	return err
}

func (s *pgEventDeliveryStore) ListSubscriptions(ctx context.Context) ([]models.EventSubscription, error) {
	var rows []models.EventSubscription
	err := s.db.SelectContext(ctx, &rows,
		`SELECT subscriber, event_type, created_at
		 FROM event_subscriptions
		 ORDER BY subscriber, event_type`)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgEventDeliveryStore) DeleteSubscriptions(ctx context.Context, subscriber string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM event_subscriptions WHERE subscriber = $1`, subscriber)
	return err
}

func (s *pgEventDeliveryStore) DeleteSubscription(ctx context.Context, subscriber, eventType string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM event_subscriptions WHERE subscriber = $1 AND event_type = $2`,
		subscriber, eventType)
	return err
}

func (s *pgEventDeliveryStore) CreateEvent(ctx context.Context, tx *sql.Tx, e *models.Event) error {
	// Send the raw JSON as text and let Postgres cast it to jsonb.
	_, err := tx.ExecContext(ctx,
		`INSERT INTO events (id, event_type, payload, source, trace_id, span_id, created_at)
		 VALUES ($1, $2, $3::jsonb, $4, $5, $6, $7)`,
		e.ID, e.EventType, string(e.Payload), e.Source, e.TraceID, e.SpanID, e.CreatedAt)
	return err
}

func (s *pgEventDeliveryStore) CreateDelivery(ctx context.Context, tx *sql.Tx, d *models.EventDelivery) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO event_deliveries (id, event_id, subscriber, status, attempts, next_run_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		d.ID, d.EventID, d.Subscriber, d.Status, d.Attempts, d.NextRunAt, d.CreatedAt)
	return err
}

func (s *pgEventDeliveryStore) FindDeliveryByID(ctx context.Context, id string) (*models.PendingDelivery, error) {
	var d models.PendingDelivery
	err := s.db.GetContext(ctx, &d,
		`SELECT `+eventDeliveryWithEventColumns+`
		 FROM event_deliveries d JOIN events e ON e.id = d.event_id
		 WHERE d.id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (s *pgEventDeliveryStore) FindDueDeliveries(ctx context.Context, now time.Time, subscribers []string, limit int) ([]models.PendingDelivery, error) {
	var rows []models.PendingDelivery
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+eventDeliveryWithEventColumns+`
		 FROM event_deliveries d JOIN events e ON e.id = d.event_id
		 WHERE d.status = $1 AND d.next_run_at <= $2 AND d.subscriber = ANY($3)
		 ORDER BY d.created_at
		 LIMIT $4`,
		models.EventDeliveryPending, now, subscribers, limit)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgEventDeliveryStore) ListDeliveries(ctx context.Context, status models.EventDeliveryStatus, limit int) ([]models.PendingDelivery, error) {
	var rows []models.PendingDelivery
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+eventDeliveryWithEventColumns+`
		 FROM event_deliveries d JOIN events e ON e.id = d.event_id
		 WHERE $1 = '' OR d.status = $1
		 ORDER BY d.created_at DESC
		 LIMIT $2`,
		status, limit)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgEventDeliveryStore) MarkDeliverySucceeded(ctx context.Context, tx *sql.Tx, id string, attempts int, finishedAt time.Time) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE event_deliveries
		 SET status = $1, attempts = $2, finished_at = $3
		 WHERE id = $4`,
		models.EventDeliverySucceeded, attempts, finishedAt, id)
	return err
}

func (s *pgEventDeliveryStore) MarkDeliveryRetry(ctx context.Context, tx *sql.Tx, id string, attempts int, nextRunAt time.Time, lastError string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE event_deliveries
		 SET attempts = $1, next_run_at = $2, last_error = $3
		 WHERE id = $4`,
		attempts, nextRunAt, lastError, id)
	return err
}

func (s *pgEventDeliveryStore) MarkDeliveryFailed(ctx context.Context, tx *sql.Tx, id string, attempts int, lastError string, finishedAt time.Time) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE event_deliveries
		 SET status = $1, attempts = $2, last_error = $3, finished_at = $4
		 WHERE id = $5`,
		models.EventDeliveryFailed, attempts, lastError, finishedAt, id)
	return err
}

func (s *pgEventDeliveryStore) ResetDeliveryToPending(ctx context.Context, tx *sql.Tx, id string, nextRunAt time.Time) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE event_deliveries
		 SET status = $1, attempts = 0, next_run_at = $2, finished_at = NULL
		 WHERE id = $3`,
		models.EventDeliveryPending, nextRunAt, id)
	return err
}
