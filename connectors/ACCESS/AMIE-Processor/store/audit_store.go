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

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
)

// AuditExtrasStore persists the AMIE-specific reference row (packet_id,
// event_id) joined to a core audit_events row via audit_event_id.
type AuditExtrasStore interface {
	Save(ctx context.Context, tx *sql.Tx, e *model.AmieAuditExtras) error
}

type mariaDBauditExtrasStore struct {
	db *sqlx.DB
}

func NewAuditExtrasStore(db *sqlx.DB) AuditExtrasStore {
	return &mariaDBauditExtrasStore{db: db}
}

func (s *mariaDBauditExtrasStore) Save(ctx context.Context, tx *sql.Tx, e *model.AmieAuditExtras) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO amie_audit_extras (audit_event_id, packet_id, event_id) VALUES (?, ?, ?)`,
		e.AuditEventID, e.PacketID, e.EventID)
	return err
}
