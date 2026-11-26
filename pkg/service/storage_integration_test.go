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
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

// storageCreates counts the storage::create events published.
const storageCreates = "SELECT COUNT(*) FROM events WHERE event_type = 'storage::create'"

// homeProjectMember are a home, a project and a member storage target.
var homeProjectMember = map[models.StorageType]models.StorageTarget{
	"user_home":    {Path: "/home/{user}", Owner: "{user}", Group: "{user}", Mode: "0700", HardLimitSize: 100},
	"project_user": {Path: "/project/{allocation}/{user}", Owner: "{user}", Group: "{allocation}", Mode: "2700"},
	"project":      {Path: "/project/{allocation}", Owner: "0", Group: "{allocation}", Mode: "2750"},
}

// newStorageCluster makes a claimed cluster with the targets.
func newStorageCluster(t *testing.T, svc *Service, targets map[models.StorageType]models.StorageTarget) *models.ComputeCluster {
	t.Helper()
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "c-" + uuid.NewString()[:8]})
	if err == nil {
		err = svc.SetStorageTargets(map[string]map[models.StorageType]models.StorageTarget{cluster.ID: targets})
	}
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	svc.ClaimStorageTargets(cluster.ID, func(context.Context) (map[string]int64, error) { return map[string]int64{"/home/jdoe": 7}, nil })
	svc.EventBus().SubscribeStorageCreated("test-subscriber", func(context.Context, models.Storage) error { return nil })
	return cluster
}

// newStorageAllocation makes an allocation on the cluster of the project, or of a new one when projectID is empty.
func newStorageAllocation(t *testing.T, svc *Service, database *sqlx.DB, clusterID, projectID string, status models.AllocationStatus) *models.ComputeAllocation {
	t.Helper()
	if projectID == "" {
		project, err := svc.CreateProject(ctx(), &models.Project{Title: "p", Origination: "TEST", ProjectPIID: seedUser(t, database, uuid.NewString()+"@example.edu"), Status: models.ProjectActive})
		if err != nil {
			t.Fatalf("create project: %v", err)
		}
		projectID = project.ID
	}
	alloc, err := svc.CreateComputeAllocation(ctx(), &models.ComputeAllocation{
		Name: "Grant", ProjectID: projectID, ComputeClusterID: clusterID, Status: status, StartTime: time.Now(), EndTime: time.Now().AddDate(1, 0, 0),
	})
	if err != nil {
		t.Fatalf("create allocation: %v", err)
	}
	return alloc
}

