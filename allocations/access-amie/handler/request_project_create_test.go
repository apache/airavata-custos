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

type mockRPCPersonService struct{ mock.Mock }

func (m *mockRPCPersonService) FindOrCreateFromPacket(ctx context.Context, tx *sql.Tx, body map[string]any) (*dmodel.Person, error) {
	args := m.Called(ctx, tx, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dmodel.Person), args.Error(1)
}

type mockRPCAccountService struct{ mock.Mock }

func (m *mockRPCAccountService) ProvisionClusterAccount(ctx context.Context, tx *sql.Tx, person *dmodel.Person) (*dmodel.ClusterAccount, error) {
	args := m.Called(ctx, tx, person)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dmodel.ClusterAccount), args.Error(1)
}

type mockRPCProjectService struct{ mock.Mock }

func (m *mockRPCProjectService) CreateOrFindProject(ctx context.Context, tx *sql.Tx, projectID, grantNumber string) (*dmodel.Project, error) {
	args := m.Called(ctx, tx, projectID, grantNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dmodel.Project), args.Error(1)
}

type mockRPCMembershipService struct{ mock.Mock }

func (m *mockRPCMembershipService) CreateMembership(ctx context.Context, tx *sql.Tx, projectID, clusterAccountID, role string) (*dmodel.ProjectMembership, error) {
	args := m.Called(ctx, tx, projectID, clusterAccountID, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dmodel.ProjectMembership), args.Error(1)
}

type mockRPCAmieClient struct{ mock.Mock }

func (m *mockRPCAmieClient) ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error {
	return m.Called(ctx, packetRecID, reply).Error(0)
}

type mockRPCAuditService struct{ mock.Mock }

func (m *mockRPCAuditService) Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error {
	return m.Called(ctx, tx, packetID, eventID, action, entityType, entityID, summary).Error(0)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newRPCHandler(
	ps *mockRPCPersonService,
	as *mockRPCAccountService,
	prj *mockRPCProjectService,
	ms *mockRPCMembershipService,
	ac *mockRPCAmieClient,
	aud *mockRPCAuditService,
) *RequestProjectCreateHandler {
	return NewRequestProjectCreateHandler(ps, as, prj, ms, ac, aud)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestRequestProjectCreateHandler_SupportsType(t *testing.T) {
	h := newRPCHandler(
		&mockRPCPersonService{}, &mockRPCAccountService{},
		&mockRPCProjectService{}, &mockRPCMembershipService{},
		&mockRPCAmieClient{}, &mockRPCAuditService{},
	)
	assert.Equal(t, "request_project_create", h.SupportsType())
}

func TestRequestProjectCreateHandler(t *testing.T) {
	validFixture := loadTestData(t, "request_project_create/incoming-request.json")

	tests := []struct {
		name       string
		input      map[string]any
		setupMocks func(ps *mockRPCPersonService, as *mockRPCAccountService, prj *mockRPCProjectService, ms *mockRPCMembershipService, ac *mockRPCAmieClient, aud *mockRPCAuditService)
		wantErr    string
	}{
		{
			name:  "valid packet processes successfully",
			input: validFixture,
			setupMocks: func(ps *mockRPCPersonService, as *mockRPCAccountService, prj *mockRPCProjectService, ms *mockRPCMembershipService, ac *mockRPCAmieClient, aud *mockRPCAuditService) {
				person := &dmodel.Person{ID: "person-123"}
				account := &dmodel.ClusterAccount{ID: "account-123", Username: "hwan"}
				project := &dmodel.Project{ID: "project-123", GrantNumber: "NNT259276"}

				ps.On("FindOrCreateFromPacket", mock.Anything, mock.Anything, mock.Anything).Return(person, nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditCreatePerson, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				as.On("ProvisionClusterAccount", mock.Anything, mock.Anything, person).Return(account, nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditCreateAccount, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				prj.On("CreateOrFindProject", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(project, nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditCreateProject, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ms.On("CreateMembership", mock.Anything, mock.Anything, project.ID, account.ID, "PI").Return(&dmodel.ProjectMembership{}, nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditCreateMembership, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				ac.On("ReplyToPacket", mock.Anything, int64(233497907), mock.Anything).Return(nil)
				aud.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, model.AuditReplySent, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "missing body returns error",
			input:   map[string]any{},
			wantErr: "packet missing 'body'",
		},
		{
			name: "missing GrantNumber returns error",
			input: map[string]any{"body": map[string]any{
				"PiGlobalID":  "PI123",
				"PiFirstName": "John",
				"PiLastName":  "Doe",
			}},
			wantErr: "'GrantNumber' must not be empty",
		},
		{
			name: "missing PiGlobalID returns error",
			input: map[string]any{"body": map[string]any{
				"GrantNumber": "NNT259276",
				"PiFirstName": "John",
				"PiLastName":  "Doe",
			}},
			wantErr: "'PiGlobalID' must not be empty",
		},
		{
			name: "missing PiFirstName returns error",
			input: map[string]any{"body": map[string]any{
				"GrantNumber": "NNT259276",
				"PiGlobalID":  "PI123",
				"PiLastName":  "Doe",
			}},
			wantErr: "'PiFirstName' must not be empty",
		},
		{
			name: "missing PiLastName returns error",
			input: map[string]any{"body": map[string]any{
				"GrantNumber": "NNT259276",
				"PiGlobalID":  "PI123",
				"PiFirstName": "John",
			}},
			wantErr: "'PiLastName' must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := &mockRPCPersonService{}
			as := &mockRPCAccountService{}
			prj := &mockRPCProjectService{}
			ms := &mockRPCMembershipService{}
			ac := &mockRPCAmieClient{}
			aud := &mockRPCAuditService{}

			if tt.setupMocks != nil {
				tt.setupMocks(ps, as, prj, ms, ac, aud)
			}

			h := newRPCHandler(ps, as, prj, ms, ac, aud)
			packet := &model.Packet{ID: "test-id", AmieID: 233497907, Type: "request_project_create"}

			err := h.Handle(context.Background(), nil, tt.input, packet, "event-1")

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)

			// Verify key reply assertions for the happy path
			if tt.name == "valid packet processes successfully" {
				ac.AssertCalled(t, "ReplyToPacket", mock.Anything, int64(233497907), mock.Anything)
				ps.AssertCalled(t, "FindOrCreateFromPacket", mock.Anything, mock.Anything, mock.Anything)
				as.AssertCalled(t, "ProvisionClusterAccount", mock.Anything, mock.Anything, mock.Anything)
				prj.AssertCalled(t, "CreateOrFindProject", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				ms.AssertCalled(t, "CreateMembership", mock.Anything, mock.Anything, mock.Anything, mock.Anything, "PI")
			}
		})
	}
}
