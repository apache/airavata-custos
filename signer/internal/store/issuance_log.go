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
	SourceIP             string
	UserAccessTokenHash  string
	RequestMetadata      map[string]interface{}
}

// InsertIssuanceLog writes a certificate issuance entry. Fails on duplicate serial_number.
func (d *DB) InsertIssuanceLog(ctx context.Context, log *IssuanceLog) error {
	var metadataJSON []byte
	if log.RequestMetadata != nil {
		var err error
		metadataJSON, err = json.Marshal(log.RequestMetadata)
		if err != nil {
			return fmt.Errorf("marshaling request_metadata: %w", err)
		}
	}

	_, err := d.ExecContext(ctx,
		`INSERT INTO certificate_issuance_logs
		 (tenant_id, client_id, serial_number, key_id, principal, user_email, public_key_fingerprint,
		  ca_fingerprint, valid_after, valid_before, source_ip, user_access_token_hash, request_metadata)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		log.TenantID, log.ClientID, log.SerialNumber, log.KeyID, log.Principal,
		log.UserEmail, log.PublicKeyFingerprint, log.CAFingerprint, log.ValidAfter, log.ValidBefore,
		log.SourceIP, log.UserAccessTokenHash, metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("inserting issuance log: %w", err)
	}
	return nil
}
