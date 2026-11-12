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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/airavata-custos/pkg/models"
)

type fakeApps struct {
	app     *models.App
	err     error
	allowed bool
}

func (f fakeApps) ResolveApp(context.Context, string) (*models.App, error) { return f.app, f.err }

func (f fakeApps) IsAppAuthorized(context.Context, string, string) (bool, error) {
	return f.allowed, nil
}

// fakeAppsClient accepts any user token except one from another client.
type fakeAppsClient struct{}

func (fakeAppsClient) Verify(_ context.Context, token string) (*Claims, error) {
	if token == "other-client-token" {
		return nil, errors.New("wrong audience")
	}
	return &Claims{Sub: "sub-1"}, nil
}

type fakeUser struct{ privileges []models.PrivilegeKey }

func (f fakeUser) ResolveCaller(context.Context, *Claims) (*Caller, []models.PrivilegeKey, error) {
	return &Caller{UserID: "u1"}, f.privileges, nil
}

type appCall struct {
	apps      fakeApps
	user      fakeUser
	userToken string
	method    string // GET when empty
	path      string
}

func serviceApp(privileges ...models.PrivilegeKey) *models.App {
	return &models.App{ID: "a1", Key: "tool", Mode: models.AppService, Privileges: privileges}
}

func delegatedApp(privileges ...models.PrivilegeKey) *models.App {
	return &models.App{ID: "a1", Key: "gateway", Mode: models.AppDelegated, Privileges: privileges}
}

// serve runs an app call through AppMiddleware. The router has a privileged route, the user's
// own routes registered as in the server, and the app authorization route.
func serve(call appCall) (int, string) {
	ok := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
	router := NewRouter(http.NewServeMux())
	router.RequirePrivilege("GET /clusters", models.ClustersWrite, ok)
	router.RequireAuth("GET /me/certificates", ok)
	router.RequireAuthWithoutUserConsent("GET /me", ok)
	router.RequireAuthWithoutUserConsent("POST /me/agreements/{id}/accept", ok)
	router.RequireAuthUsersOnly("GET /me/app-authorizations", ok)
	nonAppCalls := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	handler := AppMiddleware(call.apps, fakeAppsClient{}, call.user, router, nonAppCalls)

	method := call.method
	if method == "" {
		method = http.MethodGet
	}
	req := httptest.NewRequest(method, call.path, nil)
	req.Header.Set("Authorization", "Bearer "+AppTokenPrefix+"tool.secret")
	if call.userToken != "" {
		req.Header.Set(UserTokenHeader, call.userToken)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var body struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec.Code, body.Code
}

// Make sure a service app can call a privileged route only when it has that privilege.
func TestServiceApp_NeedsTheRoutePrivilege(t *testing.T) {
	if code, errorCode := serve(appCall{apps: fakeApps{app: serviceApp(models.ClustersWrite)}, path: "/clusters"}); code != http.StatusOK {
		t.Errorf("app with the privilege: got %d %s", code, errorCode)
	}
	if code, errorCode := serve(appCall{apps: fakeApps{app: serviceApp(models.ClustersRead)}, path: "/clusters"}); code != http.StatusForbidden || errorCode != "insufficient_privilege" {
		t.Errorf("app without the privilege: got %d %s", code, errorCode)
	}
}

// Make sure a service app can't use routes without a privilege, since those work on the caller's own data.
func TestServiceApp_RefusedOnMeRoutes(t *testing.T) {
	code, errorCode := serve(appCall{apps: fakeApps{app: serviceApp(models.ClustersWrite)}, path: "/me"})
	if code != http.StatusForbidden || errorCode != "app_needs_user" {
		t.Fatalf("got %d %s", code, errorCode)
	}
}

// Make sure a service app can't act for a user by sending a user token.
func TestServiceApp_RefusesUserToken(t *testing.T) {
	code, errorCode := serve(appCall{apps: fakeApps{app: serviceApp(models.ClustersWrite)}, userToken: "user-token", path: "/clusters"})
	if code != http.StatusForbidden || errorCode != "user_token_not_allowed" {
		t.Fatalf("got %d %s", code, errorCode)
	}
}

// Make sure a delegated app can't call anything without a user token.
func TestDelegatedApp_NeedsUserToken(t *testing.T) {
	code, errorCode := serve(appCall{apps: fakeApps{app: delegatedApp(models.ClustersWrite), allowed: true}, path: "/clusters"})
	if code != http.StatusUnauthorized || errorCode != "user_token_required" {
		t.Fatalf("got %d %s", code, errorCode)
	}
}

// Make sure a user token from another client, including the portal, is refused.
func TestDelegatedApp_RefusesTokenFromAnotherClient(t *testing.T) {
	code, errorCode := serve(appCall{apps: fakeApps{app: delegatedApp(), allowed: true}, userToken: "other-client-token", path: "/me"})
	if code != http.StatusUnauthorized || errorCode != "invalid_user_token" {
		t.Fatalf("got %d %s", code, errorCode)
	}
}

// Make sure an app can't act for a user who hasn't allowed it.
func TestDelegatedApp_NeedsUsersAllow(t *testing.T) {
	code, errorCode := serve(appCall{apps: fakeApps{app: delegatedApp()}, userToken: "user-token", path: "/me"})
	if code != http.StatusForbidden || errorCode != "app_not_authorized" {
		t.Fatalf("got %d %s", code, errorCode)
	}
}

// Make sure a delegated app can call a privileged route only when both the app and the user have the privilege.
func TestDelegatedApp_GetsOnlyPrivilegesBothHold(t *testing.T) {
	tests := []struct {
		name string
		app  []models.PrivilegeKey
		user []models.PrivilegeKey
		want int
	}{
		{"both hold it", []models.PrivilegeKey{models.ClustersWrite}, []models.PrivilegeKey{models.ClustersWrite}, http.StatusOK},
		{"only the app", []models.PrivilegeKey{models.ClustersWrite}, nil, http.StatusForbidden},
		{"only the user", nil, []models.PrivilegeKey{models.ClustersWrite}, http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := appCall{apps: fakeApps{app: delegatedApp(tt.app...), allowed: true}, user: fakeUser{privileges: tt.user}, userToken: "user-token", path: "/clusters"}
			if code, errorCode := serve(call); code != tt.want {
				t.Fatalf("got %d %s, want %d", code, errorCode, tt.want)
			}
		})
	}
}

