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

package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"github.com/apache/airavata-custos/signer/internal/store"
)

const (
	lookupCertificateQuery  = "SELECT tenant_id, client_id, key_id, ca_fingerprint.*WHERE serial_number = \\$1"
	existingRevocationQuery = "SELECT revoked_at, reason.*WHERE serial_number = \\$1"
	insertRevocationQuery   = "INSERT INTO revocation_events.*VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6, \\$7, \\$8\\)"
)

func newRevocationService(t *testing.T, now time.Time) (*RevocationService, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	service := NewRevocationService(&store.DB{DB: db})
	service.now = func() time.Time { return now }
	return service, mock
}

func certificateRow(email string, validAfter, validBefore time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"tenant_id", "client_id", "key_id", "ca_fingerprint", "user_email", "valid_after", "valid_before",
	}).AddRow("tenant-1", "client-1", "key-1", "SHA256:ca", email, validAfter, validBefore)
}

func activeCommand() RevokeCommand {
	return RevokeCommand{
		SerialNumber: 42, Reason: "compromised", RevokedBy: "subject-1",
		OwnerEmail: "alice@example.org", RequireActive: true,
	}
}

func TestRevocationService_RevokeActiveOwner(t *testing.T) {
	now := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	service, mock := newRevocationService(t, now)
	mock.ExpectBegin()
	mock.ExpectQuery(lookupCertificateQuery).WithArgs(int64(42)).
		WillReturnRows(certificateRow("alice@example.org", now.Add(-time.Hour), now.Add(time.Hour)))
	mock.ExpectQuery(existingRevocationQuery).WithArgs(int64(42)).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(insertRevocationQuery).WithArgs(
		"tenant-1", "client-1", int64(42), "key-1", "SHA256:ca", now,
		"compromised", "subject-1",
	).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	result, err := service.Revoke(context.Background(), activeCommand())
	if err != nil {
		t.Fatal(err)
	}
	if result.AlreadyRevoked || !result.RevokedAt.Equal(now) {
		t.Fatalf("unexpected result: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRevocationService_PrivilegedNonOwner(t *testing.T) {
	now := time.Now().UTC()
	service, mock := newRevocationService(t, now)
	command := activeCommand()
	command.CanRevokeAny = true
	command.RevokedBy = "admin-id"
	mock.ExpectBegin()
	mock.ExpectQuery(lookupCertificateQuery).WillReturnRows(certificateRow("other@example.org", now.Add(-time.Hour), now.Add(time.Hour)))
	mock.ExpectQuery(existingRevocationQuery).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(insertRevocationQuery).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	if _, err := service.Revoke(context.Background(), command); err != nil {
		t.Fatal(err)
	}
}

func TestRevocationService_RejectsNonOwner(t *testing.T) {
	now := time.Now().UTC()
	service, mock := newRevocationService(t, now)
	mock.ExpectBegin()
	mock.ExpectQuery(lookupCertificateQuery).WillReturnRows(certificateRow("other@example.org", now.Add(-time.Hour), now.Add(time.Hour)))
	mock.ExpectRollback()
	_, err := service.Revoke(context.Background(), activeCommand())
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}
}

func TestRevocationService_ActiveOnlyAndIdempotency(t *testing.T) {
	now := time.Now().UTC()
	t.Run("expired first revoke", func(t *testing.T) {
		service, mock := newRevocationService(t, now)
		mock.ExpectBegin()
		mock.ExpectQuery(lookupCertificateQuery).WillReturnRows(certificateRow("alice@example.org", now.Add(-2*time.Hour), now.Add(-time.Hour)))
		mock.ExpectQuery(existingRevocationQuery).WillReturnError(sql.ErrNoRows)
		mock.ExpectRollback()
		_, err := service.Revoke(context.Background(), activeCommand())
		if !errors.Is(err, ErrCertificateNotActive) {
			t.Fatalf("error = %v, want ErrCertificateNotActive", err)
		}
	})
	t.Run("future first revoke", func(t *testing.T) {
		service, mock := newRevocationService(t, now)
		mock.ExpectBegin()
		mock.ExpectQuery(lookupCertificateQuery).WillReturnRows(certificateRow("alice@example.org", now.Add(time.Hour), now.Add(2*time.Hour)))
		mock.ExpectQuery(existingRevocationQuery).WillReturnError(sql.ErrNoRows)
		mock.ExpectRollback()
		_, err := service.Revoke(context.Background(), activeCommand())
		if !errors.Is(err, ErrCertificateNotActive) {
			t.Fatalf("error = %v, want ErrCertificateNotActive", err)
		}
	})
	t.Run("existing revoke succeeds after expiry", func(t *testing.T) {
		service, mock := newRevocationService(t, now)
		original := now.Add(-2 * time.Hour)
		mock.ExpectBegin()
		mock.ExpectQuery(lookupCertificateQuery).WillReturnRows(certificateRow("alice@example.org", now.Add(-3*time.Hour), now.Add(-time.Hour)))
		mock.ExpectQuery(existingRevocationQuery).WillReturnRows(sqlmock.NewRows([]string{"revoked_at", "reason"}).AddRow(original, "original"))
		mock.ExpectCommit()
		result, err := service.Revoke(context.Background(), activeCommand())
		if err != nil || !result.AlreadyRevoked || result.Reason != "original" || !result.RevokedAt.Equal(original) {
			t.Fatalf("result=%+v err=%v", result, err)
		}
	})
}

func TestRevocationService_FailuresRollback(t *testing.T) {
	now := time.Now().UTC()
	t.Run("begin", func(t *testing.T) {
		service, mock := newRevocationService(t, now)
		mock.ExpectBegin().WillReturnError(errors.New("begin failed"))
		if _, err := service.Revoke(context.Background(), activeCommand()); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("lookup", func(t *testing.T) {
		service, mock := newRevocationService(t, now)
		mock.ExpectBegin()
		mock.ExpectQuery(lookupCertificateQuery).WillReturnError(errors.New("lookup failed"))
		mock.ExpectRollback()
		if _, err := service.Revoke(context.Background(), activeCommand()); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("insert", func(t *testing.T) {
		service, mock := newRevocationService(t, now)
		mock.ExpectBegin()
		mock.ExpectQuery(lookupCertificateQuery).WillReturnRows(certificateRow("alice@example.org", now.Add(-time.Hour), now.Add(time.Hour)))
		mock.ExpectQuery(existingRevocationQuery).WillReturnError(sql.ErrNoRows)
		mock.ExpectExec(insertRevocationQuery).WillReturnError(errors.New("insert failed"))
		mock.ExpectRollback()
		if _, err := service.Revoke(context.Background(), activeCommand()); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("commit", func(t *testing.T) {
		service, mock := newRevocationService(t, now)
		mock.ExpectBegin()
		mock.ExpectQuery(lookupCertificateQuery).WillReturnRows(certificateRow("alice@example.org", now.Add(-time.Hour), now.Add(time.Hour)))
		mock.ExpectQuery(existingRevocationQuery).WillReturnError(sql.ErrNoRows)
		mock.ExpectExec(insertRevocationQuery).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit().WillReturnError(errors.New("commit failed"))
		if _, err := service.Revoke(context.Background(), activeCommand()); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestRevocationService_UnknownCertificate(t *testing.T) {
	service, mock := newRevocationService(t, time.Now().UTC())
	mock.ExpectBegin()
	mock.ExpectQuery(lookupCertificateQuery).WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()
	_, err := service.Revoke(context.Background(), activeCommand())
	if !errors.Is(err, store.ErrCertificateNotFound) {
		t.Fatalf("error = %v, want ErrCertificateNotFound", err)
	}
}
