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

package service

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
)

func TestAuditEventTraceIDs_PopulatedFromActiveSpan(t *testing.T) {
	prev := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	ctx, span := tracing.Start(context.Background(), "test.root")
	defer span.End()

	wantTrace := span.SpanContext().TraceID()
	wantSpan := span.SpanContext().SpanID()

	e := &models.AuditEvent{EventType: "X", EntityID: "y"}
	tracing.PopulateAuditIDs(ctx, &e.TraceID, &e.SpanID, &e.ParentSpanID)

	if e.TraceID != wantTrace.String() {
		t.Fatalf("trace_id mismatch: got %s want %s", e.TraceID, wantTrace.String())
	}
	if e.SpanID != wantSpan.String() {
		t.Fatalf("span_id mismatch: got %s want %s", e.SpanID, wantSpan.String())
	}
}

func TestAuditEventTraceIDs_NilWhenNoSpan(t *testing.T) {
	e := &models.AuditEvent{EventType: "X", EntityID: "y"}
	tracing.PopulateAuditIDs(context.Background(), &e.TraceID, &e.SpanID, &e.ParentSpanID)

	if e.TraceID != "" || e.SpanID != "" {
		t.Fatalf("expected empty trace/span IDs, got trace=%s span=%s", e.TraceID, e.SpanID)
	}
}

func TestAuditEventTraceIDs_NotOverwrittenWhenPreset(t *testing.T) {
	prev := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	ctx, span := tracing.Start(context.Background(), "test.root")
	defer span.End()

	preset := "0102030405060708090a0b0c0d0e0f10"
	presetSpan := "0102030405060708"
	e := &models.AuditEvent{
		EventType: "X",
		EntityID:  "y",
		TraceID:   preset,
		SpanID:    presetSpan,
	}
	tracing.PopulateAuditIDs(ctx, &e.TraceID, &e.SpanID, &e.ParentSpanID)

	if e.TraceID != preset {
		t.Fatalf("preset trace_id was overwritten: got %s want %s", e.TraceID, preset)
	}
	if e.SpanID != presetSpan {
		t.Fatalf("preset span_id was overwritten: got %s want %s", e.SpanID, presetSpan)
	}
}
