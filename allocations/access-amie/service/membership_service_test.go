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
// Mock implementations
// ---------------------------------------------------------------------------

type mockMembershipStore struct {
	mock.Mock
}

func (m *mockMembershipStore) FindByProjectAndAccount(ctx context.Context, projectID, accountID string) (*model.ProjectMembership, error) {
	args := m.Called(ctx, projectID, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProjectMembership), args.Error(1)
}

func (m *mockMembershipStore) FindByProject(ctx context.Context, projectID string) ([]model.ProjectMembership, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ProjectMembership), args.Error(1)
}

func (m *mockMembershipStore) FindByProjectAndRole(ctx context.Context, projectID, role string) ([]model.ProjectMembership, error) {
	args := m.Called(ctx, projectID, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ProjectMembership), args.Error(1)
}

func (m *mockMembershipStore) FindByProjectAndPerson(ctx context.Context, projectID, personID string) ([]model.ProjectMembership, error) {
	args := m.Called(ctx, projectID, personID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ProjectMembership), args.Error(1)
}

func (m *mockMembershipStore) Save(ctx context.Context, tx *sql.Tx, mem *model.ProjectMembership) error {
	args := m.Called(ctx, tx, mem)
	return args.Error(0)
}

func (m *mockMembershipStore) Update(ctx context.Context, tx *sql.Tx, mem *model.ProjectMembership) error {
	args := m.Called(ctx, tx, mem)
	return args.Error(0)
}

type mockMembershipProjectStore struct {
	mock.Mock
}

func (m *mockMembershipProjectStore) FindByID(ctx context.Context, id string) (*model.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Project), args.Error(1)
}

type mockMembershipAccountStore struct {
	mock.Mock
}

func (m *mockMembershipAccountStore) FindByPerson(ctx context.Context, personID string) ([]model.ClusterAccount, error) {
	args := m.Called(ctx, personID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ClusterAccount), args.Error(1)
}

// ---------------------------------------------------------------------------
// CreateMembership tests
// ---------------------------------------------------------------------------

func TestCreateMembership_ReturnsExistingActive(t *testing.T) {
	ctx := context.Background()
	memberships := new(mockMembershipStore)
	svc := NewProjectMembershipService(memberships, new(mockMembershipProjectStore), new(mockMembershipAccountStore))

	role := "member"
	existing := &model.ProjectMembership{ID: "m1", ProjectID: "proj-1", ClusterAccountID: "acct-1", IsActive: true, Role: &role}
	memberships.On("FindByProjectAndAccount", ctx, "proj-1", "acct-1").Return(existing, nil)

	got, err := svc.CreateMembership(ctx, nil, "proj-1", "acct-1", "member")

	require.NoError(t, err)
	assert.Equal(t, existing, got)
	memberships.AssertExpectations(t)
}

func TestCreateMembership_ReactivatesInactiveMembership(t *testing.T) {
	ctx := context.Background()
	memberships := new(mockMembershipStore)
	svc := NewProjectMembershipService(memberships, new(mockMembershipProjectStore), new(mockMembershipAccountStore))

	role := "member"
	inactive := &model.ProjectMembership{ID: "m2", ProjectID: "proj-1", ClusterAccountID: "acct-1", IsActive: false, Role: &role}
	memberships.On("FindByProjectAndAccount", ctx, "proj-1", "acct-1").Return(inactive, nil)
	memberships.On("Update", ctx, mock.Anything, mock.MatchedBy(func(m *model.ProjectMembership) bool {
		return m.IsActive
	})).Return(nil)

	got, err := svc.CreateMembership(ctx, nil, "proj-1", "acct-1", "PI")

	require.NoError(t, err)
	assert.True(t, got.IsActive)
	memberships.AssertExpectations(t)
}

