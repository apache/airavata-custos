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

//go:build integration

package service

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
)

// Make sure a storage policy's templates must fit its type, and a parent is listed before what is under it.
func TestStoragePolicies(t *testing.T) {
	svc := newTestService(setupTestDB(t))
	cluster := storageCluster(t, svc)
	_, err := svc.CreateStoragePolicy(ctx(), &models.StoragePolicy{ComputeClusterID: cluster.ID, Type: models.StorageProject, PathPattern: "/bad/{allocation}", Owner: "{user}", Group: "{allocation}", Mode: "2700"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("project owned by {user}: got %v, want ErrInvalidInput", err)
	}
	got, err := svc.ListStoragePolicies(ctx(), cluster.ID)
	if err != nil || len(got) != 3 || got[1].Type != models.StorageProject {
		t.Fatalf("list: err=%v got=%+v, want home, project, then member", err, got)
	}
	home := got[0]
	home.QuotaBytes = 200
	if updated, err := svc.UpdateStoragePolicy(ctx(), &home); err != nil || updated.QuotaBytes != 200 || updated.ComputeClusterID != cluster.ID {
		t.Errorf("update: err=%v got=%+v, want the home policy with a quota of 200", err, updated)
	}
	home.Type, home.PathPattern, home.Owner, home.Group = models.StorageProject, "/home/{allocation}", "0", "{allocation}"
	if _, err := svc.UpdateStoragePolicy(ctx(), &home); !errors.Is(err, ErrNotFound) {
		t.Errorf("update to another type: got %v, want ErrNotFound", err)
	}
}

// storageCluster makes a cluster with a home, a project and a member storage policy.
func storageCluster(t *testing.T, svc *Service) *models.ComputeCluster {
	t.Helper()
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "c-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	for _, p := range []models.StoragePolicy{
		{Type: models.StorageUser, PathPattern: "/home/{user}", Owner: "{user}", Group: "{user}", Mode: "0700", QuotaBytes: 100},
		{Type: models.StorageProjectUser, PathPattern: "/project/{allocation}/{user}", Owner: "{user}", Group: "{allocation}", Mode: "2700"},
		{Type: models.StorageProject, PathPattern: "/project/{allocation}", Owner: "0", Group: "{allocation}", Mode: "2750"},
	} {
		p.ComputeClusterID = cluster.ID
		if _, err := svc.CreateStoragePolicy(ctx(), &p); err != nil {
			t.Fatalf("create path %s: %v", p.PathPattern, err)
		}
	}
	return cluster
}

// storageAllocation makes an allocation of a new project on the cluster.
func storageAllocation(t *testing.T, svc *Service, database *sqlx.DB, clusterID string, status models.AllocationStatus) *models.ComputeAllocation {
	t.Helper()
	project, err := svc.CreateProject(ctx(), &models.Project{Title: "p", Origination: "TEST", ProjectPIID: seedUser(t, database, uuid.NewString()+"@example.edu"), Status: models.ProjectActive})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	alloc, err := svc.CreateComputeAllocation(ctx(), &models.ComputeAllocation{
		Name: "Grant", ProjectID: project.ID, ComputeClusterID: clusterID, Status: status, StartTime: time.Now(), EndTime: time.Now().AddDate(1, 0, 0),
	})
	if err != nil {
		t.Fatalf("create allocation: %v", err)
	}
	return alloc
}

