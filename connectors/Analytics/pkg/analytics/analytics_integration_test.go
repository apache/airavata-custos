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

package analytics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

func seedCluster(t *testing.T, db *sqlx.DB) string {
	t.Helper()
	id := uuid.NewString()
	if _, err := db.Exec(`INSERT INTO compute_clusters (id, name) VALUES (?, ?)`, id, "Cluster "+id[:8]); err != nil {
		t.Fatalf("seed cluster: %v", err)
	}
	return id
}

func seedProject(t *testing.T, db *sqlx.DB, piUserID string) string {
	t.Helper()
	id := uuid.NewString()
	if _, err := db.Exec(
		`INSERT INTO projects (id, originated_id, title, origination, project_pi_id, status, created_time)
		 VALUES (?, ?, ?, ?, ?, ?, NOW(6))`,
		id, "REC-"+id[:8], "Project "+id[:8], "TEST", piUserID, string(models.ProjectActive),
	); err != nil {
		t.Fatalf("seed project: %v", err)
	}
	return id
}

func seedProjectRole(t *testing.T, db *sqlx.DB, projectID, userID string, role models.ProjectRole) {
	t.Helper()
	if _, err := db.Exec(
		`INSERT INTO project_memberships (project_id, user_id, role, added_time) VALUES (?, ?, ?, NOW(6))`,
		projectID, userID, string(role),
	); err != nil {
		t.Fatalf("seed project role: %v", err)
	}
}

func seedAllocation(t *testing.T, db *sqlx.DB, projectID, clusterID string, initialSU int64, start, end time.Time) string {
	t.Helper()
	id := uuid.NewString()
	if _, err := db.Exec(
		`INSERT INTO compute_allocations
		     (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, projectID, "alloc-"+id[:8], string(models.ACTIVE), clusterID, initialSU, start, end,
	); err != nil {
		t.Fatalf("seed allocation: %v", err)
	}
	return id
}

func seedAllocMember(t *testing.T, db *sqlx.DB, allocID, userID string) {
	t.Helper()
	now := time.Now().UTC()
	if _, err := db.Exec(
		`INSERT INTO compute_allocation_memberships
		     (id, compute_allocation_id, user_id, start_time, end_time, membership_status)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.NewString(), allocID, userID, now, now.AddDate(1, 0, 0), string(models.ACTIVE),
	); err != nil {
		t.Fatalf("seed alloc member: %v", err)
	}
}

func seedResource(t *testing.T, db *sqlx.DB, clusterID, name, resourceType string) string {
	t.Helper()
	id := uuid.NewString()
	if _, err := db.Exec(
		`INSERT INTO compute_allocation_resources (id, name, resource_type, resource_amount, compute_cluster_id)
		 VALUES (?, ?, ?, ?, ?)`,
		id, name, resourceType, 0, clusterID,
	); err != nil {
		t.Fatalf("seed resource: %v", err)
	}
	return id
}

