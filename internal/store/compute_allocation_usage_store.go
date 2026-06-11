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

const computeAllocationUsageColumns = "id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id"

type mysqlComputeAllocationUsageStore struct {
	db *sqlx.DB
}

// NewComputeAllocationUsageStore returns a MySQL-backed
// ComputeAllocationUsageStore.
func NewComputeAllocationUsageStore(db *sqlx.DB) ComputeAllocationUsageStore {
	return &mysqlComputeAllocationUsageStore{db: db}
}

func (s *mysqlComputeAllocationUsageStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocationUsage, error) {
	var u models.ComputeAllocationUsage
	err := s.db.GetContext(ctx, &u,
		`SELECT `+computeAllocationUsageColumns+` FROM compute_allocation_usages WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (s *mysqlComputeAllocationUsageStore) FindByAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationUsage, error) {
	var rows []models.ComputeAllocationUsage
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationUsageColumns+`
		 FROM compute_allocation_usages
		 WHERE compute_allocation_id = ?
		 ORDER BY calculated_time`, allocationID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlComputeAllocationUsageStore) FindByUser(ctx context.Context, userID string) ([]models.ComputeAllocationUsage, error) {
	var rows []models.ComputeAllocationUsage
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationUsageColumns+`
		 FROM compute_allocation_usages
		 WHERE user_id = ?
		 ORDER BY calculated_time`, userID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlComputeAllocationUsageStore) FindByComputeAllocationIDAndJobID(ctx context.Context, allocationID, jobID string) (*models.ComputeAllocationUsage, error) {
	var u models.ComputeAllocationUsage
	err := s.db.GetContext(ctx, &u,
		`SELECT `+computeAllocationUsageColumns+` FROM compute_allocation_usages WHERE compute_allocation_id = ? AND job_id = ?`, allocationID, jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (s *mysqlComputeAllocationUsageStore) SumSUForAllocation(ctx context.Context, allocationID string) (int64, error) {
	var total sql.NullInt64
	err := s.db.GetContext(ctx, &total,
		`SELECT COALESCE(SUM(used_su_amount), 0)
		 FROM compute_allocation_usages
		 WHERE compute_allocation_id = ?`, allocationID)
	if err != nil {
		return 0, err
	}
	return total.Int64, nil
}

func (s *mysqlComputeAllocationUsageStore) SumSUForUserInAllocation(ctx context.Context, allocationID, userID string) (int64, error) {
	var total sql.NullInt64
	err := s.db.GetContext(ctx, &total,
		`SELECT COALESCE(SUM(used_su_amount), 0)
		 FROM compute_allocation_usages
		 WHERE compute_allocation_id = ? AND user_id = ?`, allocationID, userID)
	if err != nil {
		return 0, err
	}
	return total.Int64, nil
}

func (s *mysqlComputeAllocationUsageStore) Create(ctx context.Context, tx *sql.Tx, u *models.ComputeAllocationUsage) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocation_usages
		     (id, compute_allocation_id, used_raw_amount, used_su_amount, calculated_time, user_id, job_id, compute_allocation_resource_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, u.ComputeAllocationID, u.UsedRawAmount, u.UsedSUAmount, u.CalculatedTime, u.UserID, u.JobID, u.ComputeAllocationResourceID)
	return err
}

func (s *mysqlComputeAllocationUsageStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_allocation_usages WHERE id = ?`, id)
	return err
}
