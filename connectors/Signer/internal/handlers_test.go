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

package internal

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/internal/connectors/health"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

const (
	testClusterID = "cluster-1"
	testSub       = "sub-of-jdoe"
)

// signerCall is what the fake signer saw on its last request.
type signerCall struct {
	method, path, query, auth, body string
}

// mockSignerServer stands in for the signer. Health is always ok. Every other path gets the status and body the test gives it.
func mockSignerServer(t *testing.T, status int, body string) (*httptest.Server, *signerCall) {
	t.Helper()
	seen := &signerCall{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sent, _ := io.ReadAll(r.Body)
		*seen = signerCall{r.Method, r.URL.Path, r.URL.RawQuery, r.Header.Get("Authorization"), string(sent)}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/health" {
			w.Write([]byte(`{"status":"ok"}`))
			return
		}
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, seen
}

// fakeCore stands in for core. The zero value is a user with no account.
type fakeCore struct {
	account     *models.ComputeClusterUser
	memberships []models.ComputeAllocationMembership
	allocation  models.ComputeAllocation
	audited     []models.AuditEvent
}

func (f *fakeCore) CreateAuditEvent(_ context.Context, e *models.AuditEvent) (*models.AuditEvent, error) {
	f.audited = append(f.audited, *e)
	return e, nil
}

func (f *fakeCore) GetComputeClusterUserByPair(context.Context, string, string) (*models.ComputeClusterUser, error) {
	if f.account == nil {
		return nil, service.ErrNotFound
	}
	return f.account, nil
}

func (f *fakeCore) ListAllocationsForUser(context.Context, string) ([]models.ComputeAllocationMembership, error) {
	return f.memberships, nil
}

func (f *fakeCore) GetComputeAllocation(context.Context, string) (*models.ComputeAllocation, error) {
	return &f.allocation, nil
}

func (f *fakeCore) GetUser(_ context.Context, id string) (*models.User, error) {
	return &models.User{ID: id, Email: "admin@example.edu", FirstName: "Site", LastName: "Admin"}, nil
}

// readyCore is a user who can generate certificates
func readyCore() *fakeCore {
	provisioned := time.Now().Add(-time.Hour)
	start, end := time.Now().Add(-time.Hour), time.Now().Add(time.Hour)
	return &fakeCore{
		account:     &models.ComputeClusterUser{LocalUsername: "jdoe", ApprovalStatus: models.ClusterAccountApproved, ProvisionedAt: &provisioned},
		memberships: []models.ComputeAllocationMembership{{ComputeAllocationID: "a1", MembershipStatus: models.ACTIVE, StartTime: start, EndTime: end}},
		allocation:  models.ComputeAllocation{ComputeClusterID: testClusterID, Status: models.ACTIVE, StartTime: start, EndTime: end},
	}
}

func serveAs(caller *identity.Caller, claims *identity.Claims, core CoreAccounts, baseURL string, r *http.Request, privs ...models.PrivilegeKey) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	NewHandlers(Config{BaseURL: baseURL, ServiceToken: "service-token", ClusterID: testClusterID, TTLSeconds: 3600, Timeout: 2 * time.Second}, core).RegisterRoutes(identity.NewRouter(mux))
	r.Header.Set("Authorization", "Bearer user-token")
	ctx := identity.WithClaims(identity.WithCaller(r.Context(), caller), claims)
	r = r.WithContext(identity.WithPrivilegesForTest(ctx, privs))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, r)
	return rec
}

func serveWith(core CoreAccounts, baseURL string, r *http.Request, privs ...models.PrivilegeKey) *httptest.ResponseRecorder {
	return serveAs(&identity.Caller{UserID: "user-1", Email: "jdoe@example.edu"}, &identity.Claims{Sub: testSub, Email: "jdoe@example.edu"}, core, baseURL, r, privs...)
}

func serve(baseURL string, r *http.Request, privs ...models.PrivilegeKey) *httptest.ResponseRecorder {
	return serveWith(readyCore(), baseURL, r, privs...)
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body %s: %v", rec.Body, err)
	}
	return body.Error
}

