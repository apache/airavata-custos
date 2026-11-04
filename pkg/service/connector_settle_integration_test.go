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

package service

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

func memberReady(t *testing.T, database *sqlx.DB, id string) bool {
	t.Helper()
	var ready bool
	if err := database.Get(&ready, `SELECT ready_at IS NOT NULL FROM compute_allocation_memberships WHERE id = $1`, id); err != nil {
		t.Fatalf("read ready_at: %v", err)
	}
	return ready
}

// Make sure approval alone readies everything when no connector is configured.
func TestNoConnectors_ApprovalReadiesAll(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	f := newLifecycleFixture(t, svc, database, 1, models.ACTIVE)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))
	if _, err := svc.ApproveComputeClusterUser(ctx(), f.users[0].ID, admin); err != nil {
		t.Fatalf("approve: %v", err)
	}
	cu, err := svc.GetComputeClusterUser(ctx(), f.users[0].ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	a, err := svc.GetComputeAllocation(ctx(), f.alloc.ID)
	if err != nil {
		t.Fatalf("get allocation: %v", err)
	}
	if cu.ProvisionedAt == nil || cu.ReadyAt == nil || a.GroupProvisionedAt == nil || !memberReady(t, database, f.members[0].ID) {
		t.Fatalf("provisioned=%v ready=%v group=%v member ready=%v, want all set",
			cu.ProvisionedAt, cu.ReadyAt, a.GroupProvisionedAt, memberReady(t, database, f.members[0].ID))
	}
}

// Make sure a request left pending by a connector that is then disabled is done by core.
func TestSettle_DisabledConnectorReleasesRequest(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	subscribeAll(svc, events.ComputeAllocationSchedulerRequestEvent)
	f := newLifecycleFixture(t, svc, database, 1, models.ACTIVE)
	if err := svc.MarkComputeClusterUserProvisioned(ctx(), f.users[0].ID); err != nil {
		t.Fatalf("mark provisioned: %v", err)
	}
	if memberReady(t, database, f.members[0].ID) {
		t.Fatal("member ready before the scheduler")
	}
	if err := svc.EventBus().Unsubscribe(ctx(), "test-subscriber"); err != nil {
		t.Fatalf("unsubscribe: %v", err)
	}
	if err := svc.settleConnectors(ctx()); err != nil {
		t.Fatalf("settle: %v", err)
	}
	if !memberReady(t, database, f.members[0].ID) {
		t.Fatal("member not ready after its scheduler was disabled")
	}
}

// Make sure a connector enabled after the fact is sent the current state once.
func TestSettle_EnabledConnectorCatchesUp(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	f := newLifecycleFixture(t, svc, database, 1, models.ACTIVE)
	if err := svc.MarkComputeClusterUserProvisioned(ctx(), f.users[0].ID); err != nil {
		t.Fatalf("mark provisioned: %v", err)
	}
	subscribeAll(svc, events.ComputeAllocationSchedulerRequestEvent, events.ComputeAllocationMembershipReadyEvent, events.ComputeClusterUserReadyEvent)
	for range 2 {
		if err := svc.settleConnectors(ctx()); err != nil {
			t.Fatalf("settle: %v", err)
		}
	}
	if got := published[models.ComputeAllocation](t, svc, events.ComputeAllocationSchedulerRequestEvent); len(got) != 1 || got[0].ID != f.alloc.ID {
		t.Fatalf("scheduler requests %+v, want one for %s", got, f.alloc.ID)
	}
	if got := published[models.ComputeAllocationMembership](t, svc, events.ComputeAllocationMembershipReadyEvent); len(got) != 1 || got[0].ID != f.members[0].ID {
		t.Fatalf("ready %+v, want one for %s", got, f.members[0].ID)
	}
	if n := countTopic(t, svc, events.ComputeClusterUserReadyEvent); n != 0 {
		t.Fatalf("%d account ready notifications resent, want 0", n)
	}
}
