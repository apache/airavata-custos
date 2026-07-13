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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/airavata-custos/pkg/models"
)

func TestGetCallerPrivileges_NoCaller_401(t *testing.T) {
	_, _, srv := setupTestStack(t)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/user/privileges", nil))
	// Route is router.RequireAuth, no privilege check. Without a caller on
	// ctx, the handler's requireCaller returns 401.
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestGetCallerPrivileges_NoGrants_ReturnsEmpty(t *testing.T) {
	_, _, srv := setupTestStack(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/privileges", nil)
	// Empty privilege set on ctx mirrors a caller the middleware resolved
	// with no direct grants and no role-derived privileges.
	req = withTestCaller(req, "u-1")
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rr.Code)
	}
	var body struct {
		Privileges []string `json:"privileges"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Privileges) != 0 {
		t.Errorf("privileges: got %v, want empty", body.Privileges)
	}
}

func TestGetCallerPrivileges_WithGrants(t *testing.T) {
	_, _, srv := setupTestStack(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/privileges", nil)
	req = withTestCaller(req, "u-1", models.PrivilegesGrant)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rr.Code)
	}
	var body struct {
		Privileges []string `json:"privileges"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Privileges) != 1 || body.Privileges[0] != string(models.PrivilegesGrant) {
		t.Errorf("privileges: got %v, want [%s]", body.Privileges, models.PrivilegesGrant)
	}
}

func TestGetCallerProfile_NoCaller_401(t *testing.T) {
	_, _, srv := setupTestStack(t)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/me", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestGetCallerProfile_ReturnsUserAndPrivileges(t *testing.T) {
	database, _, srv := setupTestStack(t)
	user := seedUser(t, database, "user@example.edu")
	// /me computes the effective set fresh from the DB, not from the request
	// context, so the grant must exist as a row.
	seedPrivilegesGrant(t, database, user)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req = withTestCaller(req, user, models.PrivilegesGrant)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200; body=%s", rr.Code, rr.Body.String())
	}
	var body CallerProfileResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.User == nil || body.User.ID != user {
		t.Errorf("user.id: got %+v, want %s", body.User, user)
	}
	if len(body.Privileges) != 1 || body.Privileges[0] != models.PrivilegesGrant {
		t.Errorf("privileges: got %v, want [%s]", body.Privileges, models.PrivilegesGrant)
	}
	if body.Roles == nil || len(body.Roles) != 0 {
		t.Errorf("roles: got %v, want empty array", body.Roles)
	}
}

func TestGetCallerProfile_IncludesOwnRoles(t *testing.T) {
	database, _, srv := setupTestStack(t)
	user := seedUser(t, database, "roleholder@example.edu")
	if _, err := database.Exec(
		"INSERT INTO roles (id, name, description, is_system) VALUES (?, ?, ?, ?)",
		"role-me-test", "Reviewer", "Reads reports", false,
	); err != nil {
		t.Fatalf("seed role: %v", err)
	}
	if _, err := database.Exec(
		"INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)",
		user, "role-me-test",
	); err != nil {
		t.Fatalf("seed grant: %v", err)
	}
	if _, err := database.Exec(
		"INSERT INTO role_privileges (role_id, privilege) VALUES (?, ?)",
		"role-me-test", "core:traces:read",
	); err != nil {
		t.Fatalf("seed role privilege: %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	// No admin privileges on purpose: /me must return the caller's own roles.
	req = withTestCaller(req, user)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200; body=%s", rr.Code, rr.Body.String())
	}
	var body CallerProfileResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Roles) != 1 || body.Roles[0].Role.Name != "Reviewer" {
		t.Fatalf("roles: got %+v, want the Reviewer role", body.Roles)
	}
	if body.Roles[0].GrantedAt.IsZero() {
		t.Errorf("granted_at is zero")
	}
	if len(body.Roles[0].Privileges) != 1 || body.Roles[0].Privileges[0] != "core:traces:read" {
		t.Errorf("role privileges: got %v, want [core:traces:read]", body.Roles[0].Privileges)
	}
	// The effective set is computed fresh, so the role-carried key shows up
	// without any cache warm-up or context injection.
	if len(body.Privileges) != 1 || body.Privileges[0] != "core:traces:read" {
		t.Errorf("effective privileges: got %v, want [core:traces:read]", body.Privileges)
	}
}

func TestRequirePrivilege_NoGrants_403(t *testing.T) {
	database, _, srv := setupTestStack(t)
	user := seedUser(t, database, "user@example.edu")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/privileges/catalog", nil)
	// Inject a caller whose ctx privilege set does NOT include privileges:grant.
	req = withTestCaller(req, user)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
}

