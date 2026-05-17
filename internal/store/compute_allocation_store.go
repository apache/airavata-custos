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

type mysqlComputeAllocationStore struct {
	db *sqlx.DB
}

// NewComputeAllocationStore returns a MySQL-backed ComputeAllocationStore.
func NewComputeAllocationStore(db *sqlx.DB) ComputeAllocationStore {
	return &mysqlComputeAllocationStore{db: db}
}

const computeAllocationColumns = `id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time`

func (s *mysqlComputeAllocationStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocation, error) {
	var a models.ComputeAllocation
	err := s.db.GetContext(ctx, &a,
		`SELECT `+computeAllocationColumns+` FROM compute_allocations WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (s *mysqlComputeAllocationStore) FindByProject(ctx context.Context, projectID string) ([]models.ComputeAllocation, error) {
	var allocs []models.ComputeAllocation
	err := s.db.SelectContext(ctx, &allocs,
		`SELECT `+computeAllocationColumns+` FROM compute_allocations WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, err
	}
	return allocs, nil
}

func (s *mysqlComputeAllocationStore) FindByCluster(ctx context.Context, clusterID string) ([]models.ComputeAllocation, error) {
	var allocs []models.ComputeAllocation
	err := s.db.SelectContext(ctx, &allocs,
		`SELECT `+computeAllocationColumns+` FROM compute_allocations WHERE compute_cluster_id = ?`, clusterID)
	if err != nil {
		return nil, err
	}
	return allocs, nil
}

func (s *mysqlComputeAllocationStore) Create(ctx context.Context, tx *sql.Tx, a *models.ComputeAllocation) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocations (`+computeAllocationColumns+`)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.ProjectID, a.Name, string(a.Status), a.ComputeClusterID, a.InitialSUAmount, a.StartTime, a.EndTime)
	return err
}

func (s *mysqlComputeAllocationStore) Update(ctx context.Context, tx *sql.Tx, a *models.ComputeAllocation) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocations
		 SET project_id = ?, name = ?, status = ?, compute_cluster_id = ?,
		     initial_su_amount = ?, start_time = ?, end_time = ?
		 WHERE id = ?`,
		a.ProjectID, a.Name, string(a.Status), a.ComputeClusterID,
		a.InitialSUAmount, a.StartTime, a.EndTime, a.ID)
	return err
}

func (s *mysqlComputeAllocationStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_allocations WHERE id = ?`, id)
	return err
}
