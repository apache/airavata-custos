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
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/drift"
	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/internal/connectors/health"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

var testCluster = models.ComputeCluster{ID: "cluster-1", Name: "testcluster"}

var isDefault = true

func normalQoS(pending bool) models.ComputeClusterQoS {
	q := models.ComputeClusterQoS{ID: "qos-normal", Name: "normal", IsDefault: true}
	if pending {
		q.SlurmPendingSince = ago(time.Minute)
	}
	return q
}

func openfoldQoS() models.ComputeClusterQoS {
	return models.ComputeClusterQoS{ID: "qos-openfold", Name: "openfold"}
}

func TestSync_SkipsClusterWhenLiveReadFails(t *testing.T) {
	core := coreMock(mockOpts{provisionedAt: ago(time.Minute), allocationPending: true})
	slurm := &fakeSlurmClient{listErr: errors.New("slurmrestd down")}
	sub := NewAssociationSubscriber(slurm, nil, core, 0, 0)
	sub.name = "slurm-live-read-test"

	if _, err := sub.syncCluster(context.Background(), testCluster); err == nil {
		t.Fatal("expected the live read error")
	}
	if len(slurm.allUpserts())+len(slurm.allDeletes())+len(slurm.qosPosts) != 0 || len(slurm.defaults) != 0 {
		t.Fatalf("no writes expected when the live read fails: %+v", slurm)
	}
	checks := health.For("slurm-live-read-test")
	if len(checks) != 1 || checks[0].OK || checks[0].Target != "cluster-1" {
		t.Fatalf("health = %+v, want one failed check for the cluster", checks)
	}
}

func TestSync_MembersGetAccountQoS(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute), allocationPending: true,
		qos:           []models.ComputeClusterQoS{normalQoS(false), openfoldQoS()},
		allocationQoS: []string{"normal", "openfold"},
	})
	slurm := &fakeSlurmClient{}
	if _, err := NewAssociationSubscriber(slurm, nil, core, 0, 0).syncCluster(context.Background(), testCluster); err != nil {
		t.Fatal(err)
	}
	members := slices.DeleteFunc(slurm.allUpserts(), func(a client.Association) bool { return a.User == "" })
	if len(members) != 1 || !slices.Equal(members[0].QoS, []string{"normal", "openfold"}) || members[0].DefaultQoS != "normal" {
		t.Fatalf("member written with %+v, want normal,openfold and default normal", members)
	}
	for _, a := range slurm.allUpserts() {
		if a.User == "" && (a.Cluster != "testcluster" || a.ParentAccount != "root") {
			t.Errorf("new account association lacks cluster or parent: %+v", a)
		}
	}
}

// A grant made on SLURM stays there until an admin decides.
func TestSync_KeepsHandGrantedQoSAsConflict(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		qos:           []models.ComputeClusterQoS{normalQoS(false)},
		allocationQoS: []string{"normal"},
	})
	slurm := &fakeSlurmClient{
		qos: []client.QoS{{Name: "normal"}},
		existing: []client.Association{
			{Account: "test-alloc", Cluster: "testcluster", QoS: []string{"normal", "openfold"}, DefaultQoS: "normal"},
			{Account: "test-alloc", Cluster: "testcluster", User: "testuser", QoS: []string{"normal", "openfold"}, DefaultQoS: "normal"},
			{Account: noAllocation, Cluster: "testcluster", User: "testuser", IsDefault: &isDefault},
		},
	}
	conflicts, err := NewAssociationSubscriber(slurm, nil, core, 0, 0).syncCluster(context.Background(), testCluster)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(slurm.allUpserts()); n != 0 {
		t.Fatalf("expected no write, got %+v", slurm.allUpserts())
	}
	if !slices.ContainsFunc(conflicts, func(c drift.Conflict) bool { return c.Key == "account:test-alloc:qos" }) {
		t.Fatalf("expected an allocation QoS conflict, got %+v", conflicts)
	}
}

