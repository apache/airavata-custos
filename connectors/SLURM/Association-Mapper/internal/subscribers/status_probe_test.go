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
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// probeFakeClient records every method invoked so tests can prove the probe
// never writes to the cluster.
type probeFakeClient struct {
	mu       sync.Mutex
	existing []client.Association
	listErr  error
	calls    []string
}

func (f *probeFakeClient) record(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, name)
}

func (f *probeFakeClient) ListAssociations(client.AssocFilter) ([]client.Association, error) {
	f.record("ListAssociations")
	return f.existing, f.listErr
}

func (f *probeFakeClient) CreateAccount(client.Account, string) error {
	f.record("CreateAccount")
	return nil
}

func (f *probeFakeClient) UpsertAssociation(client.Association) error {
	f.record("UpsertAssociation")
	return nil
}

func (f *probeFakeClient) DeleteAssociation(client.AssocFilter) error {
	f.record("DeleteAssociation")
	return nil
}

type recordedCheck struct {
	allocationID string
	userID       string
	checkType    models.AccessCheckType
	ok           bool
	detail       string
}

// probeMock builds a core mock for one cluster user with one ACTIVE
// membership, capturing every recorded check result.
func probeMock(o mockOpts, results *[]recordedCheck) *service.CoreServiceMock {
	m := coreMock(o)
	memberships := o.memberships
	if memberships == nil {
		memberships = []models.ComputeAllocationMembership{
			{ID: "mem-1", ComputeAllocationID: "alloc-1", UserID: "user-1", MembershipStatus: models.ACTIVE},
		}
	}
	m.ListComputeClusterUsersByClusterFunc = func(ctx context.Context, clusterID string) ([]models.ComputeClusterUser, error) {
		return []models.ComputeClusterUser{
			{ID: "csu-1", ComputeClusterID: clusterID, UserID: "user-1", LocalUsername: "testuser", ProvisionedAt: o.provisionedAt},
		}, nil
	}
	m.ListAllocationsForUserFunc = func(ctx context.Context, userID string) ([]models.ComputeAllocationMembership, error) {
		return memberships, nil
	}
	m.RecordAccessCheckResultFunc = func(ctx context.Context, allocationID, userID string, checkType models.AccessCheckType, ok bool, detail string) error {
		*results = append(*results, recordedCheck{allocationID, userID, checkType, ok, detail})
		return nil
	}
	return m
}

func provisionedLongAgo() *time.Time {
	t := time.Now().Add(-time.Hour)
	return &t
}

func runProbe(t *testing.T, fake *probeFakeClient, mock *service.CoreServiceMock) {
	t.Helper()
	sub := NewAssociationSubscriber(fake, nil, mock, 0, 0)
	sub.probeOnce(context.Background())
}

func assertReadOnly(t *testing.T, fake *probeFakeClient) {
	t.Helper()
	for _, call := range fake.calls {
		if call != "ListAssociations" {
			t.Fatalf("probe wrote to the cluster: %v", fake.calls)
		}
	}
}

func TestStatusProbe_AssociationPresentRecordsOK(t *testing.T) {
	var results []recordedCheck
	fake := &probeFakeClient{existing: []client.Association{
		{Account: "test-alloc", Cluster: "testcluster", User: "testuser", Partition: "compute"},
	}}
	runProbe(t, fake, probeMock(mockOpts{provisionedAt: provisionedLongAgo()}, &results))

	if len(results) != 1 {
		t.Fatalf("results: got %+v, want 1", results)
	}
	r := results[0]
	if !r.ok || r.checkType != models.AccessCheckJobSubmission || r.allocationID != "alloc-1" || r.userID != "user-1" {
		t.Errorf("recorded: %+v, want ok JOB_SUBMISSION for alloc-1/user-1", r)
	}
	assertReadOnly(t, fake)
}

func TestStatusProbe_AssociationMissingRecordsFailure(t *testing.T) {
	var results []recordedCheck
	fake := &probeFakeClient{} // cluster has no associations at all
	runProbe(t, fake, probeMock(mockOpts{provisionedAt: provisionedLongAgo()}, &results))

	if len(results) != 1 || results[0].ok || results[0].detail != "association missing" {
		t.Fatalf("recorded: %+v, want one failure 'association missing'", results)
	}
	assertReadOnly(t, fake)
}

