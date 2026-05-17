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

type mysqlComputeAllocationResourceRateStore struct {
	db *sqlx.DB
}

// NewComputeAllocationResourceRateStore returns a MySQL-backed
// ComputeAllocationResourceRateStore.
func NewComputeAllocationResourceRateStore(db *sqlx.DB) ComputeAllocationResourceRateStore {
	return &mysqlComputeAllocationResourceRateStore{db: db}
}

func (s *mysqlComputeAllocationResourceRateStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocationResourceRate, error) {
	var r models.ComputeAllocationResourceRate
	err := s.db.GetContext(ctx, &r,
		`SELECT `+computeAllocationResourceRateColumns+`
		 FROM compute_allocation_resource_rates WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

func (s *mysqlComputeAllocationResourceRateStore) FindByResource(ctx context.Context, resourceID string) ([]models.ComputeAllocationResourceRate, error) {
	var rates []models.ComputeAllocationResourceRate
	err := s.db.SelectContext(ctx, &rates,
		`SELECT `+computeAllocationResourceRateColumns+`
		 FROM compute_allocation_resource_rates
		 WHERE compute_allocation_resource_id = ?
		 ORDER BY start_time`, resourceID)
	if err != nil {
		return nil, err
	}
	return rates, nil
}

func (s *mysqlComputeAllocationResourceRateStore) FindEffective(ctx context.Context, resourceID string, at time.Time) (*models.ComputeAllocationResourceRate, error) {
	var r models.ComputeAllocationResourceRate
	err := s.db.GetContext(ctx, &r,
		`SELECT `+computeAllocationResourceRateColumns+`
		 FROM compute_allocation_resource_rates
		 WHERE compute_allocation_resource_id = ?
		   AND start_time <= ?
		   AND end_time   >  ?
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

func (s *mysqlComputeAllocationResourceRateStore) Create(ctx context.Context, tx *sql.Tx, r *models.ComputeAllocationResourceRate) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocation_resource_rates
		     (id, compute_allocation_resource_id, rate, start_time, end_time)
		 VALUES (?, ?, ?, ?, ?)`,
		r.ID, r.ComputeAllocationResourceID, r.Rate, r.StartTime, r.EndTime)
	return err
}

func (s *mysqlComputeAllocationResourceRateStore) Update(ctx context.Context, tx *sql.Tx, r *models.ComputeAllocationResourceRate) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocation_resource_rates
		 SET rate = ?, start_time = ?, end_time = ?
		 WHERE id = ?`,
		r.Rate, r.StartTime, r.EndTime, r.ID)
	return err
}

func (s *mysqlComputeAllocationResourceRateStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_allocation_resource_rates WHERE id = ?`, id)
	return err
}
