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

const accessCheckColumns = "id, compute_allocation_id, user_id, check_type, status, COALESCE(detail, '') AS detail, infrastructure, last_checked_at, last_ok_at, failing_since"

type mysqlAccessCheckStore struct {
	db *sqlx.DB
}

// NewAccessCheckStore returns a MySQL-backed AccessCheckStore.
func NewAccessCheckStore(db *sqlx.DB) AccessCheckStore {
	return &mysqlAccessCheckStore{db: db}
}

func (s *mysqlAccessCheckStore) FindByTarget(ctx context.Context, allocationID, userID string, checkType models.AccessCheckType) (*models.AccessCheck, error) {
	var c models.AccessCheck
	err := s.db.GetContext(ctx, &c,
		`SELECT `+accessCheckColumns+` FROM access_checks
		 WHERE compute_allocation_id = ? AND user_id = ? AND check_type = ?`,
		allocationID, userID, checkType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *mysqlAccessCheckStore) FindByUserAndAllocation(ctx context.Context, allocationID, userID string) ([]models.AccessCheck, error) {
	var rows []models.AccessCheck
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+accessCheckColumns+` FROM access_checks
		 WHERE compute_allocation_id = ? AND user_id = ?
		 ORDER BY check_type`, allocationID, userID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlAccessCheckStore) FindByAllocation(ctx context.Context, allocationID string) ([]models.AccessCheck, error) {
	var rows []models.AccessCheck
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+accessCheckColumns+` FROM access_checks
		 WHERE compute_allocation_id = ?
		 ORDER BY user_id, check_type`, allocationID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlAccessCheckStore) Create(ctx context.Context, tx *sql.Tx, c *models.AccessCheck) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO access_checks
		     (id, compute_allocation_id, user_id, check_type, status, detail, infrastructure, last_checked_at, last_ok_at, failing_since)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.ComputeAllocationID, c.UserID, c.CheckType, c.Status, c.Detail, c.Infrastructure, c.LastCheckedAt, c.LastOKAt, c.FailingSince)
	return err
}

func (s *mysqlAccessCheckStore) Update(ctx context.Context, tx *sql.Tx, c *models.AccessCheck) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE access_checks
		 SET status = ?, detail = ?, infrastructure = ?, last_checked_at = ?, last_ok_at = ?, failing_since = ?
		 WHERE id = ?`,
		c.Status, c.Detail, c.Infrastructure, c.LastCheckedAt, c.LastOKAt, c.FailingSince, c.ID)
	return err
}

func (s *mysqlAccessCheckStore) CreateEvent(ctx context.Context, tx *sql.Tx, e *models.AccessCheckEvent) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO access_check_events (id, access_check_id, event_type, occurred_at)
		 VALUES (?, ?, ?, ?)`,
		e.ID, e.AccessCheckID, e.EventType, e.OccurredAt)
	return err
}

func (s *mysqlAccessCheckStore) FindEventsByChecks(ctx context.Context, checkIDs []string) ([]models.AccessCheckEvent, error) {
	if len(checkIDs) == 0 {
		return nil, nil
	}
	query, args, err := sqlx.In(
		`SELECT id, access_check_id, event_type, occurred_at
		 FROM access_check_events
		 WHERE access_check_id IN (?)
		 ORDER BY occurred_at DESC, id`, checkIDs)
	if err != nil {
		return nil, err
	}
	var rows []models.AccessCheckEvent
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(query), args...); err != nil {
		return nil, err
	}
	return rows, nil
}
