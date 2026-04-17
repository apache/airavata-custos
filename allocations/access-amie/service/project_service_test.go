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

	"github.com/apache/airavata-custos/allocations/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock implementation
// ---------------------------------------------------------------------------

type mockProjectStore struct {
	mock.Mock
}

func (m *mockProjectStore) FindByID(ctx context.Context, id string) (*model.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Project), args.Error(1)
}

func (m *mockProjectStore) Save(ctx context.Context, tx *sql.Tx, p *model.Project) error {
	args := m.Called(ctx, tx, p)
	return args.Error(0)
}

func (m *mockProjectStore) Update(ctx context.Context, tx *sql.Tx, p *model.Project) error {
	args := m.Called(ctx, tx, p)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// CreateOrFindProject tests
// ---------------------------------------------------------------------------

func TestCreateOrFindProject_ReturnsExisting(t *testing.T) {
	ctx := context.Background()
	store := new(mockProjectStore)
	svc := NewProjectService(store)

	existing := &model.Project{ID: "proj-1", GrantNumber: "G-100", IsActive: true}
	store.On("FindByID", ctx, "proj-1").Return(existing, nil)

	got, err := svc.CreateOrFindProject(ctx, nil, "proj-1", "G-100")

	require.NoError(t, err)
	assert.Equal(t, existing, got)
	store.AssertExpectations(t)
}

func TestCreateOrFindProject_CreatesNew(t *testing.T) {
	ctx := context.Background()
	store := new(mockProjectStore)
	svc := NewProjectService(store)

	store.On("FindByID", ctx, "proj-new").Return(nil, nil)
	store.On("Save", ctx, mock.Anything, mock.MatchedBy(func(p *model.Project) bool {
		return p.ID == "proj-new" && p.GrantNumber == "G-200" && p.IsActive
	})).Return(nil)

	got, err := svc.CreateOrFindProject(ctx, nil, "proj-new", "G-200")

	require.NoError(t, err)
	assert.Equal(t, "proj-new", got.ID)
	assert.Equal(t, "G-200", got.GrantNumber)
	assert.True(t, got.IsActive)
	store.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// InactivateProject tests
// ---------------------------------------------------------------------------

func TestInactivateProject_Success(t *testing.T) {
	ctx := context.Background()
	store := new(mockProjectStore)
	svc := NewProjectService(store)

	p := &model.Project{ID: "proj-1", IsActive: true}
	store.On("FindByID", ctx, "proj-1").Return(p, nil)
	store.On("Update", ctx, mock.Anything, mock.MatchedBy(func(updated *model.Project) bool {
		return !updated.IsActive
	})).Return(nil)

	err := svc.InactivateProject(ctx, nil, "proj-1")

	require.NoError(t, err)
	assert.False(t, p.IsActive)
	store.AssertExpectations(t)
}

func TestInactivateProject_ErrorOnNotFound(t *testing.T) {
	ctx := context.Background()
	store := new(mockProjectStore)
	svc := NewProjectService(store)

	store.On("FindByID", ctx, "missing").Return(nil, nil)

	err := svc.InactivateProject(ctx, nil, "missing")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	store.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// ReactivateProject tests
// ---------------------------------------------------------------------------

func TestReactivateProject_Success(t *testing.T) {
	ctx := context.Background()
	store := new(mockProjectStore)
	svc := NewProjectService(store)

	p := &model.Project{ID: "proj-2", IsActive: false}
	store.On("FindByID", ctx, "proj-2").Return(p, nil)
	store.On("Update", ctx, mock.Anything, mock.MatchedBy(func(updated *model.Project) bool {
		return updated.IsActive
	})).Return(nil)

	err := svc.ReactivateProject(ctx, nil, "proj-2")

	require.NoError(t, err)
	assert.True(t, p.IsActive)
	store.AssertExpectations(t)
}

func TestReactivateProject_ErrorOnNotFound(t *testing.T) {
	ctx := context.Background()
	store := new(mockProjectStore)
	svc := NewProjectService(store)

	store.On("FindByID", ctx, "ghost").Return(nil, nil)

	err := svc.ReactivateProject(ctx, nil, "ghost")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	store.AssertExpectations(t)
}
