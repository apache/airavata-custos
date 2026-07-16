//go:build e2e

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

// Package e2e drives the full access-request flow against a real local
// identity provider and database, with the identity-provisioner registry
// mocked in-process.
package e2e

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	comanage "github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/pkg/comanage"
	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/internal/server"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/service"
)

const (
	kcBase      = "http://localhost:8081"
	kcRealm     = "custos"
	e2eClientID = "custos-e2e"
	clusterID   = "00000000-0000-0000-0000-000000000001"
	seedDir     = "../../dev-ops/compose/seeds/pearc26"
	defaultDSN  = "admin:admin@tcp(127.0.0.1:3306)/custos_test?parseTime=true&multiStatements=true"
)

func TestAccessRequestE2E(t *testing.T) {
	os.Setenv("POSIX_USERNAME_PREFIX", "nexus")

	admin := kcAdminToken(t)
	database := openTestDB(t)
	resetAndSeed(t, database)
	ensureE2EClient(t, admin)

	mock := newRegistryMock(t)
	api := bootBackend(t, database, mockConnectorConfig(mock.srv.URL))

	suffix := fmt.Sprintf("%d", time.Now().UnixNano()%1e8)

	// --- user A: happy path ---
	userA := kcNewUser(t, admin, "e2e-a-"+suffix, "Trial", "User"+suffix)
	tokenA := ropcToken(t, userA.username)
	subA := jwtSub(t, tokenA)

	// b. unlinked identity is rejected on core routes
	status, body := call(t, api, http.MethodGet, "/me", tokenA, nil)
	if status != http.StatusUnauthorized || !strings.Contains(body, "identity_not_linked") {
		t.Fatalf("GET /me unlinked: want 401 identity_not_linked, got %d %s", status, body)
	}

	// c. event code resolves for a token-only caller
	status, body = call(t, api, http.MethodGet, "/access-requests/events/PEARC26", tokenA, nil)
	if status != http.StatusOK {
		t.Fatalf("GET event PEARC26: want 200, got %d %s", status, body)
	}
	var ev struct{ Code, Name string }
	mustDecode(t, body, &ev)
	if ev.Code != "PEARC26" || ev.Name != "pearc26-tutorial" {
		t.Fatalf("event payload: want {PEARC26 pearc26-tutorial}, got %+v", ev)
	}

	// d. unknown event code
	status, body = call(t, api, http.MethodGet, "/access-requests/events/NOPE", tokenA, nil)
	if status != http.StatusNotFound {
		t.Fatalf("GET event NOPE: want 404, got %d %s", status, body)
	}

	// e. submit + duplicate
	reqBody := map[string]string{"institution": "Test University", "event_code": "PEARC26", "reason": "workshop"}
	status, body = call(t, api, http.MethodPost, "/access-requests", tokenA, reqBody)
	if status != http.StatusCreated {
		t.Fatalf("POST access request: want 201, got %d %s", status, body)
	}
	var reqA accessRequest
	mustDecode(t, body, &reqA)
	if reqA.OIDCSub != subA || reqA.Status != "PENDING" {
		t.Fatalf("created request: want sub=%s status=PENDING, got %+v", subA, reqA)
	}
	status, body = call(t, api, http.MethodPost, "/access-requests", tokenA, reqBody)
	if status != http.StatusConflict {
		t.Fatalf("duplicate POST: want 409, got %d %s", status, body)
	}

	// f. own request shows PENDING
	status, body = call(t, api, http.MethodGet, "/access-requests/me", tokenA, nil)
	if status != http.StatusOK {
		t.Fatalf("GET /access-requests/me: want 200, got %d %s", status, body)
	}
	var mine accessRequest
	mustDecode(t, body, &mine)
	if mine.ID != reqA.ID || mine.Status != "PENDING" {
		t.Fatalf("own request: want id=%s PENDING, got %+v", reqA.ID, mine)
	}

	// g. approver: throwaway IdP user bound to the seeded approver account
	approver := kcNewUser(t, admin, "e2e-approver-"+suffix, "Workshop", "Approver"+suffix)
	tokenAppr := ropcToken(t, approver.username)
	subAppr := jwtSub(t, tokenAppr)
	bindIdentity(t, database, "pearc26-approver", subAppr, approver.email)

	status, body = call(t, api, http.MethodGet, "/access-requests?status=PENDING", tokenAppr, nil)
	if status != http.StatusOK {
		t.Fatalf("list PENDING: want 200, got %d %s", status, body)
	}
	if !strings.Contains(body, reqA.ID) {
		t.Fatalf("list PENDING does not contain request %s: %s", reqA.ID, body)
	}

	status, body = call(t, api, http.MethodPut, "/access-requests/"+reqA.ID, tokenAppr, map[string]string{"status": "APPROVED"})
	if status != http.StatusOK {
		t.Fatalf("PUT APPROVED: want 200, got %d %s", status, body)
	}
	var approved accessRequest
	mustDecode(t, body, &approved)
	if approved.Status != "APPROVED" || approved.CreatedUserID == "" || approved.ExpiresAt == nil {
		t.Fatalf("approved payload incomplete: %+v", approved)
	}

	// h. database state after approval
	assertApprovedState(t, database, reqA.ID, subA, userA.email)

	// i. registry mock received the person-create sequence and the account write
	username := clusterUsername(t, database, approved.CreatedUserID)
	if !strings.HasPrefix(username, "nexus-") {
		t.Fatalf("cluster username: want nexus- prefix, got %q", username)
	}
	mock.awaitAccountWrite(t, username)

	// j. own request shows APPROVED for the now-linked caller
	status, body = call(t, api, http.MethodGet, "/access-requests/me", tokenA, nil)
	if status != http.StatusOK || !strings.Contains(body, `"status":"APPROVED"`) {
		t.Fatalf("GET /access-requests/me after approve: want 200 APPROVED, got %d %s", status, body)
	}

	// k. user B: deny path
	userB := kcNewUser(t, admin, "e2e-b-"+suffix, "Denied", "User"+suffix)
	tokenB := ropcToken(t, userB.username)
	status, body = call(t, api, http.MethodPost, "/access-requests", tokenB, reqBody)
	if status != http.StatusCreated {
		t.Fatalf("POST request (B): want 201, got %d %s", status, body)
	}
	var reqB accessRequest
	mustDecode(t, body, &reqB)
	status, body = call(t, api, http.MethodPut, "/access-requests/"+reqB.ID, tokenAppr,
		map[string]string{"status": "DENIED", "deny_reason": "not eligible"})
	if status != http.StatusOK {
		t.Fatalf("PUT DENIED: want 200, got %d %s", status, body)
	}
	status, body = call(t, api, http.MethodGet, "/access-requests/me", tokenB, nil)
	if status != http.StatusOK || !strings.Contains(body, `"status":"DENIED"`) || !strings.Contains(body, "not eligible") {
		t.Fatalf("GET /access-requests/me (B): want DENIED + reason, got %d %s", status, body)
	}
	var nB int
	if err := database.Get(&nB, "SELECT COUNT(*) FROM users WHERE email = ?", userB.email); err != nil || nB != 0 {
		t.Fatalf("denied user must not be created: count=%d err=%v", nB, err)
	}

	// l. approve is PENDING-only
	status, body = call(t, api, http.MethodPut, "/access-requests/"+reqA.ID, tokenAppr, map[string]string{"status": "APPROVED"})
	if status < 400 || status >= 500 {
		t.Fatalf("re-approve: want 4xx, got %d %s", status, body)
	}
}

