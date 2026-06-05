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
	"bytes"
	"context"
	"database/sql"
	"testing"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// ---------------------------------------------------------------------------
// Mock implementation
// ---------------------------------------------------------------------------

type mockAuditStore struct {
	mock.Mock
}

func (m *mockAuditStore) Save(ctx context.Context, tx *sql.Tx, a *model.AuditLog) error {
	args := m.Called(ctx, tx, a)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// AuditService.Log tests
// ---------------------------------------------------------------------------

func TestAuditLog_WithBothPacketAndEvent(t *testing.T) {
	ctx := context.Background()
	store := new(mockAuditStore)
	svc := NewAuditService(store)

	store.On("Save", ctx, mock.Anything, mock.MatchedBy(func(a *model.AuditLog) bool {
		return a.PacketID == "packet-1" &&
			a.EventID != nil && *a.EventID == "event-1" &&
			a.Action == model.AuditCreatePerson &&
			a.EntityType == "person" &&
			a.EntityID != nil && *a.EntityID == "p1" &&
			a.Summary != nil && *a.Summary == "created person"
	})).Return(nil)

	err := svc.Log(ctx, nil, "packet-1", "event-1", model.AuditCreatePerson, "person", "p1", "created person")

	require.NoError(t, err)
	store.AssertExpectations(t)
}

func TestAuditLog_WithNullEvent(t *testing.T) {
	ctx := context.Background()
	store := new(mockAuditStore)
	svc := NewAuditService(store)

	store.On("Save", ctx, mock.Anything, mock.MatchedBy(func(a *model.AuditLog) bool {
		return a.PacketID == "packet-2" &&
			a.EventID == nil // empty event ID yields nil pointer
	})).Return(nil)

	// Empty eventID -> EventID pointer should be nil
	err := svc.Log(ctx, nil, "packet-2", "", model.AuditReplySent, "packet", "pkt-2", "reply sent")

	require.NoError(t, err)
	store.AssertExpectations(t)
}

func TestAuditLog_WithNullEntityFields(t *testing.T) {
	ctx := context.Background()
	store := new(mockAuditStore)
	svc := NewAuditService(store)

	store.On("Save", ctx, mock.Anything, mock.MatchedBy(func(a *model.AuditLog) bool {
		// Both entityID and summary are empty strings -> should be nil pointers
		return a.PacketID == "packet-3" &&
			a.EntityID == nil &&
			a.Summary == nil
	})).Return(nil)

	err := svc.Log(ctx, nil, "packet-3", "event-3", model.AuditTransactionComplete, "packet", "", "")

	require.NoError(t, err)
	store.AssertExpectations(t)
}

func TestAuditLog_PropagatesStoreError(t *testing.T) {
	ctx := context.Background()
	store := new(mockAuditStore)
	svc := NewAuditService(store)

	store.On("Save", ctx, mock.Anything, mock.AnythingOfType("*model.AuditLog")).Return(assert.AnError)

	err := svc.Log(ctx, nil, "packet-err", "evt", model.AuditCreatePerson, "person", "p-err", "summary")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "audit_service")
	store.AssertExpectations(t)
}

func TestAuditLog_PersistsTraceAndSpanIDs(t *testing.T) {
	prev := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	store := new(mockAuditStore)
	svc := NewAuditService(store)

	ctx, span := tracing.Start(context.Background(), "test.root")
	defer span.End()

	wantTrace := span.SpanContext().TraceID()
	wantSpan := span.SpanContext().SpanID()

	var captured *model.AuditLog
	store.On("Save", mock.Anything, mock.Anything, mock.MatchedBy(func(a *model.AuditLog) bool {
		captured = a
		return true
	})).Return(nil)

	err := svc.Log(ctx, nil, "packet-trace", "event-trace", model.AuditCreatePerson, "person", "p1", "with trace")
	require.NoError(t, err)
	require.NotNil(t, captured)
	if !bytes.Equal(captured.TraceID, wantTrace[:]) {
		t.Fatalf("trace_id mismatch: got %x want %x", captured.TraceID, wantTrace[:])
	}
	if !bytes.Equal(captured.SpanID, wantSpan[:]) {
		t.Fatalf("span_id mismatch: got %x want %x", captured.SpanID, wantSpan[:])
	}
	store.AssertExpectations(t)
}

func TestAuditLog_NilTraceWhenNoSpan(t *testing.T) {
	store := new(mockAuditStore)
	svc := NewAuditService(store)

	var captured *model.AuditLog
	store.On("Save", mock.Anything, mock.Anything, mock.MatchedBy(func(a *model.AuditLog) bool {
		captured = a
		return true
	})).Return(nil)

	err := svc.Log(trace.ContextWithSpanContext(context.Background(), trace.SpanContext{}),
		nil, "p", "e", model.AuditReplySent, "reply", "", "")
	require.NoError(t, err)
	require.NotNil(t, captured)
	if captured.TraceID != nil || captured.SpanID != nil {
		t.Fatalf("expected nil trace/span IDs when no active span")
	}
}
