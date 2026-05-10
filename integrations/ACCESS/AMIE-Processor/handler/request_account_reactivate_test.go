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

type mockRARMembershipService struct{ mock.Mock }

func (m *mockRARMembershipService) ReactivateMembershipsByPersonAndProject(ctx context.Context, tx *sql.Tx, projectID, personID string) (int, error) {
	args := m.Called(ctx, tx, projectID, personID)
	return args.Int(0), args.Error(1)
}

type mockRARAmieClient struct{ mock.Mock }

func (m *mockRARAmieClient) ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error {
	return m.Called(ctx, packetRecID, reply).Error(0)
}

type mockRARAuditService struct{ mock.Mock }

func (m *mockRARAuditService) Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error {
	return m.Called(ctx, tx, packetID, eventID, action, entityType, entityID, summary).Error(0)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRequestAccountReactivateHandler_SupportsType(t *testing.T) {
	h := NewRequestAccountReactivateHandler(
		&mockRARMembershipService{}, &mockRARAmieClient{}, &mockRARAuditService{},
	)
	assert.Equal(t, "request_account_reactivate", h.SupportsType())
}

func TestRequestAccountReactivateHandler(t *testing.T) {
	validFixture := loadTestData(t, "request_account_reactivate/incoming-request.json")

	tests := []struct {
		name       string
		input      map[string]any
		setupMocks func(ms *mockRARMembershipService, ac *mockRARAmieClient, aud *mockRARAuditService)
		wantErr    string
	}{
		{
			name:  "valid packet processes successfully",
			input: validFixture,
			setupMocks: func(ms *mockRARMembershipService, ac *mockRARAmieClient, aud *mockRARAuditService) {
				ms.On("ReactivateMembershipsByPersonAndProject", mock.Anything, mock.Anything, "test-project-456", "test-user-person-123").Return(1, nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditReactivateMembership, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ac.On("ReplyToPacket", mock.Anything, int64(233497923), mock.Anything).Return(nil)
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
				"PersonID": "test-user-person-123",
			}},
			wantErr: "'ProjectID' must not be empty",
		},
		{
			name: "missing PersonID returns error",
			input: map[string]any{"body": map[string]any{
				"ProjectID": "test-project-456",
			}},
			wantErr: "'PersonID' must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &mockRARMembershipService{}
			ac := &mockRARAmieClient{}
			aud := &mockRARAuditService{}

			if tt.setupMocks != nil {
				tt.setupMocks(ms, ac, aud)
			}

			h := NewRequestAccountReactivateHandler(ms, ac, aud)
			packet := &model.Packet{ID: "test-id", AmieID: 233497923, Type: "request_account_reactivate"}

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
