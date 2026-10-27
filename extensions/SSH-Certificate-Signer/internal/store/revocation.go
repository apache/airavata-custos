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
	"fmt"
	"time"
)

// ErrCertificateNotFound is returned when no issuance record matches the serial.
var ErrCertificateNotFound = errors.New("certificate not found")

type RevocationEvent struct {
	TenantID      string
	ClientID      string
	SerialNumber  *int64
	KeyID         *string
	CAFingerprint *string
	Reason        string
	RevokedBy     string
}

type CertificateForRevocation struct {
	TenantID      string
	ClientID      string
	KeyID         string
	CAFingerprint string
	UserEmail     string
	ValidAfter    time.Time
	ValidBefore   time.Time
}

type ExistingRevocation struct {
	RevokedAt time.Time
	Reason    string
}

func (d *DB) InsertRevocationEvent(ctx context.Context, ev *RevocationEvent) error {
	var serialNumber sql.NullInt64
	if ev.SerialNumber != nil {
		serialNumber = sql.NullInt64{Int64: *ev.SerialNumber, Valid: true}
	}
	var keyID sql.NullString
	if ev.KeyID != nil {
		keyID = sql.NullString{String: *ev.KeyID, Valid: true}
	}
	var caFingerprint sql.NullString
	if ev.CAFingerprint != nil {
		caFingerprint = sql.NullString{String: *ev.CAFingerprint, Valid: true}
	}

	_, err := d.ExecContext(ctx,
		`INSERT INTO revocation_events
		 (tenant_id, client_id, serial_number, key_id, ca_fingerprint, reason, revoked_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		ev.TenantID, ev.ClientID, serialNumber, keyID, caFingerprint, ev.Reason, ev.RevokedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting revocation event: %w", err)
	}
	return nil
}

type RevocationDBTX interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func GetCertificateForRevocation(ctx context.Context, db RevocationDBTX, serialNumber int64) (*CertificateForRevocation, error) {
	var certificate CertificateForRevocation
	err := db.QueryRowContext(ctx,
		`SELECT tenant_id, client_id, key_id, ca_fingerprint, COALESCE(user_email, ''), valid_after, valid_before
		 FROM certificate_issuance_logs
		 WHERE serial_number = $1
		 LIMIT 1
		 FOR UPDATE`,
		serialNumber,
	).Scan(
		&certificate.TenantID, &certificate.ClientID, &certificate.KeyID,
		&certificate.CAFingerprint, &certificate.UserEmail, &certificate.ValidAfter, &certificate.ValidBefore,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCertificateNotFound
		}
		return nil, fmt.Errorf("looking up certificate: %w", err)
	}
	return &certificate, nil
}

func GetLatestRevocation(ctx context.Context, db RevocationDBTX, serialNumber int64) (*ExistingRevocation, error) {
	var existing ExistingRevocation
	err := db.QueryRowContext(ctx,
		`SELECT revoked_at, reason
		 FROM revocation_events
		 WHERE serial_number = $1
		 ORDER BY revoked_at DESC, id DESC
		 LIMIT 1`,
		serialNumber,
	).Scan(&existing.RevokedAt, &existing.Reason)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("checking existing revocation: %w", err)
	}
	return &existing, nil
}

func InsertRevocationEventAt(ctx context.Context, db RevocationDBTX, ev *RevocationEvent, revokedAt time.Time) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO revocation_events
		 (tenant_id, client_id, serial_number, key_id, ca_fingerprint, revoked_at, reason, revoked_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		ev.TenantID, ev.ClientID, ev.SerialNumber, ev.KeyID, ev.CAFingerprint, revokedAt, ev.Reason, ev.RevokedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting revocation event: %w", err)
	}
	return nil
}
