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

	"github.com/apache/airavata-custos/allocations/access-amie/model"
	"github.com/jmoiron/sqlx"
)

type AuditStore interface {
	Save(ctx context.Context, tx *sql.Tx, a *model.AuditLog) error
}

type mariaDBauditStore struct {
	db *sqlx.DB
}

func NewAuditStore(db *sqlx.DB) AuditStore {
	return &mariaDBauditStore{db: db}
}

func (s *mariaDBauditStore) Save(ctx context.Context, tx *sql.Tx, a *model.AuditLog) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO amie_audit_logs (packet_id, event_id, action, entity_type, entity_id, summary, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		a.PacketID, a.EventID, a.Action, a.EntityType, a.EntityID, a.Summary, a.CreatedAt)
	return err
}
