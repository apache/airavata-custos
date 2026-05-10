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
	"encoding/json"
	"fmt"
	"time"
)

type CertificateWithStatus struct {
	ID                   int64
	TenantID             string
	ClientID             string
	SerialNumber         int64
	KeyID                string
	Principal            string
	UserEmail            string
	PublicKeyFingerprint string
	CAFingerprint        string
	ValidAfter           time.Time
	ValidBefore          time.Time
	IssuedAt             time.Time
	SourceIP             string
	GrantedExtensions    []string
	ForceCommand         *string
	Revoked              bool
	RevokedAt            *time.Time
	RevocationReason     string
}

type CertificateListResult struct {
	Certificates []CertificateWithStatus
	Total        int
}

// ListCertificatesByEmail returns certificates issued to a user email, ordered by
// issued_at descending (newest first). Includes revocation status via LEFT JOIN.
func (d *DB) ListCertificatesByEmail(ctx context.Context, email string, limit, offset int) (*CertificateListResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// Count total matching certificates
	var total int
	err := d.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM certificate_issuance_logs WHERE user_email = ?`,
		email,
	).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("counting certificates: %w", err)
	}

	// Fetch page with revocation status
	rows, err := d.QueryContext(ctx,
		`SELECT
			c.id, c.tenant_id, c.client_id, c.serial_number, c.key_id,
			c.principal, COALESCE(c.user_email, ''), c.public_key_fingerprint, c.ca_fingerprint,
			c.valid_after, c.valid_before, c.issued_at, COALESCE(c.source_ip, ''),
			c.granted_extensions, c.force_command,
			CASE WHEN r.id IS NOT NULL THEN TRUE ELSE FALSE END AS revoked,
			r.revoked_at,
			COALESCE(r.reason, '') AS revocation_reason
		 FROM certificate_issuance_logs c
		 LEFT JOIN revocation_events r ON r.serial_number = c.serial_number
		 WHERE c.user_email = ?
		 ORDER BY c.issued_at DESC
		 LIMIT ? OFFSET ?`,
		email, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("querying certificates: %w", err)
	}
	defer rows.Close()

	var certs []CertificateWithStatus
	for rows.Next() {
		var cert CertificateWithStatus
		var revokedAt *time.Time
		var grantedExtensionsJSON []byte
		var forceCommand sql.NullString

		if err := rows.Scan(
			&cert.ID, &cert.TenantID, &cert.ClientID, &cert.SerialNumber, &cert.KeyID,
			&cert.Principal, &cert.UserEmail, &cert.PublicKeyFingerprint, &cert.CAFingerprint,
			&cert.ValidAfter, &cert.ValidBefore, &cert.IssuedAt, &cert.SourceIP,
			&grantedExtensionsJSON, &forceCommand,
			&cert.Revoked, &revokedAt, &cert.RevocationReason,
		); err != nil {
			return nil, fmt.Errorf("scanning certificate row: %w", err)
		}

		if grantedExtensionsJSON != nil {
			if err := json.Unmarshal(grantedExtensionsJSON, &cert.GrantedExtensions); err != nil {
				return nil, fmt.Errorf("unmarshaling granted_extensions: %w", err)
			}
		}
		if forceCommand.Valid {
			cert.ForceCommand = &forceCommand.String
		}
		cert.RevokedAt = revokedAt
		certs = append(certs, cert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating certificate rows: %w", err)
	}

	return &CertificateListResult{
		Certificates: certs,
		Total:        total,
	}, nil
}

func (d *DB) GetCertificateBySerial(ctx context.Context, serial int64) (*CertificateWithStatus, error) {
	var cert CertificateWithStatus
	var revokedAt *time.Time
	var grantedExtensionsJSON []byte
	var forceCommand sql.NullString

	err := d.QueryRowContext(ctx,
		`SELECT
			c.id, c.tenant_id, c.client_id, c.serial_number, c.key_id,
			c.principal, COALESCE(c.user_email, ''), c.public_key_fingerprint, c.ca_fingerprint,
			c.valid_after, c.valid_before, c.issued_at, COALESCE(c.source_ip, ''),
			c.granted_extensions, c.force_command,
			CASE WHEN r.id IS NOT NULL THEN TRUE ELSE FALSE END AS revoked,
			r.revoked_at,
			COALESCE(r.reason, '') AS revocation_reason
		 FROM certificate_issuance_logs c
		 LEFT JOIN revocation_events r ON r.serial_number = c.serial_number
		 WHERE c.serial_number = ?`,
		serial,
	).Scan(
		&cert.ID, &cert.TenantID, &cert.ClientID, &cert.SerialNumber, &cert.KeyID,
		&cert.Principal, &cert.UserEmail, &cert.PublicKeyFingerprint, &cert.CAFingerprint,
		&cert.ValidAfter, &cert.ValidBefore, &cert.IssuedAt, &cert.SourceIP,
		&grantedExtensionsJSON, &forceCommand,
		&cert.Revoked, &revokedAt, &cert.RevocationReason,
	)
	if err != nil {
		return nil, fmt.Errorf("querying certificate: %w", err)
	}

	if grantedExtensionsJSON != nil {
		if err := json.Unmarshal(grantedExtensionsJSON, &cert.GrantedExtensions); err != nil {
			return nil, fmt.Errorf("unmarshaling granted_extensions: %w", err)
		}
	}
	if forceCommand.Valid {
		cert.ForceCommand = &forceCommand.String
	}
	cert.RevokedAt = revokedAt
	return &cert, nil
}
