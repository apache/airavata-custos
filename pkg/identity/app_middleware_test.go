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

type fakeApps struct {
	app *models.App
	err error
}

func (f fakeApps) ResolveApp(context.Context, string) (*models.App, error) { return f.app, f.err }

func appWith(privileges ...models.PrivilegeKey) fakeApps {
	return fakeApps{app: &models.App{ID: "a1", Key: "tool", Privileges: privileges}}
}

// serve runs an app call through AppMiddleware. The router has one route of each kind,
// registered the way the server registers them.
func serve(apps fakeApps, method, path string) (int, string) {
	ok := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
	router := NewRouter(http.NewServeMux())
	router.RequirePrivilege("GET /clusters", models.ClustersWrite, ok)
	router.RequireAuth("GET /me/certificates", ok)
	router.RequireScoped("GET /projects/{id}", func(http.ResponseWriter, *http.Request) bool { return true }, ok)
	router.RequireAuthWithoutUserConsent("GET /me", ok)
	nonAppCalls := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	handler := AppMiddleware(apps, router, nonAppCalls)

	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+AppTokenPrefix+"tool.secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var body struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec.Code, body.Code
}

// Make sure an app can call a privileged route only when it has that privilege.
func TestApp_NeedsTheRoutePrivilege(t *testing.T) {
	if code, errorCode := serve(appWith(models.ClustersWrite), http.MethodGet, "/clusters"); code != http.StatusOK {
		t.Errorf("app with the privilege: got %d %s", code, errorCode)
	}
	if code, errorCode := serve(appWith(models.ClustersRead), http.MethodGet, "/clusters"); code != http.StatusForbidden || errorCode != "insufficient_privilege" {
		t.Errorf("app without the privilege: got %d %s", code, errorCode)
	}
}

// Make sure an app can't use any route without a privilege, such as /me or a scoped route,
// since those work on a person's own data and an app never acts as a person.
func TestApp_RefusedOnRoutesWithoutAPrivilege(t *testing.T) {
	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/me/certificates"},
		{http.MethodGet, "/projects/p1"},
		{http.MethodGet, "/me"},
	} {
		if code, errorCode := serve(appWith(models.ClustersWrite), route.method, route.path); code != http.StatusForbidden || errorCode != "app_not_allowed" {
			t.Errorf("%s %s: got %d %s", route.method, route.path, code, errorCode)
		}
	}
}

// Make sure an unknown token and a disabled app are refused before any route runs.
func TestAppMiddleware_RefusesBadKeyAndDisabledApp(t *testing.T) {
	if code, errorCode := serve(fakeApps{err: ErrInvalidAppToken}, http.MethodGet, "/clusters"); code != http.StatusUnauthorized || errorCode != "invalid_token" {
		t.Errorf("unknown token: got %d %s", code, errorCode)
	}
	if code, errorCode := serve(fakeApps{err: ErrAppDisabled}, http.MethodGet, "/clusters"); code != http.StatusForbidden || errorCode != "app_disabled" {
		t.Errorf("disabled app: got %d %s", code, errorCode)
	}
}

// Make sure a call without an app token, such as a person's sign-in token, skips the app path.
func TestAppMiddleware_PassesOtherTokensOn(t *testing.T) {
	nonAppCalls := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	handler := AppMiddleware(appWith(models.ClustersWrite), http.NotFoundHandler(), nonAppCalls)
	req := httptest.NewRequest(http.MethodGet, "/clusters", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot {
		t.Fatalf("got %d", rec.Code)
	}
}
