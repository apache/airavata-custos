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

package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/apache/airavata-custos/signer/internal/auth"
	"github.com/apache/airavata-custos/signer/internal/httputil"
	signerservice "github.com/apache/airavata-custos/signer/internal/service"
	"github.com/apache/airavata-custos/signer/internal/store"
)

type certificateStoreStub struct {
	get  func(context.Context, int64) (*store.CertificateWithStatus, error)
	list func(context.Context, int, *store.CertificateCursor) (*store.CertificatePageResult, error)
}

func (s certificateStoreStub) ListCertificatesByEmail(context.Context, string, int, int) (*store.CertificateListResult, error) {
	return nil, errors.New("unexpected owner list")
}

func (s certificateStoreStub) ListCertificates(ctx context.Context, limit int, cursor *store.CertificateCursor) (*store.CertificatePageResult, error) {
	if s.list == nil {
		return nil, errors.New("unexpected admin list")
	}
	return s.list(ctx, limit, cursor)
}

func (s certificateStoreStub) GetCertificateBySerial(ctx context.Context, serial int64) (*store.CertificateWithStatus, error) {
	return s.get(ctx, serial)
}

type revokerStub struct {
	command signerservice.RevokeCommand
	result  *signerservice.RevokedCertificate
	err     error
}

func (s *revokerStub) Revoke(_ context.Context, command signerservice.RevokeCommand) (*signerservice.RevokedCertificate, error) {
	s.command = command
	return s.result, s.err
}

type coreAuthorizerStub struct {
	called int
	caller *auth.CoreCaller
	err    error
}

func (s *coreAuthorizerStub) ResolveCaller(context.Context, string) (*auth.CoreCaller, error) {
	s.called++
	return s.caller, s.err
}

func portalRevokeRequest(email, subject string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/certificates/42/revoke", strings.NewReader(`{"reason":"compromised"}`))
	req.Header.Set("Authorization", "Bearer token")
	ctx := httputil.WithUserIdentity(req.Context(), &httputil.UserIdentityContext{Email: email, Subject: subject})
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("serial", "42")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, routeCtx)
	return req.WithContext(ctx)
}

func newPortalHandler(owner string, authorizer auth.CoreAuthorizer, revoker CertificateRevoker) *CertificatesHandler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	db := certificateStoreStub{get: func(context.Context, int64) (*store.CertificateWithStatus, error) {
		return &store.CertificateWithStatus{SerialNumber: 42, UserEmail: owner}, nil
	}}
	return NewCertificatesHandler(db, logger).WithRevocation(authorizer, revoker)
}

func TestCertificatesHandler_RevokeOwnerDoesNotDependOnCore(t *testing.T) {
	core := &coreAuthorizerStub{err: auth.ErrCoreUnavailable}
	revoker := &revokerStub{result: &signerservice.RevokedCertificate{
		SerialNumber: 42, Reason: "compromised", RevokedAt: time.Unix(100, 0),
	}}
	handler := newPortalHandler("alice@example.org", core, revoker)
	recorder := httptest.NewRecorder()
	handler.HandleRevoke(recorder, portalRevokeRequest("alice@example.org", "oidc-subject"))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if core.called != 0 {
		t.Fatalf("Core called %d times for owner revoke", core.called)
	}
	if revoker.command.RevokedBy != "oidc-subject" || revoker.command.CanRevokeAny || !revoker.command.RequireActive {
		t.Fatalf("unexpected command: %+v", revoker.command)
	}
}

func TestCertificatesHandler_RevokePrivilegedNonOwner(t *testing.T) {
	core := &coreAuthorizerStub{caller: &auth.CoreCaller{
		ID: "admin-id", Email: "admin@example.org", Privileges: []string{auth.SignerCertificatesWrite},
	}}
	revoker := &revokerStub{result: &signerservice.RevokedCertificate{
		SerialNumber: 42, Reason: "compromised", RevokedAt: time.Unix(100, 0),
	}}
	handler := newPortalHandler("alice@example.org", core, revoker)
	recorder := httptest.NewRecorder()
	handler.HandleRevoke(recorder, portalRevokeRequest("admin@example.org", "admin-subject"))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if core.called != 1 || !revoker.command.CanRevokeAny || revoker.command.RevokedBy != "admin-id" {
		t.Fatalf("core calls=%d command=%+v", core.called, revoker.command)
	}
}

