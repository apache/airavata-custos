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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/apache/airavata-custos/internal/server"
	"github.com/apache/airavata-custos/internal/store"
)

// TestTracingCleanSlate_EndToEnd asserts the audit-driven tracing contract
// against a wiped dev DB: every audit row gets trace_id, parent_span_id is
// populated for non-roots, and the /audit/* endpoints surface the resulting
// trace tree shape. Mock COmanage server keeps traffic local.
func TestTracingCleanSlate_EndToEnd(t *testing.T) {
	comanageMock := newComanageMockServer(t)
	defer comanageMock.Close()

	pipe := newTestPipelineWithComanage(t, comanageMock.URL)
	defer pipe.stop()

	pipe.fireScenario(t, "baseline")
	pipe.waitForDrain(t, 9, 90*time.Second)
	waitForComanageAudit(t, pipe)
	flushTracing()

	var distinctTraces int
	if err := pipe.db.Get(&distinctTraces,
		"SELECT COUNT(DISTINCT trace_id) FROM amie_audit_log"); err != nil {
		t.Fatalf("count distinct traces: %v", err)
	}
	if distinctTraces < 9 {
		t.Errorf("expected >=9 distinct trace_ids across baseline scenario, got %d", distinctTraces)
	}

	var amieNullTrace int
	if err := pipe.db.Get(&amieNullTrace,
		"SELECT COUNT(*) FROM amie_audit_log WHERE trace_id IS NULL"); err != nil {
		t.Fatalf("count amie null trace: %v", err)
	}
	if amieNullTrace != 0 {
		t.Errorf("expected 0 amie_audit_log rows with NULL trace_id, got %d", amieNullTrace)
	}

	var comanageNullTrace int
	if err := pipe.db.Get(&comanageNullTrace, `
		SELECT COUNT(*) FROM audit_events
		WHERE trace_id IS NULL AND event_type LIKE 'Comanage%'`); err != nil {
		t.Fatalf("count comanage null trace: %v", err)
	}
	if comanageNullTrace != 0 {
		t.Errorf("expected 0 Comanage audit_events rows with NULL trace_id, got %d", comanageNullTrace)
	}

	assertParentSpanPopulated(t, pipe,
		"amie_audit_log",
		"SELECT COUNT(*) FROM amie_audit_log WHERE parent_span_id IS NOT NULL")
	assertParentSpanPopulated(t, pipe,
		"audit_events Comanage",
		"SELECT COUNT(*) FROM audit_events WHERE event_type LIKE 'Comanage%' AND parent_span_id IS NOT NULL")

	httpSrv := startAuditHTTPServer(t, pipe)
	defer httpSrv.Close()

	traces := getTraces(t, httpSrv.URL+"/audit/traces?limit=50")
	if len(traces) < 9 {
		t.Errorf("expected /audit/traces to return >=9 trace summaries, got %d", len(traces))
	}

	// Pick one trace and verify the tree comes back with at least one root.
	// Intermediate spans (bus.publish/bus.subscribe) are not audited, so cross-
	// connector traces may surface as sibling roots rather than nested children.
	tree := getTraceTree(t, httpSrv.URL+"/audit/traces/"+traces[0])
	if len(tree) == 0 {
		t.Fatalf("expected /audit/traces/%s to return at least one root node", traces[0])
	}

	sources := getSources(t, httpSrv.URL+"/audit/sources")
	for _, want := range []string{"amie", "comanage", "core"} {
		if !contains(sources, want) {
			t.Errorf("expected /audit/sources to include %q, got %v", want, sources)
		}
	}
}

// waitForComanageAudit blocks until COmanage subscribers have written at
// least one audit row so the per-trace assertions have data to read.
func waitForComanageAudit(t *testing.T, pipe *testPipeline) {
	t.Helper()
	end := time.Now().Add(30 * time.Second)
	for {
		var n int
		if err := pipe.db.Get(&n,
			"SELECT COUNT(*) FROM audit_events WHERE event_type LIKE 'Comanage%'"); err != nil {
			t.Fatalf("count comanage audit rows: %v", err)
		}
		if n > 0 {
			return
		}
		if time.Now().After(end) {
			t.Fatalf("timed out waiting for Comanage audit rows")
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func assertParentSpanPopulated(t *testing.T, pipe *testPipeline, label, query string) {
	t.Helper()
	var n int
	if err := pipe.db.Get(&n, query); err != nil {
		t.Fatalf("count %s parent_span: %v", label, err)
	}
	if n == 0 {
		t.Errorf("expected at least one %s row with non-null parent_span_id", label)
	}
}

// startAuditHTTPServer stands up the real server.Server over a httptest server
// so the test exercises the /audit/* contract end-to-end.
func startAuditHTTPServer(t *testing.T, pipe *testPipeline) *httptest.Server {
	t.Helper()
	srv := server.New(nil, &server.AdminDeps{AuditTraces: store.NewAuditTraceStore(pipe.db)})
	return httptest.NewServer(srv)
}

func getTraces(t *testing.T, url string) []string {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status %d", url, resp.StatusCode)
	}
	var body struct {
		Traces []struct {
			TraceID string `json:"trace_id"`
		} `json:"traces"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode traces: %v", err)
	}
	out := make([]string, 0, len(body.Traces))
	for _, r := range body.Traces {
		out = append(out, r.TraceID)
	}
	return out
}

type treeNode struct {
	SpanID       string     `json:"span_id"`
	ParentSpanID string     `json:"parent_span_id,omitempty"`
	Children     []treeNode `json:"children"`
}

func getTraceTree(t *testing.T, url string) []treeNode {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status %d", url, resp.StatusCode)
	}
	var body struct {
		Tree []treeNode `json:"tree"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode tree: %v", err)
	}
	return body.Tree
}

func getSources(t *testing.T, url string) []string {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status %d", url, resp.StatusCode)
	}
	var body struct {
		Sources []string `json:"sources"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode sources: %v", err)
	}
	return body.Sources
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
