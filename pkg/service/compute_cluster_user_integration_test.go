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
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

func TestMarkComputeClusterUserProvisioned_RoundTrip(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)

	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{
		Name: "provisioned-" + uuid.NewString()[:8],
	})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	org, err := svc.CreateOrganization(ctx(), &models.Organization{
		OriginatedID: uuid.NewString(),
		Name:         "provisioned-test-org",
	})
	if err != nil {
		t.Fatalf("create org: %v", err)
	}
	user, err := svc.CreateUser(ctx(), &models.User{
		OrganizationID: org.ID,
		FirstName:      "Prov",
		LastName:       "Target",
		Email:          fmt.Sprintf("prov-%s@example.invalid", uuid.NewString()),
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	ccu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{
		ComputeClusterID: cluster.ID,
		UserID:           user.ID,
		LocalUsername:    "prov-" + uuid.NewString()[:8],
	})
	if err != nil {
		t.Fatalf("create compute cluster user: %v", err)
	}
	if ccu.ProvisionedAt != nil {
		t.Fatalf("new mapping should not be provisioned yet")
	}

	if err := svc.MarkComputeClusterUserProvisioned(ctx(), ccu.ID); err != nil {
		t.Fatalf("mark provisioned: %v", err)
	}

	got, err := svc.GetComputeClusterUser(ctx(), ccu.ID)
	if err != nil {
		t.Fatalf("get compute cluster user: %v", err)
	}
	if got.ProvisionedAt == nil {
		t.Fatalf("provisioned_at should be set after MarkComputeClusterUserProvisioned")
	}

	if err := svc.MarkComputeClusterUserProvisioned(ctx(), "missing-"+uuid.NewString()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for unknown id, got %v", err)
	}
}

// Make sure `provisioned_at` is stamped and the provision event published only the first time.
func TestMarkComputeClusterUserProvisioned_OnlyTheFirstTimeCounts(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	svc.EventBus().SubscribeComputeClusterUserProvisioned("test-subscriber", func(context.Context, models.ComputeClusterUser) error { return nil })
	cu := newClusterUser(t, svc, database)

	if err := svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID); err != nil {
		t.Fatalf("mark provisioned: %v", err)
	}
	first, err := svc.GetComputeClusterUser(ctx(), cu.ID)
	if err != nil {
		t.Fatalf("get compute cluster user: %v", err)
	}
	if err := svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID); err != nil {
		t.Fatalf("mark provisioned again: %v", err)
	}
	again, err := svc.GetComputeClusterUser(ctx(), cu.ID)
	if err != nil {
		t.Fatalf("get compute cluster user: %v", err)
	}

	if !again.ProvisionedAt.Equal(*first.ProvisionedAt) {
		t.Fatalf("provisioned_at moved from %v to %v", first.ProvisionedAt, again.ProvisionedAt)
	}
	if n := countDeliveries(t, svc); n != 1 {
		t.Fatalf("expected one delivery of the provision event, got %d", n)
	}
}

// Make sure the group provision event fires once and the stamp keeps its first time.
func TestMarkComputeAllocationGroupProvisioned_OnlyTheFirstTimeCounts(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "grp-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	pi := seedUser(t, database, fmt.Sprintf("pi-%s@example.edu", uuid.NewString()))
	project, err := svc.CreateProject(ctx(), &models.Project{Title: "grp", Origination: "TEST", ProjectPIID: pi, Status: models.ProjectActive})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	alloc, err := svc.CreateComputeAllocation(ctx(), &models.ComputeAllocation{
		Name: "Grant", ProjectID: project.ID, ComputeClusterID: cluster.ID, StartTime: time.Now(), EndTime: time.Now().AddDate(1, 0, 0),
	})
	if err != nil {
		t.Fatalf("create allocation: %v", err)
	}
	svc.EventBus().SubscribeComputeAllocationGroupProvisioned("test-subscriber", func(context.Context, models.ComputeAllocation) error { return nil })

	if err := svc.MarkComputeAllocationGroupProvisioned(ctx(), alloc.ID); err != nil {
		t.Fatalf("mark group provisioned: %v", err)
	}
	first, err := svc.GetComputeAllocation(ctx(), alloc.ID)
	if err != nil || first.GroupProvisionedAt == nil {
		t.Fatalf("group_provisioned_at not set: %v", err)
	}
	if err := svc.MarkComputeAllocationGroupProvisioned(ctx(), alloc.ID); err != nil {
		t.Fatalf("mark group provisioned again: %v", err)
	}
	again, err := svc.GetComputeAllocation(ctx(), alloc.ID)
	if err != nil {
		t.Fatalf("get compute allocation: %v", err)
	}

	if !again.GroupProvisionedAt.Equal(*first.GroupProvisionedAt) {
		t.Fatalf("group_provisioned_at moved from %v to %v", first.GroupProvisionedAt, again.GroupProvisionedAt)
	}
	if n := countDeliveries(t, svc); n != 1 {
		t.Fatalf("expected one delivery of the group provision event, got %d", n)
	}
}

