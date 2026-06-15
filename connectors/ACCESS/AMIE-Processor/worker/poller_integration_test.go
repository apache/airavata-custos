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
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/amieclient"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
)

// makePollerPacket constructs the JSON node shape the AMIE server returns
// for one in-progress packet.
func makePollerPacket(amieID int64, packetType string) map[string]any {
	return map[string]any{
		"type": packetType,
		"header": map[string]any{
			"packet_rec_id": float64(amieID),
		},
		"body": map[string]any{"GrantNumber": "TG-POLL-001"},
	}
}

func newPoller(t *testing.T, client pollerAmieClient) (*Poller, *stubMetrics) {
	t.Helper()
	database := setupTestDB(t)
	met := &stubMetrics{}
	cfg := amieclient.Config{
		PollInterval:  10 * time.Millisecond,
		PollerEnabled: true,
	}
	return NewPoller(client, store.NewPacketStore(database), store.NewEventStore(database), met, database, cfg), met
}

// TestPoller_SinglePacketIngestCreatesPacketAndEvent, one well-formed packet
// in, one amie_packets row + one amie_processing_events row, both NEW.
func TestPoller_SinglePacketIngestCreatesPacketAndEvent(t *testing.T) {
	database := setupTestDB(t)
	stub := &stubAmieClient{responses: [][]map[string]any{{makePollerPacket(1001, "request_project_create")}}}
	p, met := newPoller(t, stub)

	p.pollForPackets(context.Background())

	if got := countRows(t, database, "amie_packets"); got != 1 {
		t.Errorf("amie_packets: got %d, want 1", got)
	}
	if got := countRows(t, database, "amie_processing_events"); got != 1 {
		t.Errorf("amie_processing_events: got %d, want 1", got)
	}

	var status string
	if err := database.Get(&status, "SELECT status FROM amie_packets LIMIT 1"); err != nil {
		t.Fatalf("read status: %v", err)
	}
	if status != "NEW" {
		t.Errorf("packet status: got %q, want NEW", status)
	}
	if err := database.Get(&status, "SELECT status FROM amie_processing_events LIMIT 1"); err != nil {
		t.Fatalf("read event status: %v", err)
	}
	if status != "NEW" {
		t.Errorf("event status: got %q, want NEW", status)
	}

	if len(met.packetsReceived) != 1 || met.packetsReceived[0] != "request_project_create" {
		t.Errorf("packetsReceived: %v, want [request_project_create]", met.packetsReceived)
	}
	if len(met.fetchCounts) != 1 || met.fetchCounts[0] != 1 {
		t.Errorf("fetchCounts: %v, want [1]", met.fetchCounts)
	}
}

// TestPoller_DuplicateAmieIDInOneBatchDedupes, same amie_id twice in one poll
// response. After processing, only one amie_packets row exists. Witness for B3
// (poller dedup race), current single-process deployment makes this safe; the
// race only exists with concurrent pollers.
func TestPoller_DuplicateAmieIDInOneBatchDedupes(t *testing.T) {
	database := setupTestDB(t)
	p1 := makePollerPacket(2001, "request_project_create")
	p2 := makePollerPacket(2001, "request_project_create")
	stub := &stubAmieClient{responses: [][]map[string]any{{p1, p2}}}
	p, _ := newPoller(t, stub)

	p.pollForPackets(context.Background())

	if got := countRows(t, database, "amie_packets"); got != 1 {
		t.Errorf("amie_packets after dup-in-batch: got %d, want 1", got)
	}
	if got := countRows(t, database, "amie_processing_events"); got != 1 {
		t.Errorf("amie_processing_events after dup-in-batch: got %d, want 1", got)
	}
}

// TestPoller_DuplicateAmieIDAcrossTicksDedupes, same packet returned by the
// server on two consecutive ticks; only one row persists.
func TestPoller_DuplicateAmieIDAcrossTicksDedupes(t *testing.T) {
	database := setupTestDB(t)
	pkt := makePollerPacket(3001, "data_project_create")
	stub := &stubAmieClient{responses: [][]map[string]any{{pkt}, {pkt}}}
	p, _ := newPoller(t, stub)

	p.pollForPackets(context.Background())
	p.pollForPackets(context.Background())

	if got := countRows(t, database, "amie_packets"); got != 1 {
		t.Errorf("amie_packets across ticks: got %d, want 1", got)
	}
}

// TestPoller_EmptyServerResponseWritesNothing, empty packet list → no rows,
// no RecordPacketReceived, no RecordPollerFetch (which only fires when > 0).
func TestPoller_EmptyServerResponseWritesNothing(t *testing.T) {
	database := setupTestDB(t)
	stub := &stubAmieClient{responses: [][]map[string]any{{}}}
	p, met := newPoller(t, stub)

	p.pollForPackets(context.Background())

	if got := countRows(t, database, "amie_packets"); got != 0 {
		t.Errorf("amie_packets on empty fetch: got %d, want 0", got)
	}
	if len(met.fetchCounts) != 0 {
		t.Errorf("fetchCounts on empty: %v, want []", met.fetchCounts)
	}
}

// TestPoller_MalformedPacketIsDropped, packet missing header.packet_rec_id
// is logged and skipped, no DB write.
func TestPoller_MalformedPacketIsDropped(t *testing.T) {
	database := setupTestDB(t)
	badPacket := map[string]any{
		"type": "request_project_create",
		// no header at all
		"body": map[string]any{"GrantNumber": "TG-BAD"},
	}
	stub := &stubAmieClient{responses: [][]map[string]any{{badPacket}}}
	p, _ := newPoller(t, stub)

	p.pollForPackets(context.Background())

	if got := countRows(t, database, "amie_packets"); got != 0 {
		t.Errorf("amie_packets after malformed: got %d, want 0", got)
	}
}

// TestPoller_MissingTypeIsDropped, packet with header but no top-level type
// is also skipped.
func TestPoller_MissingTypeIsDropped(t *testing.T) {
	database := setupTestDB(t)
	noType := map[string]any{
		"header": map[string]any{"packet_rec_id": float64(4001)},
		"body":   map[string]any{},
	}
	stub := &stubAmieClient{responses: [][]map[string]any{{noType}}}
	p, _ := newPoller(t, stub)

	p.pollForPackets(context.Background())

	if got := countRows(t, database, "amie_packets"); got != 0 {
		t.Errorf("amie_packets after missing type: got %d, want 0", got)
	}
}

// TestPoller_DisabledRunReturnsImmediately, when PollerEnabled=false the
// Run loop must return without calling the AMIE client.
func TestPoller_DisabledRunReturnsImmediately(t *testing.T) {
	database := setupTestDB(t)
	stub := &stubAmieClient{responses: [][]map[string]any{{makePollerPacket(5001, "request_project_create")}}}
	met := &stubMetrics{}
	cfg := amieclient.Config{
		PollInterval:  10 * time.Millisecond,
		PollerEnabled: false,
	}
	p := NewPoller(stub, store.NewPacketStore(database), store.NewEventStore(database), met, database, cfg)

	done := make(chan struct{})
	go func() {
		p.Run(context.Background())
		close(done)
	}()
	select {
	case <-done:
		// ok. Run returned without polling
	case <-time.After(500 * time.Millisecond):
		t.Fatal("disabled Run did not return immediately")
	}

	if stub.callCount() != 0 {
		t.Errorf("disabled poller called client %d times, want 0", stub.callCount())
	}
}
