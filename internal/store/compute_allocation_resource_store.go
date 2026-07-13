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

type mysqlComputeAllocationResourceStore struct {
	db *sqlx.DB
}

// NewComputeAllocationResourceStore returns a MySQL-backed
// ComputeAllocationResourceStore.
func NewComputeAllocationResourceStore(db *sqlx.DB) ComputeAllocationResourceStore {
	return &mysqlComputeAllocationResourceStore{db: db}
}

func (s *mysqlComputeAllocationResourceStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocationResource, error) {
	var r models.ComputeAllocationResource
	err := s.db.GetContext(ctx, &r,
		`SELECT id, name, resource_type, resource_amount, compute_cluster_id FROM compute_allocation_resources WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *mysqlComputeAllocationResourceStore) FindByNameAndCluster(ctx context.Context, name, clusterID string) (*models.ComputeAllocationResource, error) {
	var r models.ComputeAllocationResource
	err := s.db.GetContext(ctx, &r,
		`SELECT id, name, resource_type, resource_amount, compute_cluster_id
		 FROM compute_allocation_resources
		 WHERE name = ? AND compute_cluster_id = ?`, name, clusterID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *mysqlComputeAllocationResourceStore) FindByTypeAndCluster(ctx context.Context, resourceType, clusterID string) ([]models.ComputeAllocationResource, error) {
	var resources []models.ComputeAllocationResource
	err := s.db.SelectContext(ctx, &resources,
		`SELECT id, name, resource_type, resource_amount, compute_cluster_id
		 FROM compute_allocation_resources
		 WHERE resource_type = ? AND compute_cluster_id = ?
		 ORDER BY name`, resourceType, clusterID)
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (s *mysqlComputeAllocationResourceStore) List(ctx context.Context) ([]models.ComputeAllocationResource, error) {
	var resources []models.ComputeAllocationResource
	err := s.db.SelectContext(ctx, &resources,
		`SELECT id, name, resource_type, resource_amount, compute_cluster_id FROM compute_allocation_resources ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (s *mysqlComputeAllocationResourceStore) ListSummaries(ctx context.Context) ([]ComputeAllocationResourceSummary, error) {
	var rows []ComputeAllocationResourceSummary
	err := s.db.SelectContext(ctx, &rows,
		`SELECT r.id, r.name, r.resource_type, r.resource_amount, r.compute_cluster_id,
		        COALESCE(m.allocation_count, 0) AS allocation_count,
		        COALESCE(m.total_allocated, 0)  AS total_allocated,
		        COALESCE(u.total_used_su, 0)    AS total_used_su,
		        COALESCE(rt.rate_count, 0)      AS rate_count
		 FROM compute_allocation_resources r
		 LEFT JOIN (SELECT compute_allocation_resource_id, COUNT(DISTINCT compute_allocation_id) AS allocation_count, SUM(resource_amount) AS total_allocated
		            FROM compute_allocation_resource_mappings
		            GROUP BY compute_allocation_resource_id) m ON m.compute_allocation_resource_id = r.id
		 LEFT JOIN (SELECT compute_allocation_resource_id, SUM(used_su_amount) AS total_used_su
		            FROM compute_allocation_usages
		            GROUP BY compute_allocation_resource_id) u ON u.compute_allocation_resource_id = r.id
		 LEFT JOIN (SELECT compute_allocation_resource_id, COUNT(*) AS rate_count
		            FROM compute_allocation_resource_rates
		            GROUP BY compute_allocation_resource_id) rt ON rt.compute_allocation_resource_id = r.id
		 ORDER BY r.name`)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlComputeAllocationResourceStore) Create(ctx context.Context, tx *sql.Tx, r *models.ComputeAllocationResource) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocation_resources (id, name, resource_type, resource_amount, compute_cluster_id)
		 VALUES (?, ?, ?, ?, ?)`,
		r.ID, r.Name, r.ResourceType, r.ResourceAmount, r.ComputeClusterID)
	return err
}

func (s *mysqlComputeAllocationResourceStore) Update(ctx context.Context, tx *sql.Tx, r *models.ComputeAllocationResource) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocation_resources
		 SET name = ?, resource_type = ?, resource_amount = ?, compute_cluster_id = ?
		 WHERE id = ?`,
		r.Name, r.ResourceType, r.ResourceAmount, r.ComputeClusterID, r.ID)
	return err
}

func (s *mysqlComputeAllocationResourceStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_allocation_resources WHERE id = ?`, id)
	return err
}