func TestRequirePrivilege_WithGrant_200(t *testing.T) {
	database, _, srv := setupTestStack(t)
	user := seedUser(t, database, "user@example.edu")
	seedPrivilegesGrant(t, database, user)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/privileges/catalog", nil)
	req = withTestCaller(req, user, models.PrivilegesGrant)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rr.Code)
	}
	var cat []string
	if err := json.NewDecoder(rr.Body).Decode(&cat); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(cat) == 0 {
		t.Errorf("catalog empty")
	}
}

func TestGrantPrivilegeEndpoint_HappyPath(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilegesGrant(t, database, granter)

	body, _ := json.Marshal(map[string]any{"privilege": "core:clusters:read", "reason": "ops view"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users/"+target+"/privileges", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestCaller(req, granter, models.PrivilegesGrant)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status: got %d, want 201, body=%s", rr.Code, rr.Body.String())
	}
	has, err := svc.HasPrivilege(context.Background(), target, models.ClustersRead)
	if err != nil || !has {
		t.Errorf("HasPrivilege after grant endpoint: has=%v err=%v", has, err)
	}
}

func TestGrantPrivilegeEndpoint_GranterWithoutMeta_403(t *testing.T) {
	database, _, srv := setupTestStack(t)
	plain := seedUser(t, database, "plain@example.edu")
	target := seedUser(t, database, "target@example.edu")
	body, _ := json.Marshal(map[string]any{"privilege": "core:clusters:read"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users/"+target+"/privileges", bytes.NewReader(body))
	// Caller does NOT carry privileges:grant in its ctx set → gate denies.
	req = withTestCaller(req, plain)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403 (granter lacks privileges:grant)", rr.Code)
	}
}

func TestRevokePrivilegeEndpoint_HappyPath(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilegesGrant(t, database, granter)
	if _, err := svc.GrantPrivilege(context.Background(), target, models.ClustersRead, granter, ""); err != nil {
		t.Fatalf("seed grant: %v", err)
	}
	body, _ := json.Marshal(map[string]any{"reason": "rotated"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/"+target+"/privileges/core:clusters:read", bytes.NewReader(body))
	req = withTestCaller(req, granter, models.PrivilegesGrant)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status: got %d, want 204, body=%s", rr.Code, rr.Body.String())
	}
	if has, err := svc.HasPrivilege(context.Background(), target, models.ClustersRead); err != nil || has {
		t.Errorf("HasPrivilege after revoke: has=%v err=%v", has, err)
	}
}

func TestRevokePrivilegeEndpoint_SelfRevokeMeta_400(t *testing.T) {
	database, _, srv := setupTestStack(t)
	user := seedUser(t, database, "user@example.edu")
	seedPrivilegesGrant(t, database, user)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/"+user+"/privileges/core:privileges:grant", nil)
	req = withTestCaller(req, user, models.PrivilegesGrant)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400 (self-revoke of meta)", rr.Code)
	}
}

func TestListUserPrivilegesEndpoint(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilegesGrant(t, database, granter)
	if _, err := svc.GrantPrivilege(context.Background(), target, models.ClustersRead, granter, ""); err != nil {
		t.Fatalf("grant: %v", err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/"+target+"/privileges", nil)
	req = withTestCaller(req, granter, models.PrivilegesGrant)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200, body=%s", rr.Code, rr.Body.String())
	}
	var rows []models.UserPrivilege
	if err := json.NewDecoder(rr.Body).Decode(&rows); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(rows) != 1 || rows[0].Privilege != models.ClustersRead {
		t.Errorf("rows: got %v, want [core:clusters:read]", rows)
	}
}

func TestListPrivilegeHoldersEndpoint(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilegesGrant(t, database, granter)
	if _, err := svc.GrantPrivilege(context.Background(), target, models.ClustersRead, granter, ""); err != nil {
		t.Fatalf("grant: %v", err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/privileges/core:clusters:read/holders", nil)
	req = withTestCaller(req, granter, models.PrivilegesGrant)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200, body=%s", rr.Code, rr.Body.String())
	}
	var rows []models.UserPrivilege
	if err := json.NewDecoder(rr.Body).Decode(&rows); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(rows) != 1 || rows[0].UserID != target {
		t.Errorf("holders: got %v, want [target=%s]", rows, target)
	}
}

func TestRevokePropagatesCacheInvalidation(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	a := seedUser(t, database, "a@example.edu")
	b := seedUser(t, database, "b@example.edu")
	seedPrivilegesGrant(t, database, a)
	seedPrivilegesGrant(t, database, b)

	// Drive the revoke and confirm a subsequent service lookup reflects it.
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/"+b+"/privileges/core:privileges:grant", nil)
	req = withTestCaller(req, a, models.PrivilegesGrant)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("revoke: got %d, want 204, body=%s", rr.Code, rr.Body.String())
	}

	if has, err := svc.HasPrivilege(context.Background(), b, models.PrivilegesGrant); err != nil || has {
		t.Errorf("HasPrivilege after revoke: has=%v err=%v", has, err)
	}
}
