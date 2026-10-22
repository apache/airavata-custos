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

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/tracing/tracingtest"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// seedClusterUser creates a cluster and a user, then a cluster user with no approval given, and returns it.
func seedClusterUser(t *testing.T, database *sqlx.DB, svc *service.Service) *models.ComputeClusterUser {
	t.Helper()
	return seedClusterUserOn(t, database, svc, seedCluster(t, svc))
}

func seedCluster(t *testing.T, svc *service.Service) string {
	t.Helper()
	cluster, err := svc.CreateComputeCluster(t.Context(), &models.ComputeCluster{Name: "appr-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	return cluster.ID
}

func seedClusterUserOn(t *testing.T, database *sqlx.DB, svc *service.Service, clusterID string) *models.ComputeClusterUser {
	t.Helper()
	userID := seedUser(t, database, fmt.Sprintf("member-%s@example.edu", uuid.NewString()))
	cu, err := svc.CreateComputeClusterUser(t.Context(), &models.ComputeClusterUser{
		ComputeClusterID: clusterID,
		UserID:           userID,
		LocalUsername:    "appr-" + uuid.NewString()[:8],
	})
	if err != nil {
		t.Fatalf("create cluster user: %v", err)
	}
	return cu
}

func seedAllocation(t *testing.T, database *sqlx.DB, svc *service.Service, clusterID, projectTitle string) *models.ComputeAllocation {
	t.Helper()
	pi := seedUser(t, database, fmt.Sprintf("pi-%s@example.edu", uuid.NewString()))
	project, err := svc.CreateProject(t.Context(), &models.Project{Title: projectTitle, ProjectPIID: pi})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	alloc, err := svc.CreateComputeAllocation(t.Context(), &models.ComputeAllocation{
		Name: "alloc-" + uuid.NewString()[:8], ProjectID: project.ID, ComputeClusterID: clusterID,
	})
	if err != nil {
		t.Fatalf("create allocation: %v", err)
	}
	return alloc
}

func addMember(t *testing.T, svc *service.Service, allocationID, userID string) {
	t.Helper()
	if _, err := svc.CreateComputeAllocationMembership(t.Context(), &models.ComputeAllocationMembership{
		ComputeAllocationID: allocationID, UserID: userID,
	}); err != nil {
		t.Fatalf("create membership: %v", err)
	}
}

func listClusterAccounts(t *testing.T, srv *Server, admin, query string) ClusterAccountListResponse {
	t.Helper()
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, withTestCaller(httptest.NewRequest(http.MethodGet, "/compute-cluster-users?"+query, nil), admin, models.ClustersRead))
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200: %s", rr.Code, rr.Body.String())
	}
	var body ClusterAccountListResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return body
}

// Make sure a request shows the user's allocation on the same cluster, not one on another
// cluster, since the account only gives access to its own cluster.
func TestListClusterAccounts_ShowsAllocationOnSameCluster(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))
	clusterID := seedCluster(t, svc)
	cu := seedClusterUserOn(t, database, svc, clusterID)
	here := seedAllocation(t, database, svc, clusterID, "Turbulent mixing")
	addMember(t, svc, here.ID, cu.UserID)
	elsewhere := seedAllocation(t, database, svc, seedCluster(t, svc), "Other cluster work")
	addMember(t, svc, elsewhere.ID, cu.UserID)

	body := listClusterAccounts(t, srv, admin, "approval_status=PENDING")
	if len(body.Items) != 1 {
		t.Fatalf("expected one account, got %+v", body)
	}
	got := body.Items[0]
	if got.AllocationID == nil || *got.AllocationID != here.ID || got.ProjectID == nil || *got.ProjectID != here.ProjectID {
		t.Fatalf("expected the allocation on the account's cluster, got %+v", got)
	}
}

