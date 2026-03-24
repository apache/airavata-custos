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
	"fmt"
)

type RevocationEvent struct {
	TenantID      string
	ClientID      string
	SerialNumber  *int64
	KeyID         *string
	CAFingerprint *string
	Reason        string
	RevokedBy     string
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
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		ev.TenantID, ev.ClientID, serialNumber, keyID, caFingerprint, ev.Reason, ev.RevokedBy,
	)
	if err != nil {
		return fmt.Errorf("inserting revocation event: %w", err)
	}
	return nil
}
