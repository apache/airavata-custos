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

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// scopedFixtures is two projects with one allocation each; memberID holds an
// active membership on allocA only, piID is PI of projectA only.
type scopedFixtures struct {
	memberID string
	piID     string
	projectA *models.Project
	projectB *models.Project
	allocA   *models.ComputeAllocation
	allocB   *models.ComputeAllocation
}

func seedScopedFixtures(t *testing.T, database *sqlx.DB, svc *service.Service) scopedFixtures {
	t.Helper()
	suffix := time.Now().UnixNano()
	memberID := seedUser(t, database, fmt.Sprintf("scope.member+%d@example.edu", suffix))
	piID := seedUser(t, database, fmt.Sprintf("scope.pi+%d@example.edu", suffix))

	cluster, err := svc.CreateComputeCluster(t.Context(), &models.ComputeCluster{
		Name: fmt.Sprintf("scope-cluster-%d", suffix),
	})
	if err != nil {
		t.Fatalf("seed cluster: %v", err)
	}

	projectA, err := svc.CreateProject(t.Context(), &models.Project{
		Title: fmt.Sprintf("Scope Project A %d", suffix), ProjectPIID: piID,
	})
	if err != nil {
		t.Fatalf("seed project A: %v", err)
	}
	projectB, err := svc.CreateProject(t.Context(), &models.Project{
		Title: fmt.Sprintf("Scope Project B %d", suffix), ProjectPIID: piID,
	})
	if err != nil {
		t.Fatalf("seed project B: %v", err)
	}
	if err := svc.EnsureProjectMembership(t.Context(), projectA.ID, piID, "PI"); err != nil {
		t.Fatalf("seed PI membership: %v", err)
	}

	allocA, err := svc.CreateComputeAllocation(t.Context(), &models.ComputeAllocation{
		Name: fmt.Sprintf("scope-alloc-a-%d", suffix), ProjectID: projectA.ID, ComputeClusterID: cluster.ID,
	})
	if err != nil {
		t.Fatalf("seed allocation A: %v", err)
	}
	allocB, err := svc.CreateComputeAllocation(t.Context(), &models.ComputeAllocation{
		Name: fmt.Sprintf("scope-alloc-b-%d", suffix), ProjectID: projectB.ID, ComputeClusterID: cluster.ID,
	})
	if err != nil {
		t.Fatalf("seed allocation B: %v", err)
	}

	if _, err := svc.CreateComputeAllocationMembership(t.Context(), &models.ComputeAllocationMembership{
		ComputeAllocationID: allocA.ID, UserID: memberID,
	}); err != nil {
		t.Fatalf("seed membership: %v", err)
	}

	return scopedFixtures{
		memberID: memberID, piID: piID,
		projectA: projectA, projectB: projectB,
		allocA: allocA, allocB: allocB,
	}
}

func doGet(t *testing.T, srv *Server, path, userID string, privs ...models.PrivilegeKey) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req = withTestCaller(req, userID, privs...)
	srv.ServeHTTP(rr, req)
	return rr
}

func decodeAllocationList(t *testing.T, rr *httptest.ResponseRecorder) ComputeAllocationListResponse {
	t.Helper()
	var body ComputeAllocationListResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return body
}

func TestListComputeAllocations_MemberSeesOnlyOwn(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)

	rr := doGet(t, srv, "/compute-allocations", fx.memberID)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
	body := decodeAllocationList(t, rr)
	if len(body.Items) != 1 || body.Total != 1 {
		t.Fatalf("expected exactly the member's allocation, got total=%d items=%d", body.Total, len(body.Items))
	}
	if body.Items[0].ID != fx.allocA.ID {
		t.Errorf("allocation id: got %s, want %s", body.Items[0].ID, fx.allocA.ID)
	}
}

func TestListComputeAllocations_GovernanceRoleSeesProjectAllocations(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)

	rr := doGet(t, srv, "/compute-allocations", fx.piID)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
	body := decodeAllocationList(t, rr)
	if len(body.Items) != 1 || body.Items[0].ID != fx.allocA.ID {
		t.Fatalf("expected only project A's allocation for the PI, got %d items", len(body.Items))
	}
}

func TestListComputeAllocations_PrivilegedSeesFullListing(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)

	rr := doGet(t, srv, "/compute-allocations", fx.memberID, models.AllocationsRead)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
	body := decodeAllocationList(t, rr)
	found := map[string]bool{}
	for _, a := range body.Items {
		found[a.ID] = true
	}
	if !found[fx.allocA.ID] || !found[fx.allocB.ID] {
		t.Errorf("privileged listing missing allocations: A=%v B=%v", found[fx.allocA.ID], found[fx.allocB.ID])
	}
}

