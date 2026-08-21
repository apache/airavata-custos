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

const auditEventColumns = "id, event_type, event_time, entity_id, entity_type, details, source, trace_id, span_id, parent_span_id"

type pgAuditEventStore struct {
	db *sqlx.DB
}

// NewAuditEventStore returns a PostgreSQL-backed AuditEventStore.
func NewAuditEventStore(db *sqlx.DB) AuditEventStore {
	return &pgAuditEventStore{db: db}
}

func (s *pgAuditEventStore) FindByID(ctx context.Context, id string) (*models.AuditEvent, error) {
	var e models.AuditEvent
	err := s.db.GetContext(ctx, &e,
		`SELECT `+auditEventColumns+` FROM audit_events WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (s *pgAuditEventStore) FindByEntity(ctx context.Context, entityID string) ([]models.AuditEvent, error) {
	var rows []models.AuditEvent
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+auditEventColumns+`
		 FROM audit_events
		 WHERE entity_id = $1
		 ORDER BY event_time`, entityID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgAuditEventStore) FindByEventType(ctx context.Context, eventType string) ([]models.AuditEvent, error) {
	var rows []models.AuditEvent
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+auditEventColumns+`
		 FROM audit_events
		 WHERE event_type = $1
		 ORDER BY event_time`, eventType)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgAuditEventStore) ListAll(ctx context.Context) ([]*models.AuditEvent, error) {
	var rows []*models.AuditEvent
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+auditEventColumns+`
		 FROM audit_events
		 ORDER BY event_time`)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgAuditEventStore) Create(ctx context.Context, tx *sql.Tx, e *models.AuditEvent) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO audit_events (id, event_type, event_time, entity_id, entity_type, details, source, trace_id, span_id, parent_span_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		e.ID, e.EventType, e.EventTime, e.EntityID, e.EntityType, e.Details, e.Source, e.TraceID, e.SpanID, e.ParentSpanID)
	return err
}

func (s *pgAuditEventStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM audit_events WHERE id = $1`, id)
	return err
}
