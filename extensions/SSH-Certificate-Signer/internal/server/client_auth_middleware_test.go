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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/airavata-custos/signer/internal/auth"
)

// ClientAuthMiddleware guards the machine-to-machine routes, including
// POST /api/v1/revoke. Missing or malformed credentials are rejected with 401
// before any database lookup, so a nil-DB authenticator exercises the rejection
// paths safely.
func TestClientAuthMiddleware_RejectsMissingCredentials(t *testing.T) {
	authenticator := auth.NewClientAuthenticator(nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK) // must not be reached
	})
	handler := ClientAuthMiddleware(authenticator)(next)

	cases := []struct {
		name     string
		clientID string
		secret   string
	}{
		{"missing both headers", "", ""},
		{"missing secret", "tenant1:webapp", ""},
		{"missing client id", "", "s3cret"},
		{"malformed client id (no colon)", "webapp", "s3cret"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/revoke", nil)
			if tc.clientID != "" {
				req.Header.Set("X-Client-Id", tc.clientID)
			}
			if tc.secret != "" {
				req.Header.Set("X-Client-Secret", tc.secret)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("status: got %d, want 401", rr.Code)
			}
			var body map[string]string
			json.NewDecoder(rr.Body).Decode(&body)
			if body["error"] != "unauthorized" {
				t.Errorf("error code: got %q, want unauthorized", body["error"])
			}
		})
	}
}
