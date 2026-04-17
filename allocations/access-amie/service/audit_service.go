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

	"github.com/apache/airavata-custos/allocations/access-amie/model"
)

type auditStore interface {
	Save(ctx context.Context, tx *sql.Tx, a *model.AuditLog) error
}

type AuditService struct {
	audits auditStore
}

func NewAuditService(audits auditStore) *AuditService {
	return &AuditService{audits: audits}
}

// Log records an audit entry for the given packet, event, and action.
func (s *AuditService) Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error {
	entry := &model.AuditLog{
		PacketID:   packetID,
		EventID:    ptrOrNil(eventID),
		Action:     action,
		EntityType: entityType,
		EntityID:   ptrOrNil(entityID),
		Summary:    ptrOrNil(summary),
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.audits.Save(ctx, tx, entry); err != nil {
		return fmt.Errorf("audit_service: saving audit log for packet %s: %w", packetID, err)
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
