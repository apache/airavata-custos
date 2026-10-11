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

package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/airavata-custos/signer/internal/auth"
	"github.com/apache/airavata-custos/signer/internal/httputil"
)

type coreAuthorizerFunc func(context.Context, string) (*auth.CoreCaller, error)

func (f coreAuthorizerFunc) ResolveCaller(ctx context.Context, bearer string) (*auth.CoreCaller, error) {
	return f(ctx, bearer)
}

func TestCorePrivilegeMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		caller     *auth.CoreCaller
		err        error
		wantStatus int
	}{
		{name: "missing bearer", wantStatus: http.StatusUnauthorized},
		{name: "core unauthorized", header: "Bearer bad", err: auth.ErrCoreUnauthorized, wantStatus: http.StatusUnauthorized},
		{name: "core unavailable", header: "Bearer ok", err: errors.New("down"), wantStatus: http.StatusServiceUnavailable},
		{name: "missing privilege", header: "Bearer ok", caller: &auth.CoreCaller{ID: "admin-1"}, wantStatus: http.StatusForbidden},
		{name: "allowed", header: "Bearer ok", caller: &auth.CoreCaller{ID: "admin-1", Email: "a@example.org", Privileges: []string{auth.SignerCertificatesRead}}, wantStatus: http.StatusNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorizer := coreAuthorizerFunc(func(_ context.Context, bearer string) (*auth.CoreCaller, error) {
				if tt.header != "" && bearer == "" {
					t.Fatal("expected bearer")
				}
				return tt.caller, tt.err
			})
			handler := CorePrivilegeMiddleware(authorizer, auth.SignerCertificatesRead)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				caller := httputil.AdminCallerFromContext(r.Context())
				if caller == nil || caller.ID != "admin-1" {
					t.Fatalf("missing caller context: %+v", caller)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/certificates", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)
			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}
