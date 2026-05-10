// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/apache/airavata-custos/signer/internal/auth"
	"github.com/apache/airavata-custos/signer/internal/config"
	"github.com/apache/airavata-custos/signer/internal/httputil"
)

// newTestOIDCValidator creates an OIDCValidator with token validation disabled
// so it returns a default identity without needing real JWKS.
func newTestOIDCValidator(devMode bool) *auth.OIDCValidator {
	v := auth.NewOIDCValidator(config.AuthConfig{}, config.DevModeConfig{
		Enabled: devMode,
	})
	if !devMode {
		// Set a custom JWKS fetcher that returns an empty key set to avoid HTTP calls
		v.JWKSFetcher = func(ctx context.Context, issuer string) (jwk.Set, error) {
			return jwk.NewSet(), nil
		}
	}
	return v
}

func TestBearerAuthMiddleware_MissingHeader(t *testing.T) {
	validator := newTestOIDCValidator(true)
	handler := BearerAuthMiddleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/certificates", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] != "unauthorized" {
		t.Errorf("expected error code 'unauthorized', got %q", body["error"])
	}
}

func TestBearerAuthMiddleware_InvalidScheme(t *testing.T) {
	validator := newTestOIDCValidator(true)
	handler := BearerAuthMiddleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/certificates", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestBearerAuthMiddleware_EmptyToken(t *testing.T) {
	validator := newTestOIDCValidator(true)
	handler := BearerAuthMiddleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/certificates", nil)
	req.Header.Set("Authorization", "Bearer ")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestBearerAuthMiddleware_DisabledValidation(t *testing.T) {
	// When dev mode is enabled, any token is accepted with default identity
	validator := newTestOIDCValidator(true)
	var capturedIdentity *httputil.UserIdentityContext

	handler := BearerAuthMiddleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedIdentity = httputil.UserIdentityFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/certificates", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	if capturedIdentity == nil {
		t.Fatal("expected user identity in context, got nil")
	}
	if capturedIdentity.Subject != "dev-user" {
		t.Errorf("expected subject 'dev-user', got %q", capturedIdentity.Subject)
	}
}
