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
	cluster, err := svc.CreateComputeCluster(t.Context(), &models.ComputeCluster{Name: "appr-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}
	userID := seedUser(t, database, fmt.Sprintf("member-%s@example.edu", uuid.NewString()))
	cu, err := svc.CreateComputeClusterUser(t.Context(), &models.ComputeClusterUser{
		ComputeClusterID: cluster.ID,
		UserID:           userID,
		LocalUsername:    "appr-" + uuid.NewString()[:8],
	})
	if err != nil {
		t.Fatalf("create cluster user: %v", err)
	}
	return cu
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
