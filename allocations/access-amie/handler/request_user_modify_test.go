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

package handler

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
// Mocks
// ---------------------------------------------------------------------------

type mockRUMPersonService struct{ mock.Mock }

func (m *mockRUMPersonService) ReplaceFromModifyPacket(ctx context.Context, tx *sql.Tx, body map[string]any) error {
	return m.Called(ctx, tx, body).Error(0)
}

func (m *mockRUMPersonService) DeleteFromModifyPacket(ctx context.Context, tx *sql.Tx, body map[string]any) error {
	return m.Called(ctx, tx, body).Error(0)
}

type mockRUMAmieClient struct{ mock.Mock }

func (m *mockRUMAmieClient) ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error {
	return m.Called(ctx, packetRecID, reply).Error(0)
}

type mockRUMAuditService struct{ mock.Mock }

func (m *mockRUMAuditService) Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error {
	return m.Called(ctx, tx, packetID, eventID, action, entityType, entityID, summary).Error(0)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRequestUserModifyHandler_SupportsType(t *testing.T) {
	h := NewRequestUserModifyHandler(
		&mockRUMPersonService{}, &mockRUMAmieClient{}, &mockRUMAuditService{},
	)
	assert.Equal(t, "request_user_modify", h.SupportsType())
}

func TestRequestUserModifyHandler(t *testing.T) {
	replaceFixture := loadTestData(t, "request_user_modify_replace/incoming-request.json")
	deleteFixture := loadTestData(t, "request_user_modify_delete/incoming-request.json")

	tests := []struct {
		name       string
		input      map[string]any
		amieID     int64
		setupMocks func(ps *mockRUMPersonService, ac *mockRUMAmieClient, aud *mockRUMAuditService)
		wantErr    string
	}{
		{
			name:   "replace action processes successfully",
			input:  replaceFixture,
			amieID: 233497921,
			setupMocks: func(ps *mockRUMPersonService, ac *mockRUMAmieClient, aud *mockRUMAuditService) {
				ps.On("ReplaceFromModifyPacket", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditUpdatePerson, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ac.On("ReplyToPacket", mock.Anything, int64(233497921), mock.MatchedBy(func(reply map[string]any) bool {
					return reply["type"] == "inform_transaction_complete"
				})).Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditReplySent, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:   "delete action processes successfully",
			input:  deleteFixture,
			amieID: 233497922,
			setupMocks: func(ps *mockRUMPersonService, ac *mockRUMAmieClient, aud *mockRUMAuditService) {
				ps.On("DeleteFromModifyPacket", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditDeletePerson, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ac.On("ReplyToPacket", mock.Anything, int64(233497922), mock.Anything).Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditReplySent, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "missing body returns error",
			input:   map[string]any{},
			wantErr: "packet missing 'body'",
		},
		{
			name: "missing ActionType returns error",
			input: map[string]any{"body": map[string]any{
				"PersonID": "test-person",
			}},
			wantErr: "'ActionType' must not be empty",
		},
		{
			name: "unsupported ActionType returns error",
			input: map[string]any{"body": map[string]any{
				"ActionType": "update",
				"PersonID":   "test-person",
			}},
			wantErr: "unsupported ActionType: update",
		},
		{
			name: "ActionType is case-insensitive for replace",
			input: map[string]any{"body": map[string]any{
				"ActionType": "REPLACE",
				"PersonID":   "test-person",
			}},
			setupMocks: func(ps *mockRUMPersonService, ac *mockRUMAmieClient, aud *mockRUMAuditService) {
				ps.On("ReplaceFromModifyPacket", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditUpdatePerson, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				ac.On("ReplyToPacket", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditReplySent, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &mockRUMPersonService{}
			ac := &mockRUMAmieClient{}
			aud := &mockRUMAuditService{}

			if tt.setupMocks != nil {
				tt.setupMocks(ps, ac, aud)
			}

			h := NewRequestUserModifyHandler(ps, ac, aud)
			amieID := tt.amieID
			if amieID == 0 {
				amieID = 12345
			}
			packet := &model.Packet{ID: "test-id", AmieID: amieID, Type: "request_user_modify"}

			err := h.Handle(context.Background(), nil, tt.input, packet, "event-1")

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
