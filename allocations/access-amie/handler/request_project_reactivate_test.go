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

type mockRPRProjectService struct{ mock.Mock }

func (m *mockRPRProjectService) ReactivateProject(ctx context.Context, tx *sql.Tx, projectID string) error {
	return m.Called(ctx, tx, projectID).Error(0)
}

type mockRPRMembershipService struct{ mock.Mock }

func (m *mockRPRMembershipService) ReactivatePiMembership(ctx context.Context, tx *sql.Tx, projectID string) error {
	return m.Called(ctx, tx, projectID).Error(0)
}

type mockRPRAmieClient struct{ mock.Mock }

func (m *mockRPRAmieClient) ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error {
	return m.Called(ctx, packetRecID, reply).Error(0)
}

type mockRPRAuditService struct{ mock.Mock }

func (m *mockRPRAuditService) Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error {
	return m.Called(ctx, tx, packetID, eventID, action, entityType, entityID, summary).Error(0)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRequestProjectReactivateHandler_SupportsType(t *testing.T) {
	h := NewRequestProjectReactivateHandler(
		&mockRPRProjectService{}, &mockRPRMembershipService{},
		&mockRPRAmieClient{}, &mockRPRAuditService{},
	)
	assert.Equal(t, "request_project_reactivate", h.SupportsType())
}

func TestRequestProjectReactivateHandler(t *testing.T) {
	validFixture := loadTestData(t, "request_project_reactivate/incoming-request.json")

	tests := []struct {
		name       string
		input      map[string]any
		setupMocks func(ps *mockRPRProjectService, ms *mockRPRMembershipService, ac *mockRPRAmieClient, aud *mockRPRAuditService)
		wantErr    string
	}{
		{
			name:  "valid packet processes successfully",
			input: validFixture,
			setupMocks: func(ps *mockRPRProjectService, ms *mockRPRMembershipService, ac *mockRPRAmieClient, aud *mockRPRAuditService) {
				ps.On("ReactivateProject", mock.Anything, mock.Anything, "test-project-123").Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditReactivateProject, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ms.On("ReactivatePiMembership", mock.Anything, mock.Anything, "test-project-123").Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditReactivateMembership, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ac.On("ReplyToPacket", mock.Anything, int64(233497914), mock.Anything).Return(nil)
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
			name: "optional PersonID is forwarded in reply",
			input: map[string]any{"body": map[string]any{
				"ProjectID": "test-project-777",
				"PersonID":  "opt-person-002",
			}},
			setupMocks: func(ps *mockRPRProjectService, ms *mockRPRMembershipService, ac *mockRPRAmieClient, aud *mockRPRAuditService) {
				ps.On("ReactivateProject", mock.Anything, mock.Anything, "test-project-777").Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditReactivateProject, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ms.On("ReactivatePiMembership", mock.Anything, mock.Anything, "test-project-777").Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditReactivateMembership, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ac.On("ReplyToPacket", mock.Anything, mock.Anything, mock.MatchedBy(func(reply map[string]any) bool {
					body, ok := reply["body"].(map[string]any)
					return ok && body["PersonID"] == "opt-person-002"
				})).Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditReplySent, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &mockRPRProjectService{}
			ms := &mockRPRMembershipService{}
			ac := &mockRPRAmieClient{}
			aud := &mockRPRAuditService{}

			if tt.setupMocks != nil {
				tt.setupMocks(ps, ms, ac, aud)
			}

			h := NewRequestProjectReactivateHandler(ps, ms, ac, aud)
			packet := &model.Packet{ID: "test-id", AmieID: 233497914, Type: "request_project_reactivate"}

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
