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

type mockRPIProjectService struct{ mock.Mock }

func (m *mockRPIProjectService) InactivateProject(ctx context.Context, tx *sql.Tx, projectID string) error {
	return m.Called(ctx, tx, projectID).Error(0)
}

type mockRPIMembershipService struct{ mock.Mock }

func (m *mockRPIMembershipService) InactivateAllForProject(ctx context.Context, tx *sql.Tx, projectID string) error {
	return m.Called(ctx, tx, projectID).Error(0)
}

type mockRPIAmieClient struct{ mock.Mock }

func (m *mockRPIAmieClient) ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error {
	return m.Called(ctx, packetRecID, reply).Error(0)
}

type mockRPIAuditService struct{ mock.Mock }

func (m *mockRPIAuditService) Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error {
	return m.Called(ctx, tx, packetID, eventID, action, entityType, entityID, summary).Error(0)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRequestProjectInactivateHandler_SupportsType(t *testing.T) {
	h := NewRequestProjectInactivateHandler(
		&mockRPIProjectService{}, &mockRPIMembershipService{},
		&mockRPIAmieClient{}, &mockRPIAuditService{},
	)
	assert.Equal(t, "request_project_inactivate", h.SupportsType())
}

func TestRequestProjectInactivateHandler(t *testing.T) {
	validFixture := loadTestData(t, "request_project_inactivate/incoming-request.json")

	tests := []struct {
		name       string
		input      map[string]any
		setupMocks func(ps *mockRPIProjectService, ms *mockRPIMembershipService, ac *mockRPIAmieClient, aud *mockRPIAuditService)
		wantErr    string
	}{
		{
			name:  "valid packet processes successfully",
			input: validFixture,
			setupMocks: func(ps *mockRPIProjectService, ms *mockRPIMembershipService, ac *mockRPIAmieClient, aud *mockRPIAuditService) {
				ps.On("InactivateProject", mock.Anything, mock.Anything, "test-project-123").Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditInactivateProject, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ms.On("InactivateAllForProject", mock.Anything, mock.Anything, "test-project-123").Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditInactivateMembership, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ac.On("ReplyToPacket", mock.Anything, int64(233497911), mock.Anything).Return(nil)
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
			name: "optional PersonID is included in reply when present",
			input: map[string]any{"body": map[string]any{
				"ProjectID": "test-project-999",
				"PersonID":  "opt-person-001",
			}},
			setupMocks: func(ps *mockRPIProjectService, ms *mockRPIMembershipService, ac *mockRPIAmieClient, aud *mockRPIAuditService) {
				ps.On("InactivateProject", mock.Anything, mock.Anything, "test-project-999").Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditInactivateProject, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ms.On("InactivateAllForProject", mock.Anything, mock.Anything, "test-project-999").Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditInactivateMembership, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ac.On("ReplyToPacket", mock.Anything, mock.Anything, mock.MatchedBy(func(reply map[string]any) bool {
					body, ok := reply["body"].(map[string]any)
					return ok && body["PersonID"] == "opt-person-001"
				})).Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditReplySent, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &mockRPIProjectService{}
			ms := &mockRPIMembershipService{}
			ac := &mockRPIAmieClient{}
			aud := &mockRPIAuditService{}

			if tt.setupMocks != nil {
				tt.setupMocks(ps, ms, ac, aud)
			}

			h := NewRequestProjectInactivateHandler(ps, ms, ac, aud)
			packet := &model.Packet{ID: "test-id", AmieID: 233497911, Type: "request_project_inactivate"}

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