func seedUsage(t *testing.T, db *sqlx.DB, allocID, resourceID, userID string, su, raw int64, at time.Time) {
	t.Helper()
	if _, err := db.Exec(
		`INSERT INTO compute_allocation_usages
		     (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		uuid.NewString(), allocID, raw, su, at, userID, "job-"+uuid.NewString()[:8], resourceID,
	); err != nil {
		t.Fatalf("seed usage: %v", err)
	}
}

func TestAnalyticsContexts_NoCaller_401(t *testing.T) {
	_, _, srv := setupTestStack(t)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/connectors/analytics/contexts", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestAnalyticsContexts_EmptyWhenNoMemberships(t *testing.T) {
	database, _, srv := setupTestStack(t)
	user := seedUser(t, database, "lonely@example.edu")
	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/contexts", nil), user)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200; body=%s", rr.Code, rr.Body.String())
	}
	var got []ProjectContext
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("contexts: got %d, want 0", len(got))
	}
}

func TestAnalyticsContexts_PIRoleAndAllocations(t *testing.T) {
	database, _, srv := setupTestStack(t)
	pi := seedUser(t, database, "pi@example.edu")
	cluster := seedCluster(t, database)
	project := seedProject(t, database, pi)
	seedProjectRole(t, database, project, pi, models.ProjectRolePI)

	start := time.Now().UTC().AddDate(0, 0, -10)
	end1 := time.Now().UTC().AddDate(0, 0, 20)
	end2 := time.Now().UTC().AddDate(0, 0, 40)
	alloc1 := seedAllocation(t, database, project, cluster, 1000, start, end1)
	_ = seedAllocation(t, database, project, cluster, 2000, start, end2)
	res := seedResource(t, database, cluster, "gpu-01", "GPU_HOURS")
	seedUsage(t, database, alloc1, res, pi, 300, 30, time.Now().UTC().AddDate(0, 0, -1))

	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/contexts", nil), pi)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200; body=%s", rr.Code, rr.Body.String())
	}
	var got []ProjectContext
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("contexts: got %d, want 1", len(got))
	}
	c := got[0]
	if c.Role != string(models.ProjectRolePI) {
		t.Errorf("role: got %q, want PI", c.Role)
	}
	if len(c.Allocations) != 2 {
		t.Fatalf("allocations: got %d, want 2", len(c.Allocations))
	}
	var usedForAlloc1 float64 = -1
	for _, a := range c.Allocations {
		if a.ID == alloc1 {
			usedForAlloc1 = a.UsedSUAmount
		}
	}
	if usedForAlloc1 != 300 {
		t.Errorf("used_su_amount for alloc1: got %v, want 300", usedForAlloc1)
	}
}

func TestAnalyticsContexts_MemberViaAllocationOnly(t *testing.T) {
	database, _, srv := setupTestStack(t)
	pi := seedUser(t, database, "pi2@example.edu")
	member := seedUser(t, database, "member@example.edu")
	cluster := seedCluster(t, database)
	project := seedProject(t, database, pi)
	alloc := seedAllocation(t, database, project, cluster, 1000,
		time.Now().UTC().AddDate(0, 0, -5), time.Now().UTC().AddDate(0, 0, 25))
	seedAllocMember(t, database, alloc, member)

	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/contexts", nil), member)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200; body=%s", rr.Code, rr.Body.String())
	}
	var got []ProjectContext
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("contexts: got %d, want 1", len(got))
	}
	if got[0].Role != roleMember {
		t.Errorf("role: got %q, want MEMBER (derived from allocation membership)", got[0].Role)
	}
}

// A plain member must see only the allocations they belong to, not sibling
// allocations in the same project (those 404 on usage-summary). A project
// manager on the same project sees all of them.
func TestAnalyticsContexts_MemberSeesOnlySiblingTheyBelongTo(t *testing.T) {
	database, _, srv := setupTestStack(t)
	pi := seedUser(t, database, "pi-sib@example.edu")
	member := seedUser(t, database, "member-sib@example.edu")
	cluster := seedCluster(t, database)
	project := seedProject(t, database, pi)
	seedProjectRole(t, database, project, pi, models.ProjectRolePI)
	start := time.Now().UTC().AddDate(0, 0, -5)
	end := time.Now().UTC().AddDate(0, 0, 25)
	mine := seedAllocation(t, database, project, cluster, 1000, start, end)
	_ = seedAllocation(t, database, project, cluster, 2000, start, end) // sibling, member not on it
	seedAllocMember(t, database, mine, member)

	// Member: only the allocation they belong to.
	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/contexts", nil), member)
	srv.ServeHTTP(rr, req)
	var asMember []ProjectContext
	if err := json.NewDecoder(rr.Body).Decode(&asMember); err != nil {
		t.Fatalf("decode member: %v", err)
	}
	if len(asMember) != 1 || len(asMember[0].Allocations) != 1 || asMember[0].Allocations[0].ID != mine {
		t.Errorf("member allocations: got %+v, want only the one allocation they belong to", asMember)
	}

	// PI: both allocations.
	rr = httptest.NewRecorder()
	req = withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/contexts", nil), pi)
	srv.ServeHTTP(rr, req)
	var asPI []ProjectContext
	if err := json.NewDecoder(rr.Body).Decode(&asPI); err != nil {
		t.Fatalf("decode pi: %v", err)
	}
	if len(asPI) != 1 || len(asPI[0].Allocations) != 2 {
		t.Errorf("pi allocations: got %+v, want both project allocations", asPI)
	}
}

func TestUsageSummary_NoCaller_401(t *testing.T) {
	_, _, srv := setupTestStack(t)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/connectors/analytics/allocations/whatever/usage-summary", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestUsageSummary_UnknownAllocation_404(t *testing.T) {
	database, _, srv := setupTestStack(t)
	user := seedUser(t, database, "u@example.edu")
	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/allocations/nope/usage-summary", nil), user)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", rr.Code)
	}
}

func TestUsageSummary_NonMember_404(t *testing.T) {
	database, _, srv := setupTestStack(t)
	pi := seedUser(t, database, "pi3@example.edu")
	stranger := seedUser(t, database, "stranger@example.edu")
	cluster := seedCluster(t, database)
	project := seedProject(t, database, pi)
	alloc := seedAllocation(t, database, project, cluster, 1000,
		time.Now().UTC().AddDate(0, 0, -5), time.Now().UTC().AddDate(0, 0, 25))

	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/allocations/"+alloc+"/usage-summary", nil), stranger)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("non-member status: got %d, want 404 (no existence leak)", rr.Code)
	}
}

func TestUsageSummary_InactiveMember_404(t *testing.T) {
	database, _, srv := setupTestStack(t)
	pi := seedUser(t, database, "pi-inact@example.edu")
	former := seedUser(t, database, "former@example.edu")
	cluster := seedCluster(t, database)
	project := seedProject(t, database, pi)
	alloc := seedAllocation(t, database, project, cluster, 1000,
		time.Now().UTC().AddDate(0, 0, -5), time.Now().UTC().AddDate(0, 0, 25))
	// An inactive membership must not grant analytics access.
	if _, err := database.Exec(
		`INSERT INTO compute_allocation_memberships
		     (id, compute_allocation_id, user_id, start_time, end_time, membership_status)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		uuid.NewString(), alloc, former, time.Now().UTC().AddDate(0, 0, -30), time.Now().UTC(), string(models.INACTIVE),
	); err != nil {
		t.Fatalf("seed inactive membership: %v", err)
	}

	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/allocations/"+alloc+"/usage-summary", nil), former)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("inactive member status: got %d, want 404", rr.Code)
	}
}

