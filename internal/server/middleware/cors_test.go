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

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/internal/server/middleware"
)

const allowedOrigin = "https://portal.example.com"

func newCORS(origins ...string) *middleware.CORS {
	return middleware.NewCORS(config.CORSConfig{AllowedOrigins: origins})
}

func varyContainsOrigin(h http.Header) bool {
	for _, v := range h.Values("Vary") {
		for _, token := range strings.Split(v, ",") {
			if strings.EqualFold(strings.TrimSpace(token), "Origin") {
				return true
			}
		}
	}
	return false
}

func TestCORS_AllowedOriginGetsHeaders(t *testing.T) {
	c := newCORS(allowedOrigin)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	req.Header.Set("Origin", allowedOrigin)

	c.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Errorf("Allow-Origin = %q, want %q", got, allowedOrigin)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Allow-Credentials = %q, want \"true\"", got)
	}
	if !varyContainsOrigin(rec.Header()) {
		t.Errorf("Vary should include Origin, got %q", rec.Header().Values("Vary"))
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestCORS_DeniedOriginGetsNoAllowOrigin(t *testing.T) {
	c := newCORS(allowedOrigin)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	req.Header.Set("Origin", "https://evil.example.com")

	called := false
	c.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin set for denied origin: %q", got)
	}
	if !varyContainsOrigin(rec.Header()) {
		t.Errorf("Vary should include Origin on denied response too, got %q", rec.Header().Values("Vary"))
	}
	if !called {
		t.Errorf("handler should still run for denied origin on non-OPTIONS requests")
	}
}

func TestCORS_NoOriginHeaderNoCORSHeaders(t *testing.T) {
	c := newCORS(allowedOrigin)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects", nil)

	c.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin set without Origin header: %q", got)
	}
	if varyContainsOrigin(rec.Header()) {
		t.Errorf("Vary: Origin set without Origin header")
	}
}

func TestCORS_PreflightFromAllowedOrigin(t *testing.T) {
	c := newCORS(allowedOrigin)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/projects", nil)
	req.Header.Set("Origin", allowedOrigin)
	req.Header.Set("Access-Control-Request-Method", "GET")

	called := false
	c.Wrap(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	})).ServeHTTP(rec, req)

	if called {
		t.Errorf("preflight should short-circuit before reaching the handler")
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != allowedOrigin {
		t.Errorf("Allow-Origin = %q, want %q", got, allowedOrigin)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "GET") {
		t.Errorf("Allow-Methods should advertise GET, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "Authorization") {
		t.Errorf("Allow-Headers should advertise Authorization, got %q", got)
	}
}

func TestCORS_PreflightFromDeniedOrigin(t *testing.T) {
	c := newCORS(allowedOrigin)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/projects", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	c.Wrap(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Errorf("preflight should not reach the handler")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin should not be set for denied preflight: %q", got)
	}
	if !varyContainsOrigin(rec.Header()) {
		t.Errorf("Vary should still include Origin on denied preflight")
	}
}

func TestCORS_EmptyAllowlistIsPassthrough(t *testing.T) {
	c := newCORS() // no origins
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	req.Header.Set("Origin", allowedOrigin)

	called := false
	c.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if !called {
		t.Errorf("handler should run when no allowlist is configured")
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Allow-Origin set with empty allowlist: %q", got)
	}
}