// Make sure a storage is made from its target for its owner, a retired owner's is active again once ensured, a storage
// is deleted only once everything under it is pending, and deleting an owner retires and unlinks its storages.
func TestStorageLifecycle(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cluster := newStorageCluster(t, svc, homeProjectMember)
	// The second allocation's group extends the first's, so its storage is beside the first's and not under it.
	allocation := func(projectID string) *models.ComputeAllocation {
		a := newStorageAllocation(t, svc, database, cluster.ID, projectID, models.ACTIVE)
		if err := svc.MarkComputeAllocationGroupProvisioned(ctx(), a.ID); err != nil {
			t.Fatalf("provision allocation: %v", err)
		}
		return a
	}
	a := allocation("")
	sibling := allocation(a.ProjectID)
	projectPath, siblingPath := "/project/"+*a.PosixGroup, "/project/"+*sibling.PosixGroup
	ensure := func() int {
		before := countRows(t, database, storageCreates)
		if _, err := svc.SyncStorages(ctx(), cluster.ID); err != nil {
			t.Fatalf("ensure: %v", err)
		}
		return countRows(t, database, storageCreates) - before
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
		list, err := svc.ListStorages(ctx())
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		got := map[string]models.Storage{}
		for _, st := range list {
			got[st.Path] = st
		}
		return got
	}
	if used := byPath()["/home/jdoe"].UsedBytes; used == nil || *used != 7 {
		t.Errorf("home use: got %v, want 7 from the storage connector", used)
	}
	// setAllocationStatus sets the first allocation's status, ensures, and returns how many storages that made or changed.
	setAllocationStatus := func(status models.AllocationStatus) int {
		before := countRows(t, database, storageCreates)
		a.Status = status
		if err := svc.UpdateComputeAllocation(ctx(), a); err != nil {
			t.Fatalf("set allocation %s: %v", status, err)
		}
		ensure()
		return countRows(t, database, storageCreates) - before
	}

	if m := byPath()[projectPath+"/jdoe"]; m.Username == nil || *m.Username != "jdoe" || m.AllocationID == nil || *m.AllocationID != a.ID || m.Group != *a.PosixGroup {
		t.Fatalf("member storage: %+v", m)
	}
	until := time.Now().Add(time.Hour)
	if err := svc.UpdateStorage(ctx(), byPath()["/home/jdoe"].ID, 500, &until, "", "admin"); err != nil || ensure() != 0 || byPath()["/home/jdoe"].QuotaBytes != 500 {
		t.Fatalf("administrator's quota: err=%v, want it to stand and nothing to change when storages are ensured again", err)
	}
	if err := svc.DeleteComputeClusterUser(ctx(), cu.ID); !errors.Is(err, ErrInUse) {
		t.Fatalf("deleting an account with an ACTIVE membership on an ACTIVE allocation: got %v, want ErrInUse", err)
	}

	setAllocationStatus(models.INACTIVE)
	if got := byPath(); got[projectPath+"/jdoe"].Status != "pending_deletion" || got[siblingPath].Status != "active" {
		t.Fatalf("after retiring: %+v", got)
	}
	if back := setAllocationStatus(models.ACTIVE); back != 2 {
		t.Fatalf("ensuring retired storages: made %d, want the project's and the member's active again", back)
	}
	if err := svc.DeleteStorage(ctx(), byPath()[projectPath].ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("delete over an active storage: got %v, want ErrNotFound", err)
	}
	setAllocationStatus(models.INACTIVE)
	svc.EventBus().SubscribeStorageDeleted("test-subscriber", func(context.Context, models.Storage) error { return nil })
	if err := svc.DeleteStorage(ctx(), byPath()[projectPath].ID); err != nil || len(byPath()) != 2 ||
		countRows(t, database, "SELECT COUNT(*) FROM events WHERE event_type = 'storage::delete'") != 1 {
		t.Fatalf("delete: err=%v left=%+v, want one delete event for the storage, which takes the member's under it", err, byPath())
	}
	if err := svc.UpdateStorage(ctx(), byPath()[siblingPath].ID, 1, &until, "", "admin"); !errors.Is(err, ErrNotFound) {
		t.Errorf("quota on a storage without one of its own: got %v, want ErrNotFound", err)
	}
	err = changeAllocationStatus(svc, sibling.ID, models.INACTIVE)
	clearPending(t, svc, SlurmPendingAllocation, sibling.ID) // SLURM has applied the revoke
	err = errors.Join(err, svc.DeleteComputeClusterUser(ctx(), cu.ID), svc.DeleteComputeAllocation(ctx(), sibling.ID))
	if got := byPath(); err != nil || got["/home/jdoe"].Status != "pending_deletion" || got[siblingPath].Status != "pending_deletion" ||
		got["/home/jdoe"].ComputeClusterUserID != nil || got[siblingPath].AllocationID != nil {
		t.Fatalf("deleting the account and an allocation: err=%v got=%+v, want theirs retired and unlinked", err, got)
	}
}

// Make sure core creates a storage only once its holder is on the cluster: the account, the active allocation's group,
// or both.
func TestStoragesFollowProvisioning(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cluster := newStorageCluster(t, svc, homeProjectMember)
	alloc := newStorageAllocation(t, svc, database, cluster.ID, "", models.INACTIVE)
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
	check := func(step string, err error, want ...string) {
		var got []string
		if err == nil {
			err = database.Select(&got, `SELECT path FROM storages WHERE compute_cluster_id = $1 ORDER BY path COLLATE "C"`, cluster.ID)
		}
		if n := countRows(t, database, storageCreates); err != nil || !slices.Equal(got, want) || n != len(want) {
			t.Fatalf("%s: err=%v storages=%v events=%d, want %v", step, err, got, n, want)
		}
	}
	projectPath, early, late := "/project/"+*alloc.PosixGroup, account("early"), account("late")
	join(early)
	check("a member before anything is provisioned", nil)
	check("the account", svc.MarkComputeClusterUserProvisioned(ctx(), early.ID), "/home/early")
	check("the group of an inactive allocation", svc.MarkComputeAllocationGroupProvisioned(ctx(), alloc.ID), "/home/early")
	alloc.Status = models.ACTIVE
	err := svc.UpdateComputeAllocation(ctx(), alloc)
	check("the allocation becoming active", err, "/home/early", projectPath, projectPath+"/early")
	err = svc.MarkComputeClusterUserProvisioned(ctx(), late.ID)
	join(late)
	check("a provisioned account joining", err, "/home/early", "/home/late", projectPath, projectPath+"/early", projectPath+"/late")
	if desired, err := svc.SyncStorages(ctx(), cluster.ID); err != nil || len(desired) != 5 || countRows(t, database, storageCreates) != 5 {
		t.Fatalf("ensuring again: err=%v desired=%d, want the five storages and nothing made or changed", err, len(desired))
	}
}