func TestUsageSummary_MemberSeesOwnSliceNoMembers(t *testing.T) {
	database, _, srv := setupTestStack(t)
	pi := seedUser(t, database, "pi4@example.edu")
	member := seedUser(t, database, "m4@example.edu")
	cluster := seedCluster(t, database)
	project := seedProject(t, database, pi)
	start := time.Now().UTC().AddDate(0, 0, -3)
	alloc := seedAllocation(t, database, project, cluster, 1000, start, time.Now().UTC().AddDate(0, 0, 27))
	seedAllocMember(t, database, alloc, member)
	res := seedResource(t, database, cluster, "gpu-01", "GPU_HOURS")
	seedUsage(t, database, alloc, res, member, 100, 10, time.Now().UTC().AddDate(0, 0, -1))
	seedUsage(t, database, alloc, res, pi, 400, 40, time.Now().UTC().AddDate(0, 0, -1))

	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/allocations/"+alloc+"/usage-summary", nil), member)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200; body=%s", rr.Code, rr.Body.String())
	}
	var got UsageSummary
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ByMember != nil {
		t.Errorf("by_member: got %v, want null for a plain member", got.ByMember)
	}
	if got.Total != 1000 {
		t.Errorf("total: got %d, want 1000", got.Total)
	}
	if got.Used != 500 {
		t.Errorf("used: got %v, want 500", got.Used)
	}
	if len(got.ByResource) != 1 {
		t.Fatalf("by_resource: got %d, want 1", len(got.ByResource))
	}
	r := got.ByResource[0]
	if r.Used != 500 {
		t.Errorf("resource used: got %v, want 500", r.Used)
	}
	if r.UsedByCaller != 100 {
		t.Errorf("used_by_caller: got %v, want 100 (member's own slice)", r.UsedByCaller)
	}
	if r.NativeUnit != "GPU-hours" {
		t.Errorf("native_unit: got %q, want GPU-hours", r.NativeUnit)
	}
	if r.Cap != nil {
		t.Errorf("cap: got %v, want null in v1", r.Cap)
	}
}

