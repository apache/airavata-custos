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

func TestGetCallerPrivileges_MissingHeader_401(t *testing.T) {
	_, _, srv := setupTestStack(t)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/user/privileges", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", rr.Code)
	}
}

func TestGetCallerPrivileges_NoGrants_ReturnsEmpty(t *testing.T) {
	database, _, srv := setupTestStack(t)
	user := seedUser(t, database, "plain@example.edu")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/privileges", nil)
	req = asCaller(req, user)
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
	database, _, srv := setupTestStack(t)
	user := seedUser(t, database, "user@example.edu")
	seedPrivilegeGrant(t, database, user)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/user/privileges", nil)
	req = asCaller(req, user)
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
	if len(body.Privileges) != 1 || body.Privileges[0] != string(models.PrivilegeGrant) {
		t.Errorf("privileges: got %v, want [%s]", body.Privileges, models.PrivilegeGrant)
	}
}

func TestRequirePrivilege_NoGrants_403(t *testing.T) {
	database, _, srv := setupTestStack(t)
	user := seedUser(t, database, "user@example.edu")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/privileges/catalog", nil)
	req = asCaller(req, user)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403", rr.Code)
	}
}

func TestRequirePrivilege_WithGrant_200(t *testing.T) {
	database, _, srv := setupTestStack(t)
	user := seedUser(t, database, "user@example.edu")
	seedPrivilegeGrant(t, database, user)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/privileges/catalog", nil)
	req = asCaller(req, user)
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
	seedPrivilegeGrant(t, database, granter)

	body, _ := json.Marshal(map[string]any{"privilege": "amie:read", "reason": "ops view"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users/"+target+"/privileges", bytes.NewReader(body))
	req = asCaller(req, granter)
	req.Header.Set("Content-Type", "application/json")
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status: got %d, want 201, body=%s", rr.Code, rr.Body.String())
	}
	has, err := svc.HasPrivilege(context.Background(), target, models.PrivilegeAMIERead)
	if err != nil || !has {
		t.Errorf("HasPrivilege after grant endpoint: has=%v err=%v", has, err)
	}
}

func TestGrantPrivilegeEndpoint_GranterWithoutMeta_403(t *testing.T) {
	database, _, srv := setupTestStack(t)
	plain := seedUser(t, database, "plain@example.edu")
	target := seedUser(t, database, "target@example.edu")
	body, _ := json.Marshal(map[string]any{"privilege": "amie:read"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users/"+target+"/privileges", bytes.NewReader(body))
	req = asCaller(req, plain)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403 (granter lacks privileges:grant)", rr.Code)
	}
}

func TestRevokePrivilegeEndpoint_HappyPath(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilegeGrant(t, database, granter)
	if _, err := svc.GrantPrivilege(context.Background(), target, models.PrivilegeAMIERead, granter, ""); err != nil {
		t.Fatalf("seed grant: %v", err)
	}
	body, _ := json.Marshal(map[string]any{"reason": "rotated"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/"+target+"/privileges/amie:read", bytes.NewReader(body))
	req = asCaller(req, granter)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status: got %d, want 204, body=%s", rr.Code, rr.Body.String())
	}
	if has, err := svc.HasPrivilege(context.Background(), target, models.PrivilegeAMIERead); err != nil || has {
		t.Errorf("HasPrivilege after revoke: has=%v err=%v", has, err)
	}
}

func TestRevokePrivilegeEndpoint_SelfRevokeMeta_400(t *testing.T) {
	database, _, srv := setupTestStack(t)
	user := seedUser(t, database, "user@example.edu")
	seedPrivilegeGrant(t, database, user)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/users/"+user+"/privileges/privileges:grant", nil)
	req = asCaller(req, user)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400 (self-revoke of meta)", rr.Code)
	}
}

func TestListUserPrivilegesEndpoint(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilegeGrant(t, database, granter)
	if _, err := svc.GrantPrivilege(context.Background(), target, models.PrivilegeAMIERead, granter, ""); err != nil {
		t.Fatalf("grant: %v", err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/"+target+"/privileges", nil)
	req = asCaller(req, granter)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200, body=%s", rr.Code, rr.Body.String())
	}
	var rows []models.UserPrivilege
	if err := json.NewDecoder(rr.Body).Decode(&rows); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(rows) != 1 || rows[0].Privilege != models.PrivilegeAMIERead {
		t.Errorf("rows: got %v, want [amie:read]", rows)
	}
}

func TestListPrivilegeHoldersEndpoint(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilegeGrant(t, database, granter)
	if _, err := svc.GrantPrivilege(context.Background(), target, models.PrivilegeAMIERead, granter, ""); err != nil {
		t.Fatalf("grant: %v", err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/privileges/amie:read/holders", nil)
	req = asCaller(req, granter)
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

func TestRequirePrivilege_StaleCacheStillReturns403AfterRevoke(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	a := seedUser(t, database, "a@example.edu")
	b := seedUser(t, database, "b@example.edu")
	seedPrivilegeGrant(t, database, a)
	seedPrivilegeGrant(t, database, b)

	// b warms the cache.
	warm := httptest.NewRecorder()
	warmReq := httptest.NewRequest(http.MethodGet, "/privileges/catalog", nil)
	warmReq = asCaller(warmReq, b)
	srv.ServeHTTP(warm, warmReq)
	if warm.Code != http.StatusOK {
		t.Fatalf("warm-up: got %d, want 200", warm.Code)
	}

	// a revokes b's privileges:grant via service (avoids the self-revoke
	// guard since a != b).
	if err := svc.RevokePrivilege(context.Background(), b, models.PrivilegeGrant, a, "rotated"); err != nil {
		t.Fatalf("revoke b's grant: %v", err)
	}
	// invalidate b's cache the way the HTTP handler would.
	srv.authCache.invalidate(b)

	// b retries and now gets 403.
	again := httptest.NewRecorder()
	againReq := httptest.NewRequest(http.MethodGet, "/privileges/catalog", nil)
	againReq = asCaller(againReq, b)
	srv.ServeHTTP(again, againReq)
	if again.Code != http.StatusForbidden {
		t.Errorf("post-revoke status for b: got %d, want 403", again.Code)
	}
}
