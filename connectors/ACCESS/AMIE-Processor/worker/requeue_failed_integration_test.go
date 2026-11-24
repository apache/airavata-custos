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
	"errors"
	"testing"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/jmoiron/sqlx"
)

// failPermanently drives a fresh packet through MaxAttempts failures.
func failPermanently(t *testing.T, database *sqlx.DB) (packetID, eventID string) {
	t.Helper()
	packetID, eventID = seedNewEvent(t, database)
	failing := make([]error, MaxAttempts)
	for i := range failing {
		failing[i] = errors.New("cluster account lookup failed")
	}
	p := newProcessor(database, &fakeRouter{errors: failing}, &stubMetrics{})
	for i := 0; i < MaxAttempts; i++ {
		p.processPendingEvents(context.Background())
		forceRetryNow(t, database, eventID)
	}
	if got := readPacket(t, database, packetID).Status; got != model.PacketStatusFailed {
		t.Fatalf("setup: packet status %q, want FAILED", got)
	}
	return packetID, eventID
}

func TestRequeueFailed_ProcessorRunsThePacketAgain(t *testing.T) {
	database := setupTestDB(t)
	packetID, eventID := failPermanently(t, database)
	packets := store.NewPacketStore(database)

	queued, err := packets.RequeueFailed(context.Background(), packetID)
	if err != nil {
		t.Fatalf("requeue: %v", err)
	}
	if queued == nil || queued.Status != model.PacketStatusNew {
		t.Fatalf("requeue returned %+v, want the packet in NEW", queued)
	}
	if ev := readEvent(t, database, eventID); ev.Status != model.ProcessingStatusNew || ev.NextRetryAt != nil {
		t.Fatalf("event after requeue: status %q next_retry_at %v, want NEW and nil", ev.Status, ev.NextRetryAt)
	}

	router := &fakeRouter{}
	newProcessor(database, router, &stubMetrics{}).processPendingEvents(context.Background())

	if router.calls != 1 {
		t.Errorf("router calls: got %d, want 1", router.calls)
	}
	if got := readPacket(t, database, packetID).Status; got != model.PacketStatusProcessed {
		t.Errorf("packet status: got %q, want PROCESSED", got)
	}
	if got := readEvent(t, database, eventID).Status; got != model.ProcessingStatusSucceeded {
		t.Errorf("event status: got %q, want SUCCEEDED", got)
	}
}

// The rerun is a single attempt: failing again puts the packet straight back in FAILED.
func TestRequeueFailed_FailedRerunReturnsToFailed(t *testing.T) {
	database := setupTestDB(t)
	packetID, eventID := failPermanently(t, database)

	if _, err := store.NewPacketStore(database).RequeueFailed(context.Background(), packetID); err != nil {
		t.Fatalf("requeue: %v", err)
	}
	newProcessor(database, &fakeRouter{errors: []error{errors.New("still broken")}}, &stubMetrics{}).
		processPendingEvents(context.Background())

	if got := readEvent(t, database, eventID).Status; got != model.ProcessingStatusPermanentlyFailed {
		t.Errorf("event status: got %q, want PERMANENTLY_FAILED", got)
	}
	if got := readPacket(t, database, packetID).Status; got != model.PacketStatusFailed {
		t.Errorf("packet status: got %q, want FAILED", got)
	}
}

func TestRequeueFailed_OnlyFailedPacketsAreQueued(t *testing.T) {
	database := setupTestDB(t)
	packets := store.NewPacketStore(database)

	packetID, _ := failPermanently(t, database)
	if p, err := packets.RequeueFailed(context.Background(), packetID); err != nil || p == nil {
		t.Fatalf("first requeue: packet %v err %v, want queued", p, err)
	}
	if p, err := packets.RequeueFailed(context.Background(), packetID); err != nil || p != nil {
		t.Errorf("second requeue: packet %v err %v, want nil (already queued)", p, err)
	}

	newPacketID, newEventID := seedNewEvent(t, database)
	if p, err := packets.RequeueFailed(context.Background(), newPacketID); err != nil || p != nil {
		t.Errorf("requeue of NEW packet: packet %v err %v, want nil", p, err)
	}
	if got := readEvent(t, database, newEventID).Status; got != model.ProcessingStatusNew {
		t.Errorf("NEW packet's event changed to %q", got)
	}
}
