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

	"go.opentelemetry.io/otel"
	otelcodes "go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace/noop"
)

func setupRecordingTracer(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	prev := otel.GetTracerProvider()
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })
	return sr
}

func setupNoopTracer(t *testing.T) {
	t.Helper()
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(noop.NewTracerProvider())
	t.Cleanup(func() { otel.SetTracerProvider(prev) })
}

func TestMiddlewareProductionEmitsRootSpanAndHeader(t *testing.T) {
	sr := setupRecordingTracer(t)

	mux := http.NewServeMux()
	var (
		innerTrace string
		innerSpan  string
	)
	mux.HandleFunc("GET /probe", func(w http.ResponseWriter, r *http.Request) {
		innerTrace, innerSpan = IDsFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	handler := Middleware(mux)
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
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
	if got, want := s.Name(), "http.GET /probe"; got != want {
		t.Fatalf("span name = %q, want %q", got, want)
	}

	attrs := map[string]string{}
	for _, kv := range s.Attributes() {
		attrs[string(kv.Key)] = kv.Value.Emit()
	}
	if attrs["http.method"] != "GET" {
		t.Fatalf("http.method attr = %q, want GET", attrs["http.method"])
	}
	if attrs["http.route"] != "/probe" {
		t.Fatalf("http.route attr = %q, want /probe", attrs["http.route"])
	}
	if attrs["http.status_code"] != "204" {
		t.Fatalf("http.status_code attr = %q, want 204", attrs["http.status_code"])
	}
	if attrs["source"] != "http" {
		t.Fatalf("source attr = %q, want http", attrs["source"])
	}
}

func TestMiddlewareNoopModeSetsNoHeader(t *testing.T) {
	setupNoopTracer(t)

	mux := http.NewServeMux()
	var (
		innerTrace string
		innerSpan  string
	)
	mux.HandleFunc("GET /probe", func(w http.ResponseWriter, r *http.Request) {
		innerTrace, innerSpan = IDsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := Middleware(mux)
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if hdr := rec.Header().Get("X-Trace-Id"); hdr != "" {
		t.Fatalf("expected no X-Trace-Id header in noop mode, got %q", hdr)
	}
	if innerTrace != "" || innerSpan != "" {
		t.Fatalf("expected empty IDs from noop ctx, got trace=%q span=%q", innerTrace, innerSpan)
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
