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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/apache/airavata-custos/pkg/models"
)

func TestAllocationMembershipRole_SetsProjectRoleAndKeepsUpstreamRoles(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)
	userID := seedUser(t, database, fmt.Sprintf("role.member+%d@example.edu", time.Now().UnixNano()))

	send := func(method, path string, body any) *httptest.ResponseRecorder {
		b, _ := json.Marshal(body)
		rr := httptest.NewRecorder()
		req := withTestCaller(httptest.NewRequest(method, path, bytes.NewReader(b)), fx.piID, models.AllocationsWrite, models.AllocationsRead)
		srv.ServeHTTP(rr, req)
		return rr
	}
	roleOf := func() string {
		var rows []AllocationMembershipResponse
		_ = json.NewDecoder(doGet(t, srv, "/compute-allocations/"+fx.allocA.ID+"/memberships", fx.piID, models.AllocationsRead).Body).Decode(&rows)
		for _, r := range rows {
			if r.UserID == userID {
				return r.Role
			}
		}
		return ""
	}

	now := time.Now().UTC()
	rr := send(http.MethodPost, "/compute-allocation-memberships", map[string]any{
		"compute_allocation_id": fx.allocA.ID, "user_id": userID,
		"start_time": now, "end_time": now.AddDate(0, 1, 0), "role": "ALLOCATION_MANAGER",
	})
	if rr.Code != http.StatusCreated {
		t.Fatalf("create: got %d (%s)", rr.Code, rr.Body.String())
	}
	var created models.ComputeAllocationMembership
	_ = json.NewDecoder(rr.Body).Decode(&created)
	if got := roleOf(); got != "ALLOCATION_MANAGER" {
		t.Fatalf("role after create: got %q", got)
	}

	if rr := send(http.MethodPut, "/compute-allocation-memberships/"+created.ID, map[string]any{"role": "MEMBER"}); rr.Code != http.StatusOK {
		t.Fatalf("update: got %d (%s)", rr.Code, rr.Body.String())
	}
	if got := roleOf(); got != "MEMBER" {
		t.Fatalf("role after update: got %q", got)
	}

	if rr := send(http.MethodPut, "/compute-allocation-memberships/"+created.ID, map[string]any{"role": "PI"}); rr.Code != http.StatusBadRequest {
		t.Fatalf("PI must not be assignable: got %d (%s)", rr.Code, rr.Body.String())
	}

	if err := svc.EnsureProjectMembership(t.Context(), fx.allocA.ProjectID, userID, "CO_PI"); err != nil {
		t.Fatal(err)
	}
	if rr := send(http.MethodPut, "/compute-allocation-memberships/"+created.ID, map[string]any{"role": "MEMBER"}); rr.Code != http.StatusOK || roleOf() != "CO_PI" {
		t.Fatalf("a CO_PI must keep the role: got %d, role %q", rr.Code, roleOf())
	}

	if err := svc.EnsureProjectMembership(t.Context(), fx.allocA.ProjectID, userID, "ALLOCATION_MANAGER"); err != nil {
		t.Fatal(err)
	}
	if rr := send(http.MethodDelete, "/compute-allocation-memberships/"+created.ID, nil); rr.Code != http.StatusNoContent {
		t.Fatalf("delete: got %d (%s)", rr.Code, rr.Body.String())
	}
	if role, err := svc.ProjectRoleForUser(t.Context(), fx.allocA.ProjectID, userID); err != nil || role != "" {
		t.Fatalf("a manager with no membership left must lose the role: got %q (%v)", role, err)
	}
}
