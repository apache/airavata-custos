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
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// Make sure a merge publishes the delete event only for the retiring user's
// cluster account that it drops, which is the one on a cluster where the
// survivor already has an account. An account that moves to the survivor is
// not deleted, so it gets no event.
func TestMergeUsers_PublishesDeleteForTheDroppedClusterAccount(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	svc.EventBus().SubscribeComputeClusterUserDeleted("test-subscriber", func(context.Context, models.ComputeClusterUser) error { return nil })

	survivor := seedUser(t, database, fmt.Sprintf("survivor-%s@example.edu", uuid.NewString()))
	retiring := seedUser(t, database, fmt.Sprintf("retiring-%s@example.edu", uuid.NewString()))
	account := func(userID string, cluster *models.ComputeCluster) *models.ComputeClusterUser {
		cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{
			ComputeClusterID: cluster.ID,
			UserID:           userID,
			LocalUsername:    "merge-" + uuid.NewString()[:8],
		})
		if err != nil {
			t.Fatalf("create compute cluster user: %v", err)
		}
		return cu
	}
	cluster := func() *models.ComputeCluster {
		c, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "merge-" + uuid.NewString()[:8]})
		if err != nil {
			t.Fatalf("create cluster: %v", err)
		}
		return c
	}
	shared, onlyRetiring, retiringProvisioned, bothProvisioned := cluster(), cluster(), cluster(), cluster()
	account(survivor, shared)
	dropped := account(retiring, shared)
	moved := account(retiring, onlyRetiring)
	deletedSurvivor := account(survivor, retiringProvisioned)
	kept := account(retiring, retiringProvisioned)
	deletedBoth := account(retiring, bothProvisioned)
	for _, cu := range []*models.ComputeClusterUser{kept, deletedBoth, account(survivor, bothProvisioned)} {
		if err := svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID); err != nil {
			t.Fatalf("mark provisioned: %v", err)
		}
	}

	if _, err := svc.MergeUsers(ctx(), survivor, retiring); err != nil {
		t.Fatalf("merge: %v", err)
	}

	deliveries, err := svc.EventBus().ListDeliveries(ctx(), "", "", 10)
	if err != nil {
		t.Fatalf("list deliveries: %v", err)
	}
	want := map[string]string{dropped.ID: dropped.LocalUsername, deletedSurvivor.ID: deletedSurvivor.LocalUsername, deletedBoth.ID: deletedBoth.LocalUsername}
	if len(deliveries) != len(want) {
		t.Fatalf("expected %d deliveries of the delete event, got %+v", len(want), deliveries)
	}
	for _, d := range deliveries {
		var got models.ComputeClusterUser
		if err := json.Unmarshal(d.Event.Payload, &got); err != nil || d.Event.EventType != string(events.ComputeClusterUserDeleteEvent) || want[got.ID] != got.LocalUsername {
			t.Errorf("%s is for %s (%s), %v, want a delete event for a deleted account %v", d.Event.EventType, got.ID, got.LocalUsername, err, want)
		}
		delete(want, got.ID)
	}

	after, err := svc.GetComputeClusterUser(ctx(), moved.ID)
	if err != nil {
		t.Fatalf("get moved account: %v", err)
	}
	if after.UserID != survivor {
		t.Errorf("moved account belongs to %s, want the survivor %s", after.UserID, survivor)
	}
	if after, err := svc.GetComputeClusterUser(ctx(), kept.ID); err != nil || after.UserID != survivor {
		t.Errorf("provisioned account: got %+v, %v, want owner %s", after, err, survivor)
	}
}

// Make sure a merge keeps the higher tag per project and that the survivor
// holds no tag on a project it leads.
func TestMergeUsers_KeepsHigherTagAndPIHoldsNoTag(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	user := func(name string) string {
		return seedUser(t, database, fmt.Sprintf("%s-%s@example.edu", name, uuid.NewString()))
	}
	survivor, retiring, other := user("survivor"), user("retiring"), user("other")
	project := func(pi string, tags map[string]string) string {
		p, err := svc.CreateProject(ctx(), &models.Project{Title: "merge-" + uuid.NewString()[:8], ProjectPIID: pi})
		if err != nil {
			t.Fatalf("create project: %v", err)
		}
		for u, role := range tags {
			if err := svc.EnsureProjectMembership(ctx(), p.ID, u, role); err != nil {
				t.Fatalf("tag %s: %v", role, err)
			}
		}
		return p.ID
	}
	led := project(retiring, map[string]string{survivor: "CO_PI"})
	own := project(survivor, map[string]string{retiring: "ALLOCATION_MANAGER"})
	shared := project(other, map[string]string{retiring: "CO_PI", survivor: "ALLOCATION_MANAGER"})

	if _, err := svc.MergeUsers(ctx(), survivor, retiring); err != nil {
		t.Fatalf("merge: %v", err)
	}

	if p, err := svc.GetProject(ctx(), led); err != nil || p.ProjectPIID != survivor {
		t.Errorf("led project: %+v (%v), want PI %s", p, err, survivor)
	}
	for p, want := range map[string]string{led: "", own: "", shared: survivor + ":CO_PI"} {
		var tags []string
		err := database.SelectContext(ctx(), &tags, `SELECT user_id || ':' || role FROM project_memberships WHERE project_id = $1`, p)
		got := strings.Join(tags, "")
		if err != nil || got != want {
			t.Errorf("project %s tags: %q (%v), want %q", p, got, err, want)
		}
	}
}

