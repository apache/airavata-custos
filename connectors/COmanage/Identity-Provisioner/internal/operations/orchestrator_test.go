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

package operations

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
)

// recordingProcessor captures finished spans for assertions.
type recordingProcessor struct {
	mu    sync.Mutex
	spans []sdktrace.ReadOnlySpan
}

func (p *recordingProcessor) OnStart(context.Context, sdktrace.ReadWriteSpan) {}
func (p *recordingProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.spans = append(p.spans, s)
}
func (p *recordingProcessor) Shutdown(context.Context) error   { return nil }
func (p *recordingProcessor) ForceFlush(context.Context) error { return nil }

func (p *recordingProcessor) names() []string {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]string, 0, len(p.spans))
	for _, s := range p.spans {
		out = append(out, s.Name())
	}
	return out
}

func installRecorder(t *testing.T) *recordingProcessor {
	t.Helper()
	rec := &recordingProcessor{}
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		_ = tp.ForceFlush(context.Background())
		otel.SetTracerProvider(prev)
	})
	return rec
}

// fakeCore is an in-memory CoreService stub that records audit-event writes
// and returns canned user/identity data.
type fakeCore struct {
	user            *models.User
	identities      []models.UserIdentity
	auditEvents     []models.AuditEvent
	createdIdentity *models.UserIdentity
}

func (f *fakeCore) GetUser(_ context.Context, _ string) (*models.User, error) {
	return f.user, nil
}

func (f *fakeCore) ListUserIdentitiesForUser(_ context.Context, _ string) ([]models.UserIdentity, error) {
	out := make([]models.UserIdentity, len(f.identities))
	copy(out, f.identities)
	return out, nil
}

func (f *fakeCore) CreateUserIdentity(_ context.Context, ui *models.UserIdentity) (*models.UserIdentity, error) {
	f.createdIdentity = ui
	return ui, nil
}

func (f *fakeCore) CreateAuditEvent(ctx context.Context, e *models.AuditEvent) (*models.AuditEvent, error) {
	tracing.PopulateAuditIDs(ctx, &e.TraceID, &e.SpanID, &e.ParentSpanID)
	if e.ID == "" {
		e.ID = fmt.Sprintf("audit-%d", len(f.auditEvents)+1)
	}
	if e.EventTime.IsZero() {
		e.EventTime = time.Now().UTC()
	}
	f.auditEvents = append(f.auditEvents, *e)
	return e, nil
}

