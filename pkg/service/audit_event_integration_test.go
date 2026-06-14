//go:build integration

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
	"testing"

	"github.com/apache/airavata-custos/pkg/models"
)

func TestCreateAuditEvent_EntityTypeRoundTrip(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)

	created, err := svc.CreateAuditEvent(ctx(), &models.AuditEvent{
		EventType:  "TEST_EVENT",
		EntityID:   "ent-1",
		EntityType: "test_entity",
		Details:    "{}",
	})
	if err != nil {
		t.Fatalf("CreateAuditEvent: %v", err)
	}

	got, err := svc.GetAuditEvent(ctx(), created.ID)
	if err != nil {
		t.Fatalf("GetAuditEvent: %v", err)
	}
	if got.EntityType != "test_entity" {
		t.Errorf("entity_type = %q, want %q", got.EntityType, "test_entity")
	}
	if got.EntityID != "ent-1" {
		t.Errorf("entity_id = %q, want %q", got.EntityID, "ent-1")
	}
}

func TestCreateAuditEvent_EmptyEntityTypeAllowed(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)

	// Callers that haven't been backfilled yet must still succeed; the column
	// defaults to '' at the DB layer.
	created, err := svc.CreateAuditEvent(ctx(), &models.AuditEvent{
		EventType: "TEST_EVENT_NO_TYPE",
		EntityID:  "ent-2",
		Details:   "{}",
	})
	if err != nil {
		t.Fatalf("CreateAuditEvent: %v", err)
	}

	got, err := svc.GetAuditEvent(ctx(), created.ID)
	if err != nil {
		t.Fatalf("GetAuditEvent: %v", err)
	}
	if got.EntityType != "" {
		t.Errorf("entity_type = %q, want empty", got.EntityType)
	}
}
