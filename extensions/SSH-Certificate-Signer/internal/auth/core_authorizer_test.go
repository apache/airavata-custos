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

package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCoreAuthorizationClientResolveCaller(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me" {
			t.Fatalf("path = %q, want /me", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-1" {
			t.Fatalf("authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"user":{"id":"admin-1","email":"admin@example.org"},"privileges":["signer:certificates:read"]}`))
	}))
	defer server.Close()

	client, err := NewCoreAuthorizationClient(server.URL, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	caller, err := client.ResolveCaller(context.Background(), "token-1")
	if err != nil {
		t.Fatal(err)
	}
	if caller.ID != "admin-1" || !caller.HasPrivilege(SignerCertificatesRead) {
		t.Fatalf("unexpected caller: %+v", caller)
	}
}

func TestCoreAuthorizationClientErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{name: "unauthorized", status: http.StatusUnauthorized, body: `{}`, want: ErrCoreUnauthorized},
		{name: "unavailable", status: http.StatusInternalServerError, body: `{}`, want: ErrCoreUnavailable},
		{name: "malformed", status: http.StatusOK, body: `{`, want: ErrCoreUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()
			client, err := NewCoreAuthorizationClient(server.URL, time.Second)
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.ResolveCaller(context.Background(), "token")
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestNewCoreAuthorizationClientRejectsInvalidURL(t *testing.T) {
	if _, err := NewCoreAuthorizationClient("not-a-url", time.Second); err == nil {
		t.Fatal("expected invalid URL error")
	}
}