// Make sure a storage is made from its type's path for its owner, deleting an allocation marks its storages and those
// under them, a returning owner's is active again, and a storage is deleted only once everything under it is pending.
func TestStorages(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cluster := storageCluster(t, svc)
	project, err := svc.CreateProject(ctx(), &models.Project{Title: "p", Origination: "TEST", ProjectPIID: seedUser(t, database, uuid.NewString()+"@example.edu"), Status: models.ProjectActive})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	// The second allocation's group extends the first's, so its storage is beside the first's and not under it.
	allocation := func() *models.ComputeAllocation {
		a, err := svc.CreateComputeAllocation(ctx(), &models.ComputeAllocation{
			Name: "Grant", ProjectID: project.ID, ComputeClusterID: cluster.ID, Status: models.ACTIVE, StartTime: time.Now(), EndTime: time.Now().AddDate(1, 0, 0),
		})
		if err == nil {
			err = svc.MarkComputeAllocationGroupProvisioned(ctx(), a.ID)
		}
		if err != nil {
			t.Fatalf("provision allocation: %v", err)
		}
		return a
	}
	a, ab := allocation(), allocation()
	group, beside := "/project/"+*a.PosixGroup, "/project/"+*ab.PosixGroup
	ensure := func() []models.Storage {
		made, err := svc.EnsureStorages(ctx(), cluster.ID)
		if err != nil {
			t.Fatalf("ensure: %v", err)
		}
		return made
	}
	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: cluster.ID, UserID: seedUser(t, database, uuid.NewString()+"@example.edu"), LocalUsername: "jdoe"})
	if err == nil {
		err = svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID)
	}
	if err != nil {
		t.Fatalf("provision account: %v", err)
	}
	if _, err := svc.CreateComputeAllocationMembership(ctx(), &models.ComputeAllocationMembership{ComputeAllocationID: a.ID, UserID: cu.UserID}); err != nil {
		t.Fatalf("create membership: %v", err)
	}
	byPath := func() map[string]models.Storage {
		list, _, err := svc.ListStorages(ctx(), cluster.ID, 200, 0)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		got := map[string]models.Storage{}
		for _, st := range list {
			got[st.Path] = st
		}
		return got
	}
	mark := func() {
		if err := svc.inTx(ctx(), func(tx *sql.Tx) error {
			return svc.storages.MarkForDeletion(ctx(), tx, cluster.ID, "", a.ID)
		}); err != nil {
			t.Fatalf("mark: %v", err)
		}
	}

	if m := byPath()[group+"/jdoe"]; m.Username != "jdoe" || m.AllocationID == nil || *m.AllocationID != a.ID || m.Group != *a.PosixGroup || m.Mode != "2700" || m.PolicyID == nil {
		t.Fatalf("member storage: %+v", m)
	}
	until := time.Now().Add(time.Hour)
	if err := svc.RecordStorageUsage(ctx(), cluster.ID, []string{"/home/jdoe"}, []int64{40}); err != nil || byPath()["/home/jdoe"].UsedBytes != 40 {
		t.Fatalf("usage: err=%v got=%+v", err, byPath()["/home/jdoe"])
	}
	if err := svc.SetStorageQuota(ctx(), byPath()["/home/jdoe"].ID, 500, &until, "", "admin"); err != nil || len(ensure()) != 0 || byPath()["/home/jdoe"].QuotaBytes != 500 {
		t.Fatalf("administrator's quota: err=%v, want it to stand and nothing to change when storages are ensured again", err)
	}

	mark()
	if got := byPath(); got[group+"/jdoe"].Status != "pending_deletion" || got[beside].Status != "active" {
		t.Fatalf("after marking: %+v", got)
	}
	if back := ensure(); len(back) != 2 {
		t.Fatalf("ensuring marked storages: got %+v, want the project's and the member's active again", back)
	}
	if err := svc.DeleteStorage(ctx(), byPath()[group].ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("delete over an active storage: got %v, want ErrNotFound", err)
	}
	mark()
	svc.EventBus().SubscribeStorageDeleted("test-subscriber", func(context.Context, models.Storage) error { return nil })
	if err := svc.DeleteStorage(ctx(), byPath()[group].ID); err != nil || len(byPath()) != 2 ||
		countRows(t, database, "SELECT COUNT(*) FROM events WHERE event_type = 'storage::delete'") != 2 {
		t.Fatalf("delete: err=%v left=%+v, want one delete event for the storage and one for the member's under it", err, byPath())
	}
	far := time.Now().AddDate(0, 0, 91)
	for name, q := range map[string]struct {
		bytes int64
		until *time.Time
	}{"zero": {0, &until}, "no end": {1, nil}, "past 90 days": {1, &far}} {
		if err := svc.SetStorageQuota(ctx(), byPath()["/home/jdoe"].ID, q.bytes, q.until, "", "admin"); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("quota %s: got %v, want ErrInvalidInput", name, err)
		}
	}
	if err := svc.SetStorageQuota(ctx(), byPath()[beside].ID, 1, &until, "", "admin"); !errors.Is(err, ErrNotFound) {
		t.Errorf("quota on a storage without one of its own: got %v, want ErrNotFound", err)
	}
	err = errors.Join(svc.DeleteComputeClusterUser(ctx(), cu.ID), svc.DeleteComputeAllocation(ctx(), ab.ID))
	if got := byPath(); err != nil || got["/home/jdoe"].Status != "pending_deletion" || got[beside].Status != "pending_deletion" {
		t.Fatalf("deleting the account and an allocation: err=%v storages=%+v, want theirs pending", err, got)
	}
}