func TestCertificatesHandler_RevokeNonOwnerAuthorizationFailures(t *testing.T) {
	tests := []struct {
		name   string
		core   *coreAuthorizerStub
		status int
	}{
		{"no privilege", &coreAuthorizerStub{caller: &auth.CoreCaller{ID: "user"}}, http.StatusForbidden},
		{"core unauthorized", &coreAuthorizerStub{err: auth.ErrCoreUnauthorized}, http.StatusUnauthorized},
		{"core unavailable", &coreAuthorizerStub{err: auth.ErrCoreUnavailable}, http.StatusServiceUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := newPortalHandler("alice@example.org", test.core, &revokerStub{})
			recorder := httptest.NewRecorder()
			handler.HandleRevoke(recorder, portalRevokeRequest("other@example.org", "other"))
			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d, body=%s", recorder.Code, test.status, recorder.Body.String())
			}
		})
	}
}

func TestCertificatesHandler_RevokeActiveOnlyError(t *testing.T) {
	revoker := &revokerStub{err: signerservice.ErrCertificateNotActive}
	handler := newPortalHandler("alice@example.org", &coreAuthorizerStub{}, revoker)
	recorder := httptest.NewRecorder()
	handler.HandleRevoke(recorder, portalRevokeRequest("alice@example.org", "subject"))
	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestCertificatesHandler_RevokeValidation(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
	}{
		{"invalid serial", "/api/v1/certificates/not-a-number/revoke", `{"reason":"valid"}`},
		{"empty reason", "/api/v1/certificates/42/revoke", `{"reason":"   "}`},
		{"unknown field", "/api/v1/certificates/42/revoke", `{"reason":"valid","extra":true}`},
		{"trailing body", "/api/v1/certificates/42/revoke", `{"reason":"valid"}{}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := newPortalHandler("alice@example.org", &coreAuthorizerStub{}, &revokerStub{})
			req := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.body))
			req = req.WithContext(httputil.WithUserIdentity(req.Context(), &httputil.UserIdentityContext{
				Email: "alice@example.org", Subject: "subject",
			}))
			recorder := httptest.NewRecorder()
			handler.HandleRevoke(recorder, req)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestCertificatesHandler_AdminCursor(t *testing.T) {
	issued := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	next := &store.CertificateCursor{IssuedAt: issued, ID: 42}
	encoded := encodeCertificateCursor(next)
	decoded, err := decodeCertificateCursor(encoded)
	if err != nil || decoded.ID != next.ID || !decoded.IssuedAt.Equal(issued) {
		t.Fatalf("decoded=%+v err=%v", decoded, err)
	}

	called := false
	handler := NewCertificatesHandler(certificateStoreStub{list: func(context.Context, int, *store.CertificateCursor) (*store.CertificatePageResult, error) {
		called = true
		return &store.CertificatePageResult{}, nil
	}}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	recorder := httptest.NewRecorder()
	handler.HandleAdminList(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/admin/certificates?cursor=bad", nil))
	if recorder.Code != http.StatusBadRequest || called {
		t.Fatalf("status=%d called=%v", recorder.Code, called)
	}

	handler = NewCertificatesHandler(certificateStoreStub{list: func(_ context.Context, limit int, cursor *store.CertificateCursor) (*store.CertificatePageResult, error) {
		if limit != 5 || cursor != nil {
			t.Fatalf("limit=%d cursor=%+v", limit, cursor)
		}
		return &store.CertificatePageResult{NextCursor: next}, nil
	}}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	recorder = httptest.NewRecorder()
	handler.HandleAdminList(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/admin/certificates?limit=5", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if _, ok := response["total"]; ok {
		t.Fatal("admin response unexpectedly contains total")
	}
	if _, ok := response["offset"]; ok {
		t.Fatal("admin response unexpectedly contains offset")
	}
	if response["next_cursor"] == "" {
		t.Fatal("admin response is missing next_cursor")
	}
}
