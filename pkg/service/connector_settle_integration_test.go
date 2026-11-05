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
	"testing"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// Make sure a request left pending by a connector that is then disabled is done by core, and its delivery dropped.
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
	if pending, err := svc.EventBus().ListDeliveries(ctx(), models.EventDeliveryPending, 100); err != nil || len(pending) != 0 {
		t.Fatalf("%d pending deliveries after disable, %v, want 0", len(pending), err)
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
	subscribeAll(svc, events.ComputeAllocationSchedulerRequestEvent, events.ComputeAllocationMembershipReadyEvent)
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
}

// Make sure a request whose retries ran out is sent once more, and not again while that one is outstanding.
func TestSettle_ExhaustedRequestReissued(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	subscribeAll(svc, events.ComputeAllocationSchedulerRequestEvent)
	// Synced already, so the pass sends only the reissue.
	if _, err := database.Exec(`UPDATE event_subscriptions SET synced_at = now()`); err != nil {
		t.Fatalf("mark synced: %v", err)
	}
	f := newLifecycleFixture(t, svc, database, 1, models.ACTIVE)
	if _, err := database.Exec(`UPDATE event_deliveries SET status = $1, finished_at = now() - interval '2 hours'`, models.EventDeliveryFailed); err != nil {
		t.Fatalf("exhaust delivery: %v", err)
	}
	for range 2 {
		if err := svc.settleConnectors(ctx()); err != nil {
			t.Fatalf("settle: %v", err)
		}
	}
	got := published[models.ComputeAllocation](t, svc, events.ComputeAllocationSchedulerRequestEvent)
	if len(got) != 2 || got[1].ID != f.alloc.ID {
		t.Fatalf("scheduler requests %+v, want the original and one resend for %s", got, f.alloc.ID)
	}
}
