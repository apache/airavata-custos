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
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// lifecycleFixture is an allocation with unprovisioned members.
type lifecycleFixture struct {
	alloc   *models.ComputeAllocation
	members []*models.ComputeAllocationMembership
	users   []*models.ComputeClusterUser
}

func newLifecycleFixture(t *testing.T, svc *Service, database *sqlx.DB, members int, status models.AllocationStatus) lifecycleFixture {
	t.Helper()
	end := time.Now().AddDate(1, 0, 0)
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "life-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	pi := seedUser(t, database, fmt.Sprintf("pi-%s@example.edu", uuid.NewString()))
	project, err := svc.CreateProject(ctx(), &models.Project{Title: "life", Origination: "TEST", ProjectPIID: pi, Status: models.ProjectActive})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	alloc, err := svc.CreateComputeAllocation(ctx(), &models.ComputeAllocation{
		Name: "Grant", ProjectID: project.ID, ComputeClusterID: cluster.ID, StartTime: time.Now().AddDate(-1, 0, 0), EndTime: end,
	})
	if err != nil {
		t.Fatalf("create allocation: %v", err)
	}
	f := lifecycleFixture{alloc: alloc}
	for range members {
		userID := seedUser(t, database, fmt.Sprintf("member-%s@example.edu", uuid.NewString()))
		cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{
			ComputeClusterID: cluster.ID, UserID: userID, LocalUsername: "life-" + uuid.NewString()[:8],
		})
		if err != nil {
			t.Fatalf("create compute cluster user: %v", err)
		}
		m, err := svc.CreateComputeAllocationMembership(ctx(), &models.ComputeAllocationMembership{
			ComputeAllocationID: alloc.ID, UserID: userID, StartTime: alloc.StartTime, EndTime: end, MembershipStatus: status,
		})
		if err != nil {
			t.Fatalf("create membership: %v", err)
		}
		f.members, f.users = append(f.members, m), append(f.users, cu)
	}
	return f
}

// newReadyFixture provisions every account and the group, so every member is ready.
func newReadyFixture(t *testing.T, svc *Service, database *sqlx.DB, members int) lifecycleFixture {
	t.Helper()
	f := newLifecycleFixture(t, svc, database, members, models.ACTIVE)
	for _, cu := range f.users {
		if err := svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID); err != nil {
			t.Fatalf("mark provisioned: %v", err)
		}
	}
	if err := svc.MarkComputeAllocationGroupProvisioned(ctx(), f.alloc.ID); err != nil {
		t.Fatalf("mark group provisioned: %v", err)
	}
	return f
}

// subscribeAll registers a no-op subscriber on each topic, so its publishes leave delivery rows.
func subscribeAll(svc *Service, topics ...events.EventType) {
	for _, topic := range topics {
		svc.EventBus().Subscribe("test-subscriber", topic, func(context.Context, events.Event, any) error { return nil })
	}
}

// published returns the payloads delivered on topic, oldest first.
func published[T any](t *testing.T, svc *Service, topic events.EventType) []T {
	t.Helper()
	rows, err := svc.EventBus().ListDeliveries(ctx(), "", 1000)
	if err != nil {
		t.Fatalf("list deliveries: %v", err)
	}
	slices.Reverse(rows)
	var out []T
	for _, r := range rows {
		if events.EventType(r.Event.EventType) != topic {
			continue
		}
		var l T
		if err := json.Unmarshal(r.Event.Payload, &l); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		out = append(out, l)
	}
	return out
}

func countTopic(t *testing.T, svc *Service, topic events.EventType) int {
	t.Helper()
	return len(published[json.RawMessage](t, svc, topic))
}

// Make sure a membership becomes ready, publishing ready once, on whichever condition arrives last.
func TestMembershipReady_OnLastCondition(t *testing.T) {
	steps := map[string]func(t *testing.T, svc *Service, f lifecycleFixture) error{
		"membership": func(t *testing.T, svc *Service, f lifecycleFixture) error {
			_, err := svc.UpdateMembershipStatus(ctx(), f.members[0].ID, models.ACTIVE)
			return err
		},
		"account": func(t *testing.T, svc *Service, f lifecycleFixture) error {
			return svc.MarkComputeClusterUserProvisioned(ctx(), f.users[0].ID)
		},
		"scheduler": func(t *testing.T, svc *Service, f lifecycleFixture) error {
			return svc.MarkComputeAllocationSchedulerProvisioned(ctx(), f.alloc.ID)
		},
	}
	for _, order := range [][]string{
		{"membership", "account", "scheduler"},
		{"account", "scheduler", "membership"},
		{"scheduler", "membership", "account"},
	} {
		t.Run(fmt.Sprint(order), func(t *testing.T) {
			database := setupTestDB(t)
			svc := newTestService(database)
			subscribeAll(svc, events.ComputeAllocationSchedulerRequestEvent, events.ComputeAllocationMembershipReadyEvent)
			f := newLifecycleFixture(t, svc, database, 1, models.INACTIVE)
			for _, step := range order {
				if n := countTopic(t, svc, events.ComputeAllocationMembershipReadyEvent); n != 0 {
					t.Fatalf("%d ready before %s, want 0", n, step)
				}
				if err := steps[step](t, svc, f); err != nil {
					t.Fatalf("%s: %v", step, err)
				}
			}
			if got := countTopic(t, svc, events.ComputeAllocationMembershipReadyEvent); got != 1 {
				t.Fatalf("%d ready deliveries, want 1", got)
			}
			m, err := svc.GetComputeAllocationMembership(ctx(), f.members[0].ID)
			if err != nil || m.ReadyAt == nil {
				t.Fatalf("ready_at not set: %+v, %v", m, err)
			}
		})
	}
}