type accessRequest struct {
	ID            string     `json:"id"`
	OIDCSub       string     `json:"oidc_sub"`
	Status        string     `json:"status"`
	ApproverID    string     `json:"approver_id"`
	DenyReason    string     `json:"deny_reason"`
	ExpiresAt     *time.Time `json:"expires_at"`
	CreatedUserID string     `json:"created_user_id"`
}

// --- backend boot ---

func mockConnectorConfig(registryURL string) *config.ConnectorConfig {
	return &config.ConnectorConfig{
		Type:    "comanage-identity-provisioner",
		Enabled: true,
		Config: map[string]interface{}{
			"registry": map[string]interface{}{
				"url": registryURL, "co_id": 2, "api_user": "co_2.test", "api_key": "k",
			},
			"unix_cluster": map[string]interface{}{
				"id": 1, "person_id_type": "comanage_id",
			},
			"provisioning": map[string]interface{}{
				"custos_cluster_id": clusterID, "default_shell": "/bin/bash",
				"homedir_prefix": "/home/", "http_timeout": "10s",
			},
		},
	}
}

func bootBackend(t *testing.T, database *sqlx.DB, cc *config.ConnectorConfig) *httptest.Server {
	t.Helper()
	ctx := context.Background()
	bus := events.New()
	svc := service.New(database, bus)
	router := identity.NewRouter(http.NewServeMux())
	srv := server.New(svc, router)

	if err := comanage.LoadConnector(ctx, database, bus, svc, nil, router, cc); err != nil {
		t.Fatalf("load identity-provisioner connector: %v", err)
	}

	verifier, err := identity.NewJWTVerifier(ctx, kcBase+"/realms/"+kcRealm, e2eClientID)
	if err != nil {
		t.Fatalf("build verifier: %v", err)
	}
	handler := identity.Middleware(verifier, svc, router.PublicPaths(), router.TokenPathMatcher(), srv)
	api := httptest.NewServer(handler)
	t.Cleanup(api.Close)
	return api
}

