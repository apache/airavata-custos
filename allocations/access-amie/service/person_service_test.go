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
	"errors"
	"testing"

	"github.com/apache/airavata-custos/allocations/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock implementations
// ---------------------------------------------------------------------------

type mockPersonStore struct {
	mock.Mock
}

func (m *mockPersonStore) FindByID(ctx context.Context, id string) (*model.Person, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Person), args.Error(1)
}

func (m *mockPersonStore) FindByAccessGlobalID(ctx context.Context, globalID string) (*model.Person, error) {
	args := m.Called(ctx, globalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Person), args.Error(1)
}

func (m *mockPersonStore) Save(ctx context.Context, tx *sql.Tx, p *model.Person) error {
	args := m.Called(ctx, tx, p)
	return args.Error(0)
}

func (m *mockPersonStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

type mockPersonDNStore struct {
	mock.Mock
}

func (m *mockPersonDNStore) ExistsByPersonAndDN(ctx context.Context, personID, dn string) (bool, error) {
	args := m.Called(ctx, personID, dn)
	return args.Bool(0), args.Error(1)
}

func (m *mockPersonDNStore) Save(ctx context.Context, tx *sql.Tx, d *model.PersonDN) error {
	args := m.Called(ctx, tx, d)
	return args.Error(0)
}

func (m *mockPersonDNStore) DeleteByPersonID(ctx context.Context, tx *sql.Tx, personID string) error {
	args := m.Called(ctx, tx, personID)
	return args.Error(0)
}

func (m *mockPersonDNStore) DeleteByPersonIDNotIn(ctx context.Context, tx *sql.Tx, personID string, dnsToKeep []string) error {
	args := m.Called(ctx, tx, personID, dnsToKeep)
	return args.Error(0)
}

func (m *mockPersonDNStore) FindByPersonID(ctx context.Context, personID string) ([]model.PersonDN, error) {
	args := m.Called(ctx, personID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.PersonDN), args.Error(1)
}

type mockPersonAccountStore struct {
	mock.Mock
}

func (m *mockPersonAccountStore) FindByPerson(ctx context.Context, personID string) ([]model.ClusterAccount, error) {
	args := m.Called(ctx, personID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ClusterAccount), args.Error(1)
}

func (m *mockPersonAccountStore) UpdatePersonID(ctx context.Context, tx *sql.Tx, accountID, newPersonID string) error {
	args := m.Called(ctx, tx, accountID, newPersonID)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// FindOrCreateFromPacket tests
// ---------------------------------------------------------------------------

func TestFindOrCreateFromPacket_FindExistingByGlobalID(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	dns := new(mockPersonDNStore)
	accounts := new(mockPersonAccountStore)
	svc := NewPersonService(persons, dns, accounts)

	existing := &model.Person{ID: "p1", AccessGlobalID: "global-123"}
	persons.On("FindByAccessGlobalID", ctx, "global-123").Return(existing, nil)

	body := map[string]any{"UserGlobalID": "global-123"}
	got, err := svc.FindOrCreateFromPacket(ctx, nil, body)

	require.NoError(t, err)
	assert.Equal(t, existing, got)
	persons.AssertExpectations(t)
}

func TestFindOrCreateFromPacket_CreateNewWithAllFields(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	dns := new(mockPersonDNStore)
	accounts := new(mockPersonAccountStore)
	svc := NewPersonService(persons, dns, accounts)

	persons.On("FindByAccessGlobalID", ctx, "global-456").Return(nil, nil)
	persons.On("Save", ctx, mock.Anything, mock.AnythingOfType("*model.Person")).Return(nil)

	body := map[string]any{
		"UserGlobalID":     "global-456",
		"UserFirstName":    "Jane",
		"UserLastName":     "Doe",
		"UserEmail":        "jane@example.com",
		"UserOrganization": "Test Org",
		"UserOrgCode":      "ORG1",
		"NsfStatusCode":    "F",
	}
	got, err := svc.FindOrCreateFromPacket(ctx, nil, body)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "global-456", got.AccessGlobalID)
	assert.Equal(t, "Jane", got.FirstName)
	assert.Equal(t, "Doe", got.LastName)
	assert.Equal(t, "jane@example.com", got.Email)
	require.NotNil(t, got.Organization)
	assert.Equal(t, "Test Org", *got.Organization)
	persons.AssertExpectations(t)
}

func TestFindOrCreateFromPacket_CreateWithDNList(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	dns := new(mockPersonDNStore)
	accounts := new(mockPersonAccountStore)
	svc := NewPersonService(persons, dns, accounts)

	persons.On("FindByAccessGlobalID", ctx, "global-789").Return(nil, nil)
	persons.On("Save", ctx, mock.Anything, mock.AnythingOfType("*model.Person")).Return(nil)
	dns.On("Save", ctx, mock.Anything, mock.AnythingOfType("*model.PersonDN")).Return(nil).Times(2)

	body := map[string]any{
		"UserGlobalID":  "global-789",
		"UserFirstName": "Bob",
		"UserLastName":  "Smith",
		"UserEmail":     "bob@example.com",
		"UserDnList":    []any{"/CN=bob/O=test", "/CN=bob2/O=test"},
	}
	got, err := svc.FindOrCreateFromPacket(ctx, nil, body)

	require.NoError(t, err)
	require.NotNil(t, got)
	persons.AssertExpectations(t)
	dns.AssertExpectations(t)
}

func TestFindOrCreateFromPacket_ErrorOnMissingUserGlobalID(t *testing.T) {
	ctx := context.Background()
	svc := NewPersonService(new(mockPersonStore), new(mockPersonDNStore), new(mockPersonAccountStore))

	body := map[string]any{"UserFirstName": "No", "UserLastName": "ID"}
	_, err := svc.FindOrCreateFromPacket(ctx, nil, body)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "UserGlobalID is required")
}