// Make sure pending requests come grouped by project and allocation with unlinked ones last,
// even when they arrived interleaved, so a page never splits one allocation's rows apart.
func TestListClusterAccounts_GroupsPendingByAllocation(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))
	clusterID := seedCluster(t, svc)
	alpha := seedAllocation(t, database, svc, clusterID, "Alpha project")
	beta := seedAllocation(t, database, svc, clusterID, "Beta project")

	seedClusterUserOn(t, database, svc, clusterID)
	for _, alloc := range []*models.ComputeAllocation{beta, alpha, beta, alpha} {
		cu := seedClusterUserOn(t, database, svc, clusterID)
		addMember(t, svc, alloc.ID, cu.UserID)
	}

	body := listClusterAccounts(t, srv, admin, "approval_status=PENDING")
	var got []string
	for _, item := range body.Items {
		id := "none"
		if item.AllocationID != nil {
			id = *item.AllocationID
		}
		got = append(got, id)
	}
	want := []string{alpha.ID, alpha.ID, beta.ID, beta.ID, "none"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("order: got %v, want %v", got, want)
	}
}

// Make sure search matches the allocation name, so an admin can find every request for one allocation.
func TestListClusterAccounts_SearchesByAllocation(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))
	clusterID := seedCluster(t, svc)
	alloc := seedAllocation(t, database, svc, clusterID, "Searchable project")
	member := seedClusterUserOn(t, database, svc, clusterID)
	addMember(t, svc, alloc.ID, member.UserID)
	seedClusterUserOn(t, database, svc, clusterID)

	body := listClusterAccounts(t, srv, admin, "q="+alloc.Name)
	if body.Total != 1 || len(body.Items) != 1 || body.Items[0].ID != member.ID {
		t.Fatalf("expected only the member of %s, got %+v", alloc.Name, body)
	}
}

// Make sure the list can be narrowed to the accounts waiting for approval and
// carries the user's name, email and the cluster name the admin needs.
func TestListClusterAccounts_FiltersByApprovalStatus(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))
	pending := seedClusterUser(t, database, svc)
	approved := seedClusterUser(t, database, svc)
	if _, err := svc.ApproveComputeClusterUser(tracingtest.Context(), approved.ID, admin); err != nil {
		t.Fatalf("approve: %v", err)
	}

	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodGet, "/compute-cluster-users?approval_status=PENDING", nil), admin, models.ClustersRead)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200: %s", rr.Code, rr.Body.String())
	}
	var body ClusterAccountListResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Total != 1 || len(body.Items) != 1 || body.Items[0].ID != pending.ID {
		t.Fatalf("expected only the pending account, got %+v", body)
	}
	if body.Items[0].DisplayName != "Test User" || body.Items[0].Email == "" || body.Items[0].ClusterName == "" {
		t.Fatalf("expected the user and cluster fields on the row, got %+v", body.Items[0])
	}
}

// Make sure approving through the API records the caller as the reviewer.
func TestApproveClusterAccount_RecordsCaller(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	admin := seedUser(t, database, fmt.Sprintf("admin-%s@example.edu", uuid.NewString()))
	cu := seedClusterUser(t, database, svc)

	rr := httptest.NewRecorder()
	req := withTestCaller(httptest.NewRequest(http.MethodPost, "/compute-cluster-users/"+cu.ID+"/approve", nil), admin, models.ClustersWrite)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200: %s", rr.Code, rr.Body.String())
	}
	var got models.ComputeClusterUser
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ApprovalStatus != models.ClusterAccountApproved || got.ReviewedBy == nil || *got.ReviewedBy != admin {
		t.Fatalf("expected approval by the caller, got %+v", got)
	}
}

// Make sure the read privilege is not enough to approve or deny an account.
func TestReviewClusterAccount_RequiresClustersWrite(t *testing.T) {
	_, _, srv := setupTestStack(t)
	for _, action := range []string{"approve", "deny"} {
		rr := httptest.NewRecorder()
		req := withTestCaller(httptest.NewRequest(http.MethodPost, "/compute-cluster-users/any/"+action, nil), "u-1", models.ClustersRead)
		srv.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("%s status: got %d, want 403", action, rr.Code)
		}
	}
}