// --- database ---

func openTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	dsn := os.Getenv("CORE_TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = defaultDSN
	}
	if !strings.Contains(dsn, "multiStatements=true") {
		sep := "?"
		if strings.Contains(dsn, "?") {
			sep = "&"
		}
		dsn += sep + "multiStatements=true"
	}
	database, err := db.Open(db.Config{DSN: dsn, MaxOpenConns: 5, MaxIdleConns: 2})
	if err != nil {
		t.Skipf("test database unreachable (%s): %v", dsn, err)
	}
	if err := database.Ping(); err != nil {
		t.Skipf("test database unreachable: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.MigrateEmbedded(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return database
}

// resetAndSeed wipes every core table this flow touches and re-applies the
// pearc26 seeds so runs are repeatable.
func resetAndSeed(t *testing.T, database *sqlx.DB) {
	t.Helper()
	tables := []string{
		"access_request_events", "access_requests", "access_events",
		"compute_allocation_memberships", "compute_cluster_users",
		"compute_allocation_resource_mappings", "compute_allocation_resource_rates",
		"compute_allocation_resources", "compute_allocations",
		"project_memberships", "projects",
		"user_roles", "role_privileges", "roles", "user_privileges",
		"audit_events", "user_identities", "users",
		"organizations", "compute_clusters",
	}
	if _, err := database.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		t.Fatalf("disable FK: %v", err)
	}
	for _, tbl := range tables {
		if _, err := database.Exec("TRUNCATE TABLE " + tbl); err != nil {
			t.Fatalf("truncate %s: %v", tbl, err)
		}
	}
	if _, err := database.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		t.Fatalf("re-enable FK: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(seedDir, "0*.sql"))
	if err != nil || len(files) == 0 {
		t.Fatalf("seed files not found in %s: %v", seedDir, err)
	}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read seed %s: %v", f, err)
		}
		if _, err := database.Exec(string(raw)); err != nil {
			t.Fatalf("apply seed %s: %v", f, err)
		}
	}
}

func bindIdentity(t *testing.T, database *sqlx.DB, userID, sub, email string) {
	t.Helper()
	if _, err := database.Exec(
		`INSERT INTO user_identities (id, user_id, source, external_id, email, oidc_sub, metadata)
		 VALUES (?, ?, 'oidc', ?, ?, ?, '')`,
		uuid.NewString(), userID, sub, email, sub,
	); err != nil {
		t.Fatalf("bind approver identity: %v", err)
	}
}

