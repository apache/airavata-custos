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
	"strings"

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

// FindByParticipant returns the allocations where the user holds an active
// membership, or a governance role on the parent project. Ordered like List
// (start_time descending).
func (s *mysqlComputeAllocationStore) FindByParticipant(ctx context.Context, userID string) ([]models.ComputeAllocation, error) {
	var rows []models.ComputeAllocation
	query, args, err := sqlx.In(
		`SELECT `+computeAllocationColumns+` FROM compute_allocations a
		  WHERE EXISTS (SELECT 1 FROM compute_allocation_memberships cam
		                 WHERE cam.compute_allocation_id = a.id AND cam.user_id = ?
		                   AND cam.membership_status = 'ACTIVE')
		     OR EXISTS (SELECT 1 FROM project_memberships pm
		                 WHERE pm.project_id = a.project_id AND pm.user_id = ?
		                   AND pm.role IN (?))
		  ORDER BY start_time DESC`, userID, userID, models.GovernanceRoles)
	if err != nil {
		return nil, err
	}
	err = s.db.SelectContext(ctx, &rows, s.db.Rebind(query), args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlComputeAllocationStore) List(ctx context.Context, f AllocationListFilter) ([]models.ComputeAllocation, int, error) {
	where := []string{}
	args := []any{}
	if f.ProjectID != "" {
		where = append(where, `project_id = ?`)
		args = append(args, f.ProjectID)
	}
	if f.Status != "" {
		where = append(where, `status = ?`)
		args = append(args, f.Status)
	}
	if f.Query != "" {
		where = append(where, `name LIKE ?`)
		args = append(args, "%"+f.Query+"%")
	}
	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}
	var total int
	if err := s.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM compute_allocations`+clause, args...); err != nil {
		return nil, 0, err
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	query := `SELECT ` + computeAllocationColumns + ` FROM compute_allocations` + clause +
		` ORDER BY start_time DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	var rows []models.ComputeAllocation
	if err := s.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
