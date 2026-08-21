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
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

const computeAllocationResourceRateColumns = "id, compute_allocation_resource_id, rate, start_time, end_time"

type pgComputeAllocationResourceRateStore struct {
	db *sqlx.DB
}

// NewComputeAllocationResourceRateStore returns a PostgreSQL-backed
// ComputeAllocationResourceRateStore.
func NewComputeAllocationResourceRateStore(db *sqlx.DB) ComputeAllocationResourceRateStore {
	return &pgComputeAllocationResourceRateStore{db: db}
}

func (s *pgComputeAllocationResourceRateStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocationResourceRate, error) {
	var r models.ComputeAllocationResourceRate
	err := s.db.GetContext(ctx, &r,
		`SELECT `+computeAllocationResourceRateColumns+`
		 FROM compute_allocation_resource_rates WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *pgComputeAllocationResourceRateStore) FindByResource(ctx context.Context, resourceID string) ([]models.ComputeAllocationResourceRate, error) {
	var rates []models.ComputeAllocationResourceRate
	err := s.db.SelectContext(ctx, &rates,
		`SELECT `+computeAllocationResourceRateColumns+`
		 FROM compute_allocation_resource_rates
		 WHERE compute_allocation_resource_id = $1
		 ORDER BY start_time`, resourceID)
	if err != nil {
		return nil, err
	}
	return rates, nil
}

func (s *pgComputeAllocationResourceRateStore) FindEffective(ctx context.Context, resourceID string, at time.Time) (*models.ComputeAllocationResourceRate, error) {
	var r models.ComputeAllocationResourceRate
	err := s.db.GetContext(ctx, &r,
		`SELECT `+computeAllocationResourceRateColumns+`
		 FROM compute_allocation_resource_rates
		 WHERE compute_allocation_resource_id = $1
		   AND start_time <= $2
		   AND end_time   >  $3
		 ORDER BY start_time DESC
		 LIMIT 1`, resourceID, at, at)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *pgComputeAllocationResourceRateStore) Create(ctx context.Context, tx *sql.Tx, r *models.ComputeAllocationResourceRate) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocation_resource_rates
		     (id, compute_allocation_resource_id, rate, start_time, end_time)
		 VALUES ($1, $2, $3, $4, $5)`,
		r.ID, r.ComputeAllocationResourceID, r.Rate, r.StartTime, r.EndTime)
	return err
}

func (s *pgComputeAllocationResourceRateStore) Update(ctx context.Context, tx *sql.Tx, r *models.ComputeAllocationResourceRate) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocation_resource_rates
		 SET rate = $1, start_time = $2, end_time = $3
		 WHERE id = $4`,
		r.Rate, r.StartTime, r.EndTime, r.ID)
	return err
}

func (s *pgComputeAllocationResourceRateStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_allocation_resource_rates WHERE id = $1`, id)
	return err
}