// newClusterUser creates a cluster, a user, and a cluster user with no approval
// given, the way an allocation source does, and returns the cluster user.
func newClusterUser(t *testing.T, svc *Service, database *sqlx.DB) *models.ComputeClusterUser {
	t.Helper()
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "appr-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	userID := seedUser(t, database, fmt.Sprintf("member-%s@example.edu", uuid.NewString()))
	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{
		ComputeClusterID: cluster.ID,
		UserID:           userID,
		LocalUsername:    "appr-" + uuid.NewString()[:8],
	})
	if err != nil {
		t.Fatalf("create compute cluster user: %v", err)
	}
	return cu
}

func countDeliveries(t *testing.T, svc *Service) int {
	t.Helper()
	rows, err := svc.EventBus().ListDeliveries(ctx(), "", 10)
	if err != nil {
		t.Fatalf("list deliveries: %v", err)
	}
	return len(rows)
}

// Make sure a cluster user created without an approval waits for an admin
// instead of being created on the cluster.
func TestCreateComputeClusterUser_WaitsForApproval(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)

	cu := newClusterUser(t, svc, database)
	if cu.ApprovalStatus != models.ClusterAccountPending || cu.ReviewedBy != nil {
		t.Fatalf("expected a pending row with no reviewer, got %+v", cu)
	}
}

// Make sure a cluster user an admin adds directly is approved and provisioning
// starts at once, without a separate approval step.
func TestCreateComputeClusterUser_ApprovedByAdminStartsProvisioning(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	svc.EventBus().SubscribeComputeClusterUserApproved("test-subscriber", func(context.Context, models.ComputeClusterUser) error { return nil })
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "appr-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))
	member := seedUser(t, database, fmt.Sprintf("member-%s@example.edu", uuid.NewString()))

	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{
		ComputeClusterID: cluster.ID,
		UserID:           member,
		LocalUsername:    "appr-" + uuid.NewString()[:8],
		ApprovalStatus:   models.ClusterAccountApproved,
		ReviewedBy:       &admin,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if cu.ApprovalStatus != models.ClusterAccountApproved || cu.ReviewedAt == nil {
		t.Fatalf("expected an approved row with a review time, got %+v", cu)
	}
	if n := countDeliveries(t, svc); n != 1 {
		t.Fatalf("expected one delivery of the approve event, got %d", n)
	}
}

// Make sure an approval stores who approved, leaves an audit row, and publishes
// the approve event so the account gets created on the cluster.
func TestApproveComputeClusterUser_PublishesApproveEvent(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	svc.EventBus().SubscribeComputeClusterUserApproved("test-subscriber", func(context.Context, models.ComputeClusterUser) error { return nil })
	cu := newClusterUser(t, svc, database)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))

	got, err := svc.ApproveComputeClusterUser(ctx(), cu.ID, admin)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if got.ApprovalStatus != models.ClusterAccountApproved || got.ReviewedBy == nil || *got.ReviewedBy != admin || got.ReviewedAt == nil {
		t.Fatalf("expected an approved row reviewed by %s, got %+v", admin, got)
	}
	if n := countDeliveries(t, svc); n != 1 {
		t.Fatalf("expected one delivery of the approve event, got %d", n)
	}
	if n := countAuditEventsOfType(t, database, clusterUserAuditApproved, cu.ID); n != 1 {
		t.Fatalf("expected one approval audit row, got %d", n)
	}
}

