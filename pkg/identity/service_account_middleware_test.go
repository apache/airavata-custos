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

package identity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/airavata-custos/pkg/models"
)

type fakeServiceAccounts struct {
	account *models.ServiceAccount
	err     error
}

func (f fakeServiceAccounts) ResolveServiceAccount(context.Context, string) (*models.ServiceAccount, error) {
	return f.account, f.err
}

func accountWith(privileges ...models.PrivilegeKey) fakeServiceAccounts {
	return fakeServiceAccounts{account: &models.ServiceAccount{ID: "a1", Key: "tool", Privileges: privileges}}
}

// serve runs a service account call through ServiceAccountMiddleware. The router has one route of each kind,
// registered the way the server registers them.
func serve(accounts fakeServiceAccounts, method, path string) (int, string) {
	ok := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
	router := NewRouter(http.NewServeMux())
	router.RequirePrivilege("GET /clusters", models.ClustersWrite, ok)
	router.RequireAuth("GET /me/certificates", ok)
	router.RequireScoped("GET /projects/{id}", func(http.ResponseWriter, *http.Request) bool { return true }, ok)
	router.RequireAuthWithoutUserConsent("GET /me", ok)
	nonServiceAccountCalls := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	handler := ServiceAccountMiddleware(accounts, router, nonServiceAccountCalls)

	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+ServiceAccountTokenPrefix+"tool.secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var body struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec.Code, body.Code
}

// Make sure a service account can call a privileged route only when it has that privilege.
func TestServiceAccount_NeedsTheRoutePrivilege(t *testing.T) {
	if code, errorCode := serve(accountWith(models.ClustersWrite), http.MethodGet, "/clusters"); code != http.StatusOK {
		t.Errorf("service account with the privilege: got %d %s", code, errorCode)
	}
	if code, errorCode := serve(accountWith(models.ClustersRead), http.MethodGet, "/clusters"); code != http.StatusForbidden || errorCode != "insufficient_privilege" {
		t.Errorf("service account without the privilege: got %d %s", code, errorCode)
	}
}

// Make sure a service account can't use any route without a privilege, such as /me or a scoped route,
// since those work on a person's own data and a service account never acts as a person.
func TestServiceAccount_RefusedOnRoutesWithoutAPrivilege(t *testing.T) {
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/me/certificates"},
		{http.MethodGet, "/projects/p1"},
		{http.MethodGet, "/me"},
	} {
		if code, errorCode := serve(accountWith(models.ClustersWrite), route.method, route.path); code != http.StatusForbidden || errorCode != "service_account_not_allowed" {
			t.Errorf("%s %s: got %d %s", route.method, route.path, code, errorCode)
		}
	}
}

// Make sure an unknown token and a disabled service account are refused before any route runs.
func TestServiceAccountMiddleware_RefusesBadKeyAndDisabledAccount(t *testing.T) {
	if code, errorCode := serve(fakeServiceAccounts{err: ErrInvalidServiceAccountToken}, http.MethodGet, "/clusters"); code != http.StatusUnauthorized || errorCode != "invalid_token" {
		t.Errorf("unknown token: got %d %s", code, errorCode)
	}
	if code, errorCode := serve(fakeServiceAccounts{err: ErrServiceAccountDisabled}, http.MethodGet, "/clusters"); code != http.StatusForbidden || errorCode != "service_account_disabled" {
		t.Errorf("disabled service account: got %d %s", code, errorCode)
	}
}

// Make sure a call without a service account token, such as a person's sign-in token, skips the service account path.
func TestServiceAccountMiddleware_PassesOtherTokensOn(t *testing.T) {
	nonServiceAccountCalls := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	handler := ServiceAccountMiddleware(accountWith(models.ClustersWrite), http.NotFoundHandler(), nonServiceAccountCalls)
	req := httptest.NewRequest(http.MethodGet, "/clusters", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot {
		t.Fatalf("got %d", rec.Code)
	}
}