func TestSync_ClearsPendingMarks(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute), allocationPending: true, membershipPending: true,
		qos:           []models.ComputeClusterQoS{normalQoS(true)},
		allocationQoS: []string{"normal"},
	})
	slurm := &fakeSlurmClient{
		qos: []client.QoS{{Name: "normal"}},
		existing: []client.Association{
			{Account: "test-alloc", Cluster: "testcluster", QoS: []string{"normal"}, DefaultQoS: "normal"},
			{Account: "test-alloc", Cluster: "testcluster", User: "testuser", QoS: []string{"normal"}, DefaultQoS: "normal"},
			{Account: noAllocation, Cluster: "testcluster", User: "testuser", IsDefault: &isDefault},
		},
	}
	if _, err := NewAssociationSubscriber(slurm, nil, core, 0, 0).syncCluster(context.Background(), testCluster); err != nil {
		t.Fatal(err)
	}
	if n := len(slurm.allUpserts()) + len(slurm.qosPosts); n != 0 {
		t.Fatalf("a satisfied mark finds no drift, got %d writes", n)
	}
	if got := cleared(core, service.SlurmPendingAllocation); !slices.Equal(got, []string{"alloc-1"}) {
		t.Errorf("allocation clears = %v, want alloc-1", got)
	}
	if got := cleared(core, service.SlurmPendingMembership); !slices.Equal(got, []string{"mem-1"}) {
		t.Errorf("membership clears = %v", got)
	}
	if got := cleared(core, service.SlurmPendingQoS); !slices.Equal(got, []string{"qos-normal"}) {
		t.Errorf("qos clears = %v", got)
	}
	for _, c := range core.SetSlurmPendingCalls() {
		if c.ClearSince == nil || c.ClearSince.IsZero() {
			t.Errorf("%s %s cleared without the mark it read", c.Kind, c.ID)
		}
	}
}

// A default on a system account, or one SLURM does not report, is not repaired.
func TestSync_LeavesSystemOrUnknownDefaultAccount(t *testing.T) {
	yes := true
	for _, tc := range []struct {
		name string
		live []client.Association
	}{
		{"system_account_default", []client.Association{
			{Account: "sysacct", Cluster: "testcluster", User: "testuser", IsDefault: &yes},
			{Account: noAllocation, Cluster: "testcluster", User: "testuser", IsDefault: new(bool)},
		}},
		{"unknown_default", []client.Association{
			{Account: noAllocation, Cluster: "testcluster", User: "testuser"},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			core := coreMock(mockOpts{provisionedAt: ago(time.Minute)})
			slurm := &fakeSlurmClient{existing: tc.live}
			if _, err := NewAssociationSubscriber(slurm, nil, core, 0, 0).syncCluster(context.Background(), testCluster); err != nil {
				t.Fatal(err)
			}
			if len(slurm.defaults) != 0 {
				t.Fatalf("defaults set = %v, want none", slurm.defaults)
			}
		})
	}
}

func on[T any](e eventSync[T], payload T) func(*AssociationSubscriber) error {
	return func(s *AssociationSubscriber) error { return e.run(context.Background(), s, payload) }
}

func sweep(s *AssociationSubscriber) error {
	s.reconcile(context.Background())
	return nil
}

// describe renders an upsert as account/user and its TRES limits, amounts before minutes.
func describe(a client.Association) string {
	s := a.Account + "/" + a.User
	for _, t := range append(slices.Clone(a.Limits.GrpTRES), a.Limits.GrpTRESMins...) {
		s += fmt.Sprintf(" %s:%s=%g", t.Type, t.Name, t.Count)
	}
	return s
}

