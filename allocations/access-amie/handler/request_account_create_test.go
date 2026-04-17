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
	dmodel "github.com/apache/airavata-custos/allocations/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockRACPersonService struct{ mock.Mock }

func (m *mockRACPersonService) FindOrCreateFromPacket(ctx context.Context, tx *sql.Tx, body map[string]any) (*dmodel.Person, error) {
	args := m.Called(ctx, tx, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dmodel.Person), args.Error(1)
}

type mockRACAccountService struct{ mock.Mock }

func (m *mockRACAccountService) ProvisionClusterAccount(ctx context.Context, tx *sql.Tx, person *dmodel.Person) (*dmodel.ClusterAccount, error) {
	args := m.Called(ctx, tx, person)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dmodel.ClusterAccount), args.Error(1)
}

type mockRACProjectService struct{ mock.Mock }

func (m *mockRACProjectService) CreateOrFindProject(ctx context.Context, tx *sql.Tx, projectID, grantNumber string) (*dmodel.Project, error) {
	args := m.Called(ctx, tx, projectID, grantNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dmodel.Project), args.Error(1)
}

type mockRACMembershipService struct{ mock.Mock }

func (m *mockRACMembershipService) CreateMembership(ctx context.Context, tx *sql.Tx, projectID, clusterAccountID, role string) (*dmodel.ProjectMembership, error) {
	args := m.Called(ctx, tx, projectID, clusterAccountID, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dmodel.ProjectMembership), args.Error(1)
}

type mockRACAmieClient struct{ mock.Mock }

func (m *mockRACAmieClient) ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error {
	return m.Called(ctx, packetRecID, reply).Error(0)
}

type mockRACAuditService struct{ mock.Mock }

func (m *mockRACAuditService) Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error {
	return m.Called(ctx, tx, packetID, eventID, action, entityType, entityID, summary).Error(0)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRequestAccountCreateHandler_SupportsType(t *testing.T) {
	h := NewRequestAccountCreateHandler(
		&mockRACPersonService{}, &mockRACAccountService{},
		&mockRACProjectService{}, &mockRACMembershipService{},
		&mockRACAmieClient{}, &mockRACAuditService{},
	)
	assert.Equal(t, "request_account_create", h.SupportsType())
}

func TestRequestAccountCreateHandler(t *testing.T) {
	validFixture := loadTestData(t, "request_account_create/incoming-request.json")

	tests := []struct {
		name       string
		input      map[string]any
		setupMocks func(ps *mockRACPersonService, as *mockRACAccountService, prj *mockRACProjectService, ms *mockRACMembershipService, ac *mockRACAmieClient, aud *mockRACAuditService)
		wantErr    string
	}{
		{
			name:  "valid packet processes successfully",
			input: validFixture,
			setupMocks: func(ps *mockRACPersonService, as *mockRACAccountService, prj *mockRACProjectService, ms *mockRACMembershipService, ac *mockRACAmieClient, aud *mockRACAuditService) {
				person := &dmodel.Person{ID: "person-123"}
				account := &dmodel.ClusterAccount{ID: "account-123", Username: "testuser"}
				project := &dmodel.Project{ID: "test-project-456", GrantNumber: "TEST123"}

				ps.On("FindOrCreateFromPacket", mock.Anything, mock.Anything, mock.Anything).Return(person, nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditCreatePerson, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				as.On("ProvisionClusterAccount", mock.Anything, mock.Anything, person).Return(account, nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditCreateAccount, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				prj.On("CreateOrFindProject", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(project, nil)

				ms.On("CreateMembership", mock.Anything, mock.Anything, "test-project-456", account.ID, "USER").Return(&dmodel.ProjectMembership{}, nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditCreateMembership, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ac.On("ReplyToPacket", mock.Anything, int64(233497917), mock.Anything).Return(nil)
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
				"GrantNumber":  "TEST123",
				"UserGlobalID": "12345",
			}},
			wantErr: "'ProjectID' must not be empty",
		},
		{
			name: "missing GrantNumber returns error",
			input: map[string]any{"body": map[string]any{
				"ProjectID":    "test-project-456",
				"UserGlobalID": "12345",
			}},
			wantErr: "'GrantNumber' must not be empty",
		},
		{
			name: "missing UserGlobalID returns error",
			input: map[string]any{"body": map[string]any{
				"ProjectID":   "test-project-456",
				"GrantNumber": "TEST123",
			}},
			wantErr: "'UserGlobalID' must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &mockRACPersonService{}
			as := &mockRACAccountService{}
			prj := &mockRACProjectService{}
			ms := &mockRACMembershipService{}
			ac := &mockRACAmieClient{}
			aud := &mockRACAuditService{}

			if tt.setupMocks != nil {
				tt.setupMocks(ps, as, prj, ms, ac, aud)
			}

			h := NewRequestAccountCreateHandler(ps, as, prj, ms, ac, aud)
			packet := &model.Packet{ID: "test-id", AmieID: 233497917, Type: "request_account_create"}

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
