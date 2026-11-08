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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/internal/connectors/health"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

// signerCall is what the fake signer saw on its last request.
type signerCall struct {
	path, query, auth string
}

// mockSignerServer stands in for the signer. Health is always ok. Every other path gets the status and body the test gives it.
func mockSignerServer(t *testing.T, status int, body string) (*httptest.Server, *signerCall) {
	t.Helper()
	seen := &signerCall{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*seen = signerCall{r.URL.Path, r.URL.RawQuery, r.Header.Get("Authorization")}
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

func serve(baseURL string, r *http.Request, privs ...models.PrivilegeKey) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	NewHandlers(baseURL, 2*time.Second).RegisterRoutes(identity.NewRouter(mux))
	r.Header.Set("Authorization", "Bearer user-token")
	r = r.WithContext(identity.WithPrivilegesForTest(r.Context(), privs))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, r)
	return rec
}

// Make sure the caller's token and paging go to the signer, and the signer's answer comes back unchanged.
func TestCertificates_ForwardsCallerTokenAndPaging(t *testing.T) {
	srv, seen := mockSignerServer(t, http.StatusOK, `{"items":[],"next_cursor":"c2"}`)

	rec := serve(srv.URL, httptest.NewRequest(http.MethodGet, "/connectors/signer/certificates?limit=20&cursor=c1", nil), CertificatesRead)

	if seen.path != "/api/v1/admin/certificates" || seen.query != "limit=20&cursor=c1" || seen.auth != "Bearer user-token" {
		t.Fatalf("signer got %+v", *seen)
	}
	if rec.Code != http.StatusOK || rec.Body.String() != `{"items":[],"next_cursor":"c2"}` {
		t.Fatalf("got %d %s", rec.Code, rec.Body)
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

// Make sure the status check calls the signer's health route without the caller's token.
func TestStatus_ReportsReachable(t *testing.T) {
	srv, seen := mockSignerServer(t, http.StatusOK, `{}`)

	rec := serve(srv.URL, httptest.NewRequest(http.MethodGet, "/connectors/signer/status", nil), CertificatesRead)

	var got statusResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !got.Reachable || got.BaseURL != srv.URL {
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

// Make sure a signer that is down does not show its internal address to the caller.
func TestForward_SignerDownHidesItsAddress(t *testing.T) {
	srv, _ := mockSignerServer(t, http.StatusOK, `{}`)
	srv.Close()

	rec := serve(srv.URL, httptest.NewRequest(http.MethodGet, "/connectors/signer/me/certificates", nil))

	if rec.Code != http.StatusBadGateway || strings.Contains(rec.Body.String(), srv.Listener.Addr().String()) {
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
