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

package worker

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/config"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/service"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	corestore "github.com/apache/airavata-custos/internal/store"
)

// State machine under test (from processor.go):
//   NEW → RUNNING → SUCCEEDED                    (success path)
//   NEW → RUNNING → (rollback) → RETRY_SCHEDULED (transient failure, attempts<MaxAttempts)
//   NEW → RUNNING → (rollback) → PERMANENTLY_FAILED (after MaxAttempts)

// fakeRouter is a programmable processorRouter. Each call consults the next
// element of `errors`; if non-nil, that call returns the error. Else nil.
type fakeRouter struct {
	errors []error
	calls  int
}

func (f *fakeRouter) Route(_ context.Context, _ *sql.Tx, _ map[string]any, _ *model.Packet, _ string) error {
	idx := f.calls
	f.calls++
	if idx < len(f.errors) && f.errors[idx] != nil {
		return f.errors[idx]
	}
	return nil
}

// seedNewEvent inserts an amie_packets row + amie_processing_events row both
// in NEW state. Returns the IDs so tests can refetch state directly.
func seedNewEvent(t *testing.T, database *sqlx.DB) (packetID, eventID string) {
	t.Helper()
	packetID = uuid.NewString()
	eventID = uuid.NewString()
	now := time.Now().UTC()

	if _, err := database.Exec(
		`INSERT INTO amie_packets (id, amie_id, type, status, raw_json, received_at, retries)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		packetID, time.Now().UnixNano(), "request_project_create", string(model.PacketStatusNew),
		`{"type":"request_project_create","body":{}}`, now, 0,
	); err != nil {
		t.Fatalf("seed packet: %v", err)
	}

	if _, err := database.Exec(
		`INSERT INTO amie_processing_events (id, packet_id, type, status, attempts, payload, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		eventID, packetID, string(model.EventTypeDecodePacket), string(model.ProcessingStatusNew), 0, []byte(""), now,
	); err != nil {
		t.Fatalf("seed event: %v", err)
	}
	return packetID, eventID
}

func newProcessor(database *sqlx.DB, router processorRouter, met *stubMetrics) *Processor {
	cfg := config.AMIEConfig{WorkerInterval: 50 * time.Millisecond}
	return NewProcessor(
		store.NewEventStore(database),
		store.NewPacketStore(database),
		store.NewProcessingErrorStore(database),
		router,
		met,
		service.NewAuditService(corestore.NewAuditEventStore(database), store.NewAuditExtrasStore(database)),
		database,
		cfg,
	)
}

func readEvent(t *testing.T, database *sqlx.DB, id string) *model.ProcessingEvent {
	t.Helper()
	var e model.ProcessingEvent
	if err := database.Get(&e,
		`SELECT id, packet_id, type, status, attempts, payload, created_at, started_at, finished_at, last_error, next_retry_at
		 FROM amie_processing_events WHERE id = ?`, id,
	); err != nil {
		t.Fatalf("read event %s: %v", id, err)
	}
	return &e
}

func readPacket(t *testing.T, database *sqlx.DB, id string) *model.Packet {
	t.Helper()
	var p model.Packet
	if err := database.Get(&p,
		`SELECT id, amie_id, type, status, raw_json, received_at, decoded_at, processed_at, retries, last_error
		 FROM amie_packets WHERE id = ?`, id,
	); err != nil {
		t.Fatalf("read packet %s: %v", id, err)
	}
	return &p
}

// forceRetryNow nudges a RETRY_SCHEDULED event to be eligible immediately so
// the next processPendingEvents call picks it up without waiting for the
// real backoff. Backoff timing itself is covered by the dedicated test.
func forceRetryNow(t *testing.T, database *sqlx.DB, eventID string) {
	t.Helper()
	if _, err := database.Exec(
		"UPDATE amie_processing_events SET next_retry_at = NULL WHERE id = ?", eventID,
	); err != nil {
		t.Fatalf("force retry: %v", err)
	}
}

// TestProcessor_SuccessFirstAttempt, router returns nil; event ends
// SUCCEEDED, packet ends DECODED, no error rows.
func TestProcessor_SuccessFirstAttempt(t *testing.T) {
	database := setupTestDB(t)
	_, eventID := seedNewEvent(t, database)

	router := &fakeRouter{}
	met := &stubMetrics{}
	p := newProcessor(database, router, met)

	p.processPendingEvents(context.Background())

	ev := readEvent(t, database, eventID)
	if ev.Status != model.ProcessingStatusSucceeded {
		t.Errorf("event status: got %q, want SUCCEEDED", ev.Status)
	}
	if ev.Attempts != 1 {
		t.Errorf("attempts: got %d, want 1", ev.Attempts)
	}
	if ev.NextRetryAt != nil {
		t.Errorf("next_retry_at: got %v, want nil", ev.NextRetryAt)
	}
	pkt := readPacket(t, database, ev.PacketID)
	if pkt.Status != model.PacketStatusDecoded {
		t.Errorf("packet status: got %q, want DECODED", pkt.Status)
	}
	if got := countRows(t, database, "amie_processing_errors"); got != 0 {
		t.Errorf("error rows: got %d, want 0", got)
	}
}

// TestProcessor_FailTwiceSucceedThird, three attempts total; first two fail,
// third succeeds. Two error rows persist; event ends SUCCEEDED.
func TestProcessor_FailTwiceSucceedThird(t *testing.T) {
	database := setupTestDB(t)
	_, eventID := seedNewEvent(t, database)

	router := &fakeRouter{errors: []error{errors.New("boom 1"), errors.New("boom 2")}}
	met := &stubMetrics{}
	p := newProcessor(database, router, met)

	p.processPendingEvents(context.Background())
	if ev := readEvent(t, database, eventID); ev.Status != model.ProcessingStatusRetryScheduled || ev.Attempts != 1 {
		t.Fatalf("after attempt 1: status=%q attempts=%d, want RETRY_SCHEDULED/1", ev.Status, ev.Attempts)
	}
	forceRetryNow(t, database, eventID)

	p.processPendingEvents(context.Background())
	if ev := readEvent(t, database, eventID); ev.Status != model.ProcessingStatusRetryScheduled || ev.Attempts != 2 {
		t.Fatalf("after attempt 2: status=%q attempts=%d, want RETRY_SCHEDULED/2", ev.Status, ev.Attempts)
	}
	forceRetryNow(t, database, eventID)

	p.processPendingEvents(context.Background())
	ev := readEvent(t, database, eventID)
	if ev.Status != model.ProcessingStatusSucceeded {
		t.Errorf("final status: got %q, want SUCCEEDED", ev.Status)
	}
	if ev.Attempts != 3 {
		t.Errorf("final attempts: got %d, want 3", ev.Attempts)
	}
	if got := countRows(t, database, "amie_processing_errors"); got != 2 {
		t.Errorf("error rows: got %d, want 2", got)
	}
}

// TestProcessor_FailThreeTimesPermanentlyFails, every attempt fails. Event
// ends PERMANENTLY_FAILED, packet ends FAILED, next_retry_at is NULL, three
// error rows.
func TestProcessor_FailThreeTimesPermanentlyFails(t *testing.T) {
	database := setupTestDB(t)
	_, eventID := seedNewEvent(t, database)

	router := &fakeRouter{errors: []error{
		errors.New("boom 1"), errors.New("boom 2"), errors.New("boom 3"),
	}}
	p := newProcessor(database, router, &stubMetrics{})

	p.processPendingEvents(context.Background())
	forceRetryNow(t, database, eventID)
	p.processPendingEvents(context.Background())
	forceRetryNow(t, database, eventID)
	p.processPendingEvents(context.Background())

	ev := readEvent(t, database, eventID)
	if ev.Status != model.ProcessingStatusPermanentlyFailed {
		t.Errorf("event status: got %q, want PERMANENTLY_FAILED", ev.Status)
	}
	if ev.Attempts != MaxAttempts {
		t.Errorf("attempts: got %d, want %d", ev.Attempts, MaxAttempts)
	}
	if ev.NextRetryAt != nil {
		t.Errorf("next_retry_at on permanent fail: got %v, want nil", ev.NextRetryAt)
	}
	pkt := readPacket(t, database, ev.PacketID)
	if pkt.Status != model.PacketStatusFailed {
		t.Errorf("packet status: got %q, want FAILED", pkt.Status)
	}
	if got := countRows(t, database, "amie_processing_errors"); got != MaxAttempts {
		t.Errorf("error rows: got %d, want %d", got, MaxAttempts)
	}
}

// TestProcessor_BackoffBlocksFutureRetry, a RETRY_SCHEDULED event with
// next_retry_at in the future MUST NOT be picked up. Verifies the
// FindTop50EventsToProcess filter respects the schedule.
func TestProcessor_BackoffBlocksFutureRetry(t *testing.T) {
	database := setupTestDB(t)
	_, eventID := seedNewEvent(t, database)

	router := &fakeRouter{errors: []error{errors.New("boom")}}
	p := newProcessor(database, router, &stubMetrics{})

	// First attempt → fails → event becomes RETRY_SCHEDULED with next_retry_at = now + 30s.
	p.processPendingEvents(context.Background())
	ev := readEvent(t, database, eventID)
	if ev.Status != model.ProcessingStatusRetryScheduled {
		t.Fatalf("setup: event not RETRY_SCHEDULED, got %q", ev.Status)
	}
	if ev.NextRetryAt == nil || !ev.NextRetryAt.After(time.Now().UTC()) {
		t.Fatalf("setup: next_retry_at not in future, got %v", ev.NextRetryAt)
	}

	// Run processor again without forcing retry, event should NOT be re-picked.
	callsBefore := router.calls
	p.processPendingEvents(context.Background())
	if router.calls != callsBefore {
		t.Errorf("router invoked while next_retry_at in future: calls went %d -> %d", callsBefore, router.calls)
	}
	ev2 := readEvent(t, database, eventID)
	if ev2.Attempts != 1 {
		t.Errorf("attempts: got %d, want 1 (no re-pick during backoff)", ev2.Attempts)
	}
}

// TestProcessor_AttemptCounterAcrossRollback verifies the subtle invariant
// in recordFailureInNewTransaction: the processing tx rolls back (attempt
// increment reverts), then the new tx re-reads the event and computes
// effectiveAttempts = stored+1. After one failure, attempts must equal
// exactly 1.
func TestProcessor_AttemptCounterAcrossRollback(t *testing.T) {
	database := setupTestDB(t)
	_, eventID := seedNewEvent(t, database)

	router := &fakeRouter{errors: []error{errors.New("boom")}}
	p := newProcessor(database, router, &stubMetrics{})

	p.processPendingEvents(context.Background())

	ev := readEvent(t, database, eventID)
	if ev.Attempts != 1 {
		t.Errorf("attempts after one failure: got %d, want 1", ev.Attempts)
	}
	if got := countRows(t, database, "amie_processing_errors"); got != 1 {
		t.Errorf("error rows after one failure: got %d, want 1", got)
	}
}
