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
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/apache/airavata-custos/internal/tracing"
)

// TestLoggingWrapsTracingProducesTraceIdHeader verifies the cmd/server stack
// composition: LoggingMiddleware(tracing.Middleware(server.New(...))) lets the
// access log see the trace_id and surfaces X-Trace-Id to the client.
func TestLoggingWrapsTracingProducesTraceIdHeader(t *testing.T) {
	prev := otel.GetTracerProvider()
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	handler := LoggingMiddleware(tracing.Middleware(New(nil, nil)))

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

	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected exactly 1 span, got %d", len(spans))
	}
	if got, want := spans[0].Name(), "http.GET /healthz"; got != want {
		t.Fatalf("span name = %q, want %q", got, want)
	}
}
