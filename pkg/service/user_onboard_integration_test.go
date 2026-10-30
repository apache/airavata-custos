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
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/posix"
)

func seedRolesManage(t *testing.T, database *sqlx.DB, userID string) {
	t.Helper()
	if _, err := database.Exec(
		`INSERT INTO user_privileges (id, user_id, privilege, granted_at, reason)
		 VALUES ($1, $2, $3, NOW(), 'seed')`,
		uuid.NewString(), userID, string(models.RolesManage),
	); err != nil {
		t.Fatalf("seed roles:manage for %s: %v", userID, err)
	}
}

func onboardInput(email string) OnboardUserInput {
	return OnboardUserInput{
		Email:     email,
		FirstName: "On",
		LastName:  "Board",
	}
}

func TestOnboardUser_Researcher(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
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
		t.Fatalf("onboard researcher: %v", err)
	}
	if created.Type != models.UserTypeClusterLocal {
		t.Fatalf("type = %s, want CLUSTER_LOCAL", created.Type)
	}
	cu, err := svc.GetComputeClusterUserByPair(ctx(), cluster.ID, created.ID)
	if err != nil || cu == nil {
		t.Fatalf("cluster user missing: %v", err)
	}
	if cu.AccessLevel != models.ClusterAccessUser {
		t.Fatalf("access level = %s, want USER", cu.AccessLevel)
	}
	if cu.LocalUsername != in.Username {
		t.Fatalf("username = %s, want %s", cu.LocalUsername, in.Username)
	}
	roles, err := svc.ListUserRoles(ctx(), created.ID)
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}
	if len(roles) != 0 {
		t.Fatalf("researcher should have no roles, got %d", len(roles))
	}
}

func TestOnboardUser_GeneratesUsernameFromName(t *testing.T) {
	t.Setenv("POSIX_USERNAME_PREFIX", "custos")
	database := setupTestDB(t)
	svc := newTestService(database)
	onboarder := seedUser(t, database, fmt.Sprintf("onboarder-%s@example.edu", uuid.NewString()))

	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	in := onboardInput(fmt.Sprintf("gen-%s@example.edu", uuid.NewString()))
	in.FirstName = "Jane"
	in.LastName = "Doe"
	in.ComputeClusterID = cluster.ID
	in.OnboardedBy = onboarder

	first, err := svc.OnboardUser(ctx(), in)
	if err != nil {
		t.Fatalf("onboard: %v", err)
	}
	cu, err := svc.GetComputeClusterUserByPair(ctx(), cluster.ID, first.ID)
	if err != nil {
		t.Fatalf("cluster user: %v", err)
	}
	if cu.LocalUsername != "custos-jdoe" {
		t.Fatalf("username = %q, want custos-jdoe", cu.LocalUsername)
	}

	// Same name again picks the next collision suffix.
	in.Email = fmt.Sprintf("gen2-%s@example.edu", uuid.NewString())
	second, err := svc.OnboardUser(ctx(), in)
	if err != nil {
		t.Fatalf("onboard second: %v", err)
	}
	cu2, err := svc.GetComputeClusterUserByPair(ctx(), cluster.ID, second.ID)
	if err != nil {
		t.Fatalf("cluster user 2: %v", err)
	}
	if cu2.LocalUsername != "custos-jdoe2" {
		t.Fatalf("username = %q, want custos-jdoe2", cu2.LocalUsername)
	}
}

