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

package store

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/apache/airavata-custos/pkg/models"
)

// Make sure a status filter narrows the trace set before counting and paging,
// and that null parent span ids read back as roots.
func TestListTracesStatusFilterCountsAndPages(t *testing.T) {
	database, ctx := setupTestDB(t), t.Context()
	const source = "trace-status-test"
	if _, err := database.Exec(`DELETE FROM audit_events WHERE source = $1`, source); err != nil {
		t.Fatalf("clean: %v", err)
	}
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	// Traces alternate error / ok.
	for i := range 6 {
		traceID := fmt.Sprintf("%032x", 0xa00+i)
		eventType := []string{"ROOT_FAILED", "ROOT_STARTED"}[i%2]
		if _, err := database.Exec(`INSERT INTO audit_events
			(id, event_type, event_time, entity_id, details, source, trace_id, span_id, parent_span_id)
			VALUES ($1, $2, $3, 'e', '', $4, $5, $6, NULL)`,
			traceID, eventType, base.Add(time.Duration(i)*time.Minute), source, traceID, traceID[:16]); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	s := NewAuditTraceStore(database)
	got, total, err := s.ListTraces(ctx, TraceFilter{Sources: []string{source}, Statuses: []string{"error"}, Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 3 || len(got) != 1 || got[0].TraceID != fmt.Sprintf("%032x", 0xa00) || got[0].Status != "error" {
		t.Fatalf("total=%d page=%+v, want total 3 and the oldest error trace", total, got)
	}

	tree, _, _, err := s.GetTraceTree(ctx, got[0].TraceID)
	if err != nil {
		t.Fatalf("tree: %v", err)
	}
	if tree.Status != "error" || len(tree.Children) != 1 || tree.Children[0].ParentSpanID != nil {
		t.Fatalf("tree status=%q children=%+v, want error with one root", tree.Status, tree.Children)
	}
}

// One span writing several rows yields one tree node per row, and an
// unsourced row is stored as core.
func TestTraceTreeKeepsEveryRowOfASpan(t *testing.T) {
	database, ctx := setupTestDB(t), t.Context()
	const traceID = "00000000000000000000000000000b00"
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer tx.Rollback()
	for i, src := range []string{"trace-span-test", ""} {
		e := &models.AuditEvent{ID: fmt.Sprint(traceID, i), EventType: "ROOT", EventTime: time.Unix(int64(i), 0), EntityID: "e", Source: src, TraceID: traceID, SpanID: traceID[:16]}
		if err := NewAuditEventStore(database).Create(ctx, tx, e); err != nil {
			t.Fatalf("create: %v", err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	tree, _, _, err := NewAuditTraceStore(database).GetTraceTree(ctx, traceID)
	if err != nil {
		t.Fatalf("tree: %v", err)
	}
	if len(tree.Children) != 2 || tree.Children[0].Source != "trace-span-test" || tree.Children[1].Source != "core" {
		t.Fatalf("tree children=%+v, want the sourced row then the core one", tree.Children)
	}
}

// Make sure the list and the detail include a trace's deliveries, and that search finds the trace by an entity id or a username.
func TestTraceReportsItsDeliveriesAndIsFoundByEntity(t *testing.T) {
	database, ctx := setupTestDB(t), t.Context()
	const source, traceID, userID = "trace-delivery-test", "00000000000000000000000000000c00", "user-c00"
	if _, err := database.Exec(`DELETE FROM audit_events WHERE source = $1`, source); err != nil {
		t.Fatalf("clean: %v", err)
	}
	if _, err := database.Exec(`INSERT INTO audit_events
		(id, event_type, event_time, entity_id, entity_type, details, source, trace_id, span_id, parent_span_id)
		VALUES ($1, 'CLUSTER_ACCOUNT_APPROVED', $2, $3, 'compute_cluster_user', '{"local_username":"jdoe"}', $4, $5, $6, NULL)`,
		traceID+"-row", time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), userID, source, traceID, traceID[:16]); err != nil {
		t.Fatalf("insert audit row: %v", err)
	}
	event := newEvent(time.Now())
	event.TraceID, event.SpanID = traceID, traceID[:16]
	inTx(t, database, func(tx *sql.Tx) error {
		if err := NewEventDeliveryStore(database).CreateEvent(ctx, tx, event); err != nil {
			return err
		}
		pending := newDelivery(event.ID, "slurm-association-mapper", time.Now().Add(time.Minute))
		pending.Attempts = 4
		if err := NewEventDeliveryStore(database).CreateDelivery(ctx, tx, pending); err != nil {
			return err
		}
		return NewEventDeliveryStore(database).CreateDelivery(ctx, tx, newDelivery(event.ID, "email-notifier", time.Now()))
	})

	s := NewAuditTraceStore(database)
	for _, q := range []string{userID, "jdoe"} {
		got, _, err := s.ListTraces(ctx, TraceFilter{Sources: []string{source}, Q: q})
		if err != nil {
			t.Fatalf("list q=%s: %v", q, err)
		}
		if len(got) != 1 || got[0].Status != "in_progress" || got[0].Deliveries != (models.DeliveryCounts{Pending: 2, Attempts: 4}) {
			t.Fatalf("q=%s: traces=%+v, want one in-progress trace with two pending deliveries at try 4", q, got)
		}
	}

	_, deliveries, _, err := s.GetTraceTree(ctx, traceID)
	if err != nil {
		t.Fatalf("tree: %v", err)
	}
	if len(deliveries) != 2 || deliveries[0].Subscriber != "email-notifier" || deliveries[1].Subscriber != "slurm-association-mapper" ||
		deliveries[1].Attempts != 4 || deliveries[0].SpanID != traceID[:16] || deliveries[0].EventType != event.EventType {
		t.Fatalf("deliveries=%+v, want both, by subscriber, under the publisher's span", deliveries)
	}
}
