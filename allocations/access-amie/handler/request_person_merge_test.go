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

type mockRPMPersonService struct{ mock.Mock }

func (m *mockRPMPersonService) MergePersons(ctx context.Context, tx *sql.Tx, survivingID, retiringID string) error {
	return m.Called(ctx, tx, survivingID, retiringID).Error(0)
}

type mockRPMAmieClient struct{ mock.Mock }

func (m *mockRPMAmieClient) ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error {
	return m.Called(ctx, packetRecID, reply).Error(0)
}

type mockRPMAuditService struct{ mock.Mock }

func (m *mockRPMAuditService) Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error {
	return m.Called(ctx, tx, packetID, eventID, action, entityType, entityID, summary).Error(0)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRequestPersonMergeHandler_SupportsType(t *testing.T) {
	h := NewRequestPersonMergeHandler(
		&mockRPMPersonService{}, &mockRPMAmieClient{}, &mockRPMAuditService{},
	)
	assert.Equal(t, "request_person_merge", h.SupportsType())
}

func TestRequestPersonMergeHandler(t *testing.T) {
	validFixture := loadTestData(t, "request_person_merge/incoming-request.json")

	tests := []struct {
		name       string
		input      map[string]any
		setupMocks func(ps *mockRPMPersonService, ac *mockRPMAmieClient, aud *mockRPMAuditService)
		wantErr    string
	}{
		{
			name:  "valid packet processes successfully",
			input: validFixture,
			setupMocks: func(ps *mockRPMPersonService, ac *mockRPMAmieClient, aud *mockRPMAuditService) {
				ps.On("MergePersons", mock.Anything, mock.Anything, "test-person-primary-123", "test-person-secondary-456").Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditMergePersons, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ac.On("ReplyToPacket", mock.Anything, int64(233497920), mock.MatchedBy(func(reply map[string]any) bool {
					return reply["type"] == "inform_transaction_complete"
				})).Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditReplySent, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "missing body returns error",
			input:   map[string]any{},
			wantErr: "packet missing 'body'",
		},
		{
			name: "missing KeepPersonID returns error",
			input: map[string]any{"body": map[string]any{
				"DeletePersonID": "test-person-secondary-456",
			}},
			wantErr: "'KeepPersonID' must not be empty",
		},
		{
			name: "missing DeletePersonID returns error",
			input: map[string]any{"body": map[string]any{
				"KeepPersonID": "test-person-primary-123",
			}},
			wantErr: "'DeletePersonID' must not be empty",
		},
		{
			name: "reply type is inform_transaction_complete",
			input: map[string]any{"body": map[string]any{
				"KeepPersonID":   "keep-person-1",
				"DeletePersonID": "delete-person-2",
			}},
			setupMocks: func(ps *mockRPMPersonService, ac *mockRPMAmieClient, aud *mockRPMAuditService) {
				ps.On("MergePersons", mock.Anything, mock.Anything, "keep-person-1", "delete-person-2").Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditMergePersons, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ac.On("ReplyToPacket", mock.Anything, mock.Anything, mock.MatchedBy(func(reply map[string]any) bool {
					body, ok := reply["body"].(map[string]any)
					return ok &&
						reply["type"] == "inform_transaction_complete" &&
						body["StatusCode"] == "Success"
				})).Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditReplySent, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &mockRPMPersonService{}
			ac := &mockRPMAmieClient{}
			aud := &mockRPMAuditService{}

			if tt.setupMocks != nil {
				tt.setupMocks(ps, ac, aud)
			}

			h := NewRequestPersonMergeHandler(ps, ac, aud)
			packet := &model.Packet{ID: "test-id", AmieID: 233497920, Type: "request_person_merge"}

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
