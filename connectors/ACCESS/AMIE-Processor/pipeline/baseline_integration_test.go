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

// TestPipeline_BaselineDeterminism, runs the canonical baseline scenario
// end-to-end against the real mock-amie server, drains, and asserts every
// total_count from
// `connectors/ACCESS/AMIE-Processor/testdata/scenarios/baseline.yaml`.
//
// This is a *determinism* test, not a *correctness* test. It locks
// in "same packets in → same row counts out" so future handler / scenario
// changes get caught. If baseline.yaml changes, this test changes with it ,
// the YAML is the source of truth here.
func TestPipeline_BaselineDeterminism(t *testing.T) {
	pipe := newTestPipeline(t)
	defer pipe.stop()

	pipe.fireScenario(t, "baseline")
	decoded := pipe.waitForDrain(t, 9, 90*time.Second)

	// Expected totals come straight from
	// connectors/ACCESS/AMIE-Processor/testdata/scenarios/baseline.yaml.
	// Field references in comments are the YAML section.
	expectations := []struct {
		table string
		want  int
	}{
		{"amie_packets", 9},                   // packets.total_count
		{"amie_processing_errors", 0},         // not_expected
		{"organizations", 1},                  // tables.organizations.total_count
		{"users", 2},                          // tables.users.total_count
		{"user_identities", 2},                // tables.user_identities.total_count
		{"projects", 2},                       // tables.projects.total_count
		{"compute_allocations", 2},            // tables.compute_allocations.total_count
		{"compute_allocation_diffs", 1},       // tables.compute_allocation_diffs.total_count
		{"amie_user_dns", 2},                  // tables.amie_user_dns.total_count
		{"amie_audit_log", 23},                // audit_log.total_count
		{"compute_cluster_users", 0},          // not_expected
		{"compute_allocation_memberships", 0}, // not_expected
	}
	if decoded != 9 {
		t.Errorf("decoded packets: got %d, want 9", decoded)
	}
	for _, e := range expectations {
		if got := countRows(t, pipe.db, e.table); got != e.want {
			t.Errorf("%s: got %d, want %d (per baseline.yaml)", e.table, got, e.want)
		}
	}

	// Audit-by-action breakdown also per baseline.yaml audit_log.by_action.
	actionCounts := []struct {
		action string
		want   int
	}{
		{"CREATE_PERSON", 3},
		{"CREATE_PROJECT", 3},
		{"CREATE_ALLOCATION", 3},
		{"REPLY_SENT", 8},
		{"PERSIST_DNS", 3},
		{"UPDATE_PERSON", 1},
		{"MERGE_PERSONS", 1},
		{"TRANSACTION_COMPLETE", 1},
	}
	for _, ac := range actionCounts {
		var n int
		if err := pipe.db.Get(&n,
			"SELECT COUNT(*) FROM amie_audit_log WHERE action = ?", ac.action,
		); err != nil {
			t.Fatalf("count audit action %s: %v", ac.action, err)
		}
		if n != ac.want {
			t.Errorf("audit %s: got %d, want %d", ac.action, n, ac.want)
		}
	}

	// Every processing event ended SUCCEEDED per baseline.yaml processing_events.
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
