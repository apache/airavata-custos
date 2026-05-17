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

const computeAllocationChangeRequestEventColumns = "id, compute_allocation_change_request_id, event_type, description, timestamp"

type mysqlComputeAllocationChangeRequestEventStore struct {
	db *sqlx.DB
}

// NewComputeAllocationChangeRequestEventStore returns a MySQL-backed
// ComputeAllocationChangeRequestEventStore.
func NewComputeAllocationChangeRequestEventStore(db *sqlx.DB) ComputeAllocationChangeRequestEventStore {
	return &mysqlComputeAllocationChangeRequestEventStore{db: db}
}

func (s *mysqlComputeAllocationChangeRequestEventStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocationChangeRequestEvent, error) {
	var e models.ComputeAllocationChangeRequestEvent
	err := s.db.GetContext(ctx, &e,
		`SELECT `+computeAllocationChangeRequestEventColumns+` FROM compute_allocation_change_request_events WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (s *mysqlComputeAllocationChangeRequestEventStore) FindByChangeRequest(ctx context.Context, changeRequestID string) ([]models.ComputeAllocationChangeRequestEvent, error) {
	var rows []models.ComputeAllocationChangeRequestEvent
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationChangeRequestEventColumns+`
		 FROM compute_allocation_change_request_events
		 WHERE compute_allocation_change_request_id = ?
		 ORDER BY timestamp`, changeRequestID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlComputeAllocationChangeRequestEventStore) FindLatestByChangeRequest(ctx context.Context, changeRequestID string) (*models.ComputeAllocationChangeRequestEvent, error) {
	var e models.ComputeAllocationChangeRequestEvent
	err := s.db.GetContext(ctx, &e,
		`SELECT `+computeAllocationChangeRequestEventColumns+`
		 FROM compute_allocation_change_request_events
		 WHERE compute_allocation_change_request_id = ?
		 ORDER BY timestamp DESC
		 LIMIT 1`, changeRequestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (s *mysqlComputeAllocationChangeRequestEventStore) Create(ctx context.Context, tx *sql.Tx, e *models.ComputeAllocationChangeRequestEvent) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocation_change_request_events
		     (id, compute_allocation_change_request_id, event_type, description, timestamp)
		 VALUES (?, ?, ?, ?, ?)`,
		e.ID, e.ComputeAllocationChangeRequestID, e.EventType, e.Description, e.Timestamp)
	return err
}

func (s *mysqlComputeAllocationChangeRequestEventStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_allocation_change_request_events WHERE id = ?`, id)
	return err
}
