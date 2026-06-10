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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
)

// ---------------------------------------------------------------------------
// Mocks for the two stores AuditService talks to
// ---------------------------------------------------------------------------

type mockCoreEventStore struct {
	mock.Mock
}

func (m *mockCoreEventStore) FindByID(ctx context.Context, id string) (*models.AuditEvent, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*models.AuditEvent), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *mockCoreEventStore) FindByEntity(ctx context.Context, entityID string) ([]models.AuditEvent, error) {
	args := m.Called(ctx, entityID)
	return args.Get(0).([]models.AuditEvent), args.Error(1)
}
func (m *mockCoreEventStore) FindByEventType(ctx context.Context, eventType string) ([]models.AuditEvent, error) {
	args := m.Called(ctx, eventType)
	return args.Get(0).([]models.AuditEvent), args.Error(1)
}
func (m *mockCoreEventStore) ListAll(ctx context.Context) ([]*models.AuditEvent, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.AuditEvent), args.Error(1)
}
func (m *mockCoreEventStore) Create(ctx context.Context, tx *sql.Tx, e *models.AuditEvent) error {
	args := m.Called(ctx, tx, e)
	return args.Error(0)
}
func (m *mockCoreEventStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

type mockExtrasStore struct {
	mock.Mock
}

func (m *mockExtrasStore) Save(ctx context.Context, tx *sql.Tx, e *model.AmieAuditExtras) error {
	args := m.Called(ctx, tx, e)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestAuditLog_WritesBothRowsWithSharedID(t *testing.T) {
	ctx := context.Background()
	coreEvents := new(mockCoreEventStore)
	extras := new(mockExtrasStore)
	svc := NewAuditService(coreEvents, extras)

	var captured *models.AuditEvent
	coreEvents.On("Create", ctx, mock.Anything, mock.MatchedBy(func(e *models.AuditEvent) bool {
		captured = e
		return e.EventType == string(model.AuditCreatePerson) &&
			e.EntityType == "person" &&
			e.EntityID == "p1" &&
			e.Details == "created person" &&
			e.Source == auditSource
	})).Return(nil)

	extras.On("Save", ctx, mock.Anything, mock.MatchedBy(func(x *model.AmieAuditExtras) bool {
		return x.PacketID == "packet-1" &&
			x.EventID != nil && *x.EventID == "event-1" &&
			captured != nil && x.AuditEventID == captured.ID
	})).Return(nil)

	err := svc.Log(ctx, nil, "packet-1", "event-1", model.AuditCreatePerson, "person", "p1", "created person")
	require.NoError(t, err)
	coreEvents.AssertExpectations(t)
	extras.AssertExpectations(t)
}

func TestAuditLog_NilEventIDWhenEmpty(t *testing.T) {
	ctx := context.Background()
	coreEvents := new(mockCoreEventStore)
	extras := new(mockExtrasStore)
	svc := NewAuditService(coreEvents, extras)

	coreEvents.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
	extras.On("Save", ctx, mock.Anything, mock.MatchedBy(func(x *model.AmieAuditExtras) bool {
		return x.PacketID == "packet-2" && x.EventID == nil
	})).Return(nil)

	require.NoError(t, svc.Log(ctx, nil, "packet-2", "", model.AuditReplySent, "reply", "", "reply sent"))
}

func TestAuditLog_PropagatesCoreStoreError(t *testing.T) {
	ctx := context.Background()
	coreEvents := new(mockCoreEventStore)
	extras := new(mockExtrasStore)
	svc := NewAuditService(coreEvents, extras)

	coreEvents.On("Create", ctx, mock.Anything, mock.Anything).Return(assert.AnError)

	err := svc.Log(ctx, nil, "p", "e", model.AuditCreatePerson, "person", "p1", "boom")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "audit_events")
	// extras must NOT be called when the core write failed
	extras.AssertNotCalled(t, "Save")
}

func TestAuditLog_PropagatesExtrasStoreError(t *testing.T) {
	ctx := context.Background()
	coreEvents := new(mockCoreEventStore)
	extras := new(mockExtrasStore)
	svc := NewAuditService(coreEvents, extras)

	coreEvents.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
	extras.On("Save", ctx, mock.Anything, mock.Anything).Return(assert.AnError)

	err := svc.Log(ctx, nil, "p", "e", model.AuditCreatePerson, "person", "p1", "boom")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "amie_audit_extras")
}

func TestAuditLog_PersistsTraceAndSpanIDs(t *testing.T) {
	prev := otel.GetTracerProvider()
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	ctx, span := tracing.Start(context.Background(), "test.root")
	defer span.End()
	wantTrace := span.SpanContext().TraceID()
	wantSpan := span.SpanContext().SpanID()

	coreEvents := new(mockCoreEventStore)
	extras := new(mockExtrasStore)
	svc := NewAuditService(coreEvents, extras)

	var captured *models.AuditEvent
	coreEvents.On("Create", mock.Anything, mock.Anything, mock.MatchedBy(func(e *models.AuditEvent) bool {
		captured = e
		return true
	})).Return(nil)
	extras.On("Save", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	require.NoError(t, svc.Log(ctx, nil, "packet-trace", "event-trace", model.AuditCreatePerson, "person", "p1", "with trace"))
	require.NotNil(t, captured)
	if !bytes.Equal(captured.TraceID, wantTrace[:]) {
		t.Fatalf("trace_id mismatch: got %x want %x", captured.TraceID, wantTrace[:])
	}
	if !bytes.Equal(captured.SpanID, wantSpan[:]) {
		t.Fatalf("span_id mismatch: got %x want %x", captured.SpanID, wantSpan[:])
	}
}

func TestAuditLog_NilTraceWhenNoSpan(t *testing.T) {
	coreEvents := new(mockCoreEventStore)
	extras := new(mockExtrasStore)
	svc := NewAuditService(coreEvents, extras)

	var captured *models.AuditEvent
	coreEvents.On("Create", mock.Anything, mock.Anything, mock.MatchedBy(func(e *models.AuditEvent) bool {
		captured = e
		return true
	})).Return(nil)
	extras.On("Save", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	require.NoError(t, svc.Log(
		trace.ContextWithSpanContext(context.Background(), trace.SpanContext{}),
		nil, "p", "e", model.AuditReplySent, "reply", "", ""))
	require.NotNil(t, captured)
	if captured.TraceID != nil || captured.SpanID != nil {
		t.Fatalf("expected nil trace/span IDs when no active span")
	}
}

func TestAuditLog_RejectsEmptyPacketID(t *testing.T) {
	svc := NewAuditService(new(mockCoreEventStore), new(mockExtrasStore))
	err := svc.Log(context.Background(), nil, "", "evt", model.AuditCreatePerson, "person", "p", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "packet_id")
}
