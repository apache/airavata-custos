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
)

type fakeApps struct {
	caller *Caller
	err    error
}

func (f fakeApps) ResolveApp(context.Context, string) (*Caller, error) { return f.caller, f.err }

// serveWith sends a request through AppMiddleware to a router with one app route
// and one signed-in route. Non-app tokens arrive as a signed-in user.
func serveWith(apps fakeApps, token, path string) (int, string) {
	ok := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
	router := NewRouter(http.NewServeMux())
	router.RequireApp("GET /app-route", "signer:app:generate", ok)
	router.RequireAuth("GET /me", ok)
	users := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		router.ServeHTTP(w, req.WithContext(WithCaller(req.Context(), &Caller{UserID: "u1"})))
	})
	handler := AppMiddleware(apps, router.AppRoutes(), router, users)

	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var body struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec.Code, body.Code
}

// Make sure an app token cannot reach a route that was not opened to apps, even one that needs no privilege.
func TestAppMiddleware_RefusesRouteNotOpenToApps(t *testing.T) {
	apps := fakeApps{caller: &Caller{AppID: "a1", AppActions: []string{"signer:app:generate"}}}
	code, errorCode := serveWith(apps, AppTokenPrefix+"key", "/me")
	if code != http.StatusForbidden || errorCode != "app_route_not_allowed" {
		t.Fatalf("got %d %s", code, errorCode)
	}
}

// Make sure an app reaches an app route only when it holds the action that route needs.
func TestRequireApp_NeedsTheRouteAction(t *testing.T) {
	allowed := fakeApps{caller: &Caller{AppID: "a1", AppActions: []string{"signer:app:generate"}}}
	if code, errorCode := serveWith(allowed, AppTokenPrefix+"key", "/app-route"); code != http.StatusOK {
		t.Errorf("app with the action: got %d %s", code, errorCode)
	}
	other := fakeApps{caller: &Caller{AppID: "a1", AppActions: []string{"signer:app:revoke"}}}
	if code, errorCode := serveWith(other, AppTokenPrefix+"key", "/app-route"); code != http.StatusForbidden || errorCode != "app_action_not_allowed" {
		t.Errorf("app without the action: got %d %s", code, errorCode)
	}
}

// Make sure a signed-in user cannot call an app route, whatever they hold.
func TestRequireApp_RefusesUser(t *testing.T) {
	code, errorCode := serveWith(fakeApps{}, "user-token", "/app-route")
	if code != http.StatusForbidden || errorCode != "app_action_not_allowed" {
		t.Fatalf("got %d %s", code, errorCode)
	}
}

// Make sure an unknown token and a disabled app are refused before any route runs.
func TestAppMiddleware_RefusesBadKeyAndDisabledApp(t *testing.T) {
	if code, errorCode := serveWith(fakeApps{err: ErrInvalidAppToken}, AppTokenPrefix+"key", "/app-route"); code != http.StatusUnauthorized || errorCode != "invalid_token" {
		t.Errorf("unknown token: got %d %s", code, errorCode)
	}
	if code, errorCode := serveWith(fakeApps{err: ErrAppDisabled}, AppTokenPrefix+"key", "/app-route"); code != http.StatusForbidden || errorCode != "app_disabled" {
		t.Errorf("disabled app: got %d %s", code, errorCode)
	}
}