// Make sure an inactive status deactivates a ready membership once.
func TestMembershipDeactivate_OnInactiveStatus(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	f := newReadyFixture(t, svc, database, 1)
	subscribeAll(svc, events.ComputeAllocationMembershipDeactivateEvent)

	if _, err := svc.UpdateMembershipStatus(ctx(), f.members[0].ID, models.INACTIVE); err != nil {
		t.Fatalf("set inactive: %v", err)
	}
	if n := countTopic(t, svc, events.ComputeAllocationMembershipDeactivateEvent); n != 1 {
		t.Fatalf("%d deactivate deliveries, want 1", n)
	}
	m, err := svc.GetComputeAllocationMembership(ctx(), f.members[0].ID)
	if err != nil || m.ReadyAt != nil {
		t.Fatalf("ready_at not cleared: %+v, %v", m, err)
	}
}

// Make sure moving a ready membership deactivates the old pair, readies the new one and resyncs both groups.
func TestMembershipMove_ReadiesOnNewAllocation(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	f := newReadyFixture(t, svc, database, 1)
	other, err := svc.CreateComputeAllocation(ctx(), &models.ComputeAllocation{Name: "Other", ProjectID: f.alloc.ProjectID,
		ComputeClusterID: f.alloc.ComputeClusterID, StartTime: f.alloc.StartTime, EndTime: f.alloc.EndTime})
	if err != nil {
		t.Fatalf("create allocation: %v", err)
	}
	subscribeAll(svc, events.ComputeAllocationMembershipDeactivateEvent, events.ComputeAllocationMembershipReadyEvent, events.ComputeAllocationGroupRequestEvent)
	if _, err := svc.UpdateComputeAllocationMembership(ctx(), &models.ComputeAllocationMembership{ID: f.members[0].ID, ComputeAllocationID: other.ID}); err != nil {
		t.Fatalf("move: %v", err)
	}
	if m, err := svc.GetComputeAllocationMembership(ctx(), f.members[0].ID); err != nil || m.ComputeAllocationID != other.ID || m.ReadyAt == nil {
		t.Fatalf("after move: %+v, %v; want ready on %s", m, err, other.ID)
	}
	if d, r := countTopic(t, svc, events.ComputeAllocationMembershipDeactivateEvent), countTopic(t, svc, events.ComputeAllocationMembershipReadyEvent); d != 1 || r != 1 {
		t.Fatalf("%d deactivate and %d ready deliveries, want 1 and 1", d, r)
	}
	if got := published[models.ComputeAllocation](t, svc, events.ComputeAllocationGroupRequestEvent); len(got) != 2 || got[0].ID == got[1].ID {
		t.Fatalf("group requests %+v, want one each for %s and %s", got, f.alloc.ID, other.ID)
	}
}

// Make sure an allocation's status deactivates and readies each member.
func TestAllocationStatus_PublishesReadiness(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	f := newReadyFixture(t, svc, database, 2)
	subscribeAll(svc, events.ComputeAllocationMembershipDeactivateEvent, events.ComputeAllocationMembershipReadyEvent)
	for _, status := range []models.AllocationStatus{models.INACTIVE, models.ACTIVE} {
		alloc := *f.alloc
		alloc.Status = status
		if err := svc.UpdateComputeAllocation(ctx(), &alloc); err != nil {
			t.Fatalf("set %s: %v", status, err)
		}
		for _, mem := range f.members {
			if m, err := svc.GetComputeAllocationMembership(ctx(), mem.ID); err != nil || (m.ReadyAt != nil) != (status == models.ACTIVE) {
				t.Fatalf("%s: %+v, %v", status, m, err)
			}
		}
	}
	if d, r := countTopic(t, svc, events.ComputeAllocationMembershipDeactivateEvent), countTopic(t, svc, events.ComputeAllocationMembershipReadyEvent); d != 2 || r != 2 {
		t.Fatalf("%d deactivate and %d ready deliveries, want 2 and 2", d, r)
	}
}

// Make sure deleting a ready member's account deactivates the membership.
func TestMembershipDeactivate_OnAccountDelete(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	f := newReadyFixture(t, svc, database, 1)
	subscribeAll(svc, events.ComputeAllocationMembershipDeactivateEvent)

	if err := svc.DeleteComputeClusterUser(ctx(), f.users[0].ID); err != nil {
		t.Fatalf("delete account: %v", err)
	}
	if n := countTopic(t, svc, events.ComputeAllocationMembershipDeactivateEvent); n != 1 {
		t.Fatalf("%d deactivate deliveries, want 1", n)
	}
	if m, err := svc.GetComputeAllocationMembership(ctx(), f.members[0].ID); err != nil || m.ReadyAt != nil {
		t.Fatalf("ready_at not cleared: %+v, %v", m, err)
	}
}