// Make sure a denial stores who denied and the note, leaves an audit row, and
// publishes only the deny event, so the account is never created on the cluster.
func TestDenyComputeClusterUser_PublishesDenyEvent(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	noop := func(context.Context, models.ComputeClusterUser) error { return nil }
	svc.EventBus().SubscribeComputeClusterUserApproved("test-subscriber", noop)
	svc.EventBus().SubscribeComputeClusterUserDenied("test-subscriber", noop)
	cu := newClusterUser(t, svc, database)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))

	got, err := svc.DenyComputeClusterUser(ctx(), cu.ID, admin, "not on the collaborator list")
	if err != nil {
		t.Fatalf("deny: %v", err)
	}
	if got.ApprovalStatus != models.ClusterAccountDenied || got.ReviewNote == nil || *got.ReviewNote != "not on the collaborator list" {
		t.Fatalf("expected a denied row with the note, got %+v", got)
	}
	if n := countDeliveries(t, svc); n != 1 {
		t.Fatalf("expected one delivery of the deny event, got %d", n)
	}
	if n := countAuditEventsOfType(t, database, clusterUserAuditDenied, cu.ID); n != 1 {
		t.Fatalf("expected one denial audit row, got %d", n)
	}
}

// Make sure the cluster account denial reason is mandatory
func TestDenyComputeClusterUser_RequiresReason(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cu := newClusterUser(t, svc, database)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))

	if _, err := svc.DenyComputeClusterUser(ctx(), cu.ID, admin, ""); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for a denial without a reason, got %v", err)
	}
	got, err := svc.GetComputeClusterUser(ctx(), cu.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ApprovalStatus != models.ClusterAccountPending {
		t.Fatalf("expected the account to stay pending, got %s", got.ApprovalStatus)
	}
}

// Make sure an approved account cannot be approved again or denied, since it
// may already exist on the cluster, while a denied one can still be approved.
func TestReviewComputeClusterUser_ApprovalIsFinal(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cu := newClusterUser(t, svc, database)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))

	if _, err := svc.DenyComputeClusterUser(ctx(), cu.ID, admin, "not on the collaborator list"); err != nil {
		t.Fatalf("deny: %v", err)
	}
	if _, err := svc.ApproveComputeClusterUser(ctx(), cu.ID, admin); err != nil {
		t.Fatalf("approve after deny: %v", err)
	}
	if _, err := svc.ApproveComputeClusterUser(ctx(), cu.ID, admin); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists on a second approval, got %v", err)
	}
	if _, err := svc.DenyComputeClusterUser(ctx(), cu.ID, admin, "not on the collaborator list"); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists denying an approved account, got %v", err)
	}
}

// Make sure a cluster user created through onboarding is approved by the
// onboarding admin and provisioning starts at once, since adding the user by hand is the approval.
func TestOnboardUser_ClusterAccountIsApproved(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	svc.EventBus().SubscribeComputeClusterUserApproved("test-subscriber", func(context.Context, models.ComputeClusterUser) error { return nil })
	onboarder := seedUser(t, database, fmt.Sprintf("onboarder-%s@example.edu", uuid.NewString()))
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	in := onboardInput(fmt.Sprintf("researcher-%s@example.edu", uuid.NewString()))
	in.Username = "onb-" + uuid.NewString()[:8]
	in.ComputeClusterID = cluster.ID
	in.OnboardedBy = onboarder

	created, err := svc.OnboardUser(ctx(), in)
	if err != nil {
		t.Fatalf("onboard: %v", err)
	}
	cu, err := svc.GetComputeClusterUserByPair(ctx(), cluster.ID, created.ID)
	if err != nil {
		t.Fatalf("cluster user: %v", err)
	}
	if cu.ApprovalStatus != models.ClusterAccountApproved || cu.ReviewedBy == nil || *cu.ReviewedBy != onboarder {
		t.Fatalf("expected approval by the onboarding admin, got %+v", cu)
	}
	if n := countDeliveries(t, svc); n != 1 {
		t.Fatalf("expected one delivery of the approve event, got %d", n)
	}
}

