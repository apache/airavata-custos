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

package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func certificateQueryRow() *sqlmock.Rows {
	issued := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	revoked := issued.Add(30 * time.Minute)
	return sqlmock.NewRows([]string{
		"id", "tenant_id", "client_id", "serial_number", "key_id", "principal", "user_email",
		"public_key_fingerprint", "ca_fingerprint", "valid_after", "valid_before", "issued_at",
		"source_ip", "granted_extensions", "force_command", "revoked", "revoked_at", "revocation_reason", "revoked_by",
	}).AddRow(
		1, "tenant-1", "client-1", 42, "key-1", "alice", "alice@example.org",
		"SHA256:key", "SHA256:ca", issued, issued.Add(time.Hour), issued,
		"192.0.2.1", []byte(`["permit-pty"]`), nil, true, revoked, "compromised", "admin@example.org",
	)
}

func newMockDB(t *testing.T) (*DB, sqlmock.Sqlmock) {
	t.Helper()
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqldb.Close() })
	return &DB{DB: sqldb}, mock
}

func TestListCertificatesDeploymentWide(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectQuery("LEFT JOIN revocation_events r ON r.id").
		WithArgs(21).
		WillReturnRows(certificateQueryRow())

	result, err := db.ListCertificates(context.Background(), 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Certificates) != 1 || result.NextCursor != nil {
		t.Fatalf("unexpected result: %+v", result)
	}
	certificate := result.Certificates[0]
	if certificate.TenantID != "tenant-1" || certificate.UserEmail != "alice@example.org" || !certificate.Revoked {
		t.Fatalf("unexpected certificate: %+v", certificate)
	}
	if certificate.RevokedBy != "admin@example.org" {
		t.Fatalf("unexpected revocation actor: %q", certificate.RevokedBy)
	}
	if len(certificate.GrantedExtensions) != 1 || certificate.GrantedExtensions[0] != "permit-pty" {
		t.Fatalf("unexpected extensions: %v", certificate.GrantedExtensions)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestListCertificatesDeploymentWideUsesCursorAndTimestampTieBreaker(t *testing.T) {
	db, mock := newMockDB(t)
	issued := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	cursor := &CertificateCursor{IssuedAt: issued, ID: 10}
	rows := certificateQueryRow()
	rows.AddRow(
		2, "tenant-1", "client-1", 41, "key-2", "alice", "alice@example.org",
		"SHA256:key2", "SHA256:ca", issued, issued.Add(time.Hour), issued,
		"192.0.2.2", []byte(`[]`), nil, false, nil, "", "",
	)
	mock.ExpectQuery("WHERE \\(c.issued_at < \\$1 OR \\(c.issued_at = \\$2 AND c.id < \\$3\\)\\)").
		WithArgs(issued, issued, int64(10), 2).
		WillReturnRows(rows)

	result, err := db.ListCertificates(context.Background(), 1, cursor)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Certificates) != 1 || result.NextCursor == nil {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.NextCursor.ID != result.Certificates[0].ID || !result.NextCursor.IssuedAt.Equal(issued) {
		t.Fatalf("unexpected next cursor: %+v", result.NextCursor)
	}
}

func TestListCertificatesDeploymentWideQueryError(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectQuery("LEFT JOIN revocation_events r ON r.id").WithArgs(101).WillReturnError(errors.New("query failed"))
	if _, err := db.ListCertificates(context.Background(), 500, nil); err == nil {
		t.Fatal("expected query error")
	}
}

func TestGetCertificateBySerialMapsNotFound(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectQuery("FROM certificate_issuance_logs c").WithArgs(int64(99)).WillReturnError(sql.ErrNoRows)
	_, err := db.GetCertificateBySerial(context.Background(), 99)
	if !errors.Is(err, ErrCertificateNotFound) {
		t.Fatalf("error = %v, want ErrCertificateNotFound", err)
	}
}
