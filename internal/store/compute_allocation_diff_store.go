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

const computeAllocationDiffColumns = "id, compute_allocation_id, diff_type, new_su_amount, status, timestamp, description"

type mysqlComputeAllocationDiffStore struct {
	db *sqlx.DB
}

// NewComputeAllocationDiffStore returns a MySQL-backed
// ComputeAllocationDiffStore.
func NewComputeAllocationDiffStore(db *sqlx.DB) ComputeAllocationDiffStore {
	return &mysqlComputeAllocationDiffStore{db: db}
}

func (s *mysqlComputeAllocationDiffStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocationDiff, error) {
	var d models.ComputeAllocationDiff
	err := s.db.GetContext(ctx, &d,
		`SELECT `+computeAllocationDiffColumns+` FROM compute_allocation_diffs WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (s *mysqlComputeAllocationDiffStore) FindByAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationDiff, error) {
	var diffs []models.ComputeAllocationDiff
	err := s.db.SelectContext(ctx, &diffs,
		`SELECT `+computeAllocationDiffColumns+`
		 FROM compute_allocation_diffs
		 WHERE compute_allocation_id = ?
		 ORDER BY timestamp`, allocationID)
	if err != nil {
		return nil, err
	}
	return diffs, nil
}

func (s *mysqlComputeAllocationDiffStore) FindLatestByAllocation(ctx context.Context, allocationID string) (*models.ComputeAllocationDiff, error) {
	var d models.ComputeAllocationDiff
	err := s.db.GetContext(ctx, &d,
		`SELECT `+computeAllocationDiffColumns+`
		 FROM compute_allocation_diffs
		 WHERE compute_allocation_id = ?
		 ORDER BY timestamp DESC
		 LIMIT 1`, allocationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (s *mysqlComputeAllocationDiffStore) Create(ctx context.Context, tx *sql.Tx, d *models.ComputeAllocationDiff) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocation_diffs
		     (id, compute_allocation_id, diff_type, new_su_amount, status, timestamp, description)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.ComputeAllocationID, d.DiffType, d.NewSUAmount, string(d.Status), d.Timestamp, d.Description)
	return err
}

func (s *mysqlComputeAllocationDiffStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_allocation_diffs WHERE id = ?`, id)
	return err
}
