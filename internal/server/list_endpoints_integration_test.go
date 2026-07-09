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
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
)

func TestListOrganizations_RequiresPrivilege(t *testing.T) {
	_, _, srv := setupTestStack(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
	req = withTestCaller(req, "u-1")
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status: got %d, want 403", rr.Code)
	}
}

func TestListOrganizations_ReturnsPage(t *testing.T) {
	_, svc, srv := setupTestStack(t)
	if _, err := svc.CreateOrganization(t.Context(), &models.Organization{
		OriginatedID: "list-test-org", Name: "List Test Org",
	}); err != nil {
		t.Fatalf("seed org: %v", err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/organizations?limit=5", nil)
	req = withTestCaller(req, "u-1", models.OrganizationsRead)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rr.Code)
	}
	var body struct {
		Items []models.Organization `json:"items"`
		Total int                   `json:"total"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Total < 1 || len(body.Items) < 1 {
		t.Errorf("expected at least one organization, got total=%d items=%d", body.Total, len(body.Items))
	}
	if len(body.Items) > 5 {
		t.Errorf("limit not applied: %d items", len(body.Items))
	}
}

func TestListUsers_RequiresPrivilege(t *testing.T) {
	_, _, srv := setupTestStack(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req = withTestCaller(req, "u-1")
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status: got %d, want 403", rr.Code)
	}
}

func TestListUsers_ReturnsPage(t *testing.T) {
	_, svc, srv := setupTestStack(t)
	org, err := svc.CreateOrganization(t.Context(), &models.Organization{
		OriginatedID: "list-test-user-org", Name: "List Test User Org",
	})
	if err != nil {
		t.Fatalf("seed org: %v", err)
	}
	if _, err := svc.CreateUser(t.Context(), &models.User{
		OrganizationID: org.ID, FirstName: "List", LastName: "Test",
		Email: fmt.Sprintf("list.test+%d@example.edu", time.Now().UnixNano()), Status: models.UserActive,
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users?limit=5", nil)
	req = withTestCaller(req, "u-1", models.UsersRead)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rr.Code)
	}
	var body struct {
		Items []models.User `json:"items"`
		Total int           `json:"total"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Total < 1 || len(body.Items) < 1 {
		t.Errorf("expected at least one user, got total=%d items=%d", body.Total, len(body.Items))
	}
	if len(body.Items) > 5 {
		t.Errorf("limit not applied: %d items", len(body.Items))
	}
}

func TestListResourceSummaries_ReturnsAggregates(t *testing.T) {
	_, svc, srv := setupTestStack(t)
	suffix := time.Now().UnixNano()
	cluster, err := svc.CreateComputeCluster(t.Context(), &models.ComputeCluster{
		Name: fmt.Sprintf("summary-test-cluster-%d", suffix),
	})
	if err != nil {
		t.Fatalf("seed cluster: %v", err)
	}
	res, err := svc.CreateComputeAllocationResource(t.Context(), &models.ComputeAllocationResource{
		Name: fmt.Sprintf("summary-test-cpu-%d", suffix), ResourceType: "CPU",
		ResourceAmount: 1000, ComputeClusterID: cluster.ID,
	})
	if err != nil {
		t.Fatalf("seed resource: %v", err)
	}
	for i, rate := range []float64{1.0, 1.25} {
		if _, err := svc.CreateComputeAllocationResourceRate(t.Context(), &models.ComputeAllocationResourceRate{
			ComputeAllocationResourceID: res.ID, Rate: rate,
			StartTime: time.Date(2025+i, 1, 1, 0, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026+i, 1, 1, 0, 0, 0, 0, time.UTC),
		}); err != nil {
			t.Fatalf("seed rate %d: %v", i, err)
		}
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/compute-allocation-resources/summary", nil)
	req = withTestCaller(req, "u-1", models.AllocationsRead)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rr.Code)
	}
	var rows []store.ComputeAllocationResourceSummary
	if err := json.NewDecoder(rr.Body).Decode(&rows); err != nil {
		t.Fatalf("decode: %v", err)
	}
	var found *store.ComputeAllocationResourceSummary
	for i := range rows {
		if rows[i].ID == res.ID {
			found = &rows[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("seeded resource %s not in summary rows", res.ID)
	}
	if found.RateCount != 2 {
		t.Errorf("rate_count: got %d, want 2", found.RateCount)
	}
	if found.AllocationCount != 0 || found.TotalAllocated != 0 || found.TotalUsedSU != 0 {
		t.Errorf("expected zero allocation aggregates, got count=%d allocated=%d used=%f",
			found.AllocationCount, found.TotalAllocated, found.TotalUsedSU)
	}
}

func TestListResourceSummaries_RequiresPrivilege(t *testing.T) {
	_, _, srv := setupTestStack(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/compute-allocation-resources/summary", nil)
	req = withTestCaller(req, "u-1")
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status: got %d, want 403", rr.Code)
	}
}

func TestListClusterUsers_EmptyListIsJSONArray(t *testing.T) {
	_, svc, srv := setupTestStack(t)
	cluster, err := svc.CreateComputeCluster(t.Context(), &models.ComputeCluster{
		Name: fmt.Sprintf("empty-list-cluster-%d", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("seed cluster: %v", err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/compute-clusters/"+cluster.ID+"/users", nil)
	req = withTestCaller(req, "u-1", models.ClustersRead)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rr.Code)
	}
	if body := strings.TrimSpace(rr.Body.String()); body != "[]" {
		t.Errorf("empty list body: got %q, want []", body)
	}
}

func TestUpdateUser_UpdatesNameFields(t *testing.T) {
	_, svc, srv := setupTestStack(t)
	org, err := svc.CreateOrganization(t.Context(), &models.Organization{
		OriginatedID: "name-update-org", Name: "Name Update Org",
	})
	if err != nil {
		t.Fatalf("seed org: %v", err)
	}
	user, err := svc.CreateUser(t.Context(), &models.User{
		OrganizationID: org.ID, FirstName: "Old", LastName: "Name",
		Email: fmt.Sprintf("name.update+%d@example.edu", time.Now().UnixNano()), Status: models.UserActive,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	body := strings.NewReader(`{"first_name":"New","middle_name":"Q","last_name":"Person"}`)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/users/"+user.ID, body)
	req = withTestCaller(req, "u-1", models.UsersWrite)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
	var got models.User
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.FirstName != "New" || got.MiddleName != "Q" || got.LastName != "Person" {
		t.Errorf("name not updated: got %q %q %q", got.FirstName, got.MiddleName, got.LastName)
	}
	if got.Email != user.Email {
		t.Errorf("email should be unchanged: got %q, want %q", got.Email, user.Email)
	}
}

func TestUpdateUser_RequiresPrivilege(t *testing.T) {
	_, _, srv := setupTestStack(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/users/some-id", strings.NewReader(`{"first_name":"X"}`))
	req = withTestCaller(req, "u-1")
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status: got %d, want 403", rr.Code)
	}
}