// Make sure a delegated app can use the user's own routes, including the profile, but can't allow or remove apps for them.
func TestDelegatedApp_CannotManageAppAuthorizations(t *testing.T) {
	call := appCall{apps: fakeApps{app: delegatedApp(), allowed: true}, userToken: "user-token"}
	for _, path := range []string{"/me/certificates", "/me"} {
		call.path = path
		if code, errorCode := serve(call); code != http.StatusOK {
			t.Errorf("%s: got %d %s", path, code, errorCode)
		}
	}
	call.path = "/me/app-authorizations"
	if code, errorCode := serve(call); code != http.StatusForbidden || errorCode != "app_not_allowed" {
		t.Errorf("app authorizations: got %d %s", code, errorCode)
	}
}

// Make sure a delegated app cannot accept a user agreement for the user, even though it can read the profile on the same kind of route.
func TestDelegatedApp_CannotAcceptAgreement(t *testing.T) {
	call := appCall{apps: fakeApps{app: delegatedApp(), allowed: true}, userToken: "user-token", method: http.MethodPost, path: "/me/agreements/a1/accept"}
	if code, errorCode := serve(call); code != http.StatusForbidden || errorCode != "app_not_allowed" {
		t.Fatalf("got %d %s", code, errorCode)
	}
}

// Make sure an unknown token and a disabled app are refused before any route runs.
func TestAppMiddleware_RefusesBadKeyAndDisabledApp(t *testing.T) {
	if code, errorCode := serve(appCall{apps: fakeApps{err: ErrInvalidAppToken}, path: "/clusters"}); code != http.StatusUnauthorized || errorCode != "invalid_token" {
		t.Errorf("unknown token: got %d %s", code, errorCode)
	}
	if code, errorCode := serve(appCall{apps: fakeApps{err: ErrAppDisabled}, path: "/clusters"}); code != http.StatusForbidden || errorCode != "app_disabled" {
		t.Errorf("disabled app: got %d %s", code, errorCode)
	}
}