// Every entry point runs the same diff; each row checks what one of them pushes.
func TestEventSync(t *testing.T) {
	deactivated := testMembership()
	deactivated.MembershipStatus = models.INACTIVE
	allocation := models.ComputeAllocation{ID: "alloc-1", Name: "test-alloc", ComputeClusterID: "cluster-1", Status: models.ACTIVE}
	inactive := allocation
	inactive.Status = models.INACTIVE
	gpu := []models.ComputeAllocationResource{{ID: "res-1", ResourceType: "gres/gpu"}, {ID: "res-2", ResourceType: "gres/gpu"}}
	override := func(resourceID string, amount, minutes int64) models.ComputeAllocationMembershipResourceOverride {
		return models.ComputeAllocationMembershipResourceOverride{ComputeAllocationMembershipID: "mem-1", ComputeAllocationResourceID: resourceID,
			OverrideResourceAmount: amount, OverrideResourceTime: minutes}
	}
	provisioned := ago(time.Minute)
	liveAccount := client.Association{Account: "test-alloc", Cluster: "testcluster"}
	liveMember := client.Association{Account: "test-alloc", Cluster: "testcluster", User: "testuser"}
	noneDefault := client.Association{Account: noAllocation, Cluster: "testcluster", User: "testuser", IsDefault: &isDefault}
	lookupErr := errors.New("core unavailable")

	for _, tc := range []struct {
		name                       string
		handle                     func(*AssociationSubscriber) error
		opts                       mockOpts
		tweak                      func(*service.CoreServiceMock)
		live                       []client.Association
		wantErr                    error
		upserts, deletes, defaults []string
	}{
		{name: "membership created", handle: on(membershipCreated, testMembership()), opts: mockOpts{provisionedAt: provisioned},
			upserts: []string{"test-alloc/", "test-alloc/testuser"}, defaults: []string{"testuser=none"}},
		{name: "account lowercased", handle: on(membershipCreated, testMembership()), opts: mockOpts{allocationName: "Q2-2026-Climate", provisionedAt: provisioned},
			upserts: []string{"q2-2026-climate/", "q2-2026-climate/testuser"}, defaults: []string{"testuser=none"}},
		{name: "override limits kept", handle: on(membershipCreated, testMembership()),
			opts:    mockOpts{provisionedAt: provisioned, overrides: []models.ComputeAllocationMembershipResourceOverride{override("res-1", 4, 120)}},
			upserts: []string{"test-alloc/", "test-alloc/testuser cpu:=4 cpu:=120"}, defaults: []string{"testuser=none"}},
		{name: "gpu tres split into type and name, summed per type", handle: on(overrideCreated, override("res-1", 2, 0)),
			opts: mockOpts{provisionedAt: provisioned, resources: gpu, overrides: []models.ComputeAllocationMembershipResourceOverride{
				override("res-1", 2, 120), override("res-2", 4, 0)}},
			tweak: func(m *service.CoreServiceMock) {
				m.GetComputeAllocationMembershipFunc = func(context.Context, string) (*models.ComputeAllocationMembership, error) {
					membership := testMembership()
					return &membership, nil
				}
			},
			upserts: []string{"test-alloc/", "test-alloc/testuser gres:gpu=6 gres:gpu=120"}, defaults: []string{"testuser=none"}},
		{name: "core error returned", handle: on(membershipCreated, testMembership()), opts: mockOpts{provisionedAt: provisioned},
			tweak: func(m *service.CoreServiceMock) {
				m.GetComputeAllocationFunc = func(context.Context, string) (*models.ComputeAllocation, error) { return nil, lookupErr }
			}, wantErr: lookupErr},
		{name: "membership deleted after its allocation", handle: on(membershipDeleted, testMembership()), opts: mockOpts{provisionedAt: provisioned},
			tweak: func(m *service.CoreServiceMock) {
				m.GetComputeAllocationFunc = func(context.Context, string) (*models.ComputeAllocation, error) { return nil, service.ErrNotFound }
			}},
		{name: "pending deactivation removes the member", handle: on(membershipUpdated, deactivated),
			opts: mockOpts{provisionedAt: provisioned, membershipPending: true, memberships: []models.ComputeAllocationMembership{deactivated}},
			live: []client.Association{liveAccount, liveMember}, deletes: []string{"test-alloc/testuser"}},
		{name: "unmarked deactivation keeps the member", handle: on(membershipUpdated, deactivated),
			opts: mockOpts{provisionedAt: provisioned, memberships: []models.ComputeAllocationMembership{deactivated}},
			live: []client.Association{liveAccount, liveMember}},
		{name: "pending allocation deactivation removes members", handle: on(allocationUpdated, inactive),
			opts: mockOpts{provisionedAt: provisioned, allocationStatus: models.INACTIVE, allocationPending: true},
			live: []client.Association{liveAccount, liveMember}, deletes: []string{"test-alloc/testuser"}},
		{name: "pending activation activates members", handle: on(allocationUpdated, allocation),
			opts: mockOpts{provisionedAt: provisioned, allocationPending: true}, live: []client.Association{liveAccount},
			upserts: []string{"test-alloc/testuser"}, defaults: []string{"testuser=none"}},
		{name: "unprovisioned member skipped", handle: on(allocationUpdated, allocation),
			opts: mockOpts{allocationPending: true}, live: []client.Association{liveAccount}},
		{name: "member inside the grace held", handle: sweep, opts: mockOpts{provisionedAt: ago(time.Second)}, live: []client.Association{liveAccount}},
		{name: "allocation deletion pushes nothing", handle: on(allocationDeleted, inactive), opts: mockOpts{unmanagedAccounts: true},
			live: []client.Association{liveMember}},
		{name: "unknown member left as a conflict", handle: sweep, opts: mockOpts{provisionedAt: provisioned, allocationPending: true},
			live: []client.Association{liveAccount, liveMember, noneDefault, {Account: "test-alloc", Cluster: "testcluster", User: "ghost"}}},
		{name: "unmanaged account untouched", handle: sweep, opts: mockOpts{provisionedAt: provisioned, unmanagedAccounts: true},
			live: []client.Association{{Account: "someone-elses-account", Cluster: "testcluster", User: "outsider"}}},
		{name: "steady state writes nothing", handle: sweep, opts: mockOpts{provisionedAt: provisioned},
			live: []client.Association{liveAccount, liveMember, noneDefault}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			core := coreMock(tc.opts)
			if tc.tweak != nil {
				tc.tweak(core)
			}
			slurm := &fakeSlurmClient{existing: tc.live}
			if err := tc.handle(NewAssociationSubscriber(slurm, nil, core, 0, 0)); !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			var upserts, deletes []string
			for _, a := range slurm.allUpserts() {
				upserts = append(upserts, describe(a))
			}
			for _, f := range slurm.allDeletes() {
				deletes = append(deletes, f.Account+"/"+f.User)
			}
			for what, got := range map[string][2][]string{"upserts": {upserts, tc.upserts}, "deletes": {deletes, tc.deletes}, "defaults": {slurm.defaults, tc.defaults}} {
				if !slices.Equal(got[0], got[1]) {
					t.Errorf("%s = %q, want %q", what, got[0], got[1])
				}
			}
		})
	}
}

