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

type pgComputeAllocationStore struct {
	db *sqlx.DB
}

// NewComputeAllocationStore returns a PostgreSQL-backed ComputeAllocationStore.
func NewComputeAllocationStore(db *sqlx.DB) ComputeAllocationStore {
	return &pgComputeAllocationStore{db: db}
}

const computeAllocationColumns = `id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time, posix_group`

func (s *pgComputeAllocationStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocation, error) {
	var a models.ComputeAllocation
	err := s.db.GetContext(ctx, &a,
		`SELECT `+computeAllocationColumns+` FROM compute_allocations WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (s *pgComputeAllocationStore) FindByProject(ctx context.Context, projectID string) ([]models.ComputeAllocation, error) {
	var allocs []models.ComputeAllocation
	err := s.db.SelectContext(ctx, &allocs,
		`SELECT `+computeAllocationColumns+` FROM compute_allocations WHERE project_id = $1`, projectID)
	if err != nil {
		return nil, err
	}
	return allocs, nil
}

func (s *pgComputeAllocationStore) FindByCluster(ctx context.Context, clusterID string) ([]models.ComputeAllocation, error) {
	var allocs []models.ComputeAllocation
	err := s.db.SelectContext(ctx, &allocs,
		`SELECT `+computeAllocationColumns+` FROM compute_allocations WHERE compute_cluster_id = $1`, clusterID)
	if err != nil {
		return nil, err
	}
	return allocs, nil
}

func (s *pgComputeAllocationStore) Create(ctx context.Context, tx *sql.Tx, a *models.ComputeAllocation) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocations (`+computeAllocationColumns+`)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		a.ID, a.ProjectID, a.Name, string(a.Status), a.ComputeClusterID, a.InitialSUAmount, a.StartTime, a.EndTime, a.PosixGroup)
	return err
}

func (s *pgComputeAllocationStore) Update(ctx context.Context, tx *sql.Tx, a *models.ComputeAllocation) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocations
		 SET project_id = $1, name = $2, status = $3, compute_cluster_id = $4,
		     initial_su_amount = $5, start_time = $6, end_time = $7
		 WHERE id = $8`,
		a.ProjectID, a.Name, string(a.Status), a.ComputeClusterID,
		a.InitialSUAmount, a.StartTime, a.EndTime, a.ID)
	return err
}

func (s *pgComputeAllocationStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_allocations WHERE id = $1`, id)
	return err
}

// FindByParticipant returns the allocations where the user holds an active
// membership, or a governance role on the parent project. Ordered like List
// (start_time descending).
func (s *pgComputeAllocationStore) FindByParticipant(ctx context.Context, userID string) ([]models.ComputeAllocation, error) {
	var rows []models.ComputeAllocation
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationColumns+` FROM compute_allocations a
		  WHERE EXISTS (SELECT 1 FROM project_roles r
		                 WHERE r.user_id = $1 AND r.project_id = a.project_id
		                   AND (r.compute_allocation_id IS NULL OR r.compute_allocation_id = a.id))
		  ORDER BY start_time DESC`, userID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// RoleForUser returns the user's role on the allocation, a project-wide role
// before MEMBER, or "" for none.
func (s *pgComputeAllocationStore) RoleForUser(ctx context.Context, allocationID, userID string) (models.ProjectRole, error) {
	var role models.ProjectRole
	err := s.db.GetContext(ctx, &role,
		`SELECT r.role FROM compute_allocations a
		   JOIN project_roles r ON r.project_id = a.project_id AND r.user_id = $2
		                       AND (r.compute_allocation_id IS NULL OR r.compute_allocation_id = a.id)
		  WHERE a.id = $1
		  ORDER BY r.compute_allocation_id NULLS FIRST LIMIT 1`, allocationID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return role, err
}

func (s *pgComputeAllocationStore) List(ctx context.Context, f AllocationListFilter) ([]models.ComputeAllocation, int, error) {
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
		where = append(where, `name ILIKE ?`)
		args = append(args, "%"+f.Query+"%")
	}
	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}
	var total int
	if err := s.db.GetContext(ctx, &total, s.db.Rebind(`SELECT COUNT(*) FROM compute_allocations`+clause), args...); err != nil {
		return nil, 0, err
	}
	var rows []models.ComputeAllocation
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(`SELECT `+computeAllocationColumns+` FROM compute_allocations`+clause+` ORDER BY start_time DESC LIMIT ? OFFSET ?`),
		append(args, PageLimit(f.Limit), max(f.Offset, 0))...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