// Make sure a merge waiting on a cluster account another transaction holds
// does not block that transaction from inserting a row that references the user.
func TestMergeUsers_LetsALockHolderReferenceTheUser(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	survivor := seedUser(t, database, fmt.Sprintf("survivor-%s@example.edu", uuid.NewString()))
	cu := newClusterUser(t, svc, database)

	hold, err := database.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = hold.Rollback() }()
	if _, err := hold.Exec("SELECT id FROM compute_cluster_users WHERE id = $1 FOR UPDATE", cu.ID); err != nil {
		t.Fatal(err)
	}
	merged := make(chan error, 1)
	go func() {
		_, err := svc.MergeUsers(ctx(), survivor, cu.UserID)
		merged <- err
	}()
	for n, deadline := 0, time.Now().Add(5*time.Second); n < 1; time.Sleep(10 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatal("merge never blocked on the cluster account lock")
		}
		if err := database.Get(&n, "SELECT count(*) FROM pg_stat_activity WHERE wait_event_type = 'Lock'"); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := hold.Exec("INSERT INTO user_identities (id, user_id, source, external_id) VALUES ($1, $2, 'test', $3)",
		uuid.NewString(), cu.UserID, uuid.NewString()); err != nil {
		t.Fatalf("insert identity while the merge waits: %v", err)
	}
	if err := hold.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := <-merged; err != nil {
		t.Fatalf("merge: %v", err)
	}
}

func TestUpdateUser_KeepsStatus(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	user := seedUser(t, database, "keeps-status@example.edu")
	if err := svc.UpdateUser(ctx(), &models.User{ID: user, FirstName: "Renamed", Status: models.UserInactive}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if after, err := svc.GetUser(ctx(), user); err != nil || after.FirstName != "Renamed" || after.Status != models.UserActive {
		t.Errorf("user: got %+v, %v, want renamed and ACTIVE", after, err)
	}
}

func TestDeleteUser_RemovesTheirAllocationMemberships(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	_, allocID := newClusterAllocation(t, svc, seedUser(t, database, "pi@example.edu"))
	member := seedUser(t, database, "member@example.edu")
	if _, err := svc.CreateComputeAllocationMembership(ctx(), &models.ComputeAllocationMembership{ComputeAllocationID: allocID, UserID: member, MembershipStatus: models.ACTIVE}); err != nil {
		t.Fatalf("seed membership: %v", err)
	}
	if err := svc.DeleteUser(ctx(), member); err != nil {
		t.Fatalf("delete user: %v", err)
	}
	if left, _ := svc.memberships.FindByUser(ctx(), member); len(left) != 0 {
		t.Fatalf("%d memberships left", len(left))
	}
}

func TestUpdateUserStatus_RefusesMergedUser(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	survivor := seedUser(t, database, "survivor@example.edu")
	retiring := seedUser(t, database, "retiring@example.edu")
	if _, err := svc.MergeUsers(ctx(), survivor, retiring); err != nil {
		t.Fatalf("merge: %v", err)
	}
	for _, status := range []models.UserStatus{models.UserMerged, models.UserActive} {
		if _, err := svc.UpdateUserStatus(ctx(), retiring, status); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("set %q on a merged user: got %v, want ErrInvalidInput", status, err)
		}
	}
	for _, status := range []models.UserStatus{"", "SUSPENDED", models.UserPending, models.UserMerged} {
		if _, err := svc.UpdateUserStatus(ctx(), survivor, status); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("set %q: got %v, want ErrInvalidInput", status, err)
		}
	}
}

func TestDeleteUser_RefusesAUserStillReferenced(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	pi := seedUser(t, database, "pi@example.edu")
	newClusterAllocation(t, svc, pi)
	if err := svc.DeleteUser(ctx(), pi); !errors.Is(err, ErrInUse) {
		t.Fatalf("delete a project PI: %v, want ErrInUse", err)
	}
}