// A missing account is created before its associations; a failed creation skips its members and keeps the mark.
func TestSync_CreatesMissingAccountFirst(t *testing.T) {
	core := coreMock(mockOpts{provisionedAt: ago(time.Minute), allocationPending: true})
	slurm := &fakeSlurmClient{}
	if _, err := NewAssociationSubscriber(slurm, nil, core, 0, 0).syncCluster(context.Background(), testCluster); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(slurm.writes, ","); got != "create none,create test-alloc,upsert test-alloc/,upsert test-alloc/testuser" {
		t.Fatalf("writes = %s", got)
	}

	slurm = &fakeSlurmClient{createErr: errors.New("refused")}
	if _, err := NewAssociationSubscriber(slurm, nil, core, 0, 0).syncCluster(context.Background(), testCluster); err == nil {
		t.Fatal("expected the create error")
	}
	if len(slurm.allUpserts()) != 0 || len(cleared(core, service.SlurmPendingAllocation)) != 0 {
		t.Fatalf("upserts %+v after a failed create", slurm.allUpserts())
	}
}

func TestSnapshot_ReturnsConflictsAndSystemAccountsWithoutWriting(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		qos:           []models.ComputeClusterQoS{normalQoS(false)},
		allocationQoS: []string{"normal"},
	})
	slurm := &fakeSlurmClient{
		qos: []client.QoS{{Name: "normal"}},
		existing: []client.Association{
			{Account: "test-alloc", Cluster: "testcluster", QoS: []string{"normal", "openfold"}, DefaultQoS: "normal"},
			{Account: "ops", Cluster: "testcluster", QoS: []string{"normal"}},
		},
	}
	snap, err := NewAssociationSubscriber(slurm, nil, core, 0, 0).Snapshot(context.Background(), testCluster)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(snap.Conflicts, func(c drift.Conflict) bool { return c.Key == "account:test-alloc:qos" }) {
		t.Fatalf("conflicts = %+v", snap.Conflicts)
	}
	if system := snap.Want.SystemAccounts; !system["ops"] || !system[rootAccount] || system["test-alloc"] {
		t.Fatalf("system accounts = %v", system)
	}
	if len(slurm.allUpserts())+len(slurm.allDeletes())+len(slurm.qosPosts) != 0 {
		t.Fatalf("a snapshot must not write: %+v", slurm)
	}
}