func TestCreateMembership_CreatesNew(t *testing.T) {
	ctx := context.Background()
	memberships := new(mockMembershipStore)
	svc := NewProjectMembershipService(memberships, new(mockMembershipProjectStore), new(mockMembershipAccountStore))

	memberships.On("FindByProjectAndAccount", ctx, "proj-1", "acct-new").Return(nil, nil)
	memberships.On("Save", ctx, mock.Anything, mock.AnythingOfType("*model.ProjectMembership")).Return(nil)

	got, err := svc.CreateMembership(ctx, nil, "proj-1", "acct-new", "member")

	require.NoError(t, err)
	assert.Equal(t, "proj-1", got.ProjectID)
	assert.Equal(t, "acct-new", got.ClusterAccountID)
	assert.True(t, got.IsActive)
	memberships.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// InactivateMembershipsByPersonAndProject tests
// ---------------------------------------------------------------------------

func TestInactivateMembershipsByPersonAndProject_ReturnsCount(t *testing.T) {
	ctx := context.Background()
	memberships := new(mockMembershipStore)
	svc := NewProjectMembershipService(memberships, new(mockMembershipProjectStore), new(mockMembershipAccountStore))

	mems := []model.ProjectMembership{
		{ID: "m1", IsActive: true},
		{ID: "m2", IsActive: true},
	}
	memberships.On("FindByProjectAndPerson", ctx, "proj-1", "person-1").Return(mems, nil)
	memberships.On("Update", ctx, mock.Anything, mock.AnythingOfType("*model.ProjectMembership")).Return(nil).Times(2)

	count, err := svc.InactivateMembershipsByPersonAndProject(ctx, nil, "proj-1", "person-1")

	require.NoError(t, err)
	assert.Equal(t, 2, count)
	memberships.AssertExpectations(t)
}

func TestInactivateMembershipsByPersonAndProject_ReturnsZeroWhenNone(t *testing.T) {
	ctx := context.Background()
	memberships := new(mockMembershipStore)
	svc := NewProjectMembershipService(memberships, new(mockMembershipProjectStore), new(mockMembershipAccountStore))

	memberships.On("FindByProjectAndPerson", ctx, "proj-1", "person-no-mem").Return([]model.ProjectMembership{}, nil)

	count, err := svc.InactivateMembershipsByPersonAndProject(ctx, nil, "proj-1", "person-no-mem")

	require.NoError(t, err)
	assert.Equal(t, 0, count)
	memberships.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// InactivateAllForProject tests
// ---------------------------------------------------------------------------

func TestInactivateAllForProject_InactivatesAll(t *testing.T) {
	ctx := context.Background()
	memberships := new(mockMembershipStore)
	svc := NewProjectMembershipService(memberships, new(mockMembershipProjectStore), new(mockMembershipAccountStore))

	mems := []model.ProjectMembership{
		{ID: "m1", IsActive: true},
		{ID: "m2", IsActive: true},
		{ID: "m3", IsActive: true},
	}
	memberships.On("FindByProject", ctx, "proj-all").Return(mems, nil)
	memberships.On("Update", ctx, mock.Anything, mock.AnythingOfType("*model.ProjectMembership")).Return(nil).Times(3)

	err := svc.InactivateAllForProject(ctx, nil, "proj-all")

	require.NoError(t, err)
	memberships.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// ReactivatePiMembership tests
// ---------------------------------------------------------------------------

func TestReactivatePiMembership_ReactivatesPiRole(t *testing.T) {
	ctx := context.Background()
	memberships := new(mockMembershipStore)
	svc := NewProjectMembershipService(memberships, new(mockMembershipProjectStore), new(mockMembershipAccountStore))

	piRole := "PI"
	mems := []model.ProjectMembership{
		{ID: "m-pi", Role: &piRole, IsActive: false},
	}
	memberships.On("FindByProjectAndRole", ctx, "proj-1", "PI").Return(mems, nil)
	memberships.On("Update", ctx, mock.Anything, mock.MatchedBy(func(m *model.ProjectMembership) bool {
		return m.IsActive
	})).Return(nil)

	err := svc.ReactivatePiMembership(ctx, nil, "proj-1")

	require.NoError(t, err)
	memberships.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// ReactivateMembershipsByPersonAndProject tests
// ---------------------------------------------------------------------------

func TestReactivateMembershipsByPersonAndProject_ReturnsCount(t *testing.T) {
	ctx := context.Background()
	memberships := new(mockMembershipStore)
	svc := NewProjectMembershipService(memberships, new(mockMembershipProjectStore), new(mockMembershipAccountStore))

	mems := []model.ProjectMembership{
		{ID: "m1", IsActive: false},
		{ID: "m2", IsActive: false},
	}
	memberships.On("FindByProjectAndPerson", ctx, "proj-1", "person-1").Return(mems, nil)
	memberships.On("Update", ctx, mock.Anything, mock.MatchedBy(func(m *model.ProjectMembership) bool {
		return m.IsActive
	})).Return(nil).Times(2)

	count, err := svc.ReactivateMembershipsByPersonAndProject(ctx, nil, "proj-1", "person-1")

	require.NoError(t, err)
	assert.Equal(t, 2, count)
	memberships.AssertExpectations(t)
}

func TestReactivateMembershipsByPersonAndProject_ReturnsZeroWhenNone(t *testing.T) {
	ctx := context.Background()
	memberships := new(mockMembershipStore)
	svc := NewProjectMembershipService(memberships, new(mockMembershipProjectStore), new(mockMembershipAccountStore))

	memberships.On("FindByProjectAndPerson", ctx, "proj-1", "person-none").Return([]model.ProjectMembership{}, nil)

	count, err := svc.ReactivateMembershipsByPersonAndProject(ctx, nil, "proj-1", "person-none")

	require.NoError(t, err)
	assert.Equal(t, 0, count)
	memberships.AssertExpectations(t)
}
