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
	"sync"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type fakeSlurmClient struct {
	mu       sync.Mutex
	upserts  []client.Association
	deletes  []client.AssocFilter
	existing []client.Association
}

func (f *fakeSlurmClient) DeleteAssociation(filter client.AssocFilter) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deletes = append(f.deletes, filter)
	return nil
}

func (f *fakeSlurmClient) allDeletes() []client.AssocFilter {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]client.AssocFilter(nil), f.deletes...)
}

func (f *fakeSlurmClient) ListAssociations(client.AssocFilter) ([]client.Association, error) {
	return f.existing, nil
}

func (f *fakeSlurmClient) CreateAccount(client.Account, string) error { return nil }

func (f *fakeSlurmClient) UpsertAssociation(a client.Association) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.upserts = append(f.upserts, a)
	return nil
}

func (f *fakeSlurmClient) all() []client.Association {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]client.Association(nil), f.upserts...)
}

type mockOpts struct {
	provisionedAt *time.Time
	resources     []models.ComputeAllocationResource
	overrides     []models.ComputeAllocationMembershipResourceOverride
	memberships   []models.ComputeAllocationMembership
}

func coreMock(o mockOpts) *service.CoreServiceMock {
	resources := o.resources
	if resources == nil {
		resources = []models.ComputeAllocationResource{{ID: "res-1", Name: "compute", ResourceType: "cpu"}}
	}
	return &service.CoreServiceMock{
		GetComputeAllocationFunc: func(ctx context.Context, id string) (*models.ComputeAllocation, error) {
			return &models.ComputeAllocation{ID: id, Name: "test-alloc", ComputeClusterID: "cluster-1"}, nil
		},
		GetComputeClusterFunc: func(ctx context.Context, id string) (*models.ComputeCluster, error) {
			return &models.ComputeCluster{ID: id, Name: "testcluster"}, nil
		},
		GetUserFunc: func(ctx context.Context, id string) (*models.User, error) {
			return &models.User{ID: id}, nil
		},
		ListResourcesForAllocationFunc: func(ctx context.Context, allocationID string) ([]models.ComputeAllocationResource, error) {
			return resources, nil
		},
		ListOverridesForMembershipFunc: func(ctx context.Context, membershipID string) ([]models.ComputeAllocationMembershipResourceOverride, error) {
			return o.overrides, nil
		},
		GetComputeClusterUserByPairFunc: func(ctx context.Context, clusterID, userID string) (*models.ComputeClusterUser, error) {
			return &models.ComputeClusterUser{
				ID: "csu-1", ComputeClusterID: clusterID, UserID: userID,
				LocalUsername: "testuser", ProvisionedAt: o.provisionedAt,
			}, nil
		},
		ListComputeClustersFunc: func(ctx context.Context) ([]models.ComputeCluster, error) {
			return []models.ComputeCluster{{ID: "cluster-1", Name: "testcluster"}}, nil
		},
		ListComputeClusterUsersByClusterFunc: func(ctx context.Context, clusterID string) ([]models.ComputeClusterUser, error) {
			return []models.ComputeClusterUser{{
				ID: "csu-1", ComputeClusterID: clusterID, UserID: "user-1",
				LocalUsername: "testuser", ProvisionedAt: o.provisionedAt,
			}}, nil
		},
		ListAllocationsForUserFunc: func(ctx context.Context, userID string) ([]models.ComputeAllocationMembership, error) {
			return o.memberships, nil
		},
		ListMembersForAllocationFunc: func(ctx context.Context, allocationID string) ([]store.MembershipWithUser, error) {
			out := make([]store.MembershipWithUser, 0, len(o.memberships))
			for _, m := range o.memberships {
				out = append(out, store.MembershipWithUser{ComputeAllocationMembership: m})
			}
			return out, nil
		},
		CreateAuditEventFunc: func(ctx context.Context, event *models.AuditEvent) (*models.AuditEvent, error) {
			return event, nil
		},
	}
}

func ago(d time.Duration) *time.Time {
	t := time.Now().Add(-d)
	return &t
}

func testMembership() models.ComputeAllocationMembership {
	return models.ComputeAllocationMembership{
		ID: "mem-1", ComputeAllocationID: "alloc-1", UserID: "user-1",
		MembershipStatus: models.ACTIVE,
	}
}