// ---------------------------------------------------------------------------
// ReplaceFromModifyPacket tests
// ---------------------------------------------------------------------------

func TestReplaceFromModifyPacket_UpdateAllFields(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	dns := new(mockPersonDNStore)
	accounts := new(mockPersonAccountStore)
	svc := NewPersonService(persons, dns, accounts)

	p := &model.Person{ID: "p1", FirstName: "Old", LastName: "Name", Email: "old@example.com"}
	persons.On("FindByID", ctx, "p1").Return(p, nil)
	persons.On("Save", ctx, mock.Anything, p).Return(nil)

	body := map[string]any{
		"PersonID":      "p1",
		"UserFirstName": "New",
		"UserLastName":  "Name2",
		"UserEmail":     "new@example.com",
	}
	err := svc.ReplaceFromModifyPacket(ctx, nil, body)

	require.NoError(t, err)
	assert.Equal(t, "New", p.FirstName)
	assert.Equal(t, "Name2", p.LastName)
	assert.Equal(t, "new@example.com", p.Email)
	persons.AssertExpectations(t)
}

func TestReplaceFromModifyPacket_PartialUpdate(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	dns := new(mockPersonDNStore)
	accounts := new(mockPersonAccountStore)
	svc := NewPersonService(persons, dns, accounts)

	p := &model.Person{ID: "p1", FirstName: "Keep", LastName: "This", Email: "keep@example.com"}
	persons.On("FindByID", ctx, "p1").Return(p, nil)
	persons.On("Save", ctx, mock.Anything, p).Return(nil)

	body := map[string]any{
		"PersonID":      "p1",
		"UserFirstName": "Updated",
	}
	err := svc.ReplaceFromModifyPacket(ctx, nil, body)

	require.NoError(t, err)
	assert.Equal(t, "Updated", p.FirstName)
	assert.Equal(t, "This", p.LastName) // unchanged
	persons.AssertExpectations(t)
}

func TestReplaceFromModifyPacket_PreserveOrgWhenAbsent(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	dns := new(mockPersonDNStore)
	accounts := new(mockPersonAccountStore)
	svc := NewPersonService(persons, dns, accounts)

	org := "Original Org"
	p := &model.Person{ID: "p1", FirstName: "F", Organization: &org}
	persons.On("FindByID", ctx, "p1").Return(p, nil)
	persons.On("Save", ctx, mock.Anything, p).Return(nil)

	// No UserOrganization key in body - org should be preserved
	body := map[string]any{
		"PersonID":      "p1",
		"UserFirstName": "Updated",
	}
	err := svc.ReplaceFromModifyPacket(ctx, nil, body)

	require.NoError(t, err)
	require.NotNil(t, p.Organization)
	assert.Equal(t, "Original Org", *p.Organization)
}

