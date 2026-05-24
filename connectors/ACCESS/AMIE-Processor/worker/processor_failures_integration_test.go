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

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/google/uuid"
)

// TestProcessor_RouterWriteRolledBackOnError: the processor wraps every
// router invocation in a single transaction. A write the router performs
// inside that tx MUST be rolled back when the router subsequently returns
// an error. This is the contract that lets handlers trust their `*sql.Tx`
// parameter.
func TestProcessor_RouterWriteRolledBackOnError(t *testing.T) {
	database := setupTestDB(t)
	packetID, eventID := seedNewEvent(t, database)

	// fakeRouter that writes a probe row to amie_audit_log inside the tx,
	// then returns an error, forcing the processor to roll the tx back.
	auditProbe := uuid.NewString()
	router := &probeRouter{
		work: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx,
				`INSERT INTO amie_audit_log (packet_id, action, entity_type, summary)
				 VALUES (?, ?, ?, ?)`,
				packetID, "CREATE_PROJECT", "probe", auditProbe,
			)
			return err
		},
		returnErr: errors.New("simulated handler failure after audit write"),
	}
	p := newProcessor(database, router, &stubMetrics{})

	p.processPendingEvents(context.Background())

	// The audit write must NOT be visible, tx rolled back.
	var hits int
	if err := database.Get(&hits,
		"SELECT COUNT(*) FROM amie_audit_log WHERE summary = ?", auditProbe,
	); err != nil {
		t.Fatalf("query audit probe: %v", err)
	}
	if hits != 0 {
		t.Errorf("probe audit rows: got %d, want 0 (tx must roll back when router errors)", hits)
	}

	// Event should still be visible in RETRY_SCHEDULED state (separate
	// "record failure" tx is what writes the error row + retry schedule).
	ev := readEvent(t, database, eventID)
	if ev.Status != model.ProcessingStatusRetryScheduled {
		t.Errorf("event status: got %q, want RETRY_SCHEDULED", ev.Status)
	}
	if got := countRows(t, database, "amie_processing_errors"); got != 1 {
		t.Errorf("processing_errors: got %d, want 1", got)
	}
}

// TestProcessor_FAILEDStatusIsUnreachable: `FAILED` is defined in
// model/event.go but no code path in the processor writes it. After driving
// an event through success, retry, and permanent-fail flows, no
// amie_processing_events row should end in FAILED.
func TestProcessor_FAILEDStatusIsUnreachable(t *testing.T) {
	database := setupTestDB(t)

	// One event that succeeds.
	_, _ = seedNewEvent(t, database)
	pOK := newProcessor(database, &fakeRouter{}, &stubMetrics{})
	pOK.processPendingEvents(context.Background())

	// One event that permanently fails after 3 attempts.
	_, permID := seedNewEvent(t, database)
	pFail := newProcessor(database,
		&fakeRouter{errors: []error{errors.New("e1"), errors.New("e2"), errors.New("e3")}},
		&stubMetrics{},
	)
	pFail.processPendingEvents(context.Background())
	forceRetryNow(t, database, permID)
	pFail.processPendingEvents(context.Background())
	forceRetryNow(t, database, permID)
	pFail.processPendingEvents(context.Background())

	if ev := readEvent(t, database, permID); ev.Status != model.ProcessingStatusPermanentlyFailed {
		t.Errorf("perm-fail event status: got %q, want PERMANENTLY_FAILED", ev.Status)
	}

	var failedCount int
	if err := database.Get(&failedCount,
		"SELECT COUNT(*) FROM amie_processing_events WHERE status = 'FAILED'",
	); err != nil {
		t.Fatalf("count FAILED events: %v", err)
	}
	if failedCount != 0 {
		t.Errorf("amie_processing_events in FAILED status: got %d, want 0 (B1: unreachable)", failedCount)
	}
}

// probeRouter does a configurable write inside the handler's tx, then
// returns a configurable error. Lets tests verify tx atomicity.
type probeRouter struct {
	work      func(ctx context.Context, tx *sql.Tx) error
	returnErr error
}

func (p *probeRouter) Route(ctx context.Context, tx *sql.Tx, _ map[string]any, _ *model.Packet, _ string) error {
	if p.work != nil {
		if err := p.work(ctx, tx); err != nil {
			return err
		}
	}
	return p.returnErr
}
