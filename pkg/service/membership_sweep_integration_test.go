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
	"testing"

	"github.com/apache/airavata-custos/pkg/events"
)

// Make sure a pass repairs drift from writes that bypass the service, publishing each readiness it changes.
func TestReadySweep_CorrectsStaleReady(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	f := newReadyFixture(t, svc, database, 2)
	g := newReadyFixture(t, svc, database, 1)
	subscribeAll(svc, events.ComputeAllocationMembershipDeactivateEvent, events.ComputeAllocationMembershipReadyEvent)
	if _, err := database.Exec("UPDATE compute_cluster_users SET provisioned_at = NULL WHERE compute_cluster_id = $1", f.alloc.ComputeClusterID); err != nil {
		t.Fatalf("unprovision: %v", err)
	}
	if _, err := database.Exec("UPDATE compute_allocation_memberships SET ready_at = NULL WHERE id = $1", g.members[0].ID); err != nil {
		t.Fatalf("clear ready: %v", err)
	}
	if err := svc.syncStaleMemberships(ctx()); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	for _, mem := range append(f.members, g.members...) {
		m, err := svc.GetComputeAllocationMembership(ctx(), mem.ID)
		if err != nil || (m.ReadyAt != nil) != (mem.ComputeAllocationID == g.alloc.ID) {
			t.Fatalf("after the sweep: %+v, %v", m, err)
		}
	}
	if ids, err := svc.memberships.FindStaleReady(ctx()); err != nil || len(ids) != 0 {
		t.Fatalf("stale after the sweep: %v, %v", ids, err)
	}
	if d, r := countTopic(t, svc, events.ComputeAllocationMembershipDeactivateEvent), countTopic(t, svc, events.ComputeAllocationMembershipReadyEvent); d != 2 || r != 1 {
		t.Fatalf("%d deactivate and %d ready deliveries, want 2 and 1", d, r)
	}
}
