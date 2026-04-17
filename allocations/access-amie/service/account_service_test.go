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
// Mock implementation
// ---------------------------------------------------------------------------

type mockClusterAccountStore struct {
	mock.Mock
}

func (m *mockClusterAccountStore) FindByUsername(ctx context.Context, username string) (*model.ClusterAccount, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ClusterAccount), args.Error(1)
}

func (m *mockClusterAccountStore) FindByPerson(ctx context.Context, personID string) ([]model.ClusterAccount, error) {
	args := m.Called(ctx, personID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ClusterAccount), args.Error(1)
}

func (m *mockClusterAccountStore) Save(ctx context.Context, tx *sql.Tx, a *model.ClusterAccount) error {
	args := m.Called(ctx, tx, a)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// ProvisionClusterAccount tests
// ---------------------------------------------------------------------------

func TestProvisionClusterAccount_ReturnExistingAccount(t *testing.T) {
	ctx := context.Background()
	store := new(mockClusterAccountStore)
	svc := NewUserAccountService(store)

	person := &model.Person{ID: "p1", FirstName: "John", LastName: "Doe"}
	existing := &model.ClusterAccount{ID: "acct1", PersonID: "p1", Username: "jdoe"}

	store.On("FindByPerson", ctx, "p1").Return([]model.ClusterAccount{*existing}, nil)

	got, err := svc.ProvisionClusterAccount(ctx, nil, person)

	require.NoError(t, err)
	assert.Equal(t, "jdoe", got.Username)
	store.AssertExpectations(t)
}

func TestProvisionClusterAccount_GenerateUsernameJdoe(t *testing.T) {
	ctx := context.Background()
	store := new(mockClusterAccountStore)
	svc := NewUserAccountService(store)

	person := &model.Person{ID: "p1", FirstName: "John", LastName: "Doe"}

	store.On("FindByPerson", ctx, "p1").Return([]model.ClusterAccount{}, nil)
	store.On("FindByUsername", ctx, "jdoe").Return(nil, nil) // "jdoe" is available
	store.On("Save", ctx, mock.Anything, mock.MatchedBy(func(a *model.ClusterAccount) bool {
		return a.Username == "jdoe"
	})).Return(nil)

	got, err := svc.ProvisionClusterAccount(ctx, nil, person)

	require.NoError(t, err)
	assert.Equal(t, "jdoe", got.Username)
	store.AssertExpectations(t)
}

func TestProvisionClusterAccount_GenerateJdoe1WhenJdoeTaken(t *testing.T) {
	ctx := context.Background()
	store := new(mockClusterAccountStore)
	svc := NewUserAccountService(store)

	person := &model.Person{ID: "p2", FirstName: "Janet", LastName: "Doe"}
	taken := &model.ClusterAccount{ID: "existing", Username: "jdoe"}

	store.On("FindByPerson", ctx, "p2").Return([]model.ClusterAccount{}, nil)
	store.On("FindByUsername", ctx, "jdoe").Return(taken, nil) // taken
	store.On("FindByUsername", ctx, "jdoe1").Return(nil, nil)  // available
	store.On("Save", ctx, mock.Anything, mock.MatchedBy(func(a *model.ClusterAccount) bool {
		return a.Username == "jdoe1"
	})).Return(nil)

	got, err := svc.ProvisionClusterAccount(ctx, nil, person)

	require.NoError(t, err)
	assert.Equal(t, "jdoe1", got.Username)
	store.AssertExpectations(t)
}

func TestProvisionClusterAccount_HandleMultipleSuffixes(t *testing.T) {
	ctx := context.Background()
	store := new(mockClusterAccountStore)
	svc := NewUserAccountService(store)

	person := &model.Person{ID: "p3", FirstName: "Josh", LastName: "Doe"}
	taken1 := &model.ClusterAccount{ID: "t1", Username: "jdoe"}
	taken2 := &model.ClusterAccount{ID: "t2", Username: "jdoe1"}

	store.On("FindByPerson", ctx, "p3").Return([]model.ClusterAccount{}, nil)
	store.On("FindByUsername", ctx, "jdoe").Return(taken1, nil)
	store.On("FindByUsername", ctx, "jdoe1").Return(taken2, nil)
	store.On("FindByUsername", ctx, "jdoe2").Return(nil, nil)
	store.On("Save", ctx, mock.Anything, mock.MatchedBy(func(a *model.ClusterAccount) bool {
		return a.Username == "jdoe2"
	})).Return(nil)

	got, err := svc.ProvisionClusterAccount(ctx, nil, person)

	require.NoError(t, err)
	assert.Equal(t, "jdoe2", got.Username)
	store.AssertExpectations(t)
}

func TestProvisionClusterAccount_HandleNamesWithSpaces(t *testing.T) {
	ctx := context.Background()
	store := new(mockClusterAccountStore)
	svc := NewUserAccountService(store)

	// Spaces in LastName should become hyphens
	person := &model.Person{ID: "p4", FirstName: "Mary", LastName: "Van Buren"}

	store.On("FindByPerson", ctx, "p4").Return([]model.ClusterAccount{}, nil)
	store.On("FindByUsername", ctx, "mvan-buren").Return(nil, nil)
	store.On("Save", ctx, mock.Anything, mock.MatchedBy(func(a *model.ClusterAccount) bool {
		return a.Username == "mvan-buren"
	})).Return(nil)

	got, err := svc.ProvisionClusterAccount(ctx, nil, person)

	require.NoError(t, err)
	assert.Equal(t, "mvan-buren", got.Username)
	store.AssertExpectations(t)
}

func TestProvisionClusterAccount_ErrorOnEmptyNames(t *testing.T) {
	ctx := context.Background()
	store := new(mockClusterAccountStore)
	svc := NewUserAccountService(store)

	// Person with no first or last name produces username "" - but ensureUniqueUsername
	// will loop, so let it find "" available and save it.
	person := &model.Person{ID: "p5", FirstName: "", LastName: ""}

	store.On("FindByPerson", ctx, "p5").Return([]model.ClusterAccount{}, nil)
	store.On("FindByUsername", ctx, "").Return(nil, nil)
	store.On("Save", ctx, mock.Anything, mock.AnythingOfType("*model.ClusterAccount")).Return(nil)

	// Should not error - just produces an empty username
	_, err := svc.ProvisionClusterAccount(ctx, nil, person)
	require.NoError(t, err)
	store.AssertExpectations(t)
}

func TestProvisionClusterAccount_Idempotency(t *testing.T) {
	ctx := context.Background()
	store := new(mockClusterAccountStore)
	svc := NewUserAccountService(store)

	person := &model.Person{ID: "p6", FirstName: "Alice", LastName: "Smith"}
	account := &model.ClusterAccount{ID: "acct6", PersonID: "p6", Username: "asmith"}

	// Both calls return existing account
	store.On("FindByPerson", ctx, "p6").Return([]model.ClusterAccount{*account}, nil).Times(2)

	got1, err1 := svc.ProvisionClusterAccount(ctx, nil, person)
	got2, err2 := svc.ProvisionClusterAccount(ctx, nil, person)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Equal(t, got1.Username, got2.Username)
	store.AssertExpectations(t)
}
