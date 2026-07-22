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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/apache/airavata-custos/pkg/models"
)

func putMemberRole(t *testing.T, srv *Server, projectID, userID, jsonBody, caller string, privs ...models.PrivilegeKey) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/projects/"+projectID+"/members/"+userID, strings.NewReader(jsonBody))
	req = withTestCaller(req, caller, privs...)
	srv.ServeHTTP(rr, req)
	return rr
}

// A trial user holds only an allocation membership; elevating them creates the
// governance membership on the parent project.
func TestSetProjectMemberRole_ElevatesAllocationMember(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)

	rr := putMemberRole(t, srv, fx.projectA.ID, fx.memberID, `{"role":"ALLOCATION_MANAGER"}`, "admin", models.ProjectsWrite)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
	var got map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["role"] != "ALLOCATION_MANAGER" {
		t.Errorf("role: got %q, want ALLOCATION_MANAGER", got["role"])
	}

	role, err := svc.ProjectRoleForUser(t.Context(), fx.projectA.ID, fx.memberID)
	if err != nil {
		t.Fatalf("role lookup: %v", err)
	}
	if role != models.ProjectRoleAllocationManager {
		t.Errorf("persisted role: got %q, want ALLOCATION_MANAGER", role)
	}
}

func TestSetProjectMemberRole_RequiresPrivilege(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)

	rr := putMemberRole(t, srv, fx.projectA.ID, fx.memberID, `{"role":"ALLOCATION_MANAGER"}`, "no-privs")
	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
}

func TestSetProjectMemberRole_CannotDemotePI(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)

	rr := putMemberRole(t, srv, fx.projectA.ID, fx.piID, `{"role":"MEMBER"}`, "admin", models.ProjectsWrite)
	if rr.Code != http.StatusConflict {
		t.Errorf("status: got %d, want 409 (%s)", rr.Code, rr.Body.String())
	}
}

func TestRemoveProjectMember_ClearsRole(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	fx := seedScopedFixtures(t, database, svc)

	if rr := putMemberRole(t, srv, fx.projectA.ID, fx.memberID, `{"role":"ALLOCATION_MANAGER"}`, "admin", models.ProjectsWrite); rr.Code != http.StatusOK {
		t.Fatalf("elevate: got %d (%s)", rr.Code, rr.Body.String())
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/projects/"+fx.projectA.ID+"/members/"+fx.memberID, nil)
	req = withTestCaller(req, "admin", models.ProjectsWrite)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("delete: got %d, want 204 (%s)", rr.Code, rr.Body.String())
	}

	role, err := svc.ProjectRoleForUser(t.Context(), fx.projectA.ID, fx.memberID)
	if err != nil {
		t.Fatalf("role lookup: %v", err)
	}
	if role != "" {
		t.Errorf("role after delete: got %q, want empty", role)
	}
}
