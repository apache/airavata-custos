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
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type accessStatusFixture struct {
	db      *sqlx.DB
	svc     *service.Service
	srv     *Server
	allocID string
	pi      string
	member  string
}

func setupAccessStatusFixture(t *testing.T) *accessStatusFixture {
	t.Helper()
	database, svc, srv := setupTestStack(t)
	pi := seedUser(t, database, "pi-"+uuid.NewString()[:8]+"@example.edu")
	member := seedUser(t, database, "member-"+uuid.NewString()[:8]+"@example.edu")

	clusterID, projectID, allocID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	mustExec := func(q string, args ...any) {
		t.Helper()
		if _, err := database.Exec(q, args...); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	mustExec(`INSERT INTO compute_clusters (id, name) VALUES (?, ?)`, clusterID, "cl-"+clusterID[:8])
	mustExec(`INSERT INTO projects (id, originated_id, title, origination, project_pi_id, status)
	          VALUES (?, ?, 'Access Status Project', 'TEST', ?, 'ACTIVE')`, projectID, uuid.NewString(), pi)
	mustExec(`INSERT INTO project_memberships (project_id, user_id, role, added_time)
	          VALUES (?, ?, 'PI', NOW(6))`, projectID, pi)
	mustExec(`INSERT INTO compute_allocations
	              (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time)
	          VALUES (?, ?, 'alloc', 'ACTIVE', ?, 1000, NOW(6), DATE_ADD(NOW(6), INTERVAL 30 DAY))`,
		allocID, projectID, clusterID)
	mustExec(`INSERT INTO compute_allocation_memberships
	              (id, compute_allocation_id, user_id, start_time, end_time, membership_status)
	          VALUES (?, ?, ?, NOW(6), DATE_ADD(NOW(6), INTERVAL 30 DAY), 'ACTIVE')`,
		uuid.NewString(), allocID, member)

	return &accessStatusFixture{db: database, svc: svc, srv: srv, allocID: allocID, pi: pi, member: member}
}

func (fx *accessStatusFixture) get(t *testing.T, path, callerID string, privs ...models.PrivilegeKey) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if callerID != "" {
		req = withTestCaller(req, callerID, privs...)
	}
	fx.srv.router.ServeHTTP(rr, req)
	return rr
}

func TestAccessStatus_AuthAndScope(t *testing.T) {
	fx := setupAccessStatusFixture(t)
	path := "/compute-allocations/" + fx.allocID + "/access-status"

	if rr := fx.get(t, path, ""); rr.Code != http.StatusUnauthorized {
		t.Errorf("no caller: got %d, want 401", rr.Code)
	}
	stranger := seedUser(t, fx.db, "stranger-"+uuid.NewString()[:8]+"@example.edu")
	if rr := fx.get(t, path, stranger); rr.Code != http.StatusNotFound {
		t.Errorf("non-member: got %d, want 404", rr.Code)
	}
	if rr := fx.get(t, "/compute-allocations/"+uuid.NewString()+"/access-status", fx.member); rr.Code != http.StatusNotFound {
		t.Errorf("missing allocation: got %d, want 404", rr.Code)
	}
}

