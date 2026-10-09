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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/airavata-custos/pkg/models"
)

func doWrite(t *testing.T, srv *Server, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	raw, _ := json.Marshal(body)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, withTestCaller(httptest.NewRequest(method, path, bytes.NewReader(raw)), "admin", models.ProjectsWrite, models.ProjectsRead))
	return rr
}

func projectRoles(t *testing.T, srv *Server, projectID string) map[string]string {
	t.Helper()
	var rows []ProjectMemberResponse
	_ = json.NewDecoder(doGet(t, srv, "/projects/"+projectID+"/members", "admin", models.ProjectsRead).Body).Decode(&rows)
	roles := map[string]string{}
	for _, r := range rows {
		roles[r.UserID] = r.Role
	}
	return roles
}

func TestAdminProject_CreateAssignRoleAndChangePI(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	f := seedScopedFixtures(t, database, svc)
	newPIID := seedUser(t, database, "admin.new.pi@example.edu")

	rr := doWrite(t, srv, http.MethodPost, "/projects", map[string]string{"title": "Admin Project", "project_pi_id": f.piID})
	var created models.Project
	_ = json.NewDecoder(rr.Body).Decode(&created)
	if rr.Code != http.StatusCreated || created.Origination != "internal" {
		t.Fatalf("create: %d %+v", rr.Code, created)
	}
	if roles := projectRoles(t, srv, created.ID); len(roles) != 1 || roles[f.piID] != "PI" {
		t.Fatalf("created project roles: %v", roles)
	}

	path := "/projects/" + f.projectA.ID
	if rr := doWrite(t, srv, http.MethodPut, path+"/members/"+newPIID, map[string]string{"role": "CO_PI"}); rr.Code != http.StatusNoContent {
		t.Fatalf("assign role: %d %s", rr.Code, rr.Body.String())
	}
	if roles := projectRoles(t, srv, f.projectA.ID); roles[newPIID] != "CO_PI" {
		t.Fatalf("assigned role: %v", roles)
	}

	if rr := doWrite(t, srv, http.MethodPut, path, map[string]string{"project_pi_id": newPIID, "previous_pi_role": "CO_PI"}); rr.Code != http.StatusOK {
		t.Fatalf("change PI: %d %s", rr.Code, rr.Body.String())
	}
	if roles := projectRoles(t, srv, f.projectA.ID); roles[newPIID] != "PI" || roles[f.piID] != "CO_PI" {
		t.Fatalf("roles after PI change: %v", roles)
	}

	for _, c := range []struct {
		user, role string
		code       int
	}{{newPIID, "MEMBER", http.StatusConflict}, {f.piID, "PI", http.StatusConflict}} {
		if rr := doWrite(t, srv, http.MethodPut, path+"/members/"+c.user, map[string]string{"role": c.role}); rr.Code != c.code {
			t.Fatalf("PI role via %s %s: %d %s", c.user, c.role, rr.Code, rr.Body.String())
		}
	}
	if rr := doWrite(t, srv, http.MethodDelete, path, nil); rr.Code != http.StatusConflict {
		t.Fatalf("delete with allocations: %d %s", rr.Code, rr.Body.String())
	}
	if rr := doWrite(t, srv, http.MethodDelete, "/projects/"+created.ID, nil); rr.Code != http.StatusNoContent {
		t.Fatalf("delete: %d %s", rr.Code, rr.Body.String())
	}
}
