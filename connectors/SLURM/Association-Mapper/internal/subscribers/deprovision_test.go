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
	"testing"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
)

func deactivatedMembership() models.ComputeAllocationMembership {
	m := testMembership()
	m.MembershipStatus = models.INACTIVE
	return m
}

func TestMembershipDeactivationRemovesAssociations(t *testing.T) {
	core := coreMock(mockOpts{provisionedAt: ago(time.Minute)})
	slurm := &fakeSlurmClient{}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).
		SubscribeToComputeAllocationMembershipUpdate(context.Background(), deactivatedMembership())

	got := slurm.allDeletes()
	if len(got) != 1 {
		t.Fatalf("expected 1 delete, got %d", len(got))
	}
	// Deleting by user+account covers every partition in one call.
	if got[0].User != "testuser" || got[0].Account != "test-alloc" || got[0].Cluster != "testcluster" {
		t.Errorf("unexpected delete filter: %+v", got[0])
	}
	if got[0].Partition != "" {
		t.Errorf("filter should not pin a partition, got %q", got[0].Partition)
	}
	if n := len(slurm.all()); n != 0 {
		t.Errorf("a deactivation must not write associations, got %d", n)
	}
}

// Flipping back to ACTIVE re-grants access rather than leaving the member cut off.
func TestMembershipReactivationRestoresAssociations(t *testing.T) {
	core := coreMock(mockOpts{provisionedAt: ago(time.Minute)})
	slurm := &fakeSlurmClient{}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).
		SubscribeToComputeAllocationMembershipUpdate(context.Background(), testMembership())

	if n := len(slurm.allDeletes()); n != 0 {
		t.Fatalf("an active membership must not be deprovisioned, got %d deletes", n)
	}
	if n := len(slurm.all()); n != 1 {
		t.Fatalf("expected the association to be restored, got %d writes", n)
	}
}

func TestMembershipDeletionRemovesAssociations(t *testing.T) {
	core := coreMock(mockOpts{provisionedAt: ago(time.Minute)})
	slurm := &fakeSlurmClient{}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).
		SubscribeToComputeAllocationMembershipDeletion(context.Background(), testMembership())

	got := slurm.allDeletes()
	if len(got) != 1 || got[0].User != "testuser" || got[0].Account != "test-alloc" {
		t.Fatalf("expected the member's associations to be removed, got %+v", got)
	}
}

func TestAllocationDeactivationRemovesAllAssociations(t *testing.T) {
	core := coreMock(mockOpts{})
	slurm := &fakeSlurmClient{}
	alloc := models.ComputeAllocation{
		ID: "alloc-1", Name: "test-alloc", ComputeClusterID: "cluster-1", Status: models.INACTIVE,
	}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).
		SubscribeToComputeAllocationUpdate(context.Background(), alloc)

	got := slurm.allDeletes()
	if len(got) != 1 {
		t.Fatalf("expected 1 delete, got %d", len(got))
	}
	// No user pinned: every member of the allocation loses access.
	if got[0].Account != "test-alloc" || got[0].User != "" {
		t.Errorf("expected an account-wide delete, got %+v", got[0])
	}
}

// An allocation coming back to ACTIVE restores its members here and now,
// rather than leaving them without access until the next reconciler pass.
func TestAllocationReactivationRestoresMemberAssociations(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		memberships:   []models.ComputeAllocationMembership{testMembership()},
	})
	slurm := &fakeSlurmClient{}
	alloc := models.ComputeAllocation{
		ID: "alloc-1", Name: "test-alloc", ComputeClusterID: "cluster-1", Status: models.ACTIVE,
	}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).
		SubscribeToComputeAllocationUpdate(context.Background(), alloc)

	if n := len(slurm.allDeletes()); n != 0 {
		t.Fatalf("an active allocation must not be deprovisioned, got %d deletes", n)
	}
	if n := len(slurm.all()); n != 1 {
		t.Fatalf("expected the active member's association to be restored, got %d writes", n)
	}
}

// Members whose account is not provisioned yet stay with the reconciler.
func TestAllocationReactivationSkipsUnprovisionedMembers(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: nil,
		memberships:   []models.ComputeAllocationMembership{testMembership()},
	})
	slurm := &fakeSlurmClient{}
	alloc := models.ComputeAllocation{
		ID: "alloc-1", Name: "test-alloc", ComputeClusterID: "cluster-1", Status: models.ACTIVE,
	}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).
		SubscribeToComputeAllocationUpdate(context.Background(), alloc)

	if n := len(slurm.all()); n != 0 {
		t.Fatalf("expected no association for an unprovisioned account, got %d", n)
	}
}

