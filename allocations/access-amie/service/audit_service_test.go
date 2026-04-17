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
	"database/sql"
	"testing"

	"github.com/apache/airavata-custos/allocations/access-amie/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
