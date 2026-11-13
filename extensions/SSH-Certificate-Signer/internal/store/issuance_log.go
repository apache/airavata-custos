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
	"time"
)

type IssuanceLog struct {
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
	GrantedExtensions    []string
	ForceCommand         *string
	SourceAddress        *string
}

// NextSerial returns a new certificate serial. A serial is never handed out twice, even when signing fails.
func (d *DB) NextSerial(ctx context.Context) (int64, error) {
	var serial int64
	if err := d.QueryRowContext(ctx, `SELECT nextval('certificate_serial_seq')`).Scan(&serial); err != nil {
		return 0, fmt.Errorf("allocating serial: %w", err)
	}
	return serial, nil
}

// InsertIssuanceLog writes a certificate issuance entry. Fails on duplicate serial_number.
func (d *DB) InsertIssuanceLog(ctx context.Context, log *IssuanceLog) error {
	grantedExtensions := log.GrantedExtensions
	if grantedExtensions == nil {
		grantedExtensions = []string{}
	}
	grantedExtensionsJSON, err := json.Marshal(grantedExtensions)
	if err != nil {
		return fmt.Errorf("marshaling granted_extensions: %w", err)
	}

	_, err = d.ExecContext(ctx,
		`INSERT INTO certificate_issuance_logs
		 (serial_number, key_id, principal, user_email, user_sub, requested_by, public_key_fingerprint,
		  ca_fingerprint, valid_after, valid_before, granted_extensions, force_command, source_address)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		log.SerialNumber, log.KeyID, log.Principal,
		log.UserEmail, log.UserSub, log.RequestedBy, log.PublicKeyFingerprint, log.CAFingerprint, log.ValidAfter, log.ValidBefore,
		grantedExtensionsJSON, log.ForceCommand, log.SourceAddress,
	)
	if err != nil {
		return fmt.Errorf("inserting issuance log: %w", err)
	}
	return nil
}