func TestOnboardUser_ResearcherWithAllocation(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	onboarder := seedUser(t, database, fmt.Sprintf("onboarder-%s@example.edu", uuid.NewString()))

	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	project, err := svc.CreateProject(ctx(), &models.Project{
		Title:       "onb-project-" + uuid.NewString()[:8],
		ProjectPIID: onboarder,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	alloc, err := svc.CreateComputeAllocation(ctx(), &models.ComputeAllocation{
		ProjectID:        project.ID,
		Name:             "onb-alloc-" + uuid.NewString()[:8],
		ComputeClusterID: cluster.ID,
		InitialSUAmount:  1000,
		StartTime:        time.Now().UTC(),
		EndTime:          time.Now().UTC().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create allocation: %v", err)
	}

	in := onboardInput(fmt.Sprintf("researcher-%s@example.edu", uuid.NewString()))
	in.Username = "onb-" + uuid.NewString()[:8]
	in.AllocationID = alloc.ID
	in.OnboardedBy = onboarder

	created, err := svc.OnboardUser(ctx(), in)
	if err != nil {
		t.Fatalf("onboard researcher with allocation: %v", err)
	}
	// Cluster is derived from the allocation.
	cu, err := svc.GetComputeClusterUserByPair(ctx(), cluster.ID, created.ID)
	if err != nil || cu == nil {
		t.Fatalf("cluster user missing: %v", err)
	}
	memberships, err := svc.ListAllocationsForUser(ctx(), created.ID)
	if err != nil {
		t.Fatalf("list memberships: %v", err)
	}
	if len(memberships) != 1 || memberships[0].ComputeAllocationID != alloc.ID {
		t.Fatalf("membership not created: %+v", memberships)
	}
	if memberships[0].MembershipStatus != models.ACTIVE {
		t.Fatalf("membership status = %s, want ACTIVE", memberships[0].MembershipStatus)
	}
}

func TestOnboardUser_PortalOnlyAdmin(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	onboarder := seedUser(t, database, fmt.Sprintf("superadmin-%s@example.edu", uuid.NewString()))
	seedRolesManage(t, database, onboarder)

	in := onboardInput(fmt.Sprintf("padmin-%s@example.edu", uuid.NewString()))
	in.PortalAdmin = true
	in.OnboardedBy = onboarder

	created, err := svc.OnboardUser(ctx(), in)
	if err != nil {
		t.Fatalf("onboard portal admin: %v", err)
	}
	if created.Type != models.UserTypeSystem {
		t.Fatalf("type = %s, want SYSTEM", created.Type)
	}
	roles, err := svc.ListUserRoles(ctx(), created.ID)
	if err != nil {
		t.Fatalf("list roles: %v", err)
	}
	if len(roles) != 1 {
		t.Fatalf("want 1 role, got %d", len(roles))
	}
	role, err := svc.GetRole(ctx(), roles[0].RoleID)
	if err != nil {
		t.Fatalf("get role: %v", err)
	}
	if role.Name != models.SystemRoleAdmin {
		t.Fatalf("role = %s, want admin", role.Name)
	}
	privs, err := svc.ListRolePrivileges(ctx(), role.ID)
	if err != nil {
		t.Fatalf("list role privileges: %v", err)
	}
	for _, p := range privs {
		if p == models.PrivilegesGrant || p == models.RolesManage {
			t.Fatalf("admin role must not carry %s", p)
		}
	}
	if len(privs) == 0 {
		t.Fatal("admin role has no privileges")
	}
}

func TestOnboardUser_PortalAndClusterAdmin(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	onboarder := seedUser(t, database, fmt.Sprintf("superadmin-%s@example.edu", uuid.NewString()))
	seedRolesManage(t, database, onboarder)

	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	in := onboardInput(fmt.Sprintf("cadmin-%s@example.edu", uuid.NewString()))
	in.Username = "onb-" + uuid.NewString()[:8]
	in.ComputeClusterID = cluster.ID
	in.PortalAdmin = true
	in.ClusterAdmin = true
	in.OnboardedBy = onboarder

	created, err := svc.OnboardUser(ctx(), in)
	if err != nil {
		t.Fatalf("onboard cluster admin: %v", err)
	}
	if created.Type != models.UserTypeClusterLocal {
		t.Fatalf("type = %s, want CLUSTER_LOCAL", created.Type)
	}
	cu, err := svc.GetComputeClusterUserByPair(ctx(), cluster.ID, created.ID)
	if err != nil || cu == nil {
		t.Fatalf("cluster user missing: %v", err)
	}
	if cu.AccessLevel != models.ClusterAccessAdmin {
		t.Fatalf("access level = %s, want ADMIN", cu.AccessLevel)
	}
	if got := countAuditEventsOfType(t, database, clusterUserAuditAdminGranted, cu.ID); got != 1 {
		t.Fatalf("%s audit rows = %d, want 1", clusterUserAuditAdminGranted, got)
	}
}

func TestOnboardUser_AdminRequiresRolesManage(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	onboarder := seedUser(t, database, fmt.Sprintf("plain-%s@example.edu", uuid.NewString()))

	in := onboardInput(fmt.Sprintf("padmin-%s@example.edu", uuid.NewString()))
	in.PortalAdmin = true
	in.OnboardedBy = onboarder

	if _, err := svc.OnboardUser(ctx(), in); err == nil {
		t.Fatal("onboarding an admin without roles:manage should fail")
	}
}

func TestOnboardUser_TakenUsernameLeavesNoUser(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	onboarder := seedUser(t, database, fmt.Sprintf("onboarder-%s@example.edu", uuid.NewString()))

	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	username := "onb-" + uuid.NewString()[:8]

	first := onboardInput(fmt.Sprintf("first-%s@example.edu", uuid.NewString()))
	first.Username = username
	first.ComputeClusterID = cluster.ID
	first.OnboardedBy = onboarder
	if _, err := svc.OnboardUser(ctx(), first); err != nil {
		t.Fatalf("onboard first: %v", err)
	}

	second := onboardInput(fmt.Sprintf("second-%s@example.edu", uuid.NewString()))
	second.Username = username
	second.ComputeClusterID = cluster.ID
	second.OnboardedBy = onboarder
	_, err = svc.OnboardUser(ctx(), second)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("want ErrAlreadyExists, got %v", err)
	}
	// The failed attempt must not leave a user row, so a retry with a new
	// username succeeds.
	if u, err := svc.GetUserByEmail(ctx(), second.Email); err == nil && u != nil {
		t.Fatalf("failed onboard left a user row: %s", u.ID)
	}
	second.Username = "onb-" + uuid.NewString()[:8]
	if _, err := svc.OnboardUser(ctx(), second); err != nil {
		t.Fatalf("retry with free username: %v", err)
	}
}

func TestOnboardUser_RefusesInactiveAllocation(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	onboarder := seedUser(t, database, fmt.Sprintf("onboarder-%s@example.edu", uuid.NewString()))

	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	project, err := svc.CreateProject(ctx(), &models.Project{
		Title:       "onb-project-" + uuid.NewString()[:8],
		ProjectPIID: onboarder,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	alloc, err := svc.CreateComputeAllocation(ctx(), &models.ComputeAllocation{
		ProjectID:        project.ID,
		Name:             "onb-alloc-" + uuid.NewString()[:8],
		ComputeClusterID: cluster.ID,
		Status:           models.INACTIVE,
		InitialSUAmount:  1000,
		StartTime:        time.Now().UTC().Add(-48 * time.Hour),
		EndTime:          time.Now().UTC().Add(-24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create allocation: %v", err)
	}

	in := onboardInput(fmt.Sprintf("inactive-%s@example.edu", uuid.NewString()))
	in.Username = "onb-" + uuid.NewString()[:8]
	in.AllocationID = alloc.ID
	in.OnboardedBy = onboarder

	if _, err := svc.OnboardUser(ctx(), in); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("want ErrInvalidInput for an inactive allocation, got %v", err)
	}
	if u, err := svc.GetUserByEmail(ctx(), in.Email); err == nil && u != nil {
		t.Fatalf("refused onboard left a user row: %s", u.ID)
	}
}

// A given username over the POSIX login cap (32 Chars) is refused before any row is
// written, so it cannot reach the cluster and fail there.
func TestOnboardUser_RejectsLongUsername(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	onboarder := seedUser(t, database, fmt.Sprintf("onboarder-%s@example.edu", uuid.NewString()))

	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	in := onboardInput(fmt.Sprintf("long-%s@example.edu", uuid.NewString()))
	in.Username = strings.Repeat("u", posix.MaxLoginLen+1)
	in.ComputeClusterID = cluster.ID
	in.OnboardedBy = onboarder

	if _, err := svc.OnboardUser(ctx(), in); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
	if u, err := svc.GetUserByEmail(ctx(), in.Email); err == nil && u != nil {
		t.Fatalf("user row written for a rejected username: %s", u.ID)
	}
}

func newClusterAccount(t *testing.T, svc *Service, userID, name string) *models.ComputeClusterUser {
	t.Helper()
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	cu, err := svc.CreateComputeClusterUser(ctx(), &models.ComputeClusterUser{ComputeClusterID: cluster.ID, UserID: userID, LocalUsername: name})
	if err != nil {
		t.Fatalf("create cluster account: %v", err)
	}
	return cu
}

// Accepting the current agreement at first sign-in binds the sub, activates the
// user, renames the cluster account, and records the acceptance together.
func TestCompleteOnboarding_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	bus := newTestBus(database)
	svc := New(database, bus)
	bus.SubscribeUserOnboarded("test-subscriber", func(context.Context, models.User) error { return nil })
	aup := seedPublishedAgreement(t, svc, seedAgreementManager(t, database), "1.0")
	email := "jdoe-" + uuid.NewString()[:8] + "@example.edu"
	user := seedPendingUserWithEmail(t, database, email)
	newClusterAccount(t, svc, user, "custos-jdoe")
	claims := &identity.Claims{Sub: "sub-" + uuid.NewString(), Email: email, EmailVerified: true}

	if _, err := svc.CompleteOnboarding(ctx(), claims, OnboardingInput{Username: "jdoe", AcceptedIDs: []string{aup.ID}}); err != nil {
		t.Fatalf("CompleteOnboarding: %v", err)
	}

	if u, err := svc.GetUser(ctx(), user); err != nil || u.Status != models.UserActive {
		t.Errorf("status: got %+v, %v, want ACTIVE", u, err)
	}
	if n := countRows(t, database, "SELECT COUNT(*) FROM user_identities WHERE user_id = $1 AND source = 'oidc' AND oidc_sub = $2", user, claims.Sub); n != 1 {
		t.Errorf("oidc identity with sub: got %d, want 1", n)
	}
	if accounts, err := svc.ListComputeClusterUsersByUser(ctx(), user); err != nil || len(accounts) != 1 || accounts[0].LocalUsername != "jdoe" {
		t.Errorf("cluster account: got %+v, %v, want one named jdoe", accounts, err)
	}
	if n := countRows(t, database, "SELECT COUNT(*) FROM user_agreement_acceptances WHERE user_id = $1 AND agreement_id = $2", user, aup.ID); n != 1 {
		t.Errorf("acceptance: got %d, want 1", n)
	}
	for _, ev := range []string{"IDENTITY_LINKED", "USER_AGREEMENT_ACCEPTED"} {
		if got := countAuditEventsOfType(t, database, ev, user); got != 1 {
			t.Errorf("audit %s: got %d, want 1", ev, got)
		}
	}
	if n := countRows(t, database, "SELECT COUNT(*) FROM events WHERE event_type = 'user::onboarded'"); n != 1 {
		t.Errorf("user::onboarded events: got %d, want 1", n)
	}
}

// First sign-in without accepting the current agreement is refused and writes nothing. No sub, no status change, no rename, no acceptance.
func TestCompleteOnboarding_MissingAcceptance(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	seedPublishedAgreement(t, svc, seedAgreementManager(t, database), "1.0")
	email := "jdoe-" + uuid.NewString()[:8] + "@example.edu"
	user := seedPendingUserWithEmail(t, database, email)
	newClusterAccount(t, svc, user, "custos-jdoe")
	claims := &identity.Claims{Sub: "sub-" + uuid.NewString(), Email: email, EmailVerified: true}

	_, err := svc.CompleteOnboarding(ctx(), claims, OnboardingInput{Username: "jdoe"})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}

	if u, err := svc.GetUser(ctx(), user); err != nil || u.Status != models.UserPending {
		t.Errorf("status: got %+v, %v, want PENDING", u, err)
	}
	if n := countRows(t, database, "SELECT COUNT(*) FROM user_identities WHERE user_id = $1", user); n != 0 {
		t.Errorf("identities: got %d, want 0", n)
	}
	if n := countRows(t, database, "SELECT COUNT(*) FROM user_agreement_acceptances WHERE user_id = $1", user); n != 0 {
		t.Errorf("acceptances: got %d, want 0", n)
	}
	if accounts, err := svc.ListComputeClusterUsersByUser(ctx(), user); err != nil || len(accounts) != 1 || accounts[0].LocalUsername != "custos-jdoe" {
		t.Errorf("cluster account: got %+v, %v, want one named custos-jdoe", accounts, err)
	}
}

// Make sure a user the registry already knows finishes onboarding with their username unchanged.
func TestCompleteOnboarding_RegisteredUserKeepsUsername(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	aup := seedPublishedAgreement(t, svc, seedAgreementManager(t, database), "1.0")
	email := "jdoe-" + uuid.NewString()[:8] + "@example.edu"
	user := seedPendingUserWithEmail(t, database, email)
	newClusterAccount(t, svc, user, "custos-jdoe")
	if _, err := svc.CreateUserIdentity(ctx(), &models.UserIdentity{UserID: user, Source: "comanage", ExternalID: "100001"}); err != nil {
		t.Fatalf("create comanage identity: %v", err)
	}
	claims := &identity.Claims{Sub: "sub-" + uuid.NewString(), Email: email, EmailVerified: true}

	if _, err := svc.CompleteOnboarding(ctx(), claims, OnboardingInput{Username: "jdoe", AcceptedIDs: []string{aup.ID}}); err != nil {
		t.Fatalf("CompleteOnboarding: %v", err)
	}
	if accounts, err := svc.ListComputeClusterUsersByUser(ctx(), user); err != nil || len(accounts) != 1 || accounts[0].LocalUsername != "custos-jdoe" {
		t.Errorf("cluster account: got %+v, %v, want one named custos-jdoe", accounts, err)
	}
}

// Make sure the setup steps follow what provisioning has recorded.
func TestSetupSteps_FollowProvisioning(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	user := seedPendingUserWithEmail(t, database, "jdoe-"+uuid.NewString()[:8]+"@example.edu")
	cu := newClusterAccount(t, svc, user, "jdoe")
	cluster, err := svc.GetComputeCluster(ctx(), cu.ComputeClusterID)
	if err != nil {
		t.Fatalf("get cluster: %v", err)
	}

	want := []SetupStep{{Key: "registry"}, {Key: "cluster_account", Name: cluster.Name}}
	if got, err := svc.SetupSteps(ctx(), user); err != nil || !slices.Equal(got, want) {
		t.Fatalf("before provisioning: got %+v, %v, want %+v", got, err, want)
	}

	if _, err := svc.CreateUserIdentity(ctx(), &models.UserIdentity{UserID: user, Source: "comanage", ExternalID: "100001"}); err != nil {
		t.Fatalf("create comanage identity: %v", err)
	}
	if err := svc.MarkComputeClusterUserProvisioned(ctx(), cu.ID); err != nil {
		t.Fatalf("mark provisioned: %v", err)
	}
	got, err := svc.SetupSteps(ctx(), user)
	if err != nil || len(got) != 2 {
		t.Fatalf("after provisioning: got %+v, %v, want registry and cluster_account", got, err)
	}
	for _, step := range got {
		if step.DoneAt == nil {
			t.Errorf("step %s: got %+v, want a done time", step.Key, step)
		}
	}
}