// Make sure core creates a storage only once its owner is on the cluster: the account, the active allocation's group,
// or both.
func TestStoragesFollowProvisioning(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cluster := storageCluster(t, svc)
	alloc := storageAllocation(t, svc, database, cluster.ID, models.INACTIVE)
	account := func(username string) *models.ComputeClusterUser {
		cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: cluster.ID, UserID: seedUser(t, database, uuid.NewString()+"@example.edu"), LocalUsername: username})
		if err != nil {
			t.Fatalf("create account: %v", err)
		}
		return cu
	}
	join := func(cu *models.ComputeClusterUser) {
		if _, err := svc.CreateComputeAllocationMembership(ctx(), &models.ComputeAllocationMembership{ComputeAllocationID: alloc.ID, UserID: cu.UserID}); err != nil {
			t.Fatalf("create membership: %v", err)
		}
	}
	svc.EventBus().SubscribeStorageCreated("test-subscriber", func(context.Context, models.Storage) error { return nil })
	check := func(step string, err error, want ...string) {
		var got []string
		if err == nil {
			err = database.Select(&got, `SELECT path FROM storages WHERE compute_cluster_id = $1 ORDER BY path COLLATE "C"`, cluster.ID)
		}
		if n := countRows(t, database, "SELECT COUNT(*) FROM events WHERE event_type = 'storage::create'"); err != nil || !slices.Equal(got, want) || n != len(want) {
			t.Fatalf("%s: err=%v storages=%v events=%d, want %v", step, err, got, n, want)
		}
	}
	group, early, late := "/project/"+*alloc.PosixGroup, account("early"), account("late")
	join(early)
	check("a member before anything is provisioned", nil)
	check("the account", svc.MarkComputeClusterUserProvisioned(ctx(), early.ID), "/home/early")
	check("the group of an inactive allocation", svc.MarkComputeAllocationGroupProvisioned(ctx(), alloc.ID), "/home/early")
	alloc.Status = models.ACTIVE
	check("the allocation becoming active", svc.UpdateComputeAllocation(ctx(), alloc), "/home/early", group, group+"/early")
	err := svc.MarkComputeClusterUserProvisioned(ctx(), late.ID)
	join(late)
	check("a provisioned account joining", err, "/home/early", "/home/late", group, group+"/early", group+"/late")
	if again, err := svc.EnsureStorages(ctx(), cluster.ID); err != nil || len(again) != 0 {
		t.Fatalf("ensuring again: err=%v got=%+v, want nothing made or changed", err, again)
	}
}