func TestReplaceFromModifyPacket_ClearDNsWhenEmpty(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	dns := new(mockPersonDNStore)
	accounts := new(mockPersonAccountStore)
	svc := NewPersonService(persons, dns, accounts)

	p := &model.Person{ID: "p1"}
	persons.On("FindByID", ctx, "p1").Return(p, nil)
	dns.On("DeleteByPersonID", ctx, mock.Anything, "p1").Return(nil)
	persons.On("Save", ctx, mock.Anything, p).Return(nil)

	body := map[string]any{
		"PersonID":   "p1",
		"UserDnList": []any{}, // empty list triggers full deletion
	}
	err := svc.ReplaceFromModifyPacket(ctx, nil, body)

	require.NoError(t, err)
	dns.AssertExpectations(t)
}

func TestReplaceFromModifyPacket_UpdateDNList(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	dns := new(mockPersonDNStore)
	accounts := new(mockPersonAccountStore)
	svc := NewPersonService(persons, dns, accounts)

	p := &model.Person{ID: "p1"}
	persons.On("FindByID", ctx, "p1").Return(p, nil)
	dns.On("DeleteByPersonIDNotIn", ctx, mock.Anything, "p1", []string{"/CN=new"}).Return(nil)
	dns.On("ExistsByPersonAndDN", ctx, "p1", "/CN=new").Return(false, nil)
	dns.On("Save", ctx, mock.Anything, mock.AnythingOfType("*model.PersonDN")).Return(nil)
	persons.On("Save", ctx, mock.Anything, p).Return(nil)

	body := map[string]any{
		"PersonID":   "p1",
		"UserDnList": []any{"/CN=new"},
	}
	err := svc.ReplaceFromModifyPacket(ctx, nil, body)

	require.NoError(t, err)
	dns.AssertExpectations(t)
	persons.AssertExpectations(t)
}

func TestReplaceFromModifyPacket_ErrorOnMissingPersonID(t *testing.T) {
	ctx := context.Background()
	svc := NewPersonService(new(mockPersonStore), new(mockPersonDNStore), new(mockPersonAccountStore))

	body := map[string]any{"UserFirstName": "No", "UserLastName": "ID"}
	err := svc.ReplaceFromModifyPacket(ctx, nil, body)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PersonID is required")
}