// Make sure the signer gets the service token.
func TestCertificates_UsesServiceTokenAndPaging(t *testing.T) {
	srv, seen := mockSignerServer(t, http.StatusOK, `{"certificates":[],"next_cursor":"c2"}`)

	rec := serve(srv.URL, httptest.NewRequest(http.MethodGet, "/connectors/signer/certificates?limit=20&cursor=c1", nil), CertificatesRead)

	if seen.path != "/api/v1/certificates" || seen.query != "limit=20&cursor=c1" || seen.auth != "Bearer service-token" {
		t.Fatalf("signer got %+v", *seen)
	}
	if rec.Code != http.StatusOK || rec.Body.String() != `{"certificates":[],"next_cursor":"c2"}` {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

// Make sure the username comes from the cluster account and who it is for comes from the verified token.
func TestGenerate_SignsForTheCallersOwnAccount(t *testing.T) {
	srv, seen := mockSignerServer(t, http.StatusOK, `{"certificate":"ssh-ed25519-cert-v01@openssh.com AAAA"}`)

	body := `{"public_key":"ssh-ed25519 AAAA","principal":"root","user_email":"someone@example.edu","user_sub":"someone"}`
	rec := serve(srv.URL, httptest.NewRequest(http.MethodPost, "/connectors/signer/me/certificates", strings.NewReader(body)))

	if rec.Code != http.StatusOK || seen.path != "/api/v1/sign" {
		t.Fatalf("got %d, signer saw %+v", rec.Code, *seen)
	}
	var sent map[string]any
	if err := json.Unmarshal([]byte(seen.body), &sent); err != nil {
		t.Fatal(err)
	}
	if sent["principal"] != "jdoe" || sent["user_email"] != "jdoe@example.edu" || sent["user_sub"] != testSub || sent["ttl_seconds"] != float64(3600) {
		t.Fatalf("signer got %v", sent)
	}
	if sent["requested_by"] != "user/user-1" {
		t.Fatalf("requested_by = %v", sent["requested_by"])
	}
}

// Make sure each issued certificate leaves an audit row under its serial, naming who asked.
func TestGenerate_RecordsTheIssueInTheAuditLog(t *testing.T) {
	srv, _ := mockSignerServer(t, http.StatusOK, `{"serial_number":42}`)
	core := readyCore()

	serveWith(core, srv.URL, httptest.NewRequest(http.MethodPost, "/connectors/signer/me/certificates", strings.NewReader(`{"public_key":"k"}`)))

	if len(core.audited) != 1 {
		t.Fatalf("audited %d events", len(core.audited))
	}
	got := core.audited[0]
	if got.EventType != AuditCertificateIssued || got.EntityID != "42" || !strings.Contains(got.Details, `"requested_by":"user/user-1"`) {
		t.Fatalf("audited %+v", got)
	}
}

// Make sure no audit row is written when the signer refuses to issue the certificate.
func TestGenerate_SignerRefusalIsNotAudited(t *testing.T) {
	srv, _ := mockSignerServer(t, http.StatusForbidden, `{"error":"principal_denied"}`)
	core := readyCore()

	rec := serveWith(core, srv.URL, httptest.NewRequest(http.MethodPost, "/connectors/signer/me/certificates", strings.NewReader(`{"public_key":"k"}`)))

	if rec.Code != http.StatusForbidden || len(core.audited) != 0 {
		t.Fatalf("got %d, audited %+v", rec.Code, core.audited)
	}
}

// Only with an active account and allocation can generate certificates.
func TestGenerate_RefusesWithoutAccountOrAllocation(t *testing.T) {
	expired := func(core *fakeCore) { core.memberships[0].EndTime = time.Now().Add(-time.Minute) }
	cases := []struct {
		name   string
		change func(*fakeCore)
		code   string
	}{
		{"no cluster account", func(c *fakeCore) { c.account = nil }, "no_cluster_account"},
		{"account not approved", func(c *fakeCore) { c.account.ApprovalStatus = models.ClusterAccountPending }, "cluster_account_not_ready"},
		{"account not provisioned", func(c *fakeCore) { c.account.ProvisionedAt = nil }, "cluster_account_not_ready"},
		{"no memberships", func(c *fakeCore) { c.memberships = nil }, "no_active_allocation"},
		{"membership expired", expired, "no_active_allocation"},
		{"allocation on another cluster", func(c *fakeCore) { c.allocation.ComputeClusterID = "other" }, "no_active_allocation"},
		{"allocation inactive", func(c *fakeCore) { c.allocation.Status = models.INACTIVE }, "no_active_allocation"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, seen := mockSignerServer(t, http.StatusOK, `{}`)
			core := readyCore()
			tc.change(core)

			rec := serveWith(core, srv.URL, httptest.NewRequest(http.MethodPost, "/connectors/signer/me/certificates", strings.NewReader(`{"public_key":"k"}`)))

			if rec.Code != http.StatusForbidden || errorCode(t, rec) != tc.code || seen.path != "" {
				t.Fatalf("got %d %s, signer saw %q", rec.Code, rec.Body, seen.path)
			}
		})
	}
}

// Make sure the own list is always filtered to the caller.
func TestListOwn_AlwaysFiltersToTheCaller(t *testing.T) {
	srv, seen := mockSignerServer(t, http.StatusOK, `{"certificates":[]}`)

	serve(srv.URL, httptest.NewRequest(http.MethodGet, "/connectors/signer/me/certificates?user_sub=someone-else&limit=5", nil))

	if seen.query != "limit=5&user_sub="+testSub {
		t.Fatalf("signer got query %q", seen.query)
	}
}

// Make sure a user cannot see or revoke someone else's certificate and gets the same response when there's none existing.
func TestOwnRoutes_HideOtherUsersCertificates(t *testing.T) {
	requests := map[string]*http.Request{
		"get":    httptest.NewRequest(http.MethodGet, "/connectors/signer/me/certificates/42", nil),
		"revoke": httptest.NewRequest(http.MethodPost, "/connectors/signer/me/certificates/42/revoke", strings.NewReader(`{}`)),
	}
	for name, r := range requests {
		t.Run(name, func(t *testing.T) {
			srv, seen := mockSignerServer(t, http.StatusOK, `{"user_sub":"someone-else"}`)

			rec := serve(srv.URL, r)

			if rec.Code != http.StatusNotFound || seen.method != http.MethodGet {
				t.Fatalf("got %d, last signer call %+v", rec.Code, *seen)
			}
		})
	}
}

// Make sure a user can revoke their own certificate with no privilege, recorded under their user ID.
func TestRevokeOwn_RecordedAsTheUser(t *testing.T) {
	srv, seen := mockSignerServer(t, http.StatusOK, `{"user_sub":"`+testSub+`"}`)

	rec := serve(srv.URL, httptest.NewRequest(http.MethodPost, "/connectors/signer/me/certificates/42/revoke", strings.NewReader(`{}`)))

	if rec.Code != http.StatusOK || seen.path != "/api/v1/certificates/42/revoke" || !strings.Contains(seen.body, `"revoked_by":"user/user-1"`) {
		t.Fatalf("got %d, signer saw %+v", rec.Code, *seen)
	}
}

// Make sure an admin revoke is recorded under the admin's user ID.
func TestRevokeAny_RecordedAsTheAdmin(t *testing.T) {
	srv, seen := mockSignerServer(t, http.StatusOK, `{}`)

	serve(srv.URL, httptest.NewRequest(http.MethodPost, "/connectors/signer/certificates/42/revoke", strings.NewReader(`{"reason":"lost laptop"}`)), CertificatesWrite)

	if !strings.Contains(seen.body, `"revoked_by":"user/user-1"`) || !strings.Contains(seen.body, `"reason":"lost laptop"`) {
		t.Fatalf("signer got %s", seen.body)
	}
}

// Make sure a caller without the write privilege cannot revoke, and the signer is never called.
func TestRevoke_WithoutWritePrivilegeNeverReachesSigner(t *testing.T) {
	srv, seen := mockSignerServer(t, http.StatusOK, `{}`)

	rec := serve(srv.URL, httptest.NewRequest(http.MethodPost, "/connectors/signer/certificates/42/revoke", nil), CertificatesRead)

	if rec.Code != http.StatusForbidden || seen.path != "" {
		t.Fatalf("got %d, signer saw %q", rec.Code, seen.path)
	}
}

// Make sure a certificate reader can see who revoked a certificate, getting the email and nothing else about that user.
func TestUserEmail_ShowsOnlyTheEmail(t *testing.T) {
	rec := serve("", httptest.NewRequest(http.MethodGet, "/connectors/signer/users/u-9", nil), CertificatesRead)

	if rec.Code != http.StatusOK || rec.Body.String() != `{"email":"admin@example.edu","id":"u-9"}`+"\n" {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

// Make sure the status check calls the signer's health route without the caller's token.
func TestStatus_ReportsReachable(t *testing.T) {
	srv, seen := mockSignerServer(t, http.StatusOK, `{}`)

	rec := serve(srv.URL, httptest.NewRequest(http.MethodGet, "/connectors/signer/status", nil), CertificatesRead)

	var got statusResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !got.Reachable {
		t.Fatalf("got %+v", got)
	}
	if seen.path != "/api/v1/health" || seen.auth != "" {
		t.Fatalf("health request was %+v", *seen)
	}
}

// Make sure a 5xx from the signer shows as a failed check on GET /connectors.
func TestForward_SignerServerErrorRecordsFailedCheck(t *testing.T) {
	srv, _ := mockSignerServer(t, http.StatusInternalServerError, `{}`)

	serve(srv.URL, httptest.NewRequest(http.MethodGet, "/connectors/signer/me/certificates", nil))

	checks := health.For(Type)
	if len(checks) != 1 || checks[0].OK {
		t.Fatalf("got %+v", checks)
	}
}

// Make sure a wrong service token shows as a failed check and the caller sees signer_unavailable.
func TestForward_RefusedServiceTokenIsAnOutage(t *testing.T) {
	srv, _ := mockSignerServer(t, http.StatusUnauthorized, `{"error":"unauthorized"}`)

	rec := serve(srv.URL, httptest.NewRequest(http.MethodGet, "/connectors/signer/me/certificates", nil))

	checks := health.For(Type)
	if rec.Code != http.StatusBadGateway || errorCode(t, rec) != "signer_unavailable" || len(checks) != 1 || checks[0].OK {
		t.Fatalf("got %d %s, checks %+v", rec.Code, rec.Body, checks)
	}
}

// Make sure a signer that is down does not show its internal address to the caller.
func TestForward_SignerDownHidesItsAddress(t *testing.T) {
	srv, _ := mockSignerServer(t, http.StatusOK, `{}`)
	srv.Close()

	rec := serve(srv.URL, httptest.NewRequest(http.MethodGet, "/connectors/signer/me/certificates", nil))

	if rec.Code != http.StatusBadGateway || errorCode(t, rec) != "signer_unavailable" || strings.Contains(rec.Body.String(), srv.Listener.Addr().String()) {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
	}
}

// Make sure a signer that is down shows as not reachable instead of failing the request.
func TestStatus_SignerDownIsNotReachable(t *testing.T) {
	srv, _ := mockSignerServer(t, http.StatusOK, `{}`)
	srv.Close()

	rec := serve(srv.URL, httptest.NewRequest(http.MethodGet, "/connectors/signer/status", nil), CertificatesRead)

	body, _ := io.ReadAll(rec.Body)
	var got statusResponse
	if err := json.Unmarshal(body, &got); err != nil || rec.Code != http.StatusOK || got.Reachable {
		t.Fatalf("got %d %s", rec.Code, body)
	}
}
