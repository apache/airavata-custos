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
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/posix"
)

type accessTestEnv struct {
	db        *sqlx.DB
	svc       *Service
	orgID     string
	adminID   string
	clusterID string
	allocID   string
	code      string
}

// setupAccessEnv builds the org/admin/cluster/project/allocation chain an
// access event depends on and inserts the access_events row.
func setupAccessEnv(t *testing.T) *accessTestEnv {
	t.Helper()
	database := setupTestDB(t)
	// setupTestDB truncates only the core auth tables; clear the access and
	// allocation chains this suite touches.
	if _, err := database.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		t.Fatalf("disable FK: %v", err)
	}
	for _, tbl := range []string{
		"access_request_events", "access_requests", "access_events",
		"compute_allocation_memberships", "compute_cluster_users",
		"compute_allocations", "projects", "compute_clusters",
	} {
		if _, err := database.Exec("TRUNCATE TABLE " + tbl); err != nil {
			t.Fatalf("truncate %s: %v", tbl, err)
		}
	}
	if _, err := database.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		t.Fatalf("re-enable FK: %v", err)
	}

	svc := newTestService(database)
	org, err := svc.CreateOrganization(ctx(), &models.Organization{
		OriginatedID: uuid.NewString(),
		Name:         "access-org-" + uuid.NewString()[:8],
	})
	if err != nil {
		t.Fatalf("create org: %v", err)
	}
	admin, err := svc.CreateUser(ctx(), &models.User{
		OrganizationID: org.ID,
		FirstName:      "Admin",
		LastName:       "Approver",
		Email:          "admin-" + uuid.NewString()[:8] + "@example.invalid",
		Status:         models.UserActive,
	})
	if err != nil {
		t.Fatalf("create admin: %v", err)
	}
	cluster, err := svc.CreateComputeCluster(ctx(), &models.ComputeCluster{
		Name: "access-cluster-" + uuid.NewString()[:8],
	})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	proj, err := svc.CreateProject(ctx(), &models.Project{
		OriginatedID: uuid.NewString(),
		Title:        "Access Test Project",
		Origination:  "TEST",
		ProjectPIID:  admin.ID,
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	alloc, err := svc.CreateComputeAllocation(ctx(), &models.ComputeAllocation{
		ProjectID:        proj.ID,
		Name:             "access-alloc-" + uuid.NewString()[:8],
		ComputeClusterID: cluster.ID,
		InitialSUAmount:  1000,
		StartTime:        nowUTC(),
		EndTime:          nowUTC().Add(365 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("create allocation: %v", err)
	}
	code := "EVT-" + uuid.NewString()[:8]
	if _, err := database.Exec(
		"INSERT INTO access_events (code, compute_allocation_id, organization_id) VALUES (?, ?, ?)",
		code, alloc.ID, org.ID,
	); err != nil {
		t.Fatalf("seed access event: %v", err)
	}
	return &accessTestEnv{
		db: database, svc: svc,
		orgID: org.ID, adminID: admin.ID,
		clusterID: cluster.ID, allocID: alloc.ID, code: code,
	}
}

func newAccessRequestFor(env *accessTestEnv, sub string) *models.AccessRequest {
	return &models.AccessRequest{
		OIDCSub:     sub,
		Email:       sub + "@example.invalid",
		Name:        "Trial Person",
		Institution: "Test University",
		EventCode:   env.code,
		Reason:      "tutorial",
	}
}

func requestEventTypes(t *testing.T, env *accessTestEnv, requestID string) []string {
	t.Helper()
	var types []string
	if err := env.db.Select(&types,
		"SELECT event_type FROM access_request_events WHERE access_request_id = ? ORDER BY timestamp, event_type",
		requestID,
	); err != nil {
		t.Fatalf("list request events: %v", err)
	}
	return types
}

func TestCreateAccessRequest_Guards(t *testing.T) {
	env := setupAccessEnv(t)
	sub := "sub-" + uuid.NewString()[:8]

	bad := newAccessRequestFor(env, sub)
	bad.EventCode = "NO-SUCH-EVENT"
	if _, err := env.svc.CreateAccessRequest(ctx(), bad); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown event: got %v, want ErrNotFound", err)
	}

	created, err := env.svc.CreateAccessRequest(ctx(), newAccessRequestFor(env, sub))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.Status != models.AccessRequestPending {
		t.Errorf("status = %s, want PENDING", created.Status)
	}
	if got := requestEventTypes(t, env, created.ID); len(got) != 1 || got[0] != models.AccessRequestEventCreated {
		t.Errorf("event rows = %v, want [CREATED]", got)
	}

	if _, err := env.svc.CreateAccessRequest(ctx(), newAccessRequestFor(env, sub)); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("duplicate pending: got %v, want ErrAlreadyExists", err)
	}
}

func TestApproveAccessRequest_HappyPath(t *testing.T) {
	env := setupAccessEnv(t)
	sub := "sub-" + uuid.NewString()[:8]
	req, err := env.svc.CreateAccessRequest(ctx(), newAccessRequestFor(env, sub))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	exp := nowUTC().Add(48 * time.Hour).Truncate(time.Second)
	approved, err := env.svc.ApproveAccessRequest(ctx(), req.ID, env.adminID, &exp)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if approved.Status != models.AccessRequestApproved {
		t.Errorf("status = %s, want APPROVED", approved.Status)
	}
	if approved.ApproverID != env.adminID {
		t.Errorf("approver = %q, want %q", approved.ApproverID, env.adminID)
	}
	if approved.CreatedUserID == "" {
		t.Fatal("created_user_id is empty")
	}

	user, err := env.svc.GetUser(ctx(), approved.CreatedUserID)
	if err != nil {
		t.Fatalf("get created user: %v", err)
	}
	if user.Status != models.UserActive {
		t.Errorf("user status = %s, want ACTIVE", user.Status)
	}
	if user.OrganizationID != env.orgID {
		t.Errorf("user org = %q, want %q", user.OrganizationID, env.orgID)
	}

	ident, err := env.svc.GetUserIdentityBySourceAndExternalID(ctx(), identitySourceOIDC, sub)
	if err != nil {
		t.Fatalf("get identity: %v", err)
	}
	if ident.UserID != user.ID {
		t.Errorf("identity user = %q, want %q", ident.UserID, user.ID)
	}

	ccu, err := env.svc.GetComputeClusterUserByPair(ctx(), env.clusterID, user.ID)
	if err != nil {
		t.Fatalf("get cluster user: %v", err)
	}
	if !strings.HasPrefix(ccu.LocalUsername, posix.Prefix()+"-") {
		t.Errorf("local username %q lacks prefix %q", ccu.LocalUsername, posix.Prefix()+"-")
	}

	var m struct {
		Status  string    `db:"membership_status"`
		EndTime time.Time `db:"end_time"`
	}
	if err := env.db.Get(&m,
		"SELECT membership_status, end_time FROM compute_allocation_memberships WHERE compute_allocation_id = ? AND user_id = ?",
		env.allocID, user.ID,
	); err != nil {
		t.Fatalf("get membership: %v", err)
	}
	if m.Status != string(models.ACTIVE) {
		t.Errorf("membership status = %s, want ACTIVE", m.Status)
	}
	if d := m.EndTime.UTC().Sub(exp); d > 2*time.Second || d < -2*time.Second {
		t.Errorf("membership end_time = %v, want %v", m.EndTime.UTC(), exp)
	}

	if got := requestEventTypes(t, env, req.ID); len(got) != 2 || got[0] != models.AccessRequestEventCreated || got[1] != models.AccessRequestEventApproved {
		t.Errorf("event rows = %v, want [CREATED APPROVED]", got)
	}

	latest, err := env.svc.GetLatestAccessRequestBySub(ctx(), sub)
	if err != nil {
		t.Fatalf("latest by sub: %v", err)
	}
	if latest.Status != models.AccessRequestApproved {
		t.Errorf("latest status = %s, want APPROVED", latest.Status)
	}
}

func TestApproveAccessRequest_ReusesExistingUser(t *testing.T) {
	env := setupAccessEnv(t)
	sub := "sub-" + uuid.NewString()[:8]
	req, err := env.svc.CreateAccessRequest(ctx(), newAccessRequestFor(env, sub))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	pre, err := env.svc.CreateUser(ctx(), &models.User{
		OrganizationID: env.orgID,
		FirstName:      "Trial",
		LastName:       "Person",
		Email:          req.Email,
		Status:         models.UserActive,
	})
	if err != nil {
		t.Fatalf("pre-create user: %v", err)
	}

	approved, err := env.svc.ApproveAccessRequest(ctx(), req.ID, env.adminID, nil)
	if err != nil {
		t.Fatalf("approve with existing user: %v", err)
	}
	if approved.CreatedUserID != pre.ID {
		t.Errorf("created_user_id = %q, want reused %q", approved.CreatedUserID, pre.ID)
	}
	if approved.ExpiresAt == nil {
		t.Fatal("expires_at not defaulted")
	}
	want := nowUTC().Add(defaultAccessRequestLifetime)
	if d := approved.ExpiresAt.Sub(want); d > time.Hour || d < -time.Hour {
		t.Errorf("default expires_at = %v, want about %v", approved.ExpiresAt, want)
	}
	if _, err := env.svc.GetComputeClusterUserByPair(ctx(), env.clusterID, pre.ID); err != nil {
		t.Errorf("cluster user for reused user: %v", err)
	}
}

func TestApproveAccessRequest_FailureLeavesPending(t *testing.T) {
	env := setupAccessEnv(t)
	sub := "sub-" + uuid.NewString()[:8]
	req, err := env.svc.CreateAccessRequest(ctx(), newAccessRequestFor(env, sub))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := env.db.Exec("UPDATE compute_allocations SET status = 'INACTIVE' WHERE id = ?", env.allocID); err != nil {
		t.Fatalf("deactivate allocation: %v", err)
	}
	if _, err := env.svc.ApproveAccessRequest(ctx(), req.ID, env.adminID, nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("approve inactive allocation: got %v, want ErrInvalidInput", err)
	}

	latest, err := env.svc.GetLatestAccessRequestBySub(ctx(), sub)
	if err != nil {
		t.Fatalf("latest by sub: %v", err)
	}
	if latest.Status != models.AccessRequestPending {
		t.Errorf("status after failed approve = %s, want PENDING", latest.Status)
	}
	got := requestEventTypes(t, env, req.ID)
	if len(got) != 2 || got[1] != models.AccessRequestEventFailed {
		t.Errorf("event rows = %v, want [CREATED FAILED]", got)
	}

	if _, err := env.db.Exec("UPDATE compute_allocations SET status = 'ACTIVE' WHERE id = ?", env.allocID); err != nil {
		t.Fatalf("reactivate allocation: %v", err)
	}
	if _, err := env.svc.ApproveAccessRequest(ctx(), req.ID, env.adminID, nil); err != nil {
		t.Fatalf("re-approve after failure: %v", err)
	}
}

func TestDenyAccessRequest(t *testing.T) {
	env := setupAccessEnv(t)
	sub := "sub-" + uuid.NewString()[:8]
	req, err := env.svc.CreateAccessRequest(ctx(), newAccessRequestFor(env, sub))
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	denied, err := env.svc.DenyAccessRequest(ctx(), req.ID, env.adminID, "not eligible")
	if err != nil {
		t.Fatalf("deny: %v", err)
	}
	if denied.Status != models.AccessRequestDenied {
		t.Errorf("status = %s, want DENIED", denied.Status)
	}
	if denied.DenyReason != "not eligible" {
		t.Errorf("deny_reason = %q", denied.DenyReason)
	}
	if got := requestEventTypes(t, env, req.ID); len(got) != 2 || got[1] != models.AccessRequestEventDenied {
		t.Errorf("event rows = %v, want [CREATED DENIED]", got)
	}

	if _, err := env.svc.ApproveAccessRequest(ctx(), req.ID, env.adminID, nil); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("approve denied request: got %v, want ErrInvalidInput", err)
	}
	if _, err := env.svc.DenyAccessRequest(ctx(), req.ID, env.adminID, "again"); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("deny denied request: got %v, want ErrInvalidInput", err)
	}
}

func TestListAccessRequests_Filter(t *testing.T) {
	env := setupAccessEnv(t)
	subA := "sub-" + uuid.NewString()[:8]
	subB := "sub-" + uuid.NewString()[:8]
	reqA, err := env.svc.CreateAccessRequest(ctx(), newAccessRequestFor(env, subA))
	if err != nil {
		t.Fatalf("create A: %v", err)
	}
	if _, err := env.svc.CreateAccessRequest(ctx(), newAccessRequestFor(env, subB)); err != nil {
		t.Fatalf("create B: %v", err)
	}
	if _, err := env.svc.DenyAccessRequest(ctx(), reqA.ID, env.adminID, "no"); err != nil {
		t.Fatalf("deny A: %v", err)
	}

	pending, err := env.svc.ListAccessRequests(ctx(), store.AccessRequestListFilter{Status: string(models.AccessRequestPending), EventCode: env.code})
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) != 1 || pending[0].OIDCSub != subB {
		t.Errorf("pending list = %+v, want only sub %s", pending, subB)
	}
	all, err := env.svc.ListAccessRequests(ctx(), store.AccessRequestListFilter{EventCode: env.code})
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("all list len = %d, want 2", len(all))
	}
}
