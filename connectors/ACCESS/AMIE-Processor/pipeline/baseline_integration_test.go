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

package pipeline

import (
	"testing"
	"time"
)

// TestPipeline_BaselineDeterminism runs the baseline scenario end-to-end and
// asserts the row counts from baseline.yaml. If baseline.yaml changes, this
// test changes with it.
func TestPipeline_BaselineDeterminism(t *testing.T) {
	pipe := newTestPipeline(t)
	defer pipe.stop()

	pipe.fireScenario(t, "baseline")
	decoded := pipe.waitForDrain(t, 12, 90*time.Second)

	// Counts match baseline.yaml.
	expectations := []struct {
		table string
		want  int
	}{
		{"amie_packets", 12},
		{"amie_processing_errors", 0},
		{"organizations", 1},
		{"users", 5},
		{"user_identities", 5},
		{"projects", 2},
		{"compute_allocations", 2},
		{"compute_allocation_diffs", 1},
		{"amie_user_dns", 2},
		{"amie_audit_extras", 52},
		{"compute_cluster_users", 4},
		{"compute_allocation_memberships", 5},
		{"project_memberships", 4},
	}
	if decoded != 12 {
		t.Errorf("decoded packets: got %d, want 12", decoded)
	}
	for _, e := range expectations {
		if got := countRows(t, pipe.db, e.table); got != e.want {
			t.Errorf("%s: got %d, want %d (per baseline.yaml)", e.table, got, e.want)
		}
	}

	// audit_events (source='amie') and amie_audit_extras carry equal counts.
	var amieAuditEvents int
	if err := pipe.db.Get(&amieAuditEvents,
		"SELECT COUNT(*) FROM audit_events WHERE source = 'amie'",
	); err != nil {
		t.Fatalf("count amie audit_events: %v", err)
	}
	if amieAuditEvents != 52 {
		t.Errorf("audit_events source='amie': got %d, want 52", amieAuditEvents)
	}

	// audit_log.by_action.
	actionCounts := []struct {
		action string
		want   int
	}{
		{"PACKET_RECEIVED", 12},
		{"CREATE_PERSON", 6},
		{"CREATE_ACCOUNT", 6},
		{"CREATE_PROJECT", 3},
		{"CREATE_ALLOCATION", 3},
		{"CREATE_MEMBERSHIP", 5},
		{"REPLY_SENT", 11},
		{"PERSIST_DNS", 3},
		{"UPDATE_PERSON", 1},
		{"MERGE_PERSONS", 1},
		{"TRANSACTION_COMPLETE", 1},
	}
	for _, ac := range actionCounts {
		var n int
		if err := pipe.db.Get(&n,
			"SELECT COUNT(*) FROM audit_events WHERE source = 'amie' AND event_type = ?", ac.action,
		); err != nil {
			t.Fatalf("count audit action %s: %v", ac.action, err)
		}
		if n != ac.want {
			t.Errorf("audit %s: got %d, want %d", ac.action, n, ac.want)
		}
	}

	// processing_events.by_status: all SUCCEEDED.
	var nonSucceeded int
	if err := pipe.db.Get(&nonSucceeded,
		"SELECT COUNT(*) FROM amie_processing_events WHERE status != 'SUCCEEDED'",
	); err != nil {
		t.Fatalf("count non-succeeded events: %v", err)
	}
	if nonSucceeded != 0 {
		t.Errorf("non-SUCCEEDED events: got %d, want 0", nonSucceeded)
	}
}
