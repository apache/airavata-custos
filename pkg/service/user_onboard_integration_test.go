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
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
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
	onboarder := seedUser(t, database, fmt.Sprintf("onboarder-%s@example.invalid", uuid.NewString()))

	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	in := onboardInput(fmt.Sprintf("researcher-%s@example.invalid", uuid.NewString()))
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
	onboarder := seedUser(t, database, fmt.Sprintf("onboarder-%s@example.invalid", uuid.NewString()))

	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	in := onboardInput(fmt.Sprintf("gen-%s@example.invalid", uuid.NewString()))
	in.FirstName = "Jordan"
	in.LastName = "Kellett"
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
	if cu.LocalUsername != "custos-jkellett" {
		t.Fatalf("username = %q, want custos-jkellett", cu.LocalUsername)
	}

	// Same name again picks the next collision suffix.
	in.Email = fmt.Sprintf("gen2-%s@example.invalid", uuid.NewString())
	second, err := svc.OnboardUser(ctx(), in)
	if err != nil {
		t.Fatalf("onboard second: %v", err)
	}
	cu2, err := svc.GetComputeClusterUserByPair(ctx(), cluster.ID, second.ID)
	if err != nil {
		t.Fatalf("cluster user 2: %v", err)
	}
	if cu2.LocalUsername != "custos-jkellett2" {
		t.Fatalf("username = %q, want custos-jkellett2", cu2.LocalUsername)
	}
}

func TestOnboardUser_ResearcherWithAllocation(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	onboarder := seedUser(t, database, fmt.Sprintf("onboarder-%s@example.invalid", uuid.NewString()))

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

	in := onboardInput(fmt.Sprintf("researcher-%s@example.invalid", uuid.NewString()))
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
	onboarder := seedUser(t, database, fmt.Sprintf("superadmin-%s@example.invalid", uuid.NewString()))
	seedRolesManage(t, database, onboarder)

	in := onboardInput(fmt.Sprintf("padmin-%s@example.invalid", uuid.NewString()))
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
	onboarder := seedUser(t, database, fmt.Sprintf("superadmin-%s@example.invalid", uuid.NewString()))
	seedRolesManage(t, database, onboarder)

	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	in := onboardInput(fmt.Sprintf("cadmin-%s@example.invalid", uuid.NewString()))
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
}

func TestOnboardUser_AdminRequiresRolesManage(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	onboarder := seedUser(t, database, fmt.Sprintf("plain-%s@example.invalid", uuid.NewString()))

	in := onboardInput(fmt.Sprintf("padmin-%s@example.invalid", uuid.NewString()))
	in.PortalAdmin = true
	in.OnboardedBy = onboarder

	if _, err := svc.OnboardUser(ctx(), in); err == nil {
		t.Fatal("onboarding an admin without roles:manage should fail")
	}
}

func TestOnboardUser_TakenUsernameLeavesNoUser(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	onboarder := seedUser(t, database, fmt.Sprintf("onboarder-%s@example.invalid", uuid.NewString()))

	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	username := "onb-" + uuid.NewString()[:8]

	first := onboardInput(fmt.Sprintf("first-%s@example.invalid", uuid.NewString()))
	first.Username = username
	first.ComputeClusterID = cluster.ID
	first.OnboardedBy = onboarder
	if _, err := svc.OnboardUser(ctx(), first); err != nil {
		t.Fatalf("onboard first: %v", err)
	}

	second := onboardInput(fmt.Sprintf("second-%s@example.invalid", uuid.NewString()))
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
	onboarder := seedUser(t, database, fmt.Sprintf("onboarder-%s@example.invalid", uuid.NewString()))

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

	in := onboardInput(fmt.Sprintf("inactive-%s@example.invalid", uuid.NewString()))
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

// A failure after the user row must take the whole thing back, or the email
// is claimed by a user that was never finished.
func TestOnboardUser_RollsBackOnLaterFailure(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	onboarder := seedUser(t, database, fmt.Sprintf("onboarder-%s@example.invalid", uuid.NewString()))

	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{Name: "onb-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	in := onboardInput(fmt.Sprintf("rollback-%s@example.invalid", uuid.NewString()))
	// Free at the pre-check, too long for the column, so the insert fails
	// only once the user row is already in the transaction.
	in.Username = strings.Repeat("u", 300)
	in.ComputeClusterID = cluster.ID
	in.OnboardedBy = onboarder

	if _, err := svc.OnboardUser(ctx(), in); err == nil {
		t.Fatal("expected the cluster user insert to fail")
	}
	if u, err := svc.GetUserByEmail(ctx(), in.Email); err == nil && u != nil {
		t.Fatalf("user row survived a failed onboard: %s", u.ID)
	}

	// The email is free again, so the admin can simply retry.
	in.Username = "onb-" + uuid.NewString()[:8]
	if _, err := svc.OnboardUser(ctx(), in); err != nil {
		t.Fatalf("retry after rollback: %v", err)
	}
}