// Make sure an account provisioned while its allocation's group is being provisioned still gets its member storage.
func TestMemberStorageSurvivesConcurrentProvisioning(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cluster := storageCluster(t, svc)
	alloc := storageAllocation(t, svc, database, cluster.ID, models.ACTIVE)
	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: cluster.ID, UserID: seedUser(t, database, uuid.NewString()+"@example.edu"), LocalUsername: "racer"})
	if err == nil {
		_, err = svc.CreateComputeAllocationMembership(ctx(), &models.ComputeAllocationMembership{ComputeAllocationID: alloc.ID, UserID: cu.UserID})
	}
	if err != nil {
		t.Fatalf("seed member: %v", err)
	}
	account := make(chan error, 1)
	if err := svc.inTx(ctx(), func(tx *sql.Tx) error {
		if err := svc.markGroupProvisionedTx(ctx(), tx, alloc); err != nil {
			return err
		}
		go func() { account <- svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID) }()
		time.Sleep(300 * time.Millisecond) // lets the account's transaction reach the allocation while the group's is open
		return nil
	}); err != nil {
		t.Fatalf("provision group: %v", err)
	}
	if err := <-account; err != nil {
		t.Fatalf("provision account: %v", err)
	}
	if n := countRows(t, database, "SELECT COUNT(*) FROM storages WHERE path = $1", "/project/"+*alloc.PosixGroup+"/racer"); n != 1 {
		t.Fatalf("member storage: got %d, want 1", n)
	}
}

// Make sure an account provisioned while it is being added to an allocation still gets its member storage.
func TestMemberStorageSurvivesConcurrentMembership(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cluster := storageCluster(t, svc)
	alloc := storageAllocation(t, svc, database, cluster.ID, models.ACTIVE)
	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: cluster.ID, UserID: seedUser(t, database, uuid.NewString()+"@example.edu"), LocalUsername: "joiner"})
	if err == nil {
		err = svc.MarkComputeAllocationGroupProvisioned(ctx(), alloc.ID)
	}
	if err != nil {
		t.Fatalf("seed account and group: %v", err)
	}
	account := make(chan error, 1)
	if err := svc.inTx(ctx(), func(tx *sql.Tx) error {
		m := &models.ComputeAllocationMembership{ID: newID(), ComputeAllocationID: alloc.ID, UserID: cu.UserID, MembershipStatus: models.ACTIVE}
		if err := svc.memberships.Create(ctx(), tx, m); err != nil {
			return err
		}
		if _, err := svc.ensureStorages(ctx(), tx, store.StorageOwners{UserID: m.UserID, AllocationID: m.ComputeAllocationID}); err != nil {
			return err
		}
		go func() { account <- svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID) }()
		time.Sleep(300 * time.Millisecond) // lets the account's transaction reach the account's row while the membership's is open
		return nil
	}); err != nil {
		t.Fatalf("add member: %v", err)
	}
	if err := <-account; err != nil {
		t.Fatalf("provision account: %v", err)
	}
	if n := countRows(t, database, "SELECT COUNT(*) FROM storages WHERE path = $1", "/project/"+*alloc.PosixGroup+"/joiner"); n != 1 {
		t.Fatalf("member storage: got %d, want 1", n)
	}
}

// Make sure members added to one allocation at the same moment do not deadlock on the allocation's row.
func TestConcurrentMembershipsDoNotDeadlock(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cluster := storageCluster(t, svc)
	alloc := storageAllocation(t, svc, database, cluster.ID, models.ACTIVE)
	const members = 8
	failed := make(chan error, members)
	for range members {
		user := seedUser(t, database, uuid.NewString()+"@example.edu")
		go func() {
			_, err := svc.CreateComputeAllocationMembership(ctx(), &models.ComputeAllocationMembership{ComputeAllocationID: alloc.ID, UserID: user})
			failed <- err
		}()
	}
	for range members {
		if err := <-failed; err != nil {
			t.Errorf("add member: %v", err)
		}
	}
}