func clusterUsername(t *testing.T, database *sqlx.DB, userID string) string {
	t.Helper()
	var username string
	if err := database.Get(&username,
		"SELECT local_username FROM compute_cluster_users WHERE user_id = ? AND compute_cluster_id = ?",
		userID, clusterID,
	); err != nil {
		t.Fatalf("cluster user row for %s: %v", userID, err)
	}
	return username
}

func assertApprovedState(t *testing.T, database *sqlx.DB, requestID, sub, email string) {
	t.Helper()
	var req struct {
		Status        string     `db:"status"`
		ApproverID    string     `db:"approver_id"`
		ExpiresAt     *time.Time `db:"expires_at"`
		CreatedUserID string     `db:"created_user_id"`
	}
	if err := database.Get(&req,
		"SELECT status, approver_id, expires_at, created_user_id FROM access_requests WHERE id = ?", requestID,
	); err != nil {
		t.Fatalf("load approved request: %v", err)
	}
	if req.Status != "APPROVED" || req.ApproverID != "pearc26-approver" || req.CreatedUserID == "" {
		t.Fatalf("request row: want APPROVED by pearc26-approver with created user, got %+v", req)
	}
	if req.ExpiresAt == nil {
		t.Fatalf("request row: expires_at not set")
	}
	if d := time.Until(*req.ExpiresAt) - 30*24*time.Hour; d < -15*time.Minute || d > 15*time.Minute {
		t.Fatalf("expires_at: want ~now+30d, got %v", *req.ExpiresAt)
	}

	var eventTypes []string
	if err := database.Select(&eventTypes,
		"SELECT event_type FROM access_request_events WHERE access_request_id = ? ORDER BY timestamp", requestID,
	); err != nil {
		t.Fatalf("load request events: %v", err)
	}
	joined := strings.Join(eventTypes, ",")
	if len(eventTypes) < 2 || !strings.Contains(joined, "CREATED") || !strings.Contains(joined, "APPROVED") {
		t.Fatalf("request events: want CREATED+APPROVED, got %v", eventTypes)
	}

	var user struct {
		ID     string `db:"id"`
		OrgID  string `db:"organization_id"`
		Status string `db:"status"`
	}
	if err := database.Get(&user,
		"SELECT id, organization_id, status FROM users WHERE email = ?", email,
	); err != nil {
		t.Fatalf("load created user: %v", err)
	}
	if user.ID != req.CreatedUserID || user.OrgID != "pearc26-org" || user.Status != "ACTIVE" {
		t.Fatalf("created user: want %s ACTIVE in pearc26-org, got %+v", req.CreatedUserID, user)
	}

	var identUserID string
	if err := database.Get(&identUserID,
		"SELECT user_id FROM user_identities WHERE source = 'oidc' AND external_id = ?", sub,
	); err != nil || identUserID != user.ID {
		t.Fatalf("identity binding for sub %s: user=%q err=%v", sub, identUserID, err)
	}

	var mem struct {
		Status  string    `db:"membership_status"`
		EndTime time.Time `db:"end_time"`
	}
	if err := database.Get(&mem,
		`SELECT membership_status, end_time FROM compute_allocation_memberships
		 WHERE compute_allocation_id = 'pearc26-allocation' AND user_id = ?`, user.ID,
	); err != nil {
		t.Fatalf("load membership: %v", err)
	}
	if mem.Status != "ACTIVE" {
		t.Fatalf("membership: want ACTIVE, got %+v", mem)
	}
	if diff := mem.EndTime.Sub(*req.ExpiresAt); diff < -2*time.Second || diff > 2*time.Second {
		t.Fatalf("membership end_time %v != request expires_at %v", mem.EndTime, *req.ExpiresAt)
	}
}

// --- registry mock ---

type recordedReq struct {
	method, path, body string
}

type registryMock struct {
	mu   sync.Mutex
	reqs []recordedReq
	srv  *httptest.Server
}

