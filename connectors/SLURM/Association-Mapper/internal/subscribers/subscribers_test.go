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
	"cmp"
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// fakeSlurm refuses to remove a busy user's associations, as Slurm does while jobs run.
type fakeSlurm struct {
	mu       sync.Mutex
	assocs   []client.Association
	accounts map[string]string
	busy     map[string]bool
	clusters []string
	writes   int
}

func newFakeSlurm(assocs ...client.Association) *fakeSlurm {
	return &fakeSlurm{assocs: assocs, accounts: map[string]string{}, busy: map[string]bool{}, clusters: []string{"testcluster"}}
}

func matches(a client.Association, f client.AssocFilter) bool {
	return (f.Cluster == "" || a.Cluster == f.Cluster) && (f.Account == "" || a.Account == f.Account) &&
		(f.User == "" || a.User == f.User) && (f.Partition == "" || a.Partition == f.Partition)
}

func (f *fakeSlurm) upsert(a client.Association) {
	f.writes++
	f.assocs = slices.DeleteFunc(f.assocs, func(x client.Association) bool { return keyOf(x) == keyOf(a) && x.Cluster == a.Cluster })
	f.assocs = append(f.assocs, a)
}

func (f *fakeSlurm) CreateAccount(a client.Account, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.accounts[a.Name] = a.Organization
	return nil
}

func (f *fakeSlurm) ListAccounts() ([]client.Account, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []client.Account
	for name, org := range f.accounts {
		out = append(out, client.Account{Name: name, Organization: org})
	}
	return out, nil
}

func (f *fakeSlurm) SetDefaultAccount(cluster, user, account string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !slices.ContainsFunc(f.assocs, func(x client.Association) bool { return x.User == user && x.Account == account }) {
		f.upsert(client.Association{Account: account, Cluster: cluster, User: user})
	}
	return nil
}

func (f *fakeSlurm) DeleteUser(name string) error {
	return f.DeleteAssociation(client.AssocFilter{User: name})
}

func (f *fakeSlurm) ListAssociations(filter client.AssocFilter) ([]client.Association, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.DeleteFunc(slices.Clone(f.assocs), func(a client.Association) bool { return !matches(a, filter) }), nil
}

func (f *fakeSlurm) UpsertAssociation(a client.Association) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.upsert(a)
	return nil
}

func (f *fakeSlurm) DeleteAssociation(filter client.AssocFilter) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.busy[filter.User] {
		return errors.New("slurmrestd 500: code=7001")
	}
	f.assocs = slices.DeleteFunc(f.assocs, func(a client.Association) bool { return a.User != "" && matches(a, filter) })
	return nil
}

func (f *fakeSlurm) ListClusters() ([]string, error) { return f.clusters, nil }

func (f *fakeSlurm) held(user string) []string {
	got, _ := f.ListAssociations(client.AssocFilter{User: user})
	var out []string
	for _, a := range got {
		s := a.Account
		if a.Limits.GrpJobs != nil {
			s += " blocked"
		}
		out = append(out, s)
	}
	slices.Sort(out)
	return out
}

type mockOpts struct {
	allocationName   string
	allocationStatus models.AllocationStatus
	allocationGone   bool
	provisionedAt    *time.Time
	resources        []models.ComputeAllocationResource
	overrides        []models.ComputeAllocationMembershipResourceOverride
	memberships      []models.ComputeAllocationMembership
	mappings         []models.ComputeAllocationResourceMapping
}

