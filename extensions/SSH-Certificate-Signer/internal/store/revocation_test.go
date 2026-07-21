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

package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// newMockDB returns a *DB backed by sqlmock. The mock's default query matcher is
// regular-expression based, so the query strings below are treated as regexps
// matched against the executed SQL (unique table/column fragments are enough).
func newMockDB(t *testing.T) (*DB, sqlmock.Sqlmock) {
	t.Helper()
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("creating sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqldb.Close() })
	return &DB{DB: sqldb}, mock
}

const (
	qLookupCert   = "SELECT tenant_id, client_id, key_id, ca_fingerprint"
	qExistingRevo = "SELECT revoked_at, reason"
	qInsertRevo   = "INSERT INTO revocation_events"
)

func certRow() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"tenant_id", "client_id", "key_id", "ca_fingerprint"}).
		AddRow("tenant1", "webapp", "alice@webapp-1700000000", "SHA256:cafp")
}

// Area 1: revoking a known serial inserts exactly one event that copies the
// owning certificate's identifiers, and the returned struct reflects the row.
func TestRevokeCertificateBySerial_KnownSerial(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(qLookupCert).WithArgs(int64(42)).WillReturnRows(certRow())
	mock.ExpectQuery(qExistingRevo).WithArgs(int64(42)).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(qInsertRevo).
		WithArgs(
			"tenant1",                 // tenant_id (inherited from cert)
			"webapp",                  // client_id (inherited from cert)
			int64(42),                 // serial_number
			"alice@webapp-1700000000", // key_id (inherited from cert)
			"SHA256:cafp",             // ca_fingerprint (inherited from cert)
			sqlmock.AnyArg(),          // revoked_at
			"compromised",             // reason
			"tenant1:webapp",          // revoked_by
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	before := time.Now().UTC().Add(-time.Second)
	got, err := db.RevokeCertificateBySerial(context.Background(), 42, "compromised", "tenant1:webapp")
	if err != nil {
		t.Fatalf("RevokeCertificateBySerial: %v", err)
	}

	if got.SerialNumber != 42 {
		t.Errorf("serial: got %d, want 42", got.SerialNumber)
	}
	if got.Reason != "compromised" {
		t.Errorf("reason: got %q, want %q", got.Reason, "compromised")
	}
	if got.AlreadyRevoked {
		t.Error("AlreadyRevoked: got true, want false")
	}
	if got.RevokedAt.Before(before) {
		t.Errorf("RevokedAt %v is older than test start %v", got.RevokedAt, before)
	}

	// ExpectationsWereMet enforces that exactly one insert (and nothing more) ran.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Area 2: revoking the same serial twice inserts one event; the second call
// succeeds, reports AlreadyRevoked, and echoes the original timestamp/reason.
func TestRevokeCertificateBySerial_Idempotent(t *testing.T) {
	db, mock := newMockDB(t)

	// First revoke: no prior event, so it inserts.
	mock.ExpectBegin()
	mock.ExpectQuery(qLookupCert).WithArgs(int64(42)).WillReturnRows(certRow())
	mock.ExpectQuery(qExistingRevo).WithArgs(int64(42)).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(qInsertRevo).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Second revoke: a prior event now exists, so it must NOT insert again.
	originalAt := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(qLookupCert).WithArgs(int64(42)).WillReturnRows(certRow())
	mock.ExpectQuery(qExistingRevo).WithArgs(int64(42)).WillReturnRows(
		sqlmock.NewRows([]string{"revoked_at", "reason"}).AddRow(originalAt, "first reason"),
	)
	mock.ExpectCommit()

	if _, err := db.RevokeCertificateBySerial(context.Background(), 42, "first reason", "tenant1:webapp"); err != nil {
		t.Fatalf("first revoke: %v", err)
	}

	second, err := db.RevokeCertificateBySerial(context.Background(), 42, "second reason", "tenant1:webapp")
	if err != nil {
		t.Fatalf("second revoke: %v", err)
	}
	if !second.AlreadyRevoked {
		t.Error("AlreadyRevoked: got false, want true on repeat revoke")
	}
	if !second.RevokedAt.Equal(originalAt) {
		t.Errorf("RevokedAt: got %v, want original %v", second.RevokedAt, originalAt)
	}
	if second.Reason != "first reason" {
		t.Errorf("reason: got %q, want original %q (not overwritten)", second.Reason, "first reason")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Area 3 (store): an unknown serial returns ErrCertificateNotFound and inserts
// no revocation event.
func TestRevokeCertificateBySerial_UnknownSerial(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(qLookupCert).WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err := db.RevokeCertificateBySerial(context.Background(), 999, "gone", "tenant1:webapp")
	if !errors.Is(err, ErrCertificateNotFound) {
		t.Fatalf("error: got %v, want ErrCertificateNotFound", err)
	}
	// No ExpectExec was registered; ExpectationsWereMet confirms no insert ran.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Area 4: an issuance-lookup error that is not sql.ErrNoRows is propagated
// (and never mistaken for "not found").
func TestRevokeCertificateBySerial_LookupError(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(qLookupCert).WithArgs(int64(42)).WillReturnError(errors.New("connection reset"))
	mock.ExpectRollback()

	_, err := db.RevokeCertificateBySerial(context.Background(), 42, "r", "tenant1:webapp")
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, ErrCertificateNotFound) {
		t.Error("lookup failure must not be reported as ErrCertificateNotFound")
	}
	if !strings.Contains(err.Error(), "looking up certificate") {
		t.Errorf("error not wrapped as expected: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Area 4: an insert failure is propagated and leaves no committed revocation
// (the transaction is rolled back, never committed).
func TestRevokeCertificateBySerial_InsertError(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(qLookupCert).WithArgs(int64(42)).WillReturnRows(certRow())
	mock.ExpectQuery(qExistingRevo).WithArgs(int64(42)).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(qInsertRevo).WillReturnError(errors.New("write failed"))
	mock.ExpectRollback() // no commit -> no partial revocation state

	_, err := db.RevokeCertificateBySerial(context.Background(), 42, "r", "tenant1:webapp")
	if err == nil || !strings.Contains(err.Error(), "inserting revocation event") {
		t.Fatalf("error: got %v, want wrapped insert failure", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Area 4: a commit failure is propagated.
func TestRevokeCertificateBySerial_CommitError(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(qLookupCert).WithArgs(int64(42)).WillReturnRows(certRow())
	mock.ExpectQuery(qExistingRevo).WithArgs(int64(42)).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(qInsertRevo).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(errors.New("commit failed"))

	_, err := db.RevokeCertificateBySerial(context.Background(), 42, "r", "tenant1:webapp")
	if err == nil || !strings.Contains(err.Error(), "committing transaction") {
		t.Fatalf("error: got %v, want wrapped commit failure", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// Area 4: a failure to begin the transaction is propagated.
func TestRevokeCertificateBySerial_BeginError(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectBegin().WillReturnError(errors.New("no connection"))

	_, err := db.RevokeCertificateBySerial(context.Background(), 42, "r", "tenant1:webapp")
	if err == nil || !strings.Contains(err.Error(), "beginning transaction") {
		t.Fatalf("error: got %v, want wrapped begin failure", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
