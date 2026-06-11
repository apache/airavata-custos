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
	"time"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
)

// auditSource is marked on every audit_events row produced by this connector.
const auditSource = "amie"

// AuditService writes one audit_events row plus the matching amie_audit_extras
// row for each handler action. Both writes happen inside the caller's
// transaction so they commit or roll back together.
type AuditService struct {
	coreEvents store.AuditEventStore
	extras     extrasStore
}

type extrasStore interface {
	Save(ctx context.Context, tx *sql.Tx, e *model.AmieAuditExtras) error
}

func NewAuditService(coreEvents store.AuditEventStore, extras extrasStore) *AuditService {
	return &AuditService{coreEvents: coreEvents, extras: extras}
}

// Log records one audit row for the given packet/event/action. The audit_events
// row carries the trace/span IDs from ctx; the amie_audit_extras row links
// back to it and preserves the AMIE-specific (packet_id, event_id) references.
func (s *AuditService) Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error {
	if packetID == "" {
		return fmt.Errorf("audit_service: packet_id is required")
	}

	auditEventID := uuid.NewString()
	event := &models.AuditEvent{
		ID:         auditEventID,
		EventType:  string(action),
		EventTime:  time.Now().UTC(),
		EntityID:   entityID,
		EntityType: entityType,
		Details:    summary,
		Source:     auditSource,
	}
	tracing.PopulateAuditIDs(ctx, &event.TraceID, &event.SpanID, &event.ParentSpanID)
	if event.TraceID == "" {
		slog.WarnContext(ctx, "audit write outside an active span",
			"packet_id", packetID,
			"event_id", eventID,
			"action", string(action),
			"entity_type", entityType,
			"entity_id", entityID,
		)
	}

	if err := s.coreEvents.Create(ctx, tx, event); err != nil {
		return fmt.Errorf("audit_service: saving audit_events row for packet %s: %w", packetID, err)
	}

	extras := &model.AmieAuditExtras{
		AuditEventID: auditEventID,
		PacketID:     packetID,
		EventID:      ptrOrNil(eventID),
	}
	if err := s.extras.Save(ctx, tx, extras); err != nil {
		return fmt.Errorf("audit_service: saving amie_audit_extras for packet %s: %w", packetID, err)
	}

	slog.DebugContext(ctx, "audit log recorded",
		"packet_id", packetID,
		"event_id", eventID,
		"action", string(action),
		"entity_type", entityType,
		"entity_id", entityID,
	)
	return nil
}

// ptrOrNil returns a pointer to s if it is non-empty, or nil otherwise.
func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
