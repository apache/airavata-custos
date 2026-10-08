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

package tracing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	otelcodes "go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func setupRecordingTracer(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	prev := SetProvider(tp)
	t.Cleanup(func() { SetProvider(prev) })
	return sr
}

func TestMiddlewareProductionEmitsRootSpanAndHeader(t *testing.T) {
	sr := setupRecordingTracer(t)

	mux := http.NewServeMux()
	var (
		innerTrace string
		innerSpan  string
	)
	mux.HandleFunc("GET /probe/{id}", func(w http.ResponseWriter, r *http.Request) {
		innerTrace, innerSpan, _ = IDsFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTP(w, r)
		SetRoute(r)
	}))
	req := httptest.NewRequest(http.MethodGet, "/probe/42", nil)
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	hdr := rec.Header().Get("X-Trace-Id")
	if hdr == "" {
		t.Fatalf("expected X-Trace-Id header to be set")
	}
	if len(hdr) != 32 {
		t.Fatalf("expected 32-char hex trace id, got %q (len=%d)", hdr, len(hdr))
	}
	if hdr == "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("client traceparent became the request's trace")
	}

	if innerTrace == "" || innerSpan == "" {
		t.Fatalf("expected inner handler to see recording span; got trace=%q span=%q", innerTrace, innerSpan)
	}
	if innerTrace != hdr {
		t.Fatalf("inner trace_id %q does not match response header %q", innerTrace, hdr)
	}

	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected exactly 1 ended span, got %d", len(spans))
	}
	s := spans[0]
	if links := s.Links(); len(links) != 1 || links[0].SpanContext.TraceID().String() != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("expected the client traceparent as the only link, got %+v", links)
	}
	if got, want := s.Name(), "http.GET /probe/{id}"; got != want {
		t.Fatalf("span name = %q, want %q", got, want)
	}

	attrs := map[string]string{}
	for _, kv := range s.Attributes() {
		attrs[string(kv.Key)] = kv.Value.Emit()
	}
	if attrs["http.method"] != "GET" {
		t.Fatalf("http.method attr = %q, want GET", attrs["http.method"])
	}
	if attrs["http.route"] != "/probe/{id}" {
		t.Fatalf("http.route attr = %q, want /probe/{id}", attrs["http.route"])
	}
	if attrs["http.status_code"] != "204" {
		t.Fatalf("http.status_code attr = %q, want 204", attrs["http.status_code"])
	}
	if attrs["source"] != "http" {
		t.Fatalf("source attr = %q, want http", attrs["source"])
	}
}

func TestMiddleware5xxMarksSpanError(t *testing.T) {
	sr := setupRecordingTracer(t)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /boom", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	handler := Middleware(mux)
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if got := spans[0].Status().Code; got != otelcodes.Error {
		t.Fatalf("expected span status Error, got %v", got)
	}
}

func TestMiddlewareSetsErrorOnHandlerPanic(t *testing.T) {
	sr := setupRecordingTracer(t)

	handler := Middleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/explode", nil)
	rec := httptest.NewRecorder()

	defer func() {
		rec.Result().Body.Close()
		if r := recover(); r == nil {
			t.Fatalf("expected re-panic to surface, got none")
		}

		spans := sr.Ended()
		if len(spans) != 1 {
			t.Fatalf("expected 1 ended span, got %d", len(spans))
		}
		s := spans[0]
		if s.Status().Code != otelcodes.Error {
			t.Fatalf("expected span status Error, got %v", s.Status().Code)
		}
		if s.Status().Description != "panic" {
			t.Fatalf("expected status description 'panic', got %q", s.Status().Description)
		}
		if len(s.Events()) == 0 {
			t.Fatalf("expected at least one recorded event (the error), got 0")
		}
	}()

	handler.ServeHTTP(rec, req)
}

func TestMiddlewareFallsBackToPathWhenPatternEmpty(t *testing.T) {
	sr := setupRecordingTracer(t)

	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/no-mux-route", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	spans := sr.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if got, want := spans[0].Name(), "http.GET /no-mux-route"; got != want {
		t.Fatalf("span name = %q, want %q", got, want)
	}
}
