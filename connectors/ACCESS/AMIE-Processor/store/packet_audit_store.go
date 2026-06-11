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
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
)

// PacketAuditStore returns every audit_events row written for one AMIE packet,
// joined through amie_audit_extras (which carries packet_id / event_id).
type PacketAuditStore interface {
	ListAuditsForPacket(ctx context.Context, packetID string) ([]models.TraceEvent, error)
}

type mysqlPacketAuditStore struct {
	db *sqlx.DB
}

func NewPacketAuditStore(db *sqlx.DB) PacketAuditStore {
	return &mysqlPacketAuditStore{db: db}
}

const packetAuditSelect = `
SELECT
    ae.span_id,
    ae.parent_span_id,
    ae.source,
    ae.event_type,
    ae.entity_type,
    ae.entity_id,
    ae.details   AS description,
    ae.event_time AS created_at
FROM audit_events ae
JOIN amie_audit_extras x ON x.audit_event_id = ae.id
WHERE x.packet_id = ?
ORDER BY ae.event_time ASC, ae.span_id ASC
`

func (s *mysqlPacketAuditStore) ListAuditsForPacket(ctx context.Context, packetID string) ([]models.TraceEvent, error) {
	type row struct {
		SpanID       string    `db:"span_id"`
		ParentSpanID string    `db:"parent_span_id"`
		Source       string    `db:"source"`
		EventType    string    `db:"event_type"`
		EntityType   string    `db:"entity_type"`
		EntityID     string    `db:"entity_id"`
		Description  string    `db:"description"`
		CreatedAt    time.Time `db:"created_at"`
	}
	var rows []row
	if err := s.db.SelectContext(ctx, &rows, packetAuditSelect, packetID); err != nil {
		return nil, fmt.Errorf("packet_audit_store: list: %w", err)
	}
	out := make([]models.TraceEvent, len(rows))
	for i, r := range rows {
		out[i] = models.TraceEvent{
			SpanID:       r.SpanID,
			ParentSpanID: r.ParentSpanID,
			Source:       r.Source,
			EventType:    r.EventType,
			EntityType:   r.EntityType,
			EntityID:     r.EntityID,
			Description:  r.Description,
			Status:       tracing.EventStatus(r.EventType),
			CreatedAt:    r.CreatedAt,
		}
	}
	return out, nil
}