func TestGetComputeAllocation_MembershipScoped(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)

	if rr := doGet(t, srv, "/compute-allocations/"+fx.allocA.ID, fx.memberID); rr.Code != http.StatusOK {
		t.Errorf("own allocation: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
	if rr := doGet(t, srv, "/compute-allocations/"+fx.allocB.ID, fx.memberID); rr.Code != http.StatusNotFound {
		t.Errorf("foreign allocation: got %d, want 404 (%s)", rr.Code, rr.Body.String())
	}
	if rr := doGet(t, srv, "/compute-allocations/"+fx.allocB.ID, fx.memberID, models.AllocationsRead); rr.Code != http.StatusOK {
		t.Errorf("privileged foreign allocation: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
}

func TestAllocationSubresources_MembershipScoped(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)

	for _, sub := range []string{"/memberships", "/resources", "/usages", "/usages/total"} {
		if rr := doGet(t, srv, "/compute-allocations/"+fx.allocA.ID+sub, fx.memberID); rr.Code != http.StatusOK {
			t.Errorf("own allocation %s: got %d, want 200 (%s)", sub, rr.Code, rr.Body.String())
		}
		if rr := doGet(t, srv, "/compute-allocations/"+fx.allocB.ID+sub, fx.memberID); rr.Code != http.StatusNotFound {
			t.Errorf("foreign allocation %s: got %d, want 404 (%s)", sub, rr.Code, rr.Body.String())
		}
	}
}

func TestListProjects_ParticipantViaAllocationMembership(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)

	rr := doGet(t, srv, "/projects", fx.memberID)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
	var body ProjectListResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 1 || body.Total != 1 {
		t.Fatalf("expected exactly the parent project, got total=%d items=%d", body.Total, len(body.Items))
	}
	if body.Items[0].ID != fx.projectA.ID {
		t.Errorf("project id: got %s, want %s", body.Items[0].ID, fx.projectA.ID)
	}
}

func TestGetProject_MembershipScoped(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)

	if rr := doGet(t, srv, "/projects/"+fx.projectA.ID, fx.memberID); rr.Code != http.StatusOK {
		t.Errorf("participant project: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
	if rr := doGet(t, srv, "/projects/"+fx.projectB.ID, fx.memberID); rr.Code != http.StatusNotFound {
		t.Errorf("foreign project: got %d, want 404 (%s)", rr.Code, rr.Body.String())
	}
	if rr := doGet(t, srv, "/projects/"+fx.projectB.ID, fx.memberID, models.ProjectsRead); rr.Code != http.StatusOK {
		t.Errorf("privileged foreign project: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
}

func TestListProjects_PrivilegedSeesFullListing(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)

	rr := doGet(t, srv, "/projects?limit=200", fx.memberID, models.ProjectsRead)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
	var body ProjectListResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	found := map[string]bool{}
	for _, p := range body.Items {
		found[p.ID] = true
	}
	if !found[fx.projectA.ID] || !found[fx.projectB.ID] {
		t.Errorf("privileged listing missing projects: A=%v B=%v", found[fx.projectA.ID], found[fx.projectB.ID])
	}
}

// The PI holds a governance role on projectA but no membership row on allocA,
// so this passes only through the role path of the scoped wrapper.
func TestGetComputeAllocation_GovernanceRoleWithoutMembership(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)

	if rr := doGet(t, srv, "/compute-allocations/"+fx.allocA.ID, fx.piID); rr.Code != http.StatusOK {
		t.Errorf("PI on parent project: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
	if rr := doGet(t, srv, "/compute-allocations/"+fx.allocB.ID, fx.piID); rr.Code != http.StatusNotFound {
		t.Errorf("PI of another project: got %d, want 404 (%s)", rr.Code, rr.Body.String())
	}
}

// A missing allocation and a foreign allocation must be indistinguishable.
func TestGetComputeAllocation_DenialMatchesMissing(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)

	missing := doGet(t, srv, "/compute-allocations/does-not-exist", fx.memberID)
	foreign := doGet(t, srv, "/compute-allocations/"+fx.allocB.ID, fx.memberID)
	if missing.Code != http.StatusNotFound || foreign.Code != http.StatusNotFound {
		t.Fatalf("codes: missing=%d foreign=%d, want 404/404", missing.Code, foreign.Code)
	}
	if missing.Body.String() != foreign.Body.String() {
		t.Errorf("denial body differs from missing-id body:\n  missing: %s\n  foreign: %s",
			missing.Body.String(), foreign.Body.String())
	}
}
