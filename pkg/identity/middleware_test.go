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

type fakeVerifier struct{}

func (fakeVerifier) Verify(context.Context, string) (*Claims, error) {
	return &Claims{Sub: "sub-1", Email: "a@example.edu", EmailVerified: true}, nil
}

type fakeResolver struct{ caller Caller }

func (f fakeResolver) ResolveCaller(context.Context, *Claims) (*Caller, []models.PrivilegeKey, error) {
	return &f.caller, nil, nil
}

func TestMiddleware_ConsentGate(t *testing.T) {
	ok := func(w http.ResponseWriter, req *http.Request) {
		if ClaimsFromContext(req.Context()) == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
	router := NewRouter(http.NewServeMux())
	router.RequireAuthWithoutUserConsent("GET /me", ok)
	router.RequireAuth("GET /projects", ok)

	tests := []struct {
		name     string
		caller   Caller
		path     string
		wantCode int
		wantErr  string
	}{
		{"caller needing onboarding is blocked on a normal route", Caller{UserID: "u1", NeedsOnboarding: true}, "/projects", http.StatusForbidden, "onboarding_required"},
		{"caller needing onboarding reaches a user agreement route with claims", Caller{UserID: "u1", NeedsOnboarding: true}, "/me", http.StatusOK, ""},
		{"consent-pending caller is blocked on a normal route", Caller{UserID: "u1", ConsentPending: true}, "/projects", http.StatusForbidden, "consent_required"},
		{"consent-pending caller reaches a consent route", Caller{UserID: "u1", ConsentPending: true}, "/me", http.StatusOK, ""},
		{"normal caller reaches a normal route", Caller{UserID: "u1"}, "/projects", http.StatusOK, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := Middleware(fakeVerifier{}, fakeResolver{caller: tc.caller}, nil, router.WithoutUserConsentRoutes(), router)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Authorization", "Bearer token")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantCode {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tc.wantCode, rec.Body.String())
			}
			if tc.wantErr == "" {
				return
			}
			var body map[string]string
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body["code"] != tc.wantErr {
				t.Fatalf("code = %q, want %q", body["code"], tc.wantErr)
			}
		})
	}
}
