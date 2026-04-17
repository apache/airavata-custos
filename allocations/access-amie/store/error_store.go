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

type ProcessingErrorStore interface {
	Save(ctx context.Context, tx *sql.Tx, e *model.ProcessingError) error
}

type mariaDBProcessingErrorStore struct {
	db *sqlx.DB
}

func NewProcessingErrorStore(db *sqlx.DB) ProcessingErrorStore {
	return &mariaDBProcessingErrorStore{db: db}
}

func (s *mariaDBProcessingErrorStore) Save(ctx context.Context, tx *sql.Tx, e *model.ProcessingError) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO amie_processing_errors (packet_id, event_id, occurred_at, summary, detail)
		 VALUES (?, ?, ?, ?, ?)`,
		e.PacketID, e.EventID, e.OccurredAt, e.Summary, e.Detail)
	return err
}
