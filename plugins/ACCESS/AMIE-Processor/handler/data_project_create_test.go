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

type mockDPCPersonService struct{ mock.Mock }

func (m *mockDPCPersonService) PersistDNsForPerson(ctx context.Context, tx *sql.Tx, personID string, dnList []string) error {
	return m.Called(ctx, tx, personID, dnList).Error(0)
}

type mockDPCAmieClient struct{ mock.Mock }

func (m *mockDPCAmieClient) ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error {
	return m.Called(ctx, packetRecID, reply).Error(0)
}

type mockDPCAuditService struct{ mock.Mock }

func (m *mockDPCAuditService) Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error {
	return m.Called(ctx, tx, packetID, eventID, action, entityType, entityID, summary).Error(0)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestDataProjectCreateHandler_SupportsType(t *testing.T) {
	h := NewDataProjectCreateHandler(
		&mockDPCPersonService{}, &mockDPCAmieClient{}, &mockDPCAuditService{},
	)
	assert.Equal(t, "data_project_create", h.SupportsType())
}

func TestDataProjectCreateHandler(t *testing.T) {
	validFixture := loadTestData(t, "data_project_create/incoming-data.json")

	tests := []struct {
		name       string
		input      map[string]any
		setupMocks func(ps *mockDPCPersonService, ac *mockDPCAmieClient, aud *mockDPCAuditService)
		wantErr    string
	}{
		{
			name:  "valid packet with DNs processes successfully",
			input: validFixture,
			setupMocks: func(ps *mockDPCPersonService, ac *mockDPCAmieClient, aud *mockDPCAuditService) {
				ps.On("PersistDNsForPerson", mock.Anything, mock.Anything, "test-person-456", mock.Anything).Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditPersistDNs, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ac.On("ReplyToPacket", mock.Anything, int64(233497909), mock.MatchedBy(func(reply map[string]any) bool {
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
			name: "missing ProjectID returns error",
			input: map[string]any{"body": map[string]any{
				"PersonID": "test-person-456",
			}},
			wantErr: "'ProjectID' must not be empty",
		},
		{
			name: "missing PersonID returns error",
			input: map[string]any{"body": map[string]any{
				"ProjectID": "test-project-123",
			}},
			wantErr: "'PersonID' must not be empty",
		},
		{
			name: "packet without DnList still sends reply",
			input: map[string]any{"body": map[string]any{
				"ProjectID": "test-project-123",
				"PersonID":  "test-person-456",
			}},
			setupMocks: func(ps *mockDPCPersonService, ac *mockDPCAmieClient, aud *mockDPCAuditService) {
				// PersistDNsForPerson should NOT be called when DnList is absent
				ac.On("ReplyToPacket", mock.Anything, mock.Anything, mock.MatchedBy(func(reply map[string]any) bool {
					return reply["type"] == "inform_transaction_complete"
				})).Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditReplySent, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &mockDPCPersonService{}
			ac := &mockDPCAmieClient{}
			aud := &mockDPCAuditService{}

			if tt.setupMocks != nil {
				tt.setupMocks(ps, ac, aud)
			}

			h := NewDataProjectCreateHandler(ps, ac, aud)
			packet := &model.Packet{ID: "test-id", AmieID: 233497909, Type: "data_project_create"}

			err := h.Handle(context.Background(), nil, tt.input, packet, "event-1")

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)

			// Verify PersistDNsForPerson was NOT called when DnList is absent
			if tt.name == "packet without DnList still sends reply" {
				ps.AssertNotCalled(t, "PersistDNsForPerson", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}
