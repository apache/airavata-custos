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

const computeAllocationMembershipResourceOverrideColumns = "id, compute_allocation_membership_id, compute_allocation_resource_id, override_resource_amount, override_resource_time"

type pgComputeAllocationMembershipResourceOverrideStore struct {
	db *sqlx.DB
}

// NewComputeAllocationMembershipResourceOverrideStore returns a PostgreSQL-backed
// ComputeAllocationMembershipResourceOverrideStore.
func NewComputeAllocationMembershipResourceOverrideStore(db *sqlx.DB) ComputeAllocationMembershipResourceOverrideStore {
	return &pgComputeAllocationMembershipResourceOverrideStore{db: db}
}

func (s *pgComputeAllocationMembershipResourceOverrideStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocationMembershipResourceOverride, error) {
	var o models.ComputeAllocationMembershipResourceOverride
	err := s.db.GetContext(ctx, &o,
		`SELECT `+computeAllocationMembershipResourceOverrideColumns+`
           FROM compute_allocation_membership_resource_overrides
          WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (s *pgComputeAllocationMembershipResourceOverrideStore) FindByPair(ctx context.Context, membershipID, resourceID string) (*models.ComputeAllocationMembershipResourceOverride, error) {
	var o models.ComputeAllocationMembershipResourceOverride
	err := s.db.GetContext(ctx, &o,
		`SELECT `+computeAllocationMembershipResourceOverrideColumns+`
           FROM compute_allocation_membership_resource_overrides
          WHERE compute_allocation_membership_id = $1 AND compute_allocation_resource_id = $2`,
		membershipID, resourceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (s *pgComputeAllocationMembershipResourceOverrideStore) FindByMembership(ctx context.Context, membershipID string) ([]models.ComputeAllocationMembershipResourceOverride, error) {
	var rows []models.ComputeAllocationMembershipResourceOverride
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationMembershipResourceOverrideColumns+`
           FROM compute_allocation_membership_resource_overrides
          WHERE compute_allocation_membership_id = $1
          ORDER BY compute_allocation_resource_id`, membershipID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgComputeAllocationMembershipResourceOverrideStore) FindByResource(ctx context.Context, resourceID string) ([]models.ComputeAllocationMembershipResourceOverride, error) {
	var rows []models.ComputeAllocationMembershipResourceOverride
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationMembershipResourceOverrideColumns+`
           FROM compute_allocation_membership_resource_overrides
          WHERE compute_allocation_resource_id = $1
          ORDER BY compute_allocation_membership_id`, resourceID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgComputeAllocationMembershipResourceOverrideStore) Create(ctx context.Context, tx *sql.Tx, o *models.ComputeAllocationMembershipResourceOverride) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocation_membership_resource_overrides
             (id, compute_allocation_membership_id, compute_allocation_resource_id, override_resource_amount, override_resource_time)
         VALUES ($1, $2, $3, $4, $5)`,
		o.ID, o.ComputeAllocationMembershipID, o.ComputeAllocationResourceID, o.OverrideResourceAmount, o.OverrideResourceTime)
	return err
}

func (s *pgComputeAllocationMembershipResourceOverrideStore) Update(ctx context.Context, tx *sql.Tx, o *models.ComputeAllocationMembershipResourceOverride) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocation_membership_resource_overrides
            SET compute_allocation_membership_id = $1,
                compute_allocation_resource_id   = $2,
                override_resource_amount         = $3,
                override_resource_time           = $4
          WHERE id = $5`,
		o.ComputeAllocationMembershipID, o.ComputeAllocationResourceID, o.OverrideResourceAmount, o.OverrideResourceTime, o.ID)
	return err
}

func (s *pgComputeAllocationMembershipResourceOverrideStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM compute_allocation_membership_resource_overrides WHERE id = $1`, id)
	return err
}
