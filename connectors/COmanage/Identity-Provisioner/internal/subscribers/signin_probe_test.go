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

package subscribers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type fakeRegistry struct {
	err   error
	calls int
}

func (f *fakeRegistry) GetPersonComposite(string) (json.RawMessage, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return json.RawMessage(`{}`), nil
}

type signInRecorded struct {
	allocationID string
	userID       string
	checkType    models.AccessCheckType
	ok           bool
	detail       string
}

type signInMockOpts struct {
	provisioned  bool
	linked       bool
	memberships  []models.ComputeAllocationMembership
	allocCluster string
}

func signInCoreMock(o signInMockOpts, results *[]signInRecorded) *service.CoreServiceMock {
	provisionedAt := (*time.Time)(nil)
	if o.provisioned {
		t := time.Now().Add(-time.Hour)
		provisionedAt = &t
	}
	memberships := o.memberships
	if memberships == nil {
		memberships = []models.ComputeAllocationMembership{
			{ID: "mem-1", ComputeAllocationID: "alloc-1", UserID: "user-1", MembershipStatus: models.ACTIVE},
		}
	}
	allocCluster := o.allocCluster
	if allocCluster == "" {
		allocCluster = "cluster-1"
	}
	return &service.CoreServiceMock{
		ListComputeClusterUsersByClusterFunc: func(ctx context.Context, clusterID string) ([]models.ComputeClusterUser, error) {
			return []models.ComputeClusterUser{
				{ID: "csu-1", ComputeClusterID: clusterID, UserID: "user-1", LocalUsername: "testuser", ProvisionedAt: provisionedAt},
			}, nil
		},
		ListAllocationsForUserFunc: func(ctx context.Context, userID string) ([]models.ComputeAllocationMembership, error) {
			return memberships, nil
		},
		GetComputeAllocationFunc: func(ctx context.Context, id string) (*models.ComputeAllocation, error) {
			return &models.ComputeAllocation{ID: id, Name: "alloc", ComputeClusterID: allocCluster, Status: models.ACTIVE}, nil
		},
		ListUserIdentitiesForUserFunc: func(ctx context.Context, userID string) ([]models.UserIdentity, error) {
			if !o.linked {
				return nil, nil
			}
			return []models.UserIdentity{{UserID: userID, Source: "comanage", ExternalID: "person-9"}}, nil
		},
		RecordAccessCheckResultFunc: func(ctx context.Context, allocationID, userID string, checkType models.AccessCheckType, ok bool, detail string) error {
			*results = append(*results, signInRecorded{allocationID, userID, checkType, ok, detail})
			return nil
		},
	}
}

func TestSignInProbe_ProvisionedAndResolvableRecordsOK(t *testing.T) {
	var results []signInRecorded
	reg := &fakeRegistry{}
	NewSignInProbe(reg, signInCoreMock(signInMockOpts{provisioned: true, linked: true}, &results), "cluster-1").probeOnce(context.Background())

	if len(results) != 1 {
		t.Fatalf("results: got %+v, want 1", results)
	}
	r := results[0]
	if !r.ok || r.checkType != models.AccessCheckSignIn || r.allocationID != "alloc-1" {
		t.Errorf("recorded: %+v, want ok SIGN_IN for alloc-1", r)
	}
}

func TestSignInProbe_UnprovisionedRecordsNotProvisioned(t *testing.T) {
	var results []signInRecorded
	reg := &fakeRegistry{}
	NewSignInProbe(reg, signInCoreMock(signInMockOpts{provisioned: false, linked: true}, &results), "cluster-1").probeOnce(context.Background())

	if len(results) != 1 || results[0].ok || results[0].detail != "account not provisioned yet" {
		t.Fatalf("recorded: %+v, want failure 'account not provisioned yet'", results)
	}
	if reg.calls != 0 {
		t.Errorf("registry consulted for an unprovisioned account: %d calls", reg.calls)
	}
}

func TestSignInProbe_RegistryOutcomes(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		detail string
	}{
		{"missing person", client.ErrNotFound, "registry record missing"},
		{"registry down", errors.New("http 503"), "registry unreachable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var results []signInRecorded
			reg := &fakeRegistry{err: tc.err}
			NewSignInProbe(reg, signInCoreMock(signInMockOpts{provisioned: true, linked: true}, &results), "cluster-1").probeOnce(context.Background())
			if len(results) != 1 || results[0].ok || results[0].detail != tc.detail {
				t.Fatalf("recorded: %+v, want failure %q", results, tc.detail)
			}
		})
	}
}

