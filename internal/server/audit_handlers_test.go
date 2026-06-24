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

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

// fakeAuditStore implements AuditTraceStore for handler tests.
type fakeAuditStore struct {
	traces        []models.TraceSummary
	total         int
	tree          *models.TraceNode
	truncated     bool
	events        []models.TraceEvent
	sources       []string
	listFilter    store.TraceFilter
	getTraceArg   string
	listEventsArg struct {
		traceID, spanID string
	}
	listErr   error
	getErr    error
	eventsErr error
}

func (f *fakeAuditStore) ListTraces(_ context.Context, filter store.TraceFilter) ([]models.TraceSummary, int, error) {
	f.listFilter = filter
	if f.listErr != nil {
		return nil, 0, f.listErr
	}
	return f.traces, f.total, nil
}

func (f *fakeAuditStore) GetTraceTree(_ context.Context, traceID string) (*models.TraceNode, bool, error) {
	f.getTraceArg = traceID
	if f.getErr != nil {
		return nil, false, f.getErr
	}
	return f.tree, f.truncated, nil
}

func (f *fakeAuditStore) ListEvents(_ context.Context, traceID, spanID string) ([]models.TraceEvent, error) {
	f.listEventsArg.traceID = traceID
	f.listEventsArg.spanID = spanID
	if f.eventsErr != nil {
		return nil, f.eventsErr
	}
	return f.events, nil
}

func (f *fakeAuditStore) ListSources(_ context.Context) ([]string, error) {
	if len(f.sources) == 0 {
		return []string{"amie", "comanage", "core", "slurm"}, nil
	}
	return f.sources, nil
}

func mustHexTraceID(t *testing.T) string {
	t.Helper()
	return strings.Repeat("ab", 16)
}

// newAuditServer boots an httptest.Server around the audit handlers with a
// fake AuditTraceStore. svc is nil, audit handlers don't reach into *Service.
func newAuditServer(t *testing.T, deps *AdminDeps) *httptest.Server {
	t.Helper()
	router := identity.NewRouter(http.NewServeMux())
	inner := New(nil, router, deps)
	wrap := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inner.ServeHTTP(w, withTestCaller(r, "test-user"))
	})
	srv := httptest.NewServer(wrap)
	t.Cleanup(srv.Close)
	return srv
}

