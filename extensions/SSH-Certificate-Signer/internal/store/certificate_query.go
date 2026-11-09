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
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type CertificateWithStatus struct {
	ID                   int64
	SerialNumber         int64
	KeyID                string
	Principal            string
	UserEmail            string
	UserSub              string
	RequestedBy          string
	PublicKeyFingerprint string
	CAFingerprint        string
	ValidAfter           time.Time
	ValidBefore          time.Time
	IssuedAt             time.Time
	GrantedExtensions    []string
	ForceCommand         *string
	SourceAddress        *string
	Revoked              bool
	RevokedAt            *time.Time
	RevocationReason     string
	RevokedBy            string
}

// CertificateFilter narrows a list. An empty field does not filter.
type CertificateFilter struct {
	UserSub     string
	RequestedBy string
}

type CertificateCursor struct {
	IssuedAt time.Time
	ID       int64
}

type CertificatePageResult struct {
	Certificates []CertificateWithStatus
	NextCursor   *CertificateCursor
}

const certificateSelect = `SELECT
			c.id, c.serial_number, c.key_id,
			c.principal, c.user_email, c.user_sub, c.requested_by, c.public_key_fingerprint, c.ca_fingerprint,
			c.valid_after, c.valid_before, c.issued_at,
			c.granted_extensions, c.force_command, c.source_address,
			CASE WHEN r.id IS NOT NULL THEN TRUE ELSE FALSE END AS revoked,
			r.revoked_at,
			COALESCE(r.reason, '') AS revocation_reason,
			COALESCE(r.revoked_by, '') AS revoked_by
		 FROM certificate_issuance_logs c
		 LEFT JOIN revocation_events r ON r.id = (
			SELECT r2.id FROM revocation_events r2
			WHERE r2.serial_number = c.serial_number
			ORDER BY r2.revoked_at DESC, r2.id DESC LIMIT 1
		 )`

func (d *DB) ListCertificates(ctx context.Context, limit int, cursor *CertificateCursor, filter CertificateFilter) (*CertificatePageResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var conditions []string
	var args []any
	arg := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}
	if filter.UserSub != "" {
		conditions = append(conditions, "c.user_sub = "+arg(filter.UserSub))
	}
	if filter.RequestedBy != "" {
		conditions = append(conditions, "c.requested_by = "+arg(filter.RequestedBy))
	}
	if cursor != nil {
		issuedAtBefore, issuedAtSame, idBefore := arg(cursor.IssuedAt), arg(cursor.IssuedAt), arg(cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(c.issued_at < %s OR (c.issued_at = %s AND c.id < %s))", issuedAtBefore, issuedAtSame, idBefore))
	}

	query := certificateSelect
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY c.issued_at DESC, c.id DESC LIMIT " + arg(limit+1)

	rows, err := d.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying certificates: %w", err)
	}
	defer rows.Close()

	certs, err := scanCertificates(rows)
	if err != nil {
		return nil, err
	}
	result := &CertificatePageResult{Certificates: certs}
	if len(certs) > limit {
		last := certs[limit-1]
		result.Certificates = certs[:limit]
		result.NextCursor = &CertificateCursor{IssuedAt: last.IssuedAt, ID: last.ID}
	}
	return result, nil
}

type certificateRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanCertificates(rows certificateRows) ([]CertificateWithStatus, error) {
	var certs []CertificateWithStatus
	for rows.Next() {
		var cert CertificateWithStatus
		var revokedAt *time.Time
		var grantedExtensionsJSON []byte
		if err := rows.Scan(
			&cert.ID, &cert.SerialNumber, &cert.KeyID,
			&cert.Principal, &cert.UserEmail, &cert.UserSub, &cert.RequestedBy, &cert.PublicKeyFingerprint, &cert.CAFingerprint,
			&cert.ValidAfter, &cert.ValidBefore, &cert.IssuedAt,
			&grantedExtensionsJSON, &cert.ForceCommand, &cert.SourceAddress,
			&cert.Revoked, &revokedAt, &cert.RevocationReason, &cert.RevokedBy,
		); err != nil {
			return nil, fmt.Errorf("scanning certificate row: %w", err)
		}
		if grantedExtensionsJSON != nil {
			if err := json.Unmarshal(grantedExtensionsJSON, &cert.GrantedExtensions); err != nil {
				return nil, fmt.Errorf("unmarshaling granted_extensions: %w", err)
			}
		}
		cert.RevokedAt = revokedAt
		certs = append(certs, cert)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating certificate rows: %w", err)
	}
	return certs, nil
}

func (d *DB) GetCertificateBySerial(ctx context.Context, serial int64) (*CertificateWithStatus, error) {
	rows, err := d.QueryContext(ctx, certificateSelect+" WHERE c.serial_number = $1", serial)
	if err != nil {
		return nil, fmt.Errorf("querying certificate: %w", err)
	}
	defer rows.Close()

	certs, err := scanCertificates(rows)
	if err != nil {
		return nil, err
	}
	if len(certs) == 0 {
		return nil, ErrCertificateNotFound
	}
	return &certs[0], nil
}