// mockComanageServer returns a happy-path COmanage REST mock that walks the
// orchestrator through every step without hitting a real registry.
func mockComanageServer(t *testing.T) *httptest.Server {
	t.Helper()
	// composite served on every /people/<id> GET.
	composite := `{
        "CoPerson":{"meta":{"id":42},"co_id":2,"status":"A"},
        "Name":[{"given":"E2E","family":"Test","type":"official","primary_name":true}],
        "EmailAddress":[{"mail":"e2e@example.invalid","type":"official","verified":false}],
        "Identifier":[
            {"identifier":"Person100099","type":"comanage_id","login":false,"status":"A"},
            {"identifier":"2000099","type":"uidnumber","login":false,"status":"A"}
        ]
    }`

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/people"):
			_, _ = io.WriteString(w, `[{"identifier":"Person100099","type":"comanage_id","login":false,"status":"A"}]`)
		case r.Method == http.MethodGet && strings.Contains(path, "/people/"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, composite)
		case r.Method == http.MethodPut && strings.Contains(path, "/people/"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, composite)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/co_people.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"CoPeople","Version":"1.0","CoPeople":[]}`)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/co_groups.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"CoGroups","Version":"1.0","CoGroups":[]}`)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/co_groups.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"CoGroup","Id":"55"}`)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/identifiers.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"Identifiers","Version":"1.0","Identifiers":[]}`)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/identifiers.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"Identifier","Id":"7"}`)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/co_group_members.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"CoGroupMembers","Version":"1.0","CoGroupMembers":[]}`)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/co_group_members.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"CoGroupMember","Id":"11"}`)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/unix_cluster/unix_cluster_groups.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"UnixClusterGroups","Version":"1.0","UnixClusterGroups":[]}`)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/unix_cluster/unix_cluster_groups.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"UnixClusterGroup","Id":"3"}`)
		default:
			t.Logf("unexpected mock request: %s %s", r.Method, path)
			http.NotFound(w, r)
		}
	}))
}

func newOrchestratorForTest(t *testing.T, srv *httptest.Server, core CoreService) *Orchestrator {
	t.Helper()
	c := client.New(client.Config{
		RegistryURL:   srv.URL,
		COID:          2,
		APIUser:       "co_2.test",
		APIKey:        "k",
		PersonIDType:  "comanage_id",
		UnixClusterID: 1,
		DefaultShell:  "/bin/bash",
		HomedirPrefix: "/home/",
		HTTPTimeout:   5 * time.Second,
	})
	return &Orchestrator{c: c, core: core}
}

func TestEnsurePOSIXAccount_EmitsComanageSpanTree(t *testing.T) {
	rec := installRecorder(t)
	srv := mockComanageServer(t)
	defer srv.Close()

	core := &fakeCore{user: &models.User{ID: "user-1", FirstName: "E2E", LastName: "Test", Email: "e2e@example.invalid"}}
	orch := newOrchestratorForTest(t, srv, core)

	ctx, root := tracing.Start(context.Background(), "test.root")
	defer root.End()

	cu := &models.ComputeClusterUser{ID: "ccu-1", UserID: "user-1", LocalUsername: "e2etest"}
	if err := orch.EnsurePOSIXAccount(ctx, cu); err != nil {
		t.Fatalf("EnsurePOSIXAccount: %v", err)
	}

	wantSpans := []string{
		"comanage.ensure_posix_account",
		"comanage.lookup_or_create_co_person",
		"comanage.find_stored_person_id",
		"comanage.store_person_id",
		"comanage.find_or_create_co_group",
		"comanage.find_or_create_identifier",
		"comanage.find_or_create_co_group_member",
		"comanage.find_or_create_unix_cluster_group",
		"comanage.get_person_composite",
		"comanage.update_person",
	}
	got := rec.names()
	for _, want := range wantSpans {
		if !contains(got, want) {
			t.Errorf("missing span %q. got=%v", want, got)
		}
	}
}

func TestEnsurePOSIXAccount_DlqAuditCarriesTraceID(t *testing.T) {
	installRecorder(t)

	// Server fails CreatePerson with 500, exhausting retries — triggers dlq.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/co_people.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"CoPeople","Version":"1.0","CoPeople":[]}`)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/people"):
			w.WriteHeader(http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	core := &fakeCore{user: &models.User{ID: "user-x", FirstName: "F", LastName: "L", Email: "fl@example.invalid"}}
	orch := newOrchestratorForTest(t, srv, core)

	ctx, root := tracing.Start(context.Background(), "test.root")
	defer root.End()
	wantTrace := root.SpanContext().TraceID()

	cu := &models.ComputeClusterUser{ID: "ccu-x", UserID: "user-x", LocalUsername: "flx"}
	if err := orch.EnsurePOSIXAccount(ctx, cu); err == nil {
		t.Fatalf("expected EnsurePOSIXAccount to fail under 500")
	}

	if len(core.auditEvents) == 0 {
		t.Fatalf("expected dlq to emit an audit event")
	}
	last := core.auditEvents[len(core.auditEvents)-1]
	if last.EventType != "ComanageProvisioningFailed" {
		t.Fatalf("expected ComanageProvisioningFailed dlq audit, got %q", last.EventType)
	}
	if last.TraceID != wantTrace.String() {
		t.Fatalf("dlq audit row trace_id != root span trace_id: got %s want %s", last.TraceID, wantTrace.String())
	}
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