func TestUsageSummary_PISeesRankedMembers(t *testing.T) {
	database, _, srv := setupTestStack(t)
	pi := seedUser(t, database, "pi5@example.edu")
	member := seedUser(t, database, "m5@example.edu")
	cluster := seedCluster(t, database)
	project := seedProject(t, database, pi)
	seedProjectRole(t, database, project, pi, models.ProjectRolePI)
	start := time.Now().UTC().AddDate(0, 0, -2)
	alloc := seedAllocation(t, database, project, cluster, 1000, start, time.Now().UTC().AddDate(0, 0, 28))
	res := seedResource(t, database, cluster, "gpu-01", "GPU_HOURS")
	seedUsage(t, database, alloc, res, member, 100, 10, time.Now().UTC().AddDate(0, 0, -1))
	seedUsage(t, database, alloc, res, pi, 400, 40, time.Now().UTC().AddDate(0, 0, -1))

	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/allocations/"+alloc+"/usage-summary", nil), pi)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200; body=%s", rr.Code, rr.Body.String())
	}
	var got UsageSummary
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ByMember == nil {
		t.Fatalf("by_member: got null, want populated for PI")
	}
	if len(got.ByMember) != 2 {
		t.Fatalf("by_member: got %d, want 2", len(got.ByMember))
	}
	if got.ByMember[0].UserID != pi || got.ByMember[0].Used != 400 {
		t.Errorf("by_member[0]: got %+v, want pi with 400 (ranked first)", got.ByMember[0])
	}
	if got.ByMember[1].Used != 100 {
		t.Errorf("by_member[1] used: got %v, want 100", got.ByMember[1].Used)
	}
}

// Access is membership-only: a site-wide privilege does not open an allocation
// the caller has no membership or role on.
func TestUsageSummary_SitePrivilegeAloneDoesNotGrantAccess_404(t *testing.T) {
	database, _, srv := setupTestStack(t)
	pi := seedUser(t, database, "pi6@example.edu")
	admin := seedUser(t, database, "admin6@example.edu")
	cluster := seedCluster(t, database)
	project := seedProject(t, database, pi)
	start := time.Now().UTC().AddDate(0, 0, -2)
	alloc := seedAllocation(t, database, project, cluster, 1000, start, time.Now().UTC().AddDate(0, 0, 28))
	res := seedResource(t, database, cluster, "gpu-01", "GPU_HOURS")
	seedUsage(t, database, alloc, res, pi, 400, 40, time.Now().UTC().AddDate(0, 0, -1))

	rr := httptest.NewRecorder()
	// admin has no membership on this allocation, only site allocations:read.
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/allocations/"+alloc+"/usage-summary", nil), admin, models.AllocationsRead)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404 (site privilege must not grant access)", rr.Code)
	}
}

