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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

const appCertificates = "/connectors/signer/app/certificates"

func serveAsApp(core CoreAccounts, baseURL string, r *http.Request, actions ...string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	NewHandlers(Config{BaseURL: baseURL, ServiceToken: "service-token", ClusterID: testClusterID, TTLSeconds: 3600, Timeout: 2 * time.Second}, core).RegisterRoutes(identity.NewRouter(mux))
	r = r.WithContext(identity.WithCaller(r.Context(), &identity.Caller{AppID: "app-uuid", AppKey: "test-gateway", AppActions: actions}))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, r)
	return rec
}

// Make sure the certificate is for the named user's own cluster account and says which app asked, whatever else the app sends.
func TestGenerateForApp_SignsForTheNamedUser(t *testing.T) {
	srv, seen := mockSignerServer(t, http.StatusOK, `{}`)

	body := `{"user_email":"jdoe@example.edu","public_key":"ssh-ed25519 AAAA","source_address":"203.0.113.0/24","principal":"root","requested_by":"user/someone"}`
	rec := serveAsApp(readyCore(), srv.URL, httptest.NewRequest(http.MethodPost, appCertificates, strings.NewReader(body)), AppActionGenerateCertificate)

	if rec.Code != http.StatusOK || seen.path != "/api/v1/sign" {
		t.Fatalf("got %d, signer saw %+v", rec.Code, *seen)
	}
	var sent map[string]any
	if err := json.Unmarshal([]byte(seen.body), &sent); err != nil {
		t.Fatal(err)
	}
	if sent["principal"] != "jdoe" || sent["user_email"] != "jdoe@example.edu" || sent["user_sub"] != testSub {
		t.Fatalf("signer got %v", sent)
	}
	if sent["requested_by"] != "app/test-gateway" || sent["source_address"] != "203.0.113.0/24" {
		t.Fatalf("signer got %v", sent)
	}
}

// Make sure an app cannot get a certificate for a user who could not get one themselves.
func TestGenerateForApp_RefusesUserWhoCannotHaveACertificate(t *testing.T) {
	cases := []struct {
		name   string
		change func(*fakeCore)
		status int
		code   string
	}{
		{"unknown email", func(c *fakeCore) { c.user = nil }, http.StatusNotFound, "user_not_found"},
		{"suspended user", func(c *fakeCore) { c.user.Status = models.UserSuspended }, http.StatusForbidden, "user_not_active"},
		{"pending user agreement", func(c *fakeCore) { c.pending = []models.UserAgreement{{ID: "a1"}} }, http.StatusForbidden, "consent_required"},
		{"never signed in", func(c *fakeCore) { c.identities = nil }, http.StatusForbidden, "user_not_signed_in"},
		{"app not allowed by the user", func(c *fakeCore) { c.appAuthorized = false }, http.StatusForbidden, "app_not_authorized"},
		{"no cluster account", func(c *fakeCore) { c.account = nil }, http.StatusForbidden, "no_cluster_account"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, seen := mockSignerServer(t, http.StatusOK, `{}`)
			core := readyCore()
			tc.change(core)

			body := `{"user_email":"jdoe@example.edu","public_key":"k"}`
			rec := serveAsApp(core, srv.URL, httptest.NewRequest(http.MethodPost, appCertificates, strings.NewReader(body)), AppActionGenerateCertificate)

			if rec.Code != tc.status || errorCode(t, rec) != tc.code || seen.path != "" {
				t.Fatalf("got %d %s, signer saw %q", rec.Code, rec.Body, seen.path)
			}
		})
	}
}

// Make sure an app's list only ever holds what it requested, whatever filter it sends.
func TestListForApp_AlwaysFiltersToTheApp(t *testing.T) {
	srv, seen := mockSignerServer(t, http.StatusOK, `{"certificates":[]}`)

	serveAsApp(readyCore(), srv.URL, httptest.NewRequest(http.MethodGet, appCertificates+"?requested_by=app/other&limit=5", nil), AppActionGenerateCertificate)

	if seen.query != "limit=5&requested_by=app%2Ftest-gateway" {
		t.Fatalf("signer got query %q", seen.query)
	}
}

// Make sure an app cannot see or revoke a certificate that another app or a user requested.
func TestAppRoutes_HideCertificatesTheAppDidNotRequest(t *testing.T) {
	for _, requester := range []string{"app/other", "user/user-1"} {
		srv, seen := mockSignerServer(t, http.StatusOK, `{"requested_by":"`+requester+`"}`)

		get := serveAsApp(readyCore(), srv.URL, httptest.NewRequest(http.MethodGet, appCertificates+"/42", nil), AppActionGenerateCertificate)
		revoke := serveAsApp(readyCore(), srv.URL, httptest.NewRequest(http.MethodPost, appCertificates+"/42/revoke", strings.NewReader(`{}`)), AppActionRevokeCertificate)

		if get.Code != http.StatusNotFound || revoke.Code != http.StatusNotFound || strings.HasSuffix(seen.path, "/revoke") {
			t.Fatalf("%s: get %d, revoke %d, signer saw %q", requester, get.Code, revoke.Code, seen.path)
		}
	}
}

// Make sure a revoke by an app is recorded under the app.
func TestRevokeForApp_RecordedAsTheApp(t *testing.T) {
	srv, seen := mockSignerServer(t, http.StatusOK, `{"requested_by":"app/test-gateway"}`)

	serveAsApp(readyCore(), srv.URL, httptest.NewRequest(http.MethodPost, appCertificates+"/42/revoke", strings.NewReader(`{"reason":"rotated"}`)), AppActionRevokeCertificate)

	if !strings.Contains(seen.body, `"revoked_by":"app/test-gateway"`) || !strings.HasSuffix(seen.path, "/revoke") {
		t.Fatalf("signer saw %+v", *seen)
	}
}

// Make sure an app only with the revoke action cannot generate a certificate.
func TestAppRoutes_NeedTheirOwnAction(t *testing.T) {
	srv, seen := mockSignerServer(t, http.StatusOK, `{}`)

	rec := serveAsApp(readyCore(), srv.URL, httptest.NewRequest(http.MethodPost, appCertificates, strings.NewReader(`{"public_key":"k"}`)), AppActionRevokeCertificate)

	if rec.Code != http.StatusForbidden || seen.path != "" {
		t.Fatalf("got %d, signer saw %q", rec.Code, seen.path)
	}
}