func TestReplaceFromModifyPacket_ErrorOnUnknownPersonID(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	svc := NewPersonService(persons, new(mockPersonDNStore), new(mockPersonAccountStore))

	persons.On("FindByID", ctx, "unknown").Return(nil, nil)

	body := map[string]any{"PersonID": "unknown"}
	err := svc.ReplaceFromModifyPacket(ctx, nil, body)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	persons.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// PersistDNsForPerson tests
// ---------------------------------------------------------------------------

func TestPersistDNsForPerson_PersistNewDN(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	dns := new(mockPersonDNStore)
	accounts := new(mockPersonAccountStore)
	svc := NewPersonService(persons, dns, accounts)

	dns.On("ExistsByPersonAndDN", ctx, "p1", "/CN=new").Return(false, nil)
	dns.On("Save", ctx, mock.Anything, mock.AnythingOfType("*model.PersonDN")).Return(nil)

	err := svc.PersistDNsForPerson(ctx, nil, "p1", []string{"/CN=new"})

	require.NoError(t, err)
	dns.AssertExpectations(t)
}

func TestPersistDNsForPerson_SkipExistingDN(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	dns := new(mockPersonDNStore)
	accounts := new(mockPersonAccountStore)
	svc := NewPersonService(persons, dns, accounts)

	dns.On("ExistsByPersonAndDN", ctx, "p1", "/CN=existing").Return(true, nil)
	// Save should NOT be called for existing DNs

	err := svc.PersistDNsForPerson(ctx, nil, "p1", []string{"/CN=existing"})

	require.NoError(t, err)
	dns.AssertExpectations(t)
	dns.AssertNotCalled(t, "Save", mock.Anything, mock.Anything, mock.Anything)
}

// ---------------------------------------------------------------------------
// MergePersons tests
// ---------------------------------------------------------------------------

func TestMergePersons_MoveAccountsAndDNs(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	dns := new(mockPersonDNStore)
	accounts := new(mockPersonAccountStore)
	svc := NewPersonService(persons, dns, accounts)

	surviving := &model.Person{ID: "survivor"}
	retiring := &model.Person{ID: "retiring"}

	persons.On("FindByID", ctx, "survivor").Return(surviving, nil)
	persons.On("FindByID", ctx, "retiring").Return(retiring, nil)
	accounts.On("FindByPerson", ctx, "retiring").Return([]model.ClusterAccount{{ID: "acct1", PersonID: "retiring"}}, nil)
	accounts.On("UpdatePersonID", ctx, mock.Anything, "acct1", "survivor").Return(nil)
	dns.On("FindByPersonID", ctx, "retiring").Return([]model.PersonDN{{PersonID: "retiring", DN: "/CN=retiring"}}, nil)
	dns.On("ExistsByPersonAndDN", ctx, "survivor", "/CN=retiring").Return(false, nil)
	dns.On("Save", ctx, mock.Anything, mock.AnythingOfType("*model.PersonDN")).Return(nil)
	persons.On("Delete", ctx, mock.Anything, "retiring").Return(nil)

	err := svc.MergePersons(ctx, nil, "survivor", "retiring")

	require.NoError(t, err)
	persons.AssertExpectations(t)
	accounts.AssertExpectations(t)
	dns.AssertExpectations(t)
}

func TestMergePersons_ErrorOnUnknownSurvivingPerson(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	svc := NewPersonService(persons, new(mockPersonDNStore), new(mockPersonAccountStore))

	persons.On("FindByID", ctx, "missing").Return(nil, nil)

	err := svc.MergePersons(ctx, nil, "missing", "retiring")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "surviving person")
	persons.AssertExpectations(t)
}

func TestMergePersons_ErrorOnUnknownRetiringPerson(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	svc := NewPersonService(persons, new(mockPersonDNStore), new(mockPersonAccountStore))

	surviving := &model.Person{ID: "survivor"}
	persons.On("FindByID", ctx, "survivor").Return(surviving, nil)
	persons.On("FindByID", ctx, "missing-retiring").Return(nil, nil)

	err := svc.MergePersons(ctx, nil, "survivor", "missing-retiring")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "retiring person")
	persons.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// DeleteFromModifyPacket tests
// ---------------------------------------------------------------------------

func TestDeleteFromModifyPacket_DeleteByID(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	svc := NewPersonService(persons, new(mockPersonDNStore), new(mockPersonAccountStore))

	persons.On("Delete", ctx, mock.Anything, "p1").Return(nil)

	body := map[string]any{"PersonID": "p1"}
	err := svc.DeleteFromModifyPacket(ctx, nil, body)

	require.NoError(t, err)
	persons.AssertExpectations(t)
}

func TestDeleteFromModifyPacket_ErrorOnMissingPersonID(t *testing.T) {
	ctx := context.Background()
	svc := NewPersonService(new(mockPersonStore), new(mockPersonDNStore), new(mockPersonAccountStore))

	body := map[string]any{}
	err := svc.DeleteFromModifyPacket(ctx, nil, body)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PersonID is required")
}

func TestDeleteFromModifyPacket_PropagatesStoreError(t *testing.T) {
	ctx := context.Background()
	persons := new(mockPersonStore)
	svc := NewPersonService(persons, new(mockPersonDNStore), new(mockPersonAccountStore))

	persons.On("Delete", ctx, mock.Anything, "p1").Return(errors.New("db error"))

	body := map[string]any{"PersonID": "p1"}
	err := svc.DeleteFromModifyPacket(ctx, nil, body)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	persons.AssertExpectations(t)
}