func TestAllocationJobs_NoCaller_401(t *testing.T) {
	_, _, srv := setupTestStack(t)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/connectors/analytics/allocations/x/jobs", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestAllocationJobs_NonMember_404(t *testing.T) {
	database, _, srv := setupTestStack(t)
	pi := seedUser(t, database, "pij@example.edu")
	stranger := seedUser(t, database, "strangerj@example.edu")
	cluster := seedCluster(t, database)
	project := seedProject(t, database, pi)
	alloc := seedAllocation(t, database, project, cluster, 1000,
		time.Now().UTC().AddDate(0, 0, -3), time.Now().UTC().AddDate(0, 0, 27))
	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/allocations/"+alloc+"/jobs", nil), stranger)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", rr.Code)
	}
}

func TestAllocationJobs_MemberSeesOnlyOwn(t *testing.T) {
	database, _, srv := setupTestStack(t)
	pi := seedUser(t, database, "pij2@example.edu")
	member := seedUser(t, database, "mj2@example.edu")
	cluster := seedCluster(t, database)
	project := seedProject(t, database, pi)
	alloc := seedAllocation(t, database, project, cluster, 1000,
		time.Now().UTC().AddDate(0, 0, -3), time.Now().UTC().AddDate(0, 0, 27))
	seedAllocMember(t, database, alloc, member)
	res := seedResource(t, database, cluster, "gpu-01", "GPU_HOURS")
	seedUsage(t, database, alloc, res, member, 100, 10, time.Now().UTC().AddDate(0, 0, -1))
	seedUsage(t, database, alloc, res, pi, 400, 40, time.Now().UTC().AddDate(0, 0, -1))

	// Member requesting mine=false is still forced to their own jobs.
	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/allocations/"+alloc+"/jobs?mine=false", nil), member)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200; body=%s", rr.Code, rr.Body.String())
	}
	var got AllocationJobs
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Total != 1 || len(got.Jobs) != 1 {
		t.Fatalf("member jobs: total=%d len=%d, want 1/1 (own only)", got.Total, len(got.Jobs))
	}
	if got.Jobs[0].UserID != member {
		t.Errorf("job user: got %s, want the member", got.Jobs[0].UserID)
	}
	if got.Jobs[0].NativeUnit != "GPU-hours" {
		t.Errorf("native_unit: got %q, want GPU-hours", got.Jobs[0].NativeUnit)
	}
}

func TestAllocationJobs_ManagerSeesAllAndCanFilterMine(t *testing.T) {
	database, _, srv := setupTestStack(t)
	pi := seedUser(t, database, "pij3@example.edu")
	member := seedUser(t, database, "mj3@example.edu")
	cluster := seedCluster(t, database)
	project := seedProject(t, database, pi)
	seedProjectRole(t, database, project, pi, models.ProjectRolePI)
	alloc := seedAllocation(t, database, project, cluster, 1000,
		time.Now().UTC().AddDate(0, 0, -3), time.Now().UTC().AddDate(0, 0, 27))
	res := seedResource(t, database, cluster, "gpu-01", "GPU_HOURS")
	seedUsage(t, database, alloc, res, member, 100, 10, time.Now().UTC().AddDate(0, 0, -2))
	seedUsage(t, database, alloc, res, pi, 400, 40, time.Now().UTC().AddDate(0, 0, -1))

	// Default: everyone's jobs, newest first.
	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/allocations/"+alloc+"/jobs", nil), pi)
	srv.ServeHTTP(rr, req)
	var all AllocationJobs
	if err := json.NewDecoder(rr.Body).Decode(&all); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if all.Total != 2 || len(all.Jobs) != 2 {
		t.Fatalf("manager jobs: total=%d len=%d, want 2/2", all.Total, len(all.Jobs))
	}
	if all.Jobs[0].UserID != pi {
		t.Errorf("ordering: got first user %s, want the pi (newest)", all.Jobs[0].UserID)
	}

	// mine=true narrows to the manager's own.
	rr = httptest.NewRecorder()
	req = withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/allocations/"+alloc+"/jobs?mine=true", nil), pi)
	srv.ServeHTTP(rr, req)
	var mine AllocationJobs
	if err := json.NewDecoder(rr.Body).Decode(&mine); err != nil {
		t.Fatalf("decode mine: %v", err)
	}
	if mine.Total != 1 || len(mine.Jobs) != 1 || mine.Jobs[0].UserID != pi {
		t.Errorf("mine filter: got total=%d len=%d, want 1/1 for the pi", mine.Total, len(mine.Jobs))
	}
}

