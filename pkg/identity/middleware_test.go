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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/airavata-custos/pkg/models"
)

type stubVerifier struct {
	claims *Claims
	err    error
}

func (s *stubVerifier) Verify(context.Context, string) (*Claims, error) { return s.claims, s.err }

type stubResolver struct {
	caller     *Caller
	privileges []models.PrivilegeKey
	err        error
}

func (s *stubResolver) ResolveCaller(context.Context, *Claims) (*Caller, []models.PrivilegeKey, error) {
	return s.caller, s.privileges, s.err
}

func newTestRouter(t *testing.T) (*Router, *Claims, *Caller) {
	t.Helper()
	router := NewRouter(http.NewServeMux())
	claims := &Claims{Sub: "sub-1", Email: "a@b.c", Name: "A B", GivenName: "A", FamilyName: "B"}
	return router, claims, &Caller{UserID: "user-1"}
}

func serve(router *Router, verifier TokenVerifier, resolver UserResolver, method, target string) *httptest.ResponseRecorder {
	handler := Middleware(verifier, resolver, router.PublicPaths(), router.TokenPathMatcher(), router)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, nil)
	req.Header.Set("Authorization", "Bearer token")
	handler.ServeHTTP(rec, req)
	return rec
}

func TestUnlinkedCallerReachesTokenRouteWithClaims(t *testing.T) {
	router, claims, _ := newTestRouter(t)
	var got *Claims
	var gotCaller *Caller
	router.RequireToken("GET /access-requests/events/{code}", func(w http.ResponseWriter, req *http.Request) {
		got = ClaimsFromContext(req.Context())
		gotCaller = CallerFromContext(req.Context())
	})

	rec := serve(router, &stubVerifier{claims: claims}, &stubResolver{err: ErrNotLinked}, http.MethodGet, "/access-requests/events/PEARC26")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body %s", rec.Code, rec.Body.String())
	}
	if got == nil || got.Sub != "sub-1" || got.GivenName != "A" {
		t.Fatalf("claims on context = %+v, want the verified claims", got)
	}
	if gotCaller != nil {
		t.Fatalf("caller on context = %+v, want nil for unlinked", gotCaller)
	}
}

func TestUnlinkedCallerStill401OnNonTokenRoute(t *testing.T) {
	router, claims, _ := newTestRouter(t)
	router.RequireToken("GET /access-requests/me", func(http.ResponseWriter, *http.Request) {})
	router.RequireAuth("GET /me", func(http.ResponseWriter, *http.Request) {})

	rec := serve(router, &stubVerifier{claims: claims}, &stubResolver{err: ErrNotLinked}, http.MethodGet, "/me")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestUnlinkedCallerMethodMismatchIsNotATokenPath(t *testing.T) {
	router, claims, _ := newTestRouter(t)
	router.RequireToken("POST /access-requests", func(http.ResponseWriter, *http.Request) {})

	rec := serve(router, &stubVerifier{claims: claims}, &stubResolver{err: ErrNotLinked}, http.MethodGet, "/access-requests")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestLinkedCallerOnTokenRouteGetsCallerAndClaims(t *testing.T) {
	router, claims, caller := newTestRouter(t)
	var gotCaller *Caller
	var gotClaims *Claims
	router.RequireToken("GET /access-requests/me", func(w http.ResponseWriter, req *http.Request) {
		gotCaller = CallerFromContext(req.Context())
		gotClaims = ClaimsFromContext(req.Context())
	})

	rec := serve(router, &stubVerifier{claims: claims}, &stubResolver{caller: caller}, http.MethodGet, "/access-requests/me")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if gotCaller == nil || gotCaller.UserID != "user-1" {
		t.Fatalf("caller = %+v, want user-1", gotCaller)
	}
	if gotClaims == nil || gotClaims.Sub != "sub-1" {
		t.Fatalf("claims = %+v, want the verified claims", gotClaims)
	}
}

func TestLinkedCallerUnaffectedOnRegularRoutes(t *testing.T) {
	router, claims, caller := newTestRouter(t)
	router.RequireToken("POST /access-requests", func(http.ResponseWriter, *http.Request) {})
	var hasPriv bool
	router.RequirePrivilege("GET /users", models.PrivilegeKey("core:users:read"), func(w http.ResponseWriter, req *http.Request) {
		hasPriv = HasPrivilege(req.Context(), models.PrivilegeKey("core:users:read"))
	})

	resolver := &stubResolver{caller: caller, privileges: []models.PrivilegeKey{"core:users:read"}}
	rec := serve(router, &stubVerifier{claims: claims}, resolver, http.MethodGet, "/users")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !hasPriv {
		t.Fatal("privilege set missing on context")
	}
}

func TestInvalidTokenIs401EvenOnTokenRoute(t *testing.T) {
	router, _, _ := newTestRouter(t)
	router.RequireToken("GET /access-requests/me", func(http.ResponseWriter, *http.Request) {})

	rec := serve(router, &stubVerifier{err: errors.New("expired")}, &stubResolver{}, http.MethodGet, "/access-requests/me")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}
