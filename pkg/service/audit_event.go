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

package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/apache/airavata-custos/internal/audit"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
)

// SourceCore is the default value stamped on audit rows when nothing
// else has tagged the ctx direct service writes, admin endpoints, etc.
const SourceCore = "core"

// CreateAuditEvent records a new audit event. EventTime defaults to the
// server's current UTC time when unset. Audit events are append-only — there
// is no update operation.
func (s *Service) CreateAuditEvent(ctx context.Context, e *models.AuditEvent) (*models.AuditEvent, error) {
	if e == nil {
		return nil, fmt.Errorf("%w: audit event is nil", ErrInvalidInput)
	}
	if e.EventType == "" {
		return nil, fmt.Errorf("%w: event_type is required", ErrInvalidInput)
	}
	if e.EntityID == "" {
		return nil, fmt.Errorf("%w: entity_id is required", ErrInvalidInput)
	}

	if e.ID == "" {
		e.ID = newID()
	}
	if e.EventTime.IsZero() {
		e.EventTime = nowUTC()
	}

	tracing.PopulateAuditIDs(ctx, &e.TraceID, &e.SpanID, &e.ParentSpanID)
	if e.TraceID == nil {
		slog.WarnContext(ctx, "audit write outside an active span",
			"event_type", e.EventType,
			"entity_id", e.EntityID,
		)
	}
	if e.Source == "" {
		if src := audit.SourceFromContext(ctx); src != "" {
			e.Source = src
		} else {
			e.Source = SourceCore
		}
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.auditEvents.Create(ctx, tx, e)
	}); err != nil {
		return nil, fmt.Errorf("create audit event: %w", err)
	}
	return e, nil
}

// GetAuditEvent retrieves an audit event by its ID. Returns ErrNotFound
// when no audit event matches.
func (s *Service) GetAuditEvent(ctx context.Context, id string) (*models.AuditEvent, error) {
	e, err := s.auditEvents.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get audit event: %w", err)
	}
	if e == nil {
		return nil, ErrNotFound
	}
	return e, nil
}

// ListAuditEventsByEntity returns every audit event recorded against the
// given entity, ordered by event_time ascending.
func (s *Service) ListAuditEventsByEntity(ctx context.Context, entityID string) ([]models.AuditEvent, error) {
	if entityID == "" {
		return nil, fmt.Errorf("%w: entity_id is required", ErrInvalidInput)
	}
	rows, err := s.auditEvents.FindByEntity(ctx, entityID)
	if err != nil {
		return nil, fmt.Errorf("list audit events by entity: %w", err)
	}
	return rows, nil
}

// ListAuditEventsByEventType returns every audit event of the given type,
// ordered by event_time ascending.
func (s *Service) ListAuditEventsByEventType(ctx context.Context, eventType string) ([]models.AuditEvent, error) {
	if eventType == "" {
		return nil, fmt.Errorf("%w: event_type is required", ErrInvalidInput)
	}
	rows, err := s.auditEvents.FindByEventType(ctx, eventType)
	if err != nil {
		return nil, fmt.Errorf("list audit events by event type: %w", err)
	}
	return rows, nil
}

// ListAllAuditEvents returns every audit event ordered by event_time ascending.
func (s *Service) ListAllAuditEvents(ctx context.Context) ([]*models.AuditEvent, error) {
	rows, err := s.auditEvents.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all audit events: %w", err)
	}
	return rows, nil
}

// DeleteAuditEvent removes an audit event by ID.
func (s *Service) DeleteAuditEvent(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: audit event id is required", ErrInvalidInput)
	}
	existing, err := s.auditEvents.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup audit event: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.auditEvents.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete audit event: %w", err)
	}
	return nil
}