func coreMock(o mockOpts) *service.CoreServiceMock {
	resources := o.resources
	if resources == nil {
		resources = []models.ComputeAllocationResource{{ID: "res-1", ResourceType: "cpu"}}
	}
	allocation := models.ComputeAllocation{ID: "alloc-1", Name: cmp.Or(o.allocationName, "test-alloc"), ComputeClusterID: "cluster-1", Status: cmp.Or(o.allocationStatus, models.ACTIVE)}
	account := models.ComputeClusterUser{ID: "csu-1", ComputeClusterID: "cluster-1", UserID: "user-1", LocalUsername: "testuser", ProvisionedAt: o.provisionedAt}
	return &service.CoreServiceMock{
		GetComputeAllocationFunc: func(context.Context, string) (*models.ComputeAllocation, error) {
			if o.allocationGone {
				return nil, service.ErrNotFound
			}
			return &allocation, nil
		},
		GetComputeClusterFunc: func(_ context.Context, id string) (*models.ComputeCluster, error) {
			return &models.ComputeCluster{ID: id, Name: "testcluster"}, nil
		},
		GetComputeAllocationResourceFunc: func(_ context.Context, id string) (*models.ComputeAllocationResource, error) {
			for _, r := range resources {
				if r.ID == id {
					return &r, nil
				}
			}
			return nil, service.ErrNotFound
		},
		ListResourcesForAllocationFunc: func(context.Context, string) ([]models.ComputeAllocationResource, error) {
			return resources, nil
		},
		ListOverridesForMembershipFunc: func(context.Context, string) ([]models.ComputeAllocationMembershipResourceOverride, error) {
			return o.overrides, nil
		},
		GetComputeClusterUserByPairFunc: func(_ context.Context, _, userID string) (*models.ComputeClusterUser, error) {
			if userID == "" {
				return nil, service.ErrInvalidInput
			}
			if o.provisionedAt == nil || userID != account.UserID {
				return nil, service.ErrNotFound
			}
			return &account, nil
		},
		GetComputeAllocationMembershipFunc: func(_ context.Context, id string) (*models.ComputeAllocationMembership, error) {
			for _, m := range o.memberships {
				if m.ID == id {
					return &m, nil
				}
			}
			return nil, service.ErrNotFound
		},
		ListComputeAllocationsByClusterFunc: func(context.Context, string) ([]models.ComputeAllocation, error) {
			if o.allocationGone {
				return nil, nil
			}
			return []models.ComputeAllocation{allocation}, nil
		},
		ListComputeClustersFunc: func(context.Context) ([]models.ComputeCluster, error) {
			return []models.ComputeCluster{{ID: "cluster-1", Name: "testcluster"}}, nil
		},
		ListComputeClusterUsersByClusterFunc: func(context.Context, string) ([]models.ComputeClusterUser, error) {
			if o.provisionedAt == nil {
				return nil, nil
			}
			return []models.ComputeClusterUser{account}, nil
		},
		ListAllocationsForUserFunc: func(context.Context, string) ([]models.ComputeAllocationMembership, error) {
			return o.memberships, nil
		},
		ListMembersForAllocationFunc: func(context.Context, string) ([]store.MembershipWithUser, error) {
			var out []store.MembershipWithUser
			for _, m := range o.memberships {
				out = append(out, store.MembershipWithUser{ComputeAllocationMembership: m, LocalUsername: "testuser", ProvisionedAt: o.provisionedAt})
			}
			return out, nil
		},
		CreateAuditEventFunc: func(_ context.Context, e *models.AuditEvent) (*models.AuditEvent, error) { return e, nil },
	}
}

type fakeMappings []models.ComputeAllocationResourceMapping

func (f fakeMappings) FindByPair(_ context.Context, _, resourceID string) (*models.ComputeAllocationResourceMapping, error) {
	for _, m := range f {
		if m.ComputeAllocationResourceID == resourceID {
			return &m, nil
		}
	}
	return nil, nil
}

func newSub(slurm SlurmClient, o mockOpts, grace time.Duration) *AssociationSubscriber {
	return NewAssociationSubscriber(slurm, nil, coreMock(o), fakeMappings(o.mappings), 0, grace)
}

func ago(d time.Duration) *time.Time {
	t := time.Now().Add(-d)
	return &t
}

func testMembership() models.ComputeAllocationMembership {
	return models.ComputeAllocationMembership{ID: "mem-1", ComputeAllocationID: "alloc-1", UserID: "user-1", MembershipStatus: models.ACTIVE}
}