func TestAllocationJobs_Pagination(t *testing.T) {
	database, _, srv := setupTestStack(t)
	pi := seedUser(t, database, "pij4@example.edu")
	cluster := seedCluster(t, database)
	project := seedProject(t, database, pi)
	seedProjectRole(t, database, project, pi, models.ProjectRolePI)
	alloc := seedAllocation(t, database, project, cluster, 100000,
		time.Now().UTC().AddDate(0, 0, -5), time.Now().UTC().AddDate(0, 0, 25))
	res := seedResource(t, database, cluster, "gpu-01", "GPU_HOURS")
	for i := 0; i < 5; i++ {
		seedUsage(t, database, alloc, res, pi, int64(100+i), 10, time.Now().UTC().AddDate(0, 0, -i))
	}

	fetch := func(offset int) AllocationJobs {
		rr := httptest.NewRecorder()
		url := "/connectors/analytics/allocations/" + alloc + "/jobs?limit=2&offset=" + strconv.Itoa(offset)
		req := withTestCaller(httptest.NewRequest(http.MethodGet, url, nil), pi)
		srv.ServeHTTP(rr, req)
		var got AllocationJobs
		if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		return got
	}

	first := fetch(0)
	if first.Total != 5 || len(first.Jobs) != 2 {
		t.Fatalf("first page: total=%d len=%d, want 5/2", first.Total, len(first.Jobs))
	}
	second := fetch(2)
	if second.Total != 5 || len(second.Jobs) != 2 {
		t.Fatalf("second page: total=%d len=%d, want 5/2", second.Total, len(second.Jobs))
	}
	// Offset must advance the window: the second page shares no ids with the first.
	firstIDs := map[string]bool{first.Jobs[0].ID: true, first.Jobs[1].ID: true}
	if firstIDs[second.Jobs[0].ID] || firstIDs[second.Jobs[1].ID] {
		t.Errorf("offset not applied: page 2 overlaps page 1")
	}
}

func TestUsageSummary_DailyBucketsContinuous(t *testing.T) {
	database, _, srv := setupTestStack(t)
	pi := seedUser(t, database, "pi7@example.edu")
	cluster := seedCluster(t, database)
	project := seedProject(t, database, pi)
	seedProjectRole(t, database, project, pi, models.ProjectRolePI)
	start := time.Now().UTC().AddDate(0, 0, -5)
	alloc := seedAllocation(t, database, project, cluster, 1000, start, time.Now().UTC().AddDate(0, 0, 25))
	res := seedResource(t, database, cluster, "gpu-01", "GPU_HOURS")
	// Usage only on two days; the rest must still appear as zero buckets.
	seedUsage(t, database, alloc, res, pi, 100, 10, time.Now().UTC().AddDate(0, 0, -4))
	seedUsage(t, database, alloc, res, pi, 250, 25, time.Now().UTC().AddDate(0, 0, -1))

	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/connectors/analytics/allocations/"+alloc+"/usage-summary", nil), pi)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200; body=%s", rr.Code, rr.Body.String())
	}
	var got UsageSummary
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// start is 5 days ago; buckets span start..today inclusive = 6.
	if len(got.Daily) != 6 {
		t.Fatalf("daily buckets: got %d, want 6 (continuous start..today)", len(got.Daily))
	}
	var total float64
	nonEmpty := 0
	for _, b := range got.Daily {
		for _, v := range b.ByResource {
			total += v
		}
		if len(b.ByResource) > 0 {
			nonEmpty++
		}
	}
	if total != 350 {
		t.Errorf("summed daily credits: got %v, want 350", total)
	}
	if nonEmpty != 2 {
		t.Errorf("non-empty days: got %d, want 2", nonEmpty)
	}
}