func TestListTracesReturnsRowsAndEcho(t *testing.T) {
	traceBytes := mustHexTraceID(t)
	fs := &fakeAuditStore{
		traces: []models.TraceSummary{{
			TraceID:       traceBytes,
			RootOperation: "request_account_create",
			Source:        "amie",
			Status:        "ok",
			StartedAt:     time.Unix(1700000000, 0).UTC(),
			EndedAt:       time.Unix(1700000010, 0).UTC(),
			EventCount:    5,
		}},
		total: 1,
	}
	srv := newAuditServer(t, &AdminDeps{AuditTraces: fs})

	resp, err := http.Get(srv.URL + "/audit/traces?source=amie&status=ok&limit=25")
	if err != nil {
		t.Fatalf("http get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body struct {
		Traces []map[string]any `json:"traces"`
		Total  int              `json:"total"`
		Limit  int              `json:"limit"`
		Offset int              `json:"offset"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Total != 1 || body.Limit != 25 || body.Offset != 0 {
		t.Errorf("echo = total=%d limit=%d offset=%d", body.Total, body.Limit, body.Offset)
	}
	if len(body.Traces) != 1 {
		t.Fatalf("traces len = %d", len(body.Traces))
	}
	if got := body.Traces[0]["trace_id"].(string); got != strings.Repeat("ab", 16) {
		t.Errorf("trace_id = %q, want 32-char hex", got)
	}
	if fs.listFilter.Sources[0] != "amie" || fs.listFilter.Statuses[0] != "ok" {
		t.Errorf("filter not threaded: %+v", fs.listFilter)
	}
}

func TestListTracesRejectsBadStatus(t *testing.T) {
	srv := newAuditServer(t, &AdminDeps{AuditTraces: &fakeAuditStore{}})
	resp, err := http.Get(srv.URL + "/audit/traces?status=garbage")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestListTracesAppliesDefaultLimit(t *testing.T) {
	fs := &fakeAuditStore{}
	srv := newAuditServer(t, &AdminDeps{AuditTraces: fs})
	resp, err := http.Get(srv.URL + "/audit/traces")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if fs.listFilter.Limit != defaultTraceLimit {
		t.Errorf("default limit = %d", fs.listFilter.Limit)
	}
}

func TestListTracesCapsLimitAt200(t *testing.T) {
	fs := &fakeAuditStore{}
	srv := newAuditServer(t, &AdminDeps{AuditTraces: fs})
	resp, err := http.Get(srv.URL + "/audit/traces?limit=5000")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if fs.listFilter.Limit != maxTraceListLimit {
		t.Errorf("limit cap = %d", fs.listFilter.Limit)
	}
}

func TestListTracesRejectsOversizeOffset(t *testing.T) {
	srv := newAuditServer(t, &AdminDeps{AuditTraces: &fakeAuditStore{}})
	resp, err := http.Get(srv.URL + "/audit/traces?offset=99999999")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestListTracesRejectsOversizeWindow(t *testing.T) {
	srv := newAuditServer(t, &AdminDeps{AuditTraces: &fakeAuditStore{}})
	resp, err := http.Get(srv.URL + "/audit/traces?from=2024-01-01T00:00:00Z&to=2026-01-02T00:00:00Z")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestGetTraceBadHex(t *testing.T) {
	srv := newAuditServer(t, &AdminDeps{AuditTraces: &fakeAuditStore{}})
	resp, err := http.Get(srv.URL + "/audit/traces/not-hex")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestGetTraceNotFound(t *testing.T) {
	raw := mustHexTraceID(t)
	srv := newAuditServer(t, &AdminDeps{AuditTraces: &fakeAuditStore{}})
	resp, err := http.Get(srv.URL + "/audit/traces/" + raw)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestGetTraceReturnsTree(t *testing.T) {
	raw := mustHexTraceID(t)
	traceBytes := raw
	rootSpan := "0102030405060708"
	childSpan := "090a0b0c0d0e0f10"
	tree := &models.TraceNode{
		Children: []*models.TraceNode{
			{
				TraceEvent: models.TraceEvent{
					SpanID:    rootSpan,
					Source:    "amie",
					EventType: "CREATE_PERSON",
					Status:    "ok",
					CreatedAt: time.Unix(1700000000, 0).UTC(),
				},
				Children: []*models.TraceNode{
					{
						TraceEvent: models.TraceEvent{
							SpanID:       childSpan,
							ParentSpanID: rootSpan,
							Source:       "comanage",
							EventType:    "ComanageProvisioningFailed",
							Status:       "error",
							CreatedAt:    time.Unix(1700000001, 0).UTC(),
						},
					},
				},
			},
		},
	}
	fs := &fakeAuditStore{tree: tree}
	srv := newAuditServer(t, &AdminDeps{AuditTraces: fs})

	resp, err := http.Get(srv.URL + "/audit/traces/" + raw)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["trace_id"].(string) != raw {
		t.Errorf("trace_id = %v", body["trace_id"])
	}
	treeNodes := body["tree"].([]any)
	if len(treeNodes) != 1 {
		t.Fatalf("tree len = %d", len(treeNodes))
	}
	root := treeNodes[0].(map[string]any)
	if root["span_id"].(string) != rootSpan {
		t.Errorf("root span_id mismatch")
	}
	children := root["children"].([]any)
	if len(children) != 1 {
		t.Fatalf("child len = %d", len(children))
	}
	child := children[0].(map[string]any)
	if child["parent_span_id"].(string) != rootSpan {
		t.Errorf("child parent_span_id missing/wrong")
	}
	if child["status"].(string) != "error" {
		t.Errorf("child status = %v", child["status"])
	}
	if string(fs.getTraceArg) != string(traceBytes) {
		t.Errorf("getTraceArg = %x", fs.getTraceArg)
	}
}

func TestListEventsFlat(t *testing.T) {
	raw := mustHexTraceID(t)
	traceBytes := raw
	span := "0101010101010101"
	fs := &fakeAuditStore{events: []models.TraceEvent{{
		SpanID:    span,
		Source:    "amie",
		EventType: "CREATE_PROJECT",
		Status:    "ok",
		CreatedAt: time.Unix(1700000000, 0).UTC(),
	}}}
	srv := newAuditServer(t, &AdminDeps{AuditTraces: fs})

	resp, err := http.Get(srv.URL + "/audit/events?trace_id=" + raw + "&span_id=" + span)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body struct {
		Events []map[string]any `json:"events"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Events) != 1 {
		t.Fatalf("events = %d", len(body.Events))
	}
	if body.Events[0]["span_id"].(string) != span {
		t.Errorf("span_id = %v", body.Events[0]["span_id"])
	}
	if string(fs.listEventsArg.traceID) != string(traceBytes) {
		t.Errorf("trace arg = %x", fs.listEventsArg.traceID)
	}
	if string(fs.listEventsArg.spanID) != string(span) {
		t.Errorf("span arg = %x", fs.listEventsArg.spanID)
	}
}

func TestListEventsRejectsMissingTraceID(t *testing.T) {
	srv := newAuditServer(t, &AdminDeps{AuditTraces: &fakeAuditStore{}})
	resp, err := http.Get(srv.URL + "/audit/events")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestListSources(t *testing.T) {
	srv := newAuditServer(t, &AdminDeps{AuditTraces: &fakeAuditStore{}})
	resp, err := http.Get(srv.URL + "/audit/sources")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body struct {
		Sources []string `json:"sources"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := []string{"amie", "comanage", "core", "slurm"}
	if len(body.Sources) != len(want) {
		t.Fatalf("sources = %v", body.Sources)
	}
	for i, s := range want {
		if body.Sources[i] != s {
			t.Errorf("sources[%d] = %q want %q", i, body.Sources[i], s)
		}
	}
}

func TestAdminRoutes503WithoutStore(t *testing.T) {
	srv := newAuditServer(t, nil)
	cases := []struct {
		method, path string
	}{
		{http.MethodGet, "/audit/traces"},
		{http.MethodGet, "/audit/traces/" + strings.Repeat("ab", 16)},
		{http.MethodGet, "/audit/events?trace_id=" + strings.Repeat("ab", 16)},
		{http.MethodGet, "/audit/sources"},
	}
	for _, c := range cases {
		req, _ := http.NewRequest(c.method, srv.URL+c.path, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("%s %s: %v", c.method, c.path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("%s %s: status = %d, want 503", c.method, c.path, resp.StatusCode)
		}
	}
}