var (
	member = []models.ComputeAllocationMembership{testMembership()}
	cpu8   = []models.ComputeAllocationResourceMapping{{ComputeAllocationResourceID: "res-1", ResourceAmount: 8}}
	cu     = models.ComputeClusterUser{ID: "csu-1", ComputeClusterID: "cluster-1", UserID: "user-1", LocalUsername: "testuser"}
)

func run(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// A second sync of an unchanged allocation writes nothing.
func TestAllocationChangeWritesAccountAndMembers(t *testing.T) {
	slurm := newFakeSlurm()
	s := newSub(slurm, mockOpts{provisionedAt: ago(time.Hour), memberships: member, mappings: cpu8}, 0)
	run(t, s.SubscribeToComputeAllocationChange(context.Background(), models.ComputeAllocation{ID: "alloc-1"}))
	records, _ := slurm.ListAssociations(client.AssocFilter{Account: "test-alloc"})
	i := slices.IndexFunc(records, func(a client.Association) bool { return a.User == "" })
	if got := slurm.held("testuser"); !slices.Equal(got, []string{"none", "test-alloc"}) || slurm.accounts["test-alloc"] != custos ||
		i < 0 || records[i].Limits.GrpTRES[0].Count != 8 {
		t.Fatalf("held=%v accounts=%v records=%+v", got, slurm.accounts, records)
	}
	writes := slurm.writes
	run(t, s.SubscribeToComputeAllocationChange(context.Background(), models.ComputeAllocation{ID: "alloc-1"}))
	if slurm.writes != writes {
		t.Fatalf("a second sync wrote %d more records", slurm.writes-writes)
	}
}

func TestAllocationRemovalRemovesMembers(t *testing.T) {
	for name, o := range map[string]mockOpts{
		"deactivated": {allocationStatus: models.INACTIVE},
		"deleted":     {allocationGone: true},
	} {
		t.Run(name, func(t *testing.T) {
			slurm := newFakeSlurm(client.Association{Account: "test-alloc", Cluster: "testcluster"},
				client.Association{Account: "test-alloc", Cluster: "testcluster", User: "testuser"})
			s := newSub(slurm, o, 0)
			run(t, s.SubscribeToComputeAllocationChange(context.Background(), models.ComputeAllocation{ID: "alloc-1", Name: "test-alloc", ComputeClusterID: "cluster-1"}))
			if got, _ := slurm.ListAssociations(client.AssocFilter{Account: "test-alloc"}); len(got) != 1 || got[0].User != "" {
				t.Fatalf("left %+v, want only the account record", got)
			}
		})
	}
}

// A late event for a member who has since left grants nothing.
func TestMembershipChange(t *testing.T) {
	inactive := testMembership()
	inactive.MembershipStatus = models.INACTIVE
	tests := []struct {
		name      string
		o         mockOpts
		had, want []string
		err       error
	}{
		{"active", mockOpts{provisionedAt: ago(time.Hour), memberships: member}, nil, []string{"none", "test-alloc"}, nil},
		{"left since the event", mockOpts{provisionedAt: ago(time.Hour), memberships: []models.ComputeAllocationMembership{inactive}}, []string{"test-alloc"}, []string{"none"}, nil},
		{"deleted", mockOpts{provisionedAt: ago(time.Hour)}, []string{"test-alloc"}, []string{"none"}, nil},
		{"settling", mockOpts{provisionedAt: ago(0), memberships: member}, nil, []string{"none"}, errNotProvisioned},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slurm := newFakeSlurm(client.Association{Account: "none", Cluster: "testcluster", User: "testuser"})
			for _, account := range tt.had {
				slurm.upsert(client.Association{Account: account, Cluster: "testcluster", User: "testuser"})
			}
			s := newSub(slurm, tt.o, time.Minute)
			err := s.SubscribeToComputeAllocationMembershipChange(context.Background(), testMembership())
			if got := slurm.held("testuser"); !errors.Is(err, tt.err) || !slices.Equal(got, tt.want) {
				t.Fatalf("err=%v held=%v, want %v", err, got, tt.want)
			}
		})
	}
}