func TestMembershipCreationWritesAssociationWhenProvisioned(t *testing.T) {
	core := coreMock(mockOpts{provisionedAt: ago(time.Minute)})
	slurm := &fakeSlurmClient{}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).
		SubscribeToComputeAllocationMembershipCreation(context.Background(), testMembership())

	got := slurm.all()
	if len(got) != 1 {
		t.Fatalf("expected 1 association, got %d", len(got))
	}
	if got[0].User != "testuser" || got[0].Account != "test-alloc" || got[0].Partition != "compute" {
		t.Errorf("unexpected association: %+v", got[0])
	}
}

// An unprovisioned account must NOT get an association: slurmctld would cache
// the failed uid lookup and reject the user's jobs. The reconciler picks it up.
func TestMembershipCreationDefersWhenNotProvisioned(t *testing.T) {
	core := coreMock(mockOpts{provisionedAt: nil})
	slurm := &fakeSlurmClient{}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).
		SubscribeToComputeAllocationMembershipCreation(context.Background(), testMembership())

	if n := len(slurm.all()); n != 0 {
		t.Fatalf("expected no association for an unprovisioned account, got %d", n)
	}
}

// Regression: the membership path once wrote a record with no limits, and
// Slurm's last-write-wins upsert erased limits a per-member override had set.
func TestMembershipCreationKeepsOverrideLimits(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		overrides: []models.ComputeAllocationMembershipResourceOverride{{
			ID: "ovr-1", ComputeAllocationMembershipID: "mem-1", ComputeAllocationResourceID: "res-1",
			OverrideResourceAmount: 4, OverrideResourceTime: 120,
		}},
	})
	slurm := &fakeSlurmClient{}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).
		SubscribeToComputeAllocationMembershipCreation(context.Background(), testMembership())

	got := slurm.all()
	if len(got) != 1 {
		t.Fatalf("expected 1 association, got %d", len(got))
	}
	if len(got[0].Limits.GrpTRES) != 1 || got[0].Limits.GrpTRES[0].Count != 4 {
		t.Errorf("override amount not carried into the association: %+v", got[0].Limits)
	}
	if len(got[0].Limits.GrpTRESMins) != 1 || got[0].Limits.GrpTRESMins[0].Count != 120 {
		t.Errorf("override time not carried into the association: %+v", got[0].Limits)
	}
}

// A GPU resource is stored joined as "gres/gpu" but SLURM wants the TRES split
// into type and name. Sending "gres/gpu" as the type yields an invalid limit.
func TestMembershipCreationSplitsGpuTres(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		resources: []models.ComputeAllocationResource{
			{ID: "res-1", Name: "gpu", ResourceType: "gres/gpu"},
		},
		overrides: []models.ComputeAllocationMembershipResourceOverride{{
			ID: "ovr-1", ComputeAllocationMembershipID: "mem-1", ComputeAllocationResourceID: "res-1",
			OverrideResourceAmount: 2, OverrideResourceTime: 120,
		}},
	})
	slurm := &fakeSlurmClient{}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).
		SubscribeToComputeAllocationMembershipCreation(context.Background(), testMembership())

	got := slurm.all()
	if len(got) != 1 {
		t.Fatalf("expected 1 association, got %d", len(got))
	}
	tres := got[0].Limits.GrpTRES
	if len(tres) != 1 || tres[0].Type != "gres" || tres[0].Name != "gpu" || tres[0].Count != 2 {
		t.Errorf("expected GrpTRES {gres, gpu, 2}, got %+v", tres)
	}
}

// An allocation with no resources has no partition to map onto; skipping beats
// panicking or inventing a partition name.
func TestMembershipCreationSkipsWhenAllocationHasNoResources(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		resources:     []models.ComputeAllocationResource{},
	})
	slurm := &fakeSlurmClient{}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).
		SubscribeToComputeAllocationMembershipCreation(context.Background(), testMembership())

	if n := len(slurm.all()); n != 0 {
		t.Fatalf("expected no association when the allocation has no resources, got %d", n)
	}
}

func TestMembershipCreationWritesOnePerResource(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		resources: []models.ComputeAllocationResource{
			{ID: "res-1", Name: "compute", ResourceType: "cpu"},
			{ID: "res-2", Name: "gpu", ResourceType: "gres/gpu"},
		},
	})
	slurm := &fakeSlurmClient{}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).
		SubscribeToComputeAllocationMembershipCreation(context.Background(), testMembership())

	if n := len(slurm.all()); n != 2 {
		t.Fatalf("expected an association per resource, got %d", n)
	}
}

