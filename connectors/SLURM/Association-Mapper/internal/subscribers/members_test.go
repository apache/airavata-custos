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
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type fakeSlurmClient struct {
	mu           sync.Mutex
	upserts      []client.Association
	defaults     []string
	deletes      []client.AssocFilter
	existing     []client.Association
	deletedUsers []string
	qos          []client.QoS
	qosPosts     []client.QoS
	listErr      error
	createErr    error    // fails every allocation account's creation
	writes       []string // creates and upserts, in order
}

func (f *fakeSlurmClient) ListQoS() ([]client.QoS, error) { return f.qos, f.listErr }

func (f *fakeSlurmClient) PostQoS(q client.QoS) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.qosPosts = append(f.qosPosts, q)
	return nil
}

func (f *fakeSlurmClient) DeleteAssociation(filter client.AssocFilter) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deletes = append(f.deletes, filter)
	return nil
}

func (f *fakeSlurmClient) DeleteUser(name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deletedUsers = append(f.deletedUsers, name)
	return nil
}

func (f *fakeSlurmClient) allDeletes() []client.AssocFilter {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]client.AssocFilter(nil), f.deletes...)
}

func (f *fakeSlurmClient) ListAssociations(client.AssocFilter) ([]client.Association, error) {
	return f.existing, f.listErr
}

func (f *fakeSlurmClient) ListPartitions(string) ([]client.Partition, error) { return nil, nil }

func (f *fakeSlurmClient) CreateAccount(a client.Account, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if a.Name != noAllocation && f.createErr != nil {
		return f.createErr
	}
	f.writes = append(f.writes, "create "+a.Name)
	return nil
}

func (f *fakeSlurmClient) SetDefaultAccount(_, user, account string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.defaults = append(f.defaults, user+"="+account)
	return nil
}

func (f *fakeSlurmClient) UpsertAssociation(a client.Association) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.upserts = append(f.upserts, a)
	f.writes = append(f.writes, "upsert "+a.Account+"/"+a.User)
	return nil
}

func (f *fakeSlurmClient) allUpserts() []client.Association {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]client.Association(nil), f.upserts...)
}

type mockOpts struct {
	allocationName    string
	allocationStatus  models.AllocationStatus
	unmanagedAccounts bool
	provisionedAt     *time.Time
	resources         []models.ComputeAllocationResource
	overrides         []models.ComputeAllocationMembershipResourceOverride
	memberships       []models.ComputeAllocationMembership
	allocationPending bool
	membershipPending bool
	qos               []models.ComputeClusterQoS
	allocationQoS     []string
}

func coreMock(o mockOpts) *service.CoreServiceMock {
	allocStatus := o.allocationStatus
	if allocStatus == "" {
		allocStatus = models.ACTIVE
	}
	resources := o.resources
	if resources == nil {
		resources = []models.ComputeAllocationResource{{ID: "res-1", Name: "compute", ResourceType: "cpu"}}
	}
	if o.memberships == nil {
		o.memberships = []models.ComputeAllocationMembership{testMembership()}
	}
	allocName := o.allocationName
	if allocName == "" {
		allocName = "test-alloc"
	}
	return &service.CoreServiceMock{
		GetComputeAllocationFunc: func(ctx context.Context, id string) (*models.ComputeAllocation, error) {
			alloc := &models.ComputeAllocation{ID: id, Name: allocName, ComputeClusterID: "cluster-1", Status: allocStatus}
			if o.allocationPending {
				alloc.SlurmPendingSince = ago(time.Minute)
			}
			return alloc, nil
		},
		GetProjectFunc: func(ctx context.Context, id string) (*models.Project, error) {
			return &models.Project{ID: id}, nil
		},
		GetOrganizationFunc: func(ctx context.Context, id string) (*models.Organization, error) {
			return &models.Organization{ID: id, Name: "org"}, nil
		},
		ListComputeClusterQoSFunc: func(ctx context.Context, clusterID, allocationID string) ([]models.ComputeClusterQoS, error) {
			if allocationID == "" {
				return o.qos, nil
			}
			granted := make([]models.ComputeClusterQoS, len(o.allocationQoS))
			for i, name := range o.allocationQoS {
				granted[i].Name = name
			}
			return granted, nil
		},
		ListResourceMappingsForAllocationFunc: func(ctx context.Context, allocationID string) ([]models.ComputeAllocationResourceMapping, error) {
			return nil, nil
		},
		SetSlurmPendingFunc: func(context.Context, service.SlurmPendingKind, string, *time.Time) error { return nil },
		GetComputeClusterFunc: func(ctx context.Context, id string) (*models.ComputeCluster, error) {
			return &models.ComputeCluster{ID: id, Name: "testcluster"}, nil
		},
		GetUserFunc: func(ctx context.Context, id string) (*models.User, error) {
			return &models.User{ID: id}, nil
		},
		GetComputeAllocationResourceFunc: func(ctx context.Context, id string) (*models.ComputeAllocationResource, error) {
			for _, r := range resources {
				if r.ID == id {
					return &r, nil
				}
			}
			return nil, service.ErrNotFound
		},
		ListOverridesForMembershipFunc: func(ctx context.Context, membershipID string) ([]models.ComputeAllocationMembershipResourceOverride, error) {
			return o.overrides, nil
		},
		ListComputeAllocationsByClusterFunc: func(ctx context.Context, clusterID string) ([]models.ComputeAllocation, error) {
			if o.unmanagedAccounts {
				return nil, nil
			}
			alloc := models.ComputeAllocation{ID: "alloc-1", Name: allocName, ComputeClusterID: clusterID, Status: allocStatus}
			if o.allocationPending {
				alloc.SlurmPendingSince = ago(time.Minute)
			}
			return []models.ComputeAllocation{alloc}, nil
		},
		ListComputeClustersFunc: func(ctx context.Context) ([]models.ComputeCluster, error) {
			return []models.ComputeCluster{{ID: "cluster-1", Name: "testcluster"}}, nil
		},
		ListMembersForAllocationFunc: func(ctx context.Context, allocationID string) ([]store.MembershipWithUser, error) {
			out := make([]store.MembershipWithUser, 0, len(o.memberships))
			for _, m := range o.memberships {
				if m.ComputeAllocationID != allocationID {
					continue
				}
				if o.membershipPending {
					m.SlurmPendingSince = ago(time.Minute)
				}
				out = append(out, store.MembershipWithUser{ComputeAllocationMembership: m, LocalUsername: "testuser", ProvisionedAt: o.provisionedAt})
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

// cleared returns the IDs whose pending mark of the kind the sync cleared.
func cleared(core *service.CoreServiceMock, kind service.SlurmPendingKind) []string {
	var ids []string
	for _, c := range core.SetSlurmPendingCalls() {
		if c.Kind == kind && c.ClearSince != nil {
			ids = append(ids, c.ID)
		}
	}
	return ids
}