func TestSignInProbe_UnlinkedRecordsNotLinked(t *testing.T) {
	var results []signInRecorded
	reg := &fakeRegistry{}
	NewSignInProbe(reg, signInCoreMock(signInMockOpts{provisioned: true, linked: false}, &results), "cluster-1").probeOnce(context.Background())

	if len(results) != 1 || results[0].ok || results[0].detail != "registry record not linked" {
		t.Fatalf("recorded: %+v, want failure 'registry record not linked'", results)
	}
	if reg.calls != 0 {
		t.Errorf("registry consulted without a stored person id: %d calls", reg.calls)
	}
}

func TestSignInProbe_OneLookupSharedAcrossMemberships(t *testing.T) {
	var results []signInRecorded
	reg := &fakeRegistry{}
	mock := signInCoreMock(signInMockOpts{
		provisioned: true, linked: true,
		memberships: []models.ComputeAllocationMembership{
			{ID: "mem-1", ComputeAllocationID: "alloc-1", UserID: "user-1", MembershipStatus: models.ACTIVE},
			{ID: "mem-2", ComputeAllocationID: "alloc-2", UserID: "user-1", MembershipStatus: models.ACTIVE},
			{ID: "mem-3", ComputeAllocationID: "alloc-3", UserID: "user-1", MembershipStatus: models.INACTIVE},
		},
	}, &results)
	NewSignInProbe(reg, mock, "cluster-1").probeOnce(context.Background())

	if len(results) != 2 {
		t.Fatalf("results: got %+v, want 2 (both ACTIVE memberships)", results)
	}
	if reg.calls != 1 {
		t.Errorf("registry lookups: got %d, want 1 shared", reg.calls)
	}
}

func TestSignInProbe_SkipsOtherClustersAllocations(t *testing.T) {
	var results []signInRecorded
	reg := &fakeRegistry{}
	mock := signInCoreMock(signInMockOpts{provisioned: true, linked: true, allocCluster: "cluster-other"}, &results)
	NewSignInProbe(reg, mock, "cluster-1").probeOnce(context.Background())

	if len(results) != 0 {
		t.Fatalf("recorded for another cluster's allocation: %+v", results)
	}
}

func TestSignInProbe_CoreLookupFailureRecordsNothing(t *testing.T) {
	var results []signInRecorded
	reg := &fakeRegistry{}
	mock := signInCoreMock(signInMockOpts{provisioned: true, linked: true}, &results)
	mock.ListUserIdentitiesForUserFunc = func(ctx context.Context, userID string) ([]models.UserIdentity, error) {
		return nil, errors.New("db down")
	}
	NewSignInProbe(reg, mock, "cluster-1").probeOnce(context.Background())

	if len(results) != 0 {
		t.Fatalf("core failure recorded %+v, want nothing (unknowable, not unhealthy)", results)
	}
}

func TestSignInProbe_SkipsInactiveAllocation(t *testing.T) {
	var results []signInRecorded
	reg := &fakeRegistry{}
	mock := signInCoreMock(signInMockOpts{provisioned: true, linked: true}, &results)
	mock.GetComputeAllocationFunc = func(ctx context.Context, id string) (*models.ComputeAllocation, error) {
		return &models.ComputeAllocation{ID: id, Name: "alloc", ComputeClusterID: "cluster-1", Status: models.INACTIVE}, nil
	}
	NewSignInProbe(reg, mock, "cluster-1").probeOnce(context.Background())

	if len(results) != 0 {
		t.Fatalf("inactive allocation recorded %+v, want nothing", results)
	}
}

func TestSignInProbe_PassSurvivesPanic(t *testing.T) {
	var results []signInRecorded
	mock := signInCoreMock(signInMockOpts{provisioned: true, linked: true}, &results)
	mock.ListComputeClusterUsersByClusterFunc = func(ctx context.Context, clusterID string) ([]models.ComputeClusterUser, error) {
		panic("boom")
	}
	NewSignInProbe(&fakeRegistry{}, mock, "cluster-1").probeOnce(context.Background())
}