func TestStatusProbe_ClusterAPIDownRecordsFailure(t *testing.T) {
	var results []recordedCheck
	fake := &probeFakeClient{listErr: errors.New("connection refused")}
	runProbe(t, fake, probeMock(mockOpts{provisionedAt: provisionedLongAgo()}, &results))

	if len(results) != 1 || results[0].ok || results[0].detail != "cluster API unreachable" {
		t.Fatalf("recorded: %+v, want one failure 'cluster API unreachable'", results)
	}
	assertReadOnly(t, fake)
}

func TestStatusProbe_UnprovisionedRecordsWaiting(t *testing.T) {
	var results []recordedCheck
	fake := &probeFakeClient{}
	runProbe(t, fake, probeMock(mockOpts{provisionedAt: nil}, &results))

	if len(results) != 1 || results[0].ok || results[0].detail != "waiting for the cluster account" {
		t.Fatalf("recorded: %+v, want one failure 'waiting for the cluster account'", results)
	}
	assertReadOnly(t, fake)
}

func TestStatusProbe_SkipsInactiveMembershipAndResourcelessAllocation(t *testing.T) {
	var results []recordedCheck
	fake := &probeFakeClient{}
	mock := probeMock(mockOpts{
		provisionedAt: provisionedLongAgo(),
		memberships: []models.ComputeAllocationMembership{
			{ID: "mem-1", ComputeAllocationID: "alloc-1", UserID: "user-1", MembershipStatus: models.INACTIVE},
		},
	}, &results)
	runProbe(t, fake, mock)
	if len(results) != 0 {
		t.Fatalf("inactive membership: recorded %+v, want nothing", results)
	}

	// An allocation with no resources leaves nothing to verify.
	results = nil
	mock = probeMock(mockOpts{
		provisionedAt: provisionedLongAgo(),
		resources:     []models.ComputeAllocationResource{},
	}, &results)
	// coreMock defaults nil resources; force empty explicitly.
	mock.ListResourcesForAllocationFunc = func(ctx context.Context, allocationID string) ([]models.ComputeAllocationResource, error) {
		return nil, nil
	}
	runProbe(t, fake, mock)
	if len(results) != 0 {
		t.Fatalf("resourceless allocation: recorded %+v, want nothing", results)
	}
	assertReadOnly(t, fake)
}

func TestStatusProbe_UnprovisionedOnAnotherClusterNotRecordedHere(t *testing.T) {
	var results []recordedCheck
	fake := &probeFakeClient{}
	mock := probeMock(mockOpts{provisionedAt: nil}, &results)
	// The membership's allocation lives on a different cluster than the one
	// being probed (coreMock probes cluster-1).
	mock.GetComputeAllocationFunc = func(ctx context.Context, id string) (*models.ComputeAllocation, error) {
		return &models.ComputeAllocation{ID: id, Name: "other-alloc", ComputeClusterID: "cluster-other", Status: models.ACTIVE}, nil
	}
	runProbe(t, fake, mock)
	if len(results) != 0 {
		t.Fatalf("cross-cluster unprovisioned membership recorded %+v, want nothing", results)
	}
}

func TestStatusProbe_CoreLookupFailureRecordsNothing(t *testing.T) {
	var results []recordedCheck
	fake := &probeFakeClient{}
	mock := probeMock(mockOpts{provisionedAt: provisionedLongAgo()}, &results)
	mock.GetComputeAllocationFunc = func(ctx context.Context, id string) (*models.ComputeAllocation, error) {
		return nil, errors.New("db down")
	}
	runProbe(t, fake, mock)
	if len(results) != 0 {
		t.Fatalf("core failure: recorded %+v, want nothing (unknowable, not unhealthy)", results)
	}
}

func TestStatusProbe_PassSurvivesPanic(t *testing.T) {
	var results []recordedCheck
	fake := &probeFakeClient{}
	mock := probeMock(mockOpts{provisionedAt: provisionedLongAgo()}, &results)
	mock.ListComputeClusterUsersByClusterFunc = func(ctx context.Context, clusterID string) ([]models.ComputeClusterUser, error) {
		panic("boom")
	}
	// Must not propagate: the shared goroutine also runs the reconciler's peer.
	runProbe(t, fake, mock)
}
