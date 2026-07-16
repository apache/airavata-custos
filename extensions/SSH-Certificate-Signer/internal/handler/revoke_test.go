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

package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"github.com/apache/airavata-custos/signer/internal/audit"
	"github.com/apache/airavata-custos/signer/internal/httputil"
	"github.com/apache/airavata-custos/signer/internal/store"
)

func TestRevokeRequest_NoIdentifier(t *testing.T) {
	req := RevokeRequest{
		Reason: "no reason",
	}
	if req.SerialNumber != nil || req.KeyID != nil || req.CAFingerprint != nil {
		t.Error("all identifiers should be nil")
	}
}

// Validate JSON marshaling
func TestRevokeResponse_JSON(t *testing.T) {
	resp := RevokeResponse{
		Success:      true,
		Message:      "Certificate(s) revoked successfully",
		RevokedCount: 1,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	if parsed["success"] != true {
		t.Error("expected success true")
	}
	if parsed["revoked_count"].(float64) != 1 {
		t.Error("expected revoked_count 1")
	}
}

// SQL fragments (regexps under sqlmock's default matcher) identifying the two
// revoke code paths.
const (
	qLookupCert = "SELECT tenant_id, client_id, key_id, ca_fingerprint"
	qInsertRevo = "INSERT INTO revocation_events"
)

// newMockRevokeHandler wires a RevokeHandler to a sqlmock-backed *store.DB (used
// for both the audit logger and the serial store path), so both paths can be
// driven without a real database.
func newMockRevokeHandler(t *testing.T) (*RevokeHandler, sqlmock.Sqlmock) {
	t.Helper()
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("creating sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqldb.Close() })

	sdb := &store.DB{DB: sqldb}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRevokeHandler(audit.NewLogger(sdb, logger), sdb, logger), mock
}

func revokeRequest(body string, withClient bool) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/v1/revoke", strings.NewReader(body))
	if withClient {
		cfg := &store.ClientConfig{TenantID: "tenant1", ClientID: "webapp"}
		r = r.WithContext(httputil.WithClientConfig(r.Context(), cfg))
	}
	return r
}

func decodeError(t *testing.T, rr *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v (%s)", err, rr.Body.String())
	}
	return body
}

// Area 3 (handler): an unknown serial falls back to the legacy audit insert so
// automated clients keep working. The store path runs, returns not-found, and
// the handler then performs the legacy revocation_events insert.
func TestRevokeHandler_UnknownSerial_FallsBackToLegacy(t *testing.T) {
	h, mock := newMockRevokeHandler(t)

	mock.ExpectBegin()
	mock.ExpectQuery(qLookupCert).WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()
	// Legacy audit insert (single, non-transactional Exec).
	mock.ExpectExec(qInsertRevo).WillReturnResult(sqlmock.NewResult(1, 1))

	rr := httptest.NewRecorder()
	h.Handle(rr, revokeRequest(`{"serial_number":999,"reason":"lost key"}`, true))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Area 4 (handler): a real database error from the serial path returns 500 and
// must NOT silently fall back to the legacy audit insert.
func TestRevokeHandler_StoreError_Returns500_NoFallback(t *testing.T) {
	h, mock := newMockRevokeHandler(t)

	mock.ExpectBegin()
	mock.ExpectQuery(qLookupCert).WithArgs(int64(5)).WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	// No ExpectExec(qInsertRevo): the legacy insert must not run for a real error.

	rr := httptest.NewRecorder()
	h.Handle(rr, revokeRequest(`{"serial_number":5,"reason":"r"}`, true))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want 500 (%s)", rr.Code, rr.Body.String())
	}
	if got := decodeError(t, rr)["error"]; got != "internal_error" {
		t.Errorf("error code: got %q, want internal_error", got)
	}
	// ExpectationsWereMet confirms no legacy insert was attempted.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (unexpected fallback insert?): %v", err)
	}
}

// Area 5: a key-ID-only request uses the legacy audit path (no serial store path).
func TestRevokeHandler_LegacyKeyIDOnly(t *testing.T) {
	h, mock := newMockRevokeHandler(t)

	mock.ExpectExec(qInsertRevo).
		WithArgs("tenant1", "webapp", sqlmock.AnyArg(), "alice@webapp-1", sqlmock.AnyArg(), "key revoke", "tenant1:webapp").
		WillReturnResult(sqlmock.NewResult(1, 1))

	rr := httptest.NewRecorder()
	h.Handle(rr, revokeRequest(`{"key_id":"alice@webapp-1","reason":"key revoke"}`, true))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Area 5: a CA-fingerprint-only request uses the legacy audit path.
func TestRevokeHandler_LegacyCAFingerprintOnly(t *testing.T) {
	h, mock := newMockRevokeHandler(t)

	mock.ExpectExec(qInsertRevo).
		WithArgs("tenant1", "webapp", sqlmock.AnyArg(), sqlmock.AnyArg(), "SHA256:cafp", "ca revoke", "tenant1:webapp").
		WillReturnResult(sqlmock.NewResult(1, 1))

	rr := httptest.NewRecorder()
	h.Handle(rr, revokeRequest(`{"ca_fingerprint":"SHA256:cafp","reason":"ca revoke"}`, true))

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200 (%s)", rr.Code, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Area 5: the handler rejects a request lacking resolved client credentials
// (client-credential auth is enforced upstream; the handler guards the context).
func TestRevokeHandler_MissingClientConfig(t *testing.T) {
	h, _ := newMockRevokeHandler(t)

	rr := httptest.NewRecorder()
	h.Handle(rr, revokeRequest(`{"serial_number":5,"reason":"r"}`, false))

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d, want 500", rr.Code)
	}
	if got := decodeError(t, rr)["error"]; got != "internal_error" {
		t.Errorf("error code: got %q, want internal_error", got)
	}
}

// Area 5: current validation behavior for malformed / incomplete requests.
func TestRevokeHandler_BadRequests(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"malformed json", `{`},
		{"no identifier", `{"reason":"r"}`},
		{"missing reason", `{"serial_number":5}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, _ := newMockRevokeHandler(t)
			rr := httptest.NewRecorder()
			h.Handle(rr, revokeRequest(tc.body, true))

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("status: got %d, want 400 (%s)", rr.Code, rr.Body.String())
			}
			if got := decodeError(t, rr)["error"]; got != "invalid_request" {
				t.Errorf("error code: got %q, want invalid_request", got)
			}
		})
	}
}