func TestAccessStatus_SelfViewWithDerivedAndRealChecks(t *testing.T) {
	fx := setupAccessStatusFixture(t)
	if err := fx.svc.RecordAccessCheckResult(t.Context(), fx.allocID, fx.member, models.AccessCheckJobSubmission, true, ""); err != nil {
		t.Fatalf("record: %v", err)
	}

	rr := fx.get(t, "/compute-allocations/"+fx.allocID+"/access-status", fx.member)
	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, body=%s", rr.Code, rr.Body.String())
	}
	var got struct {
		Checks []struct {
			Type    string `json:"type"`
			Status  string `json:"status"`
			UIState string `json:"ui_state"`
		} `json:"checks"`
		Events []struct {
			CheckType string `json:"check_type"`
			EventType string `json:"event_type"`
		} `json:"events"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got.Checks) != 2 {
		t.Fatalf("checks: got %d, want 2", len(got.Checks))
	}
	states := map[string]string{}
	for _, c := range got.Checks {
		states[c.Type] = c.UIState
	}
	if states["JOB_SUBMISSION"] != "ok" || states["SIGN_IN"] != "setting_up" {
		t.Errorf("ui states: got %v", states)
	}
	if len(got.Events) != 2 {
		t.Fatalf("events: got %d, want STARTED+ONLINE", len(got.Events))
	}
	for _, e := range got.Events {
		if e.CheckType != "JOB_SUBMISSION" {
			t.Errorf("event check type: got %+v", e)
		}
	}
}

func TestAccessStatus_StuckDerivation(t *testing.T) {
	fx := setupAccessStatusFixture(t)
	rec := func(ok bool) {
		t.Helper()
		if err := fx.svc.RecordAccessCheckResult(t.Context(), fx.allocID, fx.member, models.AccessCheckJobSubmission, ok, "x"); err != nil {
			t.Fatalf("record: %v", err)
		}
	}
	rec(true)
	rec(false)

	rr := fx.get(t, "/compute-allocations/"+fx.allocID+"/access-status", fx.member)
	body := rr.Body.String()
	if rr.Code != http.StatusOK || !jsonHasUIState(t, body, "JOB_SUBMISSION", "retrying") {
		t.Fatalf("fresh failure: want retrying, got %s", body)
	}

	// Backdate the episode past the default threshold: it derives as stuck.
	if _, err := fx.db.Exec(
		`UPDATE access_checks SET failing_since = ? WHERE compute_allocation_id = ? AND user_id = ?`,
		time.Now().UTC().Add(-2*time.Hour), fx.allocID, fx.member); err != nil {
		t.Fatalf("backdate: %v", err)
	}
	rr = fx.get(t, "/compute-allocations/"+fx.allocID+"/access-status", fx.member)
	body = rr.Body.String()
	if !jsonHasUIState(t, body, "JOB_SUBMISSION", "stuck") {
		t.Fatalf("old failure: want stuck, got %s", body)
	}
}

func jsonHasUIState(t *testing.T, body, checkType, uiState string) bool {
	t.Helper()
	var got struct {
		Checks []struct {
			Type    string `json:"type"`
			UIState string `json:"ui_state"`
		} `json:"checks"`
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, c := range got.Checks {
		if c.Type == checkType {
			return c.UIState == uiState
		}
	}
	return false
}

func TestAccessStatus_MembersParam(t *testing.T) {
	fx := setupAccessStatusFixture(t)
	path := "/compute-allocations/" + fx.allocID + "/access-status?members=true"

	decodeMembers := func(rr *httptest.ResponseRecorder) (membersLen int, hasMembersKey bool) {
		t.Helper()
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(rr.Body.Bytes(), &raw); err != nil {
			t.Fatalf("decode: %v", err)
		}
		m, ok := raw["members"]
		if !ok {
			return 0, false
		}
		var members []json.RawMessage
		if err := json.Unmarshal(m, &members); err != nil {
			t.Fatalf("decode members: %v", err)
		}
		return len(members), true
	}

	// Manager (PI) gets the grid with the one ACTIVE member.
	rr := fx.get(t, path, fx.pi)
	if rr.Code != http.StatusOK {
		t.Fatalf("pi members: got %d, body=%s", rr.Code, rr.Body.String())
	}
	if n, ok := decodeMembers(rr); !ok || n != 1 {
		t.Errorf("pi members view: got n=%d hasKey=%v, want 1 member", n, ok)
	}

	// Read-privileged staff (non-member) also gets the grid.
	staff := seedUser(t, fx.db, "staff-"+uuid.NewString()[:8]+"@example.edu")
	rr = fx.get(t, path, staff, models.AllocationsRead)
	if rr.Code != http.StatusOK {
		t.Fatalf("staff members: got %d", rr.Code)
	}
	if n, ok := decodeMembers(rr); !ok || n != 1 {
		t.Errorf("staff members view: got n=%d hasKey=%v, want 1 member", n, ok)
	}

	// A plain member silently degrades to the self view.
	rr = fx.get(t, path, fx.member)
	if rr.Code != http.StatusOK {
		t.Fatalf("member members=true: got %d", rr.Code)
	}
	if _, ok := decodeMembers(rr); ok {
		t.Errorf("member must degrade to self view, got members key: %s", rr.Body.String())
	}
}