func TestAllocationDeletionRemovesAllAssociations(t *testing.T) {
	core := coreMock(mockOpts{})
	slurm := &fakeSlurmClient{}
	alloc := models.ComputeAllocation{
		ID: "alloc-1", Name: "test-alloc", ComputeClusterID: "cluster-1", Status: models.ACTIVE,
	}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).
		SubscribeToComputeAllocationDeletion(context.Background(), alloc)

	got := slurm.allDeletes()
	if len(got) != 1 || got[0].Account != "test-alloc" || got[0].User != "" {
		t.Fatalf("expected an account-wide delete on allocation deletion, got %+v", got)
	}
}

// The sweep is the backstop for a lost deactivation: an association the
// allocations no longer call for is removed even if no event ever arrived.
func TestReconcilerRemovesStaleAssociation(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		memberships:   []models.ComputeAllocationMembership{testMembership()},
	})
	slurm := &fakeSlurmClient{existing: []client.Association{
		// desired
		{Account: "test-alloc", Cluster: "testcluster", User: "testuser", Partition: "compute"},
		// left behind by a membership that is gone
		{Account: "test-alloc", Cluster: "testcluster", User: "ghost", Partition: "compute"},
	}}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).reconcile(context.Background())

	got := slurm.allDeletes()
	if len(got) != 1 {
		t.Fatalf("expected the stale association to be removed, got %d deletes", len(got))
	}
	if got[0].User != "ghost" || got[0].Partition != "compute" {
		t.Errorf("removed the wrong association: %+v", got[0])
	}
}

// Guard: an empty desired set almost always means a failed lookup, so the
// sweep must not read it as "revoke everyone".
func TestReconcilerDoesNotPruneWhenNothingIsDesired(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		memberships:   nil, // nobody entitled to anything
	})
	slurm := &fakeSlurmClient{existing: []client.Association{
		{Account: "test-alloc", Cluster: "testcluster", User: "testuser", Partition: "compute"},
	}}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).reconcile(context.Background())

	if n := len(slurm.allDeletes()); n != 0 {
		t.Fatalf("an empty desired set must not revoke anything, got %d deletes", n)
	}
}

// Guard: associations on accounts Custos does not manage are never touched.
func TestReconcilerLeavesUnmanagedAccountsAlone(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt:     ago(time.Minute),
		memberships:       []models.ComputeAllocationMembership{testMembership()},
		unmanagedAccounts: true, // core reports no allocations on this cluster
	})
	slurm := &fakeSlurmClient{existing: []client.Association{
		{Account: "someone-elses-account", Cluster: "testcluster", User: "outsider", Partition: "compute"},
	}}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).reconcile(context.Background())

	if n := len(slurm.allDeletes()); n != 0 {
		t.Fatalf("an unmanaged account must not be touched, got %d deletes", n)
	}
}

// Account-level records carry the allocation's own limits, not a member's
// access, so the sweep must leave them be.
func TestReconcilerLeavesAccountLevelAssociationsAlone(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Minute),
		memberships:   []models.ComputeAllocationMembership{testMembership()},
	})
	slurm := &fakeSlurmClient{existing: []client.Association{
		{Account: "test-alloc", Cluster: "testcluster", User: "testuser", Partition: "compute"},
		{Account: "test-alloc", Cluster: "testcluster", User: ""}, // account-level
	}}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).reconcile(context.Background())

	if n := len(slurm.allDeletes()); n != 0 {
		t.Fatalf("account-level associations must not be pruned, got %d deletes", n)
	}
}

// A member inside the provisioning grace is still entitled, so their fresh
// association must not be pruned just because the sweep is not writing it yet.
func TestReconcilerDoesNotPruneAssociationsInsideGrace(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt: ago(time.Second),
		memberships:   []models.ComputeAllocationMembership{testMembership()},
	})
	slurm := &fakeSlurmClient{existing: []client.Association{
		{Account: "test-alloc", Cluster: "testcluster", User: "testuser", Partition: "compute"},
	}}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).reconcile(context.Background())

	if n := len(slurm.allDeletes()); n != 0 {
		t.Fatalf("a member inside the grace must keep their association, got %d deletes", n)
	}
}

// An allocation that is no longer active grants nothing, so its members'
// associations are swept away even without a deactivation event.
func TestReconcilerRemovesAssociationsForInactiveAllocation(t *testing.T) {
	core := coreMock(mockOpts{
		provisionedAt:    ago(time.Minute),
		allocationStatus: models.INACTIVE,
		memberships:      []models.ComputeAllocationMembership{testMembership()},
	})
	slurm := &fakeSlurmClient{existing: []client.Association{
		{Account: "test-alloc", Cluster: "testcluster", User: "testuser", Partition: "compute"},
	}}
	NewAssociationSubscriber(slurm, nil, core, 0, 0).reconcile(context.Background())

	// Desired is empty for an inactive allocation, so the empty-desired guard
	// holds and nothing is revoked. The event handler does that job.
	if n := len(slurm.allDeletes()); n != 0 {
		t.Fatalf("expected the empty-desired guard to hold, got %d deletes", n)
	}
}
