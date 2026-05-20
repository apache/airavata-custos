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

type mysqlComputeAllocationMembershipResourceOverrideStore struct {
	db *sqlx.DB
}

// NewComputeAllocationMembershipResourceOverrideStore returns a MySQL-backed
// ComputeAllocationMembershipResourceOverrideStore.
func NewComputeAllocationMembershipResourceOverrideStore(db *sqlx.DB) ComputeAllocationMembershipResourceOverrideStore {
	return &mysqlComputeAllocationMembershipResourceOverrideStore{db: db}
}

func (s *mysqlComputeAllocationMembershipResourceOverrideStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocationMembershipResourceOverride, error) {
	var o models.ComputeAllocationMembershipResourceOverride
	err := s.db.GetContext(ctx, &o,
		`SELECT `+computeAllocationMembershipResourceOverrideColumns+`
           FROM compute_allocation_membership_resource_overrides
          WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (s *mysqlComputeAllocationMembershipResourceOverrideStore) FindByPair(ctx context.Context, membershipID, resourceID string) (*models.ComputeAllocationMembershipResourceOverride, error) {
	var o models.ComputeAllocationMembershipResourceOverride
	err := s.db.GetContext(ctx, &o,
		`SELECT `+computeAllocationMembershipResourceOverrideColumns+`
           FROM compute_allocation_membership_resource_overrides
          WHERE compute_allocation_membership_id = ? AND compute_allocation_resource_id = ?`,
		membershipID, resourceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (s *mysqlComputeAllocationMembershipResourceOverrideStore) FindByMembership(ctx context.Context, membershipID string) ([]models.ComputeAllocationMembershipResourceOverride, error) {
	var rows []models.ComputeAllocationMembershipResourceOverride
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationMembershipResourceOverrideColumns+`
           FROM compute_allocation_membership_resource_overrides
          WHERE compute_allocation_membership_id = ?
          ORDER BY compute_allocation_resource_id`, membershipID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlComputeAllocationMembershipResourceOverrideStore) FindByResource(ctx context.Context, resourceID string) ([]models.ComputeAllocationMembershipResourceOverride, error) {
	var rows []models.ComputeAllocationMembershipResourceOverride
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationMembershipResourceOverrideColumns+`
           FROM compute_allocation_membership_resource_overrides
          WHERE compute_allocation_resource_id = ?
          ORDER BY compute_allocation_membership_id`, resourceID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlComputeAllocationMembershipResourceOverrideStore) Create(ctx context.Context, tx *sql.Tx, o *models.ComputeAllocationMembershipResourceOverride) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocation_membership_resource_overrides
             (id, compute_allocation_membership_id, compute_allocation_resource_id, override_resource_amount, override_resource_time)
         VALUES (?, ?, ?, ?, ?)`,
		o.ID, o.ComputeAllocationMembershipID, o.ComputeAllocationResourceID, o.OverrideResourceAmount, o.OverrideResourceTime)
	return err
}

func (s *mysqlComputeAllocationMembershipResourceOverrideStore) Update(ctx context.Context, tx *sql.Tx, o *models.ComputeAllocationMembershipResourceOverride) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocation_membership_resource_overrides
            SET compute_allocation_membership_id = ?,
                compute_allocation_resource_id   = ?,
                override_resource_amount         = ?,
                override_resource_time           = ?
          WHERE id = ?`,
		o.ComputeAllocationMembershipID, o.ComputeAllocationResourceID, o.OverrideResourceAmount, o.OverrideResourceTime, o.ID)
	return err
}

func (s *mysqlComputeAllocationMembershipResourceOverrideStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM compute_allocation_membership_resource_overrides WHERE id = ?`, id)
	return err
}