// Make sure an account provisioned while its member storage's other owner is being made still gets that storage.
func TestMemberStorageSurvivesConcurrentProvisioning(t *testing.T) {
	for name, open := range map[string]func(svc *Service, tx *sql.Tx, alloc *models.ComputeAllocation, cu *models.ComputeClusterUser) error{
		"group provisioned": func(svc *Service, tx *sql.Tx, alloc *models.ComputeAllocation, cu *models.ComputeClusterUser) error {
			if _, err := svc.CreateComputeAllocationMembership(ctx(), &models.ComputeAllocationMembership{ComputeAllocationID: alloc.ID, UserID: cu.UserID}); err != nil {
				return err
			}
			return svc.markGroupProvisionedTx(ctx(), tx, alloc)
		},
		"membership added": func(svc *Service, tx *sql.Tx, alloc *models.ComputeAllocation, cu *models.ComputeClusterUser) error {
			if err := svc.MarkComputeAllocationGroupProvisioned(ctx(), alloc.ID); err != nil {
				return err
			}
			m := &models.ComputeAllocationMembership{ID: newID(), ComputeAllocationID: alloc.ID, UserID: cu.UserID, MembershipStatus: models.ACTIVE}
			if err := svc.memberships.Create(ctx(), tx, m); err != nil {
				return err
			}
			return svc.syncStoragesTx(ctx(), tx, store.StorageScope{UserID: m.UserID, AllocationID: m.ComputeAllocationID})
		},
	} {
		t.Run(name, func(t *testing.T) {
			database := setupTestDB(t)
			svc := newTestService(database)
			cluster := newStorageCluster(t, svc, homeProjectMember)
			alloc := newStorageAllocation(t, svc, database, cluster.ID, "", models.ACTIVE)
			cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: cluster.ID, UserID: seedUser(t, database, uuid.NewString()+"@example.edu"), LocalUsername: "racer"})
			if err != nil {
				t.Fatalf("seed account: %v", err)
			}
			account := make(chan error, 1)
			if err := svc.inTx(ctx(), func(tx *sql.Tx) error {
				if err := open(svc, tx, alloc, cu); err != nil {
					return err
				}
				go func() { account <- svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID) }()
				time.Sleep(300 * time.Millisecond) // lets the account's transaction reach the locked row while this one is open
				return nil
			}); err != nil {
				t.Fatalf("open: %v", err)
			}
			if err := <-account; err != nil {
				t.Fatalf("provision account: %v", err)
			}
			if n := countRows(t, database, "SELECT COUNT(*) FROM storages WHERE path = $1", "/project/"+*alloc.PosixGroup+"/racer"); n != 1 {
				t.Fatalf("member storage: got %d, want 1", n)
			}
		})
	}
}