func TestReconcilerWritesAssociationsForProvisionedMembers(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		memberships:   []models.ComputeAllocationMembership{testMembership()},
	})
	slurm := &fakeSlurmClient{}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).reconcile(context.Background())

	if n := len(slurm.all()); n != 1 {
		t.Fatalf("expected the reconciler to write 1 association, got %d", n)
	}
}

func TestReconcilerSkipsUnprovisionedAndTooFreshAccounts(t *testing.T) {
	cases := map[string]*time.Time{
		"never provisioned": nil,
		"inside grace":      ago(time.Second),
	}
	for name, provisioned := range cases {
		t.Run(name, func(t *testing.T) {
			core := coreMock(mockOpts{
				provisionedAt: provisioned,
				memberships:   []models.ComputeAllocationMembership{testMembership()},
			})
			slurm := &fakeSlurmClient{}
			NewAssociationSubscriber(slurm, nil, core, 0, 0).reconcile(context.Background())

			if n := len(slurm.all()); n != 0 {
				t.Fatalf("expected no association, got %d", n)
			}
		})
	}
}

func TestReconcilerSkipsInactiveMemberships(t *testing.T) {
	inactive := testMembership()
	inactive.MembershipStatus = models.INACTIVE
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		memberships:   []models.ComputeAllocationMembership{inactive},
	})
	slurm := &fakeSlurmClient{}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).reconcile(context.Background())

	if n := len(slurm.all()); n != 0 {
		t.Fatalf("expected no association for an inactive membership, got %d", n)
	}
}

// The sweep reads what the cluster already has, so a steady-state pass writes
// nothing instead of re-declaring every member every interval.
func TestReconcilerSkipsAssociationsAlreadyCorrect(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		memberships:   []models.ComputeAllocationMembership{testMembership()},
	})
	slurm := &fakeSlurmClient{existing: []client.Association{{
		Account: "test-alloc", Cluster: "testcluster", User: "testuser", Partition: "compute",
	}}}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).reconcile(context.Background())

	if n := len(slurm.all()); n != 0 {
		t.Fatalf("expected no writes when the association already matches, got %d", n)
	}
}

// Drifted limits are repaired: same key, different limits, so it rewrites.
func TestReconcilerRewritesAssociationWithDriftedLimits(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		memberships:   []models.ComputeAllocationMembership{testMembership()},
		overrides: []models.ComputeAllocationMembershipResourceOverride{{
			ID: "ovr-1", ComputeAllocationMembershipID: "mem-1", ComputeAllocationResourceID: "res-1",
			OverrideResourceAmount: 4,
		}},
	})
	// Cluster has the association but without the override's limits.
	slurm := &fakeSlurmClient{existing: []client.Association{{
		Account: "test-alloc", Cluster: "testcluster", User: "testuser", Partition: "compute",
	}}}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).reconcile(context.Background())

	got := slurm.all()
	if len(got) != 1 {
		t.Fatalf("expected the drifted association to be rewritten, got %d writes", len(got))
	}
	if len(got[0].Limits.GrpTRES) != 1 || got[0].Limits.GrpTRES[0].Count != 4 {
		t.Errorf("rewrite did not restore the override limits: %+v", got[0].Limits)
	}
}

// A missing association is still written even though others already exist.
func TestReconcilerWritesOnlyTheMissingAssociation(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		memberships:   []models.ComputeAllocationMembership{testMembership()},
		resources: []models.ComputeAllocationResource{
			{ID: "res-1", Name: "compute", ResourceType: "cpu"},
			{ID: "res-2", Name: "gpu", ResourceType: "gres/gpu"},
		},
	})
	slurm := &fakeSlurmClient{existing: []client.Association{{
		Account: "test-alloc", Cluster: "testcluster", User: "testuser", Partition: "compute",
	}}}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).reconcile(context.Background())

	got := slurm.all()
	if len(got) != 1 {
		t.Fatalf("expected only the missing partition to be written, got %d", len(got))
	}
	if got[0].Partition != "gpu" {
		t.Errorf("expected the gpu association to be written, got %q", got[0].Partition)
	}
}
