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

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

type mysqlAccessEventStore struct {
	db *sqlx.DB
}

// NewAccessEventStore returns a MySQL-backed AccessEventStore.
func NewAccessEventStore(db *sqlx.DB) AccessEventStore {
	return &mysqlAccessEventStore{db: db}
}

func (s *mysqlAccessEventStore) FindByCode(ctx context.Context, code string) (*models.AccessEvent, error) {
	var e models.AccessEvent
	err := s.db.GetContext(ctx, &e,
		`SELECT code, compute_allocation_id, organization_id FROM access_events WHERE code = ?`, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}
