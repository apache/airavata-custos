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

const computeAllocationChangeRequestColumns = "id, compute_allocation_id, requested_su_amount, requested_status, reason, change_status, requester_id, approver_id, timestamp"

type mysqlComputeAllocationChangeRequestStore struct {
	db *sqlx.DB
}

// NewComputeAllocationChangeRequestStore returns a MySQL-backed
// ComputeAllocationChangeRequestStore.
func NewComputeAllocationChangeRequestStore(db *sqlx.DB) ComputeAllocationChangeRequestStore {
	return &mysqlComputeAllocationChangeRequestStore{db: db}
}

func (s *mysqlComputeAllocationChangeRequestStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocationChangeRequest, error) {
	var c models.ComputeAllocationChangeRequest
	err := s.db.GetContext(ctx, &c,
		`SELECT `+computeAllocationChangeRequestColumns+` FROM compute_allocation_change_requests WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *mysqlComputeAllocationChangeRequestStore) FindByAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationChangeRequest, error) {
	var rows []models.ComputeAllocationChangeRequest
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationChangeRequestColumns+`
		 FROM compute_allocation_change_requests
		 WHERE compute_allocation_id = ?
		 ORDER BY timestamp`, allocationID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlComputeAllocationChangeRequestStore) FindByRequester(ctx context.Context, requesterID string) ([]models.ComputeAllocationChangeRequest, error) {
	var rows []models.ComputeAllocationChangeRequest
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationChangeRequestColumns+`
		 FROM compute_allocation_change_requests
		 WHERE requester_id = ?
		 ORDER BY timestamp`, requesterID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlComputeAllocationChangeRequestStore) Create(ctx context.Context, tx *sql.Tx, c *models.ComputeAllocationChangeRequest) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocation_change_requests
		     (id, compute_allocation_id, requested_su_amount, requested_status, reason, change_status, requester_id, approver_id, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.ComputeAllocationID, c.RequestedSUAmount, string(c.RequestedStatus), c.Reason, c.ChangeStatus, c.RequesterID, c.ApproverID, c.Timestamp)
	return err
}

func (s *mysqlComputeAllocationChangeRequestStore) Update(ctx context.Context, tx *sql.Tx, c *models.ComputeAllocationChangeRequest) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocation_change_requests
		    SET compute_allocation_id = ?,
		        requested_su_amount   = ?,
		        requested_status      = ?,
		        reason                = ?,
		        change_status         = ?,
		        requester_id          = ?,
		        approver_id           = ?,
		        timestamp             = ?
		  WHERE id = ?`,
		c.ComputeAllocationID, c.RequestedSUAmount, string(c.RequestedStatus), c.Reason, c.ChangeStatus, c.RequesterID, c.ApproverID, c.Timestamp, c.ID)
	return err
}

func (s *mysqlComputeAllocationChangeRequestStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_allocation_change_requests WHERE id = ?`, id)
	return err
}

func (s *mysqlComputeAllocationChangeRequestStore) List(ctx context.Context, f ChangeRequestListFilter) ([]models.ComputeAllocationChangeRequest, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if f.Status != "" {
		var rows []models.ComputeAllocationChangeRequest
		err := s.db.SelectContext(ctx, &rows,
			`SELECT `+computeAllocationChangeRequestColumns+
				` FROM compute_allocation_change_requests WHERE change_status = ?`+
				` ORDER BY timestamp DESC LIMIT ?`, f.Status, limit)
		return rows, err
	}
	var rows []models.ComputeAllocationChangeRequest
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationChangeRequestColumns+
			` FROM compute_allocation_change_requests`+
			` ORDER BY timestamp DESC LIMIT ?`, limit)
	return rows, err
}
