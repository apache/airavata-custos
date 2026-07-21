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
