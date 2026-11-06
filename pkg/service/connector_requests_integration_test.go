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
	"testing"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

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
	if err != nil || cu.ProvisionedAt == nil || cu.StorageProvisionedAt == nil {
		t.Fatalf("account %+v, %v: want provisioned with storage", cu, err)
	}
	if a, err := svc.GetComputeAllocation(ctx(), f.alloc.ID); err != nil || a.GroupProvisionedAt == nil {
		t.Fatalf("allocation %+v, %v: want group provisioned", a, err)
	}
	if m, err := svc.GetComputeAllocationMembership(ctx(), f.members[0].ID); err != nil || m.ReadyAt == nil {
		t.Fatalf("member %+v, %v: want ready", m, err)
	}
}

// Make sure a request left pending by a connector that is then disabled is done by core, and its delivery dropped.
func TestSettle_DisabledConnectorReleasesRequest(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	subscribeAll(svc, events.ComputeAllocationSchedulerRequestEvent)
	f := newLifecycleFixture(t, svc, database, 1, models.ACTIVE)
	if err := svc.MarkComputeClusterUserProvisioned(ctx(), f.users[0].ID); err != nil {
		t.Fatalf("mark provisioned: %v", err)
	}
	if err := svc.EventBus().Unsubscribe(ctx(), "test-subscriber"); err != nil {
		t.Fatalf("unsubscribe: %v", err)
	}
	if err := svc.SettleConnectors(ctx()); err != nil {
		t.Fatalf("settle: %v", err)
	}
	if m, err := svc.GetComputeAllocationMembership(ctx(), f.members[0].ID); err != nil || m.ReadyAt == nil {
		t.Fatalf("member not ready after its scheduler was disabled: %+v, %v", m, err)
	}
	if pending, err := svc.EventBus().ListDeliveries(ctx(), models.EventDeliveryPending, 100); err != nil || len(pending) != 0 {
		t.Fatalf("%d pending deliveries after disable, %v, want 0", len(pending), err)
	}
}
