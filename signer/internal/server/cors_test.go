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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/airavata-custos/signer/internal/config"
)

func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	cfg := config.CORSConfig{
		Enabled:        true,
		AllowedOrigins: []string{"http://localhost:5173", "https://custos.example.com"},
	}

	handler := CORSMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/certificates", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	acao := rec.Header().Get("Access-Control-Allow-Origin")
	if acao != "http://localhost:5173" {
		t.Errorf("expected ACAO 'http://localhost:5173', got %q", acao)
	}
}

func TestCORSMiddleware_DisallowedOrigin(t *testing.T) {
	cfg := config.CORSConfig{
		Enabled:        true,
		AllowedOrigins: []string{"http://localhost:5173"},
	}

	handler := CORSMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/certificates", nil)
	req.Header.Set("Origin", "https://evil.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	acao := rec.Header().Get("Access-Control-Allow-Origin")
	if acao != "" {
		t.Errorf("expected no ACAO header for disallowed origin, got %q", acao)
	}
}

func TestCORSMiddleware_PreflightOptions(t *testing.T) {
	cfg := config.CORSConfig{
		Enabled:        true,
		AllowedOrigins: []string{"http://localhost:5173"},
	}

	nextCalled := false
	handler := CORSMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/api/v1/certificates", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", rec.Code)
	}
	if nextCalled {
		t.Error("next handler should not be called for preflight")
	}

	acam := rec.Header().Get("Access-Control-Allow-Methods")
	if acam == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
}

func TestCORSMiddleware_NoOriginHeader(t *testing.T) {
	cfg := config.CORSConfig{
		Enabled:        true,
		AllowedOrigins: []string{"http://localhost:5173"},
	}

	handler := CORSMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/certificates", nil)
	// No Origin header
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	acao := rec.Header().Get("Access-Control-Allow-Origin")
	if acao != "" {
		t.Errorf("expected no ACAO header without Origin, got %q", acao)
	}
}