// A user provisioned moments ago is retried until the cluster can resolve it.
func TestClusterUserProvisioned(t *testing.T) {
	slurm := newFakeSlurm()
	err := newSub(slurm, mockOpts{provisionedAt: ago(0), memberships: member}, time.Minute).SubscribeToComputeClusterUserChange(context.Background(), cu)
	if got := slurm.held("testuser"); !errors.Is(err, errNotProvisioned) || len(got) != 0 {
		t.Fatalf("err=%v held=%v while settling", err, got)
	}
	run(t, newSub(slurm, mockOpts{provisionedAt: ago(time.Hour), memberships: member}, time.Minute).SubscribeToComputeClusterUserChange(context.Background(), cu))
	if got := slurm.held("testuser"); !slices.Equal(got, []string{"none", "test-alloc"}) {
		t.Fatalf("held %v, want the allocation", got)
	}
}

func TestRemovalWhileJobsRun(t *testing.T) {
	slurm := newFakeSlurm(client.Association{Account: "none", Cluster: "testcluster", User: "testuser"},
		client.Association{Account: "test-alloc", Cluster: "testcluster", User: "testuser"})
	slurm.busy["testuser"] = true
	s := newSub(slurm, mockOpts{}, 0)
	run(t, s.SubscribeToComputeClusterUserChange(context.Background(), cu))
	if got := slurm.held("testuser"); !slices.Equal(got, []string{"none", "test-alloc blocked"}) {
		t.Fatalf("held %v while jobs run, want the allocation blocked", got)
	}
	slurm.busy["testuser"] = false
	s.reconcile(context.Background())
	if got := slurm.held("testuser"); len(got) != 0 {
		t.Fatalf("held %v after the jobs ended, want nothing", got)
	}
}

// Accounts Custos did not create are left alone.
func TestReconcilerRemovesOrphans(t *testing.T) {
	slurm := newFakeSlurm(client.Association{Account: "old-name", Cluster: "testcluster", User: "alice"},
		client.Association{Account: "theirs", Cluster: "testcluster", User: "alice"},
		client.Association{Account: "none", Cluster: "testcluster"}, client.Association{Account: "none", Cluster: "testcluster", User: "ghost"})
	slurm.accounts["old-name"], slurm.accounts["theirs"] = custos, ""
	s := newSub(slurm, mockOpts{}, 0)
	s.reconcile(context.Background())
	if alice, ghost := slurm.held("alice"), slurm.held("ghost"); !slices.Equal(alice, []string{"theirs"}) || len(ghost) != 0 {
		t.Fatalf("alice holds %v, ghost holds %v", alice, ghost)
	}
}

func TestUnhostedClusterIsSkipped(t *testing.T) {
	slurm := newFakeSlurm()
	slurm.clusters = []string{"elsewhere"}
	s := newSub(slurm, mockOpts{provisionedAt: ago(time.Hour), memberships: member, mappings: cpu8}, 0)
	run(t, s.SubscribeToComputeAllocationChange(context.Background(), models.ComputeAllocation{ID: "alloc-1"}))
	if slurm.writes != 0 {
		t.Fatalf("wrote %d records to a cluster Slurm does not host", slurm.writes)
	}
}

func TestMemberLimits(t *testing.T) {
	core := coreMock(mockOpts{
		resources: []models.ComputeAllocationResource{{ID: "a", ResourceType: "gres/gpu"}, {ID: "b", ResourceType: "gres/gpu"}},
		overrides: []models.ComputeAllocationMembershipResourceOverride{
			{ComputeAllocationResourceID: "a", OverrideResourceAmount: 2}, {ComputeAllocationResourceID: "b", OverrideResourceAmount: 3}},
	})
	got, err := NewAssociationSubscriber(nil, nil, core, fakeMappings(nil), 0, 0).
		userAssociation(context.Background(), "testcluster", models.ComputeAllocation{Name: "Test-Alloc"}, "testuser", "mem-1")
	if want := (client.TRES{Type: "gres", Name: "gpu", Count: 5}); err != nil || got.Account != "test-alloc" || !slices.Equal(got.Limits.GrpTRES, []client.TRES{want}) {
		t.Fatalf("err=%v got %+v", err, got)
	}
}