// newRegistryMock serves the happy-path REST fixtures the identity
// provisioner walks through, recording every request for assertions.
func newRegistryMock(t *testing.T) *registryMock {
	t.Helper()
	m := &registryMock{}
	composite := `{
		"CoPerson":{"meta":{"id":42},"co_id":2,"status":"A"},
		"Name":[{"given":"Trial","family":"User","type":"official","primary_name":true}],
		"EmailAddress":[{"mail":"e2e@example.invalid","type":"official","verified":false}],
		"Identifier":[
			{"identifier":"Person100099","type":"comanage_id","login":false,"status":"A"},
			{"identifier":"2000099","type":"uidnumber","login":false,"status":"A"}
		]
	}`
	m.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		m.mu.Lock()
		m.reqs = append(m.reqs, recordedReq{r.Method, r.URL.Path, string(raw)})
		m.mu.Unlock()

		path := r.URL.Path
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/people"):
			_, _ = io.WriteString(w, `[{"identifier":"Person100099","type":"comanage_id","login":false,"status":"A"}]`)
		case r.Method == http.MethodGet && strings.Contains(path, "/people/"):
			_, _ = io.WriteString(w, composite)
		case r.Method == http.MethodPut && strings.Contains(path, "/people/"):
			_, _ = io.WriteString(w, composite)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/co_people.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"CoPeople","Version":"1.0","CoPeople":[]}`)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/co_groups.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"CoGroups","Version":"1.0","CoGroups":[]}`)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/co_groups.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"CoGroup","Id":"55"}`)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/identifiers.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"Identifiers","Version":"1.0","Identifiers":[]}`)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/identifiers.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"Identifier","Id":"7"}`)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/co_group_members.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"CoGroupMembers","Version":"1.0","CoGroupMembers":[]}`)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/co_group_members.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"CoGroupMember","Id":"11"}`)
		case r.Method == http.MethodGet && strings.HasSuffix(path, "/unix_cluster/unix_cluster_groups.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"UnixClusterGroups","Version":"1.0","UnixClusterGroups":[]}`)
		case r.Method == http.MethodPost && strings.HasSuffix(path, "/unix_cluster/unix_cluster_groups.json"):
			_, _ = io.WriteString(w, `{"ResponseType":"NewObject","Version":"1.0","ObjectType":"UnixClusterGroup","Id":"3"}`)
		default:
			t.Logf("registry mock: unexpected %s %s", r.Method, path)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(m.srv.Close)
	return m
}

