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

type pgComputeAllocationResourceMappingStore struct {
	db *sqlx.DB
}

// NewComputeAllocationResourceMappingStore returns a PostgreSQL-backed
// ComputeAllocationResourceMappingStore.
func NewComputeAllocationResourceMappingStore(db *sqlx.DB) ComputeAllocationResourceMappingStore {
	return &pgComputeAllocationResourceMappingStore{db: db}
}

func (s *pgComputeAllocationResourceMappingStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocationResourceMapping, error) {
	var m models.ComputeAllocationResourceMapping
	err := s.db.GetContext(ctx, &m,
		`SELECT id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time
		 FROM compute_allocation_resource_mappings
		 WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (s *pgComputeAllocationResourceMappingStore) FindByPair(ctx context.Context, allocationID, resourceID string) (*models.ComputeAllocationResourceMapping, error) {
	var m models.ComputeAllocationResourceMapping
	err := s.db.GetContext(ctx, &m,
		`SELECT id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time
		 FROM compute_allocation_resource_mappings
		 WHERE compute_allocation_id = $1 AND compute_allocation_resource_id = $2`,
		allocationID, resourceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (s *pgComputeAllocationResourceMappingStore) FindResourcesByAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationResource, error) {
	var resources []models.ComputeAllocationResource
	err := s.db.SelectContext(ctx, &resources,
		`SELECT r.id, r.name, r.resource_type, r.resource_amount, r.compute_cluster_id
		 FROM compute_allocation_resources r
		 JOIN compute_allocation_resource_mappings m
		     ON m.compute_allocation_resource_id = r.id
		 WHERE m.compute_allocation_id = $1
		 ORDER BY r.name`, allocationID)
	if err != nil {
		return nil, err
	}
	return resources, nil
}

func (s *pgComputeAllocationResourceMappingStore) FindAllocationsByResource(ctx context.Context, resourceID string) ([]models.ComputeAllocation, error) {
	var allocs []models.ComputeAllocation
	err := s.db.SelectContext(ctx, &allocs,
		`SELECT a.id, a.project_id, a.name, a.status, a.compute_cluster_id,
		        a.initial_su_amount, a.start_time, a.end_time
		 FROM compute_allocations a
		 JOIN compute_allocation_resource_mappings m
		     ON m.compute_allocation_id = a.id
		 WHERE m.compute_allocation_resource_id = $1
		 ORDER BY a.name`, resourceID)
	if err != nil {
		return nil, err
	}
	return allocs, nil
}

func (s *pgComputeAllocationResourceMappingStore) Create(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationResourceMapping) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocation_resource_mappings
		     (id, compute_allocation_id, compute_allocation_resource_id, resource_amount, resource_time)
		 VALUES ($1, $2, $3, $4, $5)`,
		m.ID, m.ComputeAllocationID, m.ComputeAllocationResourceID, m.ResourceAmount, m.ResourceTime)
	return err
}

func (s *pgComputeAllocationResourceMappingStore) Update(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationResourceMapping) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocation_resource_mappings
		    SET resource_amount = $1,
		        resource_time   = $2
		  WHERE id = $3`,
		m.ResourceAmount, m.ResourceTime, m.ID)
	return err
}

func (s *pgComputeAllocationResourceMappingStore) DeleteByPair(ctx context.Context, tx *sql.Tx, allocationID, resourceID string) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM compute_allocation_resource_mappings
		 WHERE compute_allocation_id = $1 AND compute_allocation_resource_id = $2`,
		allocationID, resourceID)
	return err
}