func TestListMembersForAllocation_CarriesTheClusterAccount(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "grp-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	pi := seedUser(t, database, fmt.Sprintf("pi-%s@example.edu", uuid.NewString()))
	project, err := svc.CreateProject(ctx(), &models.Project{Title: "grp", Origination: "TEST", ProjectPIID: pi, Status: models.ProjectActive, OriginatedID: uuid.NewString()})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	alloc, err := svc.CreateComputeAllocation(ctx(), &models.ComputeAllocation{
		Name: "Grant", ProjectID: project.ID, ComputeClusterID: cluster.ID, StartTime: time.Now(), EndTime: time.Now().AddDate(1, 0, 0),
	})
	if err != nil {
		t.Fatalf("create allocation: %v", err)
	}

	// account: "" none, "p" provisioned, "u" unprovisioned
	member := func(account string) string {
		userID := seedUser(t, database, fmt.Sprintf("m-%s@example.edu", uuid.NewString()))
		if account != "" {
			cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: cluster.ID, UserID: userID, LocalUsername: "m-" + uuid.NewString()[:8]})
			if err == nil && account == "p" {
				err = svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID)
			}
			if err != nil {
				t.Fatalf("seed account: %v", err)
			}
		}
		if _, err := svc.CreateComputeAllocationMembership(ctx(), &models.ComputeAllocationMembership{ComputeAllocationID: alloc.ID, UserID: userID, MembershipStatus: models.ACTIVE}); err != nil {
			t.Fatalf("seed membership: %v", err)
		}
		return userID
	}
	want := map[string]string{member("p"): "p", member("u"): "u", member(""): ""}

	rows, err := svc.ListMembersForAllocation(ctx(), alloc.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		got := ""
		if r.LocalUsername != "" {
			got = "u"
		}
		if r.ProvisionedAt != nil {
			got = "p"
		}
		if acct, ok := want[r.UserID]; ok && got != acct {
			t.Errorf("member %s: account %q, want %q", r.UserID, got, acct)
		}
	}
}

func TestCreateComputeAllocation_SuffixesACollidingPosixGroup(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "grp-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	pi := seedUser(t, database, fmt.Sprintf("pi-%s@example.edu", uuid.NewString()))
	// Refs that agree up to the 32-character limit collide once cut.
	grant := "CIS" + uuid.NewString()[:8] + strings.Repeat("X", 16)
	var groups []string
	for _, originated := range []string{grant + "a", grant + "b", grant + "c", ""} {
		project, err := svc.CreateProject(ctx(), &models.Project{Title: "grp", Origination: "TEST", ProjectPIID: pi, Status: models.ProjectActive, OriginatedID: originated})
		if err != nil {
			t.Fatalf("create project: %v", err)
		}
		alloc, err := svc.CreateComputeAllocation(ctx(), &models.ComputeAllocation{
			Name: "Grant", ProjectID: project.ID, ComputeClusterID: cluster.ID, StartTime: time.Now(), EndTime: time.Now().AddDate(1, 0, 0),
		})
		if err != nil {
			t.Fatalf("create allocation for %q: %v", originated, err)
		}
		groups = append(groups, *alloc.PosixGroup)
		if originated == "" && groups[3] != "test-"+project.ID[:8] {
			t.Fatalf("group %s, want one named after project id %s", groups[3], project.ID)
		}
	}
	if want := "test-" + strings.ToLower(grant); !slices.Equal(groups[:3], []string{want, want[:30] + "-2", want[:30] + "-3"}) {
		t.Fatalf("groups %v, want %s, -2, -3", groups, want)
	}
}

func TestDenyComputeClusterUser_ConcurrentDenialsRecordOnce(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	cu := newClusterUser(t, svc, database)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))

	hold, err := database.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := hold.Exec("SELECT id FROM compute_cluster_users WHERE id = $1 FOR UPDATE", cu.ID); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 2 {
		wg.Go(func() { svc.DenyComputeClusterUser(ctx(), cu.ID, admin, "no") })
	}
	for n, deadline := 0, time.Now().Add(5*time.Second); n < 2; time.Sleep(10 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatal("reviewers never blocked on the row lock")
		}
		if err := database.Get(&n, "SELECT count(*) FROM pg_stat_activity WHERE wait_event_type = 'Lock'"); err != nil {
			t.Fatal(err)
		}
	}
	if err := hold.Commit(); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if n := countAuditEventsOfType(t, database, clusterUserAuditDenied, cu.ID); n != 1 {
		t.Fatalf("denial audits: got %d, want 1", n)
	}
}
