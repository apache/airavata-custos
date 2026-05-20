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

type mysqlComputeAllocationResourceMappingStore struct {
	db *sqlx.DB
}

// NewComputeAllocationResourceMappingStore returns a MySQL-backed
// ComputeAllocationResourceMappingStore.
func NewComputeAllocationResourceMappingStore(db *sqlx.DB) ComputeAllocationResourceMappingStore {
	return &mysqlComputeAllocationResourceMappingStore{db: db}
}

func (s *mysqlComputeAllocationResourceMappingStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocationResourceMapping, error) {
	var m models.ComputeAllocationResourceMapping
	err := s.db.GetContext(ctx, &m,
		`SELECT id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time
		 FROM compute_allocation_resource_mappings
		 WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (s *mysqlComputeAllocationResourceMappingStore) FindByPair(ctx context.Context, allocationID, resourceID string) (*models.ComputeAllocationResourceMapping, error) {
	var m models.ComputeAllocationResourceMapping
	err := s.db.GetContext(ctx, &m,
		`SELECT id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time
		 FROM compute_allocation_resource_mappings
		 WHERE compute_allocation_id = ? AND compute_allocation_resource_id = ?`,
		allocationID, resourceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (s *mysqlComputeAllocationResourceMappingStore) FindResourcesByAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationResource, error) {
	var resources []models.ComputeAllocationResource
	err := s.db.SelectContext(ctx, &resources,
		`SELECT r.id, r.name, r.resource_type, r.resource_amount
		 FROM compute_allocation_resources r
		 JOIN compute_allocation_resource_mappings m
		     ON m.compute_allocation_resource_id = r.id
		 WHERE m.compute_allocation_id = ?
		 ORDER BY r.name`, allocationID)
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (s *mysqlComputeAllocationResourceMappingStore) FindAllocationsByResource(ctx context.Context, resourceID string) ([]models.ComputeAllocation, error) {
	var allocs []models.ComputeAllocation
	err := s.db.SelectContext(ctx, &allocs,
		`SELECT a.id, a.project_id, a.name, a.status, a.compute_cluster_id,
		        a.initial_su_amount, a.start_time, a.end_time
		 FROM compute_allocations a
		 JOIN compute_allocation_resource_mappings m
		     ON m.compute_allocation_id = a.id
		 WHERE m.compute_allocation_resource_id = ?
		 ORDER BY a.name`, resourceID)
	if err != nil {
		return nil, err
	}
	return allocs, nil
}

func (s *mysqlComputeAllocationResourceMappingStore) Create(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationResourceMapping) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocation_resource_mappings
		     (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time)
		 VALUES (?, ?, ?, ?, ?)`,
		m.ID, m.ComputeAllocationID, m.ComputeAllocationResourceID, m.ResourceAmount, m.ResourceTime)
	return err
}

func (s *mysqlComputeAllocationResourceMappingStore) Update(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationResourceMapping) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocation_resource_mappings
		    SET resource_amount = ?,
		        resource_time   = ?
		  WHERE id = ?`,
		m.ResourceAmount, m.ResourceTime, m.ID)
	return err
}

func (s *mysqlComputeAllocationResourceMappingStore) DeleteByPair(ctx context.Context, tx *sql.Tx, allocationID, resourceID string) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM compute_allocation_resource_mappings
		 WHERE compute_allocation_id = ? AND compute_allocation_resource_id = ?`,
		allocationID, resourceID)
	return err
}