// awaitAccountWrite polls for the async provisioning: a person create followed
// by the account PUT carrying the allocated username and homedir.
func (m *registryMock) awaitAccountWrite(t *testing.T, username string) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for {
		m.mu.Lock()
		var created, wrote bool
		for _, r := range m.reqs {
			if r.method == http.MethodPost && strings.HasSuffix(r.path, "/people") {
				created = true
			}
			if r.method == http.MethodPut && strings.Contains(r.path, "/people/") &&
				strings.Contains(r.body, `"username":"`+username+`"`) &&
				strings.Contains(r.body, `"home_directory":"/home/`+username+`"`) {
				wrote = true
			}
		}
		m.mu.Unlock()
		if created && wrote {
			return
		}
		if time.Now().After(deadline) {
			m.mu.Lock()
			defer m.mu.Unlock()
			for _, r := range m.reqs {
				t.Logf("registry mock saw: %s %s %s", r.method, r.path, r.body)
			}
			t.Fatalf("registry mock never received person create + account write for %s", username)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// --- identity provider (test infra) ---

type kcUser struct {
	id, username, email string
}

func kcAdminToken(t *testing.T) string {
	t.Helper()
	resp, err := http.PostForm(kcBase+"/realms/master/protocol/openid-connect/token", url.Values{
		"client_id": {"admin-cli"}, "grant_type": {"password"},
		"username": {"admin"}, "password": {"admin"},
	})
	if err != nil {
		t.Skipf("keycloak unreachable at %s: %v", kcBase, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Skipf("keycloak admin token: status %d", resp.StatusCode)
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || out.AccessToken == "" {
		t.Skipf("keycloak admin token decode: %v", err)
	}
	return out.AccessToken
}

func kcDo(t *testing.T, admin, method, path string, payload any) *http.Response {
	t.Helper()
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, kcBase+path, body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+admin)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

// ensureE2EClient creates the public direct-grant client with an audience
// mapper once; later runs find it by clientId and leave it alone.
func ensureE2EClient(t *testing.T, admin string) {
	t.Helper()
	resp := kcDo(t, admin, http.MethodGet, "/admin/realms/"+kcRealm+"/clients?clientId="+e2eClientID, nil)
	raw, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var existing []map[string]any
	if err := json.Unmarshal(raw, &existing); err != nil {
		t.Fatalf("list clients: %v (%s)", err, raw)
	}
	if len(existing) > 0 {
		return
	}
	resp = kcDo(t, admin, http.MethodPost, "/admin/realms/"+kcRealm+"/clients", map[string]any{
		"clientId": e2eClientID, "enabled": true, "publicClient": true,
		"directAccessGrantsEnabled": true, "standardFlowEnabled": false,
		"protocolMappers": []map[string]any{{
			"name": "aud-" + e2eClientID, "protocol": "openid-connect",
			"protocolMapper": "oidc-audience-mapper",
			"config": map[string]string{
				"included.client.audience": e2eClientID,
				"access.token.claim":       "true",
			},
		}},
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("create client: status %d %s", resp.StatusCode, raw)
	}
}

func kcNewUser(t *testing.T, admin, username, first, last string) kcUser {
	t.Helper()
	email := username + "@e2e.invalid"
	resp := kcDo(t, admin, http.MethodPost, "/admin/realms/"+kcRealm+"/users", map[string]any{
		"username": username, "email": email, "firstName": first, "lastName": last,
		"enabled": true, "emailVerified": true,
		"credentials": []map[string]any{{"type": "password", "value": "e2e-pass", "temporary": false}},
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("create idp user %s: status %d %s", username, resp.StatusCode, raw)
	}
	loc := resp.Header.Get("Location")
	id := loc[strings.LastIndex(loc, "/")+1:]
	t.Cleanup(func() {
		// Fresh admin token: the flow may outlive the original one's lifetime.
		del := kcDo(t, kcAdminToken(t), http.MethodDelete, "/admin/realms/"+kcRealm+"/users/"+id, nil)
		del.Body.Close()
	})
	return kcUser{id: id, username: username, email: email}
}

func ropcToken(t *testing.T, username string) string {
	t.Helper()
	resp, err := http.PostForm(kcBase+"/realms/"+kcRealm+"/protocol/openid-connect/token", url.Values{
		"client_id": {e2eClientID}, "grant_type": {"password"},
		"username": {username}, "password": {"e2e-pass"},
		"scope": {"openid profile email"},
	})
	if err != nil {
		t.Fatalf("token request for %s: %v", username, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("token for %s: status %d %s", username, resp.StatusCode, raw)
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || out.AccessToken == "" {
		t.Fatalf("token decode for %s: %v", username, err)
	}
	return out.AccessToken
}

func jwtSub(t *testing.T, token string) string {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("malformed jwt")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode jwt payload: %v", err)
	}
	var claims struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Sub == "" {
		t.Fatalf("jwt sub: %v", err)
	}
	return claims.Sub
}

// --- http helpers ---

func call(t *testing.T, api *httptest.Server, method, path, token string, payload any) (int, string) {
	t.Helper()
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, api.URL+path, body)
	if err != nil {
		t.Fatalf("build %s %s: %v", method, path, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(raw)
}

func mustDecode(t *testing.T, body string, out any) {
	t.Helper()
	if err := json.Unmarshal([]byte(body), out); err != nil {
		t.Fatalf("decode %q: %v", body, err)
	}
}
