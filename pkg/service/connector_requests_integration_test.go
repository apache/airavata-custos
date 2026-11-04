//go:build integration

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
	"fmt"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// Make sure only an approved account is sent to the identity registry.
func TestIdentityRequest_OnApprove(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	subscribeAll(svc, events.ComputeClusterUserIdentityRequestEvent)
	f := newLifecycleFixture(t, svc, database, 1, models.ACTIVE)
	if n := countTopic(t, svc, events.ComputeClusterUserIdentityRequestEvent); n != 0 {
		t.Fatalf("%d identity requests for a pending account, want 0", n)
	}
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))
	if _, err := svc.ApproveComputeClusterUser(ctx(), f.users[0].ID, admin); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if got := published[models.ComputeClusterUser](t, svc, events.ComputeClusterUserIdentityRequestEvent); len(got) != 1 || got[0].ID != f.users[0].ID {
		t.Fatalf("identity requests %+v, want one for %s", got, f.users[0].ID)
	}
}

// Make sure a provisioned account resyncs each of its allocation groups once.
func TestGroupRequest_OnProvision(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	f := newLifecycleFixture(t, svc, database, 1, models.ACTIVE)
	subscribeAll(svc, events.ComputeAllocationGroupRequestEvent)
	if err := svc.MarkComputeClusterUserProvisioned(ctx(), f.users[0].ID); err != nil {
		t.Fatalf("mark provisioned: %v", err)
	}
	if got := published[models.ComputeAllocation](t, svc, events.ComputeAllocationGroupRequestEvent); len(got) != 1 || got[0].ID != f.alloc.ID {
		t.Fatalf("group requests %+v, want one for %s", got, f.alloc.ID)
	}
}

// Make sure moving a membership resyncs both the group it left and the one it joined.
func TestGroupRequest_OnMove(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	from := newLifecycleFixture(t, svc, database, 1, models.ACTIVE)
	to := newLifecycleFixture(t, svc, database, 0, models.ACTIVE)
	subscribeAll(svc, events.ComputeAllocationGroupRequestEvent)
	m := *from.members[0]
	m.ComputeAllocationID = to.alloc.ID
	if _, err := svc.UpdateComputeAllocationMembership(ctx(), &m); err != nil {
		t.Fatalf("move: %v", err)
	}
	var ids []string
	for _, a := range published[models.ComputeAllocation](t, svc, events.ComputeAllocationGroupRequestEvent) {
		ids = append(ids, a.ID)
	}
	if len(ids) != 2 || !slices.Contains(ids, from.alloc.ID) || !slices.Contains(ids, to.alloc.ID) {
		t.Fatalf("group requests for %v, want %s and %s", ids, from.alloc.ID, to.alloc.ID)
	}
}

// Make sure storage is requested once per member, after both its readiness and its allocation group.
func TestStorageRequest_OnReadyAndGroup(t *testing.T) {
	steps := map[string]func(svc *Service, f lifecycleFixture) error{
		"accounts": func(svc *Service, f lifecycleFixture) error {
			for _, cu := range f.users {
				if err := svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID); err != nil {
					return err
				}
			}
			return nil
		},
		"group": func(svc *Service, f lifecycleFixture) error {
			return svc.MarkComputeAllocationGroupProvisioned(ctx(), f.alloc.ID)
		},
	}
	for _, order := range [][]string{{"accounts", "group"}, {"group", "accounts"}} {
		t.Run(fmt.Sprint(order), func(t *testing.T) {
			database := setupTestDB(t)
			svc := newTestService(database)
			subscribeAll(svc, events.ComputeAllocationMembershipStorageRequestEvent, events.ComputeAllocationGroupRequestEvent)
			f := newLifecycleFixture(t, svc, database, 2, models.ACTIVE)
			if err := steps[order[0]](svc, f); err != nil {
				t.Fatalf("%s: %v", order[0], err)
			}
			if n := countTopic(t, svc, events.ComputeAllocationMembershipStorageRequestEvent); n != 0 {
				t.Fatalf("%d storage requests after %s alone, want 0", n, order[0])
			}
			if err := steps[order[1]](svc, f); err != nil {
				t.Fatalf("%s: %v", order[1], err)
			}
			if n := countTopic(t, svc, events.ComputeAllocationMembershipStorageRequestEvent); n != 2 {
				t.Fatalf("%d storage requests, want 2", n)
			}
		})
	}
}

// Make sure an account is ready once, after both its identity and its storage.
func TestAccountReady_WaitsOnStorage(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	subscribeAll(svc, events.ComputeClusterUserStorageRequestEvent, events.ComputeClusterUserReadyEvent)
	f := newLifecycleFixture(t, svc, database, 1, models.ACTIVE)
	id := f.users[0].ID
	if err := svc.MarkComputeClusterUserProvisioned(ctx(), id); err != nil {
		t.Fatalf("mark provisioned: %v", err)
	}
	if n := countTopic(t, svc, events.ComputeClusterUserReadyEvent); n != 0 {
		t.Fatalf("%d ready before storage, want 0", n)
	}
	for range 2 {
		if err := svc.MarkComputeClusterUserStorageProvisioned(ctx(), id); err != nil {
			t.Fatalf("mark storage: %v", err)
		}
	}
	if n := countTopic(t, svc, events.ComputeClusterUserReadyEvent); n != 1 {
		t.Fatalf("%d ready, want 1", n)
	}
}

// Make sure an account is ready on identity alone when no storage connector subscribes.
func TestAccountReady_NoStorageConnector(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	subscribeAll(svc, events.ComputeClusterUserReadyEvent)
	f := newLifecycleFixture(t, svc, database, 1, models.ACTIVE)
	if err := svc.MarkComputeClusterUserProvisioned(ctx(), f.users[0].ID); err != nil {
		t.Fatalf("mark provisioned: %v", err)
	}
	if n := countTopic(t, svc, events.ComputeClusterUserReadyEvent); n != 1 {
		t.Fatalf("%d ready, want 1", n)
	}
}

// Make sure members are not ready until the scheduler has the allocation's account.
func TestMembershipReady_WaitsOnScheduler(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	subscribeAll(svc, events.ComputeAllocationSchedulerRequestEvent, events.ComputeAllocationMembershipReadyEvent)
	f := newLifecycleFixture(t, svc, database, 2, models.ACTIVE)
	for _, cu := range f.users {
		if err := svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID); err != nil {
			t.Fatalf("mark provisioned: %v", err)
		}
	}
	if n := countTopic(t, svc, events.ComputeAllocationMembershipReadyEvent); n != 0 {
		t.Fatalf("%d ready before the scheduler, want 0", n)
	}
	if err := svc.MarkComputeAllocationSchedulerProvisioned(ctx(), f.alloc.ID); err != nil {
		t.Fatalf("mark scheduler: %v", err)
	}
	if n := countTopic(t, svc, events.ComputeAllocationMembershipReadyEvent); n != 2 {
		t.Fatalf("%d ready, want 2", n)
	}
}