// Make sure members added to one allocation at the same moment do not deadlock on the allocation's row.
func TestConcurrentMembershipsDoNotDeadlock(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cluster := newStorageCluster(t, svc, homeProjectMember)
	alloc := newStorageAllocation(t, svc, database, cluster.ID, "", models.ACTIVE)
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

// Make sure a connector's claimed type follows its holders: a held user gets one on every claimed cluster, losing the
// hold retires it, regaining restores it, deleting the user retires and unlinks it, and an unclaimed type is kept.
func TestClaimedStorageHolders(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cluster := newStorageCluster(t, svc, map[models.StorageType]models.StorageTarget{
		"trial_home": {Path: "/trial-home/{name}", Owner: "0", Group: "0", Mode: "0700"},
		"unclaimed":  {Path: "/unclaimed/{name}", Owner: "0", Group: "0", Mode: "0700"},
	})
	svc.ClaimStorageHolders("trial_home", func(u models.User) (string, bool) { return u.ID, u.Status == models.UserActive })
	if _, err := database.Exec(`INSERT INTO storages (id, compute_cluster_id, type, path, owner, "group", mode) VALUES ($1, $2, 'unclaimed', '/unclaimed/x', '0', '0', '0700')`,
		uuid.NewString(), cluster.ID); err != nil {
		t.Fatalf("seed unclaimed storage: %v", err)
	}
	user, err := svc.CreateUser(ctx(), &models.User{Email: uuid.NewString() + "@example.edu", OrganizationID: seedOrg(t, database), Status: models.UserActive})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	storage := func(path string) (status string, userID *string) {
		_ = database.QueryRow(`SELECT status, user_id FROM storages WHERE compute_cluster_id = $1 AND path = $2`, cluster.ID, path).Scan(&status, &userID)
		return
	}
	home := "/trial-home/" + user.ID
	for _, step := range []struct {
		status models.UserStatus
		want   string
	}{{"", "active"}, {models.UserInactive, "pending_deletion"}, {models.UserActive, "active"}} {
		if step.status != "" {
			if _, err := svc.UpdateUserStatus(ctx(), user.ID, step.status); err != nil {
				t.Fatalf("set %s: %v", step.status, err)
			}
		}
		if got, _ := storage(home); got != step.want {
			t.Fatalf("after %q: %s, want %s", step.status, got, step.want)
		}
	}
	if _, err := svc.SyncStorages(ctx(), cluster.ID); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if got, _ := storage("/unclaimed/x"); got != "active" {
		t.Fatalf("unclaimed: %s, want active", got)
	}
	if err := svc.DeleteUser(ctx(), user.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if got, userID := storage(home); got != "pending_deletion" || userID != nil {
		t.Fatalf("after delete: %s %v, want pending_deletion and unlinked", got, userID)
	}
}

// Make sure a {name} storage stays with its user when another gets the same name, and a name that would leave its
// parent or name the parent itself gives no storage.
func TestClaimedStorageNames(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cluster := newStorageCluster(t, svc, map[models.StorageType]models.StorageTarget{
		"trial_home": {Path: "/trial-home/{name}", Owner: "0", Group: "0", Mode: "0700"},
	})
	names := map[string]string{}
	svc.ClaimStorageHolders("trial_home", func(u models.User) (string, bool) { name, ok := names[u.ID]; return name, ok })
	org := seedOrg(t, database)
	user := func(name string) *models.User {
		u, err := svc.CreateUser(ctx(), &models.User{Email: uuid.NewString() + "@example.edu", OrganizationID: org})
		if err == nil {
			names[u.ID] = name
			u, err = svc.UpdateUserStatus(ctx(), u.ID, models.UserActive)
		}
		if err != nil {
			t.Fatalf("create user: %v", err)
		}
		return u
	}
	holder := func(path string) (userID string) {
		_ = database.QueryRow(`SELECT COALESCE(user_id, '') FROM storages WHERE compute_cluster_id = $1 AND path = $2 AND status = 'active'`, cluster.ID, path).Scan(&userID)
		return
	}
	name := "n" + uuid.NewString()[:8]
	older, newer := user(name), user(name)
	if got := holder("/trial-home/" + name); got != older.ID {
		t.Fatalf("shared name: held by %q, want the older user %q", got, older.ID)
	}
	if err := svc.DeleteUser(ctx(), newer.ID); err != nil || holder("/trial-home/"+name) != older.ID {
		t.Fatalf("deleting the newer user: %v, want the older user's storage active", err)
	}
	for _, bad := range []string{"", "..", "a/b"} {
		user(bad)
	}
	if n := countRows(t, database, `SELECT COUNT(*) FROM storages WHERE path IN ('/trial-home/', '/trial-home/..', '/trial-home/a/b')`); n != 0 {
		t.Fatalf("bad names gave %d storages, want none", n)
	}
}

// Make sure a provisioned account's home is active only while its user is ACTIVE, including after onboarding and a merge.
func TestStorages_ActiveOnlyWhileUserIsActive(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cluster := newStorageCluster(t, svc, homeProjectMember)
	home := func() string {
		list, err := svc.ListStorages(ctx())
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		for _, st := range list {
			if st.Path == "/home/jdoe" {
				return st.Status
			}
		}
		return ""
	}
	aup := seedPublishedAgreement(t, svc, seedAgreementManager(t, database), "1.0")
	email := "jdoe-" + uuid.NewString()[:8] + "@example.edu"
	pending := seedPendingUserWithEmail(t, database, email)
	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: cluster.ID, UserID: pending, LocalUsername: "custos-jdoe"})
	if err == nil {
		err = svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID)
	}
	if err != nil {
		t.Fatalf("provision account: %v", err)
	}
	claims := &identity.Claims{Sub: "sub-" + uuid.NewString(), Email: email, EmailVerified: true}
	if _, err := svc.CompleteOnboarding(ctx(), claims, OnboardingInput{Username: "jdoe", AcceptedIDs: []string{aup.ID}}); err != nil || home() != "active" {
		t.Fatalf("onboarding: err=%v home=%q, want active", err, home())
	}

	survivor := seedPendingUserWithEmail(t, database, "survivor-"+uuid.NewString()[:8]+"@example.edu")
	if _, err := svc.MergeUsers(ctx(), survivor, pending); err != nil || home() != "active" {
		t.Fatalf("merging into a PENDING survivor: err=%v home=%q, want active", err, home())
	}
	for _, status := range []models.UserStatus{models.UserInactive, models.UserActive} {
		if _, err := svc.UpdateUserStatus(ctx(), survivor, status); err != nil || (home() == "active") != (status == models.UserActive) {
			t.Fatalf("setting the user %s: err=%v home=%q, want it active only while ACTIVE", status, err, home())
		}
	}
}
