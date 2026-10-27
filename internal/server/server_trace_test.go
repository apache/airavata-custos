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
	"net/http"
	"net/http/httptest"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

// TestLoggingWrapsTracingProducesTraceIdHeader verifies the cmd/server stack
// composition: LoggingMiddleware(tracing.Middleware(identity.Middleware(server.New(...)))) lets the
// access log see the trace_id and surfaces X-Trace-Id to the client.
func TestLoggingWrapsTracingProducesTraceIdHeader(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	prev := tracing.SetProvider(tp)
	t.Cleanup(func() { tracing.SetProvider(prev) })

	router := identity.NewRouter(http.NewServeMux())
	srv := New(nil, router)
	handler := LoggingMiddleware(tracing.Middleware(identity.Middleware(stubAuth{}, stubAuth{}, router.PublicPaths(), srv)))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /healthz, got %d", rec.Code)
	}
	hdr := rec.Header().Get("X-Trace-Id")
	if hdr == "" || len(hdr) != 32 {
		t.Fatalf("expected 32-char X-Trace-Id header, got %q", hdr)
	}

	// Authenticated route: the span takes the mux pattern, not the raw path.
	req = httptest.NewRequest(http.MethodGet, "/users/abc", nil)
	req.Header.Set("Authorization", "Bearer t")
	handler.ServeHTTP(httptest.NewRecorder(), req)

	spans := sr.Ended()
	if len(spans) != 2 {
		t.Fatalf("expected exactly 2 spans, got %d", len(spans))
	}
	if got, want := spans[0].Name(), "http.GET /healthz"; got != want {
		t.Fatalf("span name = %q, want %q", got, want)
	}
	if got, want := spans[1].Name(), "http.GET /users/{id}"; got != want {
		t.Fatalf("span name = %q, want %q", got, want)
	}
}

type stubAuth struct{}

func (stubAuth) Verify(context.Context, string) (*identity.Claims, error) {
	return &identity.Claims{}, nil
}

func (stubAuth) ResolveCaller(context.Context, *identity.Claims) (*identity.Caller, []models.PrivilegeKey, error) {
	return &identity.Caller{}, nil, nil
}
