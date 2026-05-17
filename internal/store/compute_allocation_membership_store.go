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

const computeAllocationMembershipColumns = "id, compute_allocation_id, user_id, allocation_amount, start_time, end_time, membership_status"

type mysqlComputeAllocationMembershipStore struct {
	db *sqlx.DB
}

// NewComputeAllocationMembershipStore returns a MySQL-backed
// ComputeAllocationMembershipStore.
func NewComputeAllocationMembershipStore(db *sqlx.DB) ComputeAllocationMembershipStore {
	return &mysqlComputeAllocationMembershipStore{db: db}
}

func (s *mysqlComputeAllocationMembershipStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocationMembership, error) {
	var m models.ComputeAllocationMembership
	err := s.db.GetContext(ctx, &m,
		`SELECT `+computeAllocationMembershipColumns+` FROM compute_allocation_memberships WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (s *mysqlComputeAllocationMembershipStore) FindByPair(ctx context.Context, allocationID, userID string) (*models.ComputeAllocationMembership, error) {
	var m models.ComputeAllocationMembership
	err := s.db.GetContext(ctx, &m,
		`SELECT `+computeAllocationMembershipColumns+`
		 FROM compute_allocation_memberships
		 WHERE compute_allocation_id = ? AND user_id = ?`, allocationID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (s *mysqlComputeAllocationMembershipStore) FindByAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationMembership, error) {
	var rows []models.ComputeAllocationMembership
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationMembershipColumns+`
		 FROM compute_allocation_memberships
		 WHERE compute_allocation_id = ?
		 ORDER BY start_time`, allocationID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlComputeAllocationMembershipStore) FindByUser(ctx context.Context, userID string) ([]models.ComputeAllocationMembership, error) {
	var rows []models.ComputeAllocationMembership
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationMembershipColumns+`
		 FROM compute_allocation_memberships
		 WHERE user_id = ?
		 ORDER BY start_time`, userID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlComputeAllocationMembershipStore) Create(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationMembership) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocation_memberships
		     (id, compute_allocation_id, user_id, allocation_amount, start_time, end_time, membership_status)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.ComputeAllocationID, m.UserID, m.AllocationAmount, m.StartTime, m.EndTime, string(m.MembershipStatus))
	return err
}

func (s *mysqlComputeAllocationMembershipStore) Update(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationMembership) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocation_memberships
		    SET compute_allocation_id = ?,
		        user_id               = ?,
		        allocation_amount     = ?,
		        start_time            = ?,
		        end_time              = ?,
		        membership_status     = ?
		  WHERE id = ?`,
		m.ComputeAllocationID, m.UserID, m.AllocationAmount, m.StartTime, m.EndTime, string(m.MembershipStatus), m.ID)
	return err
}

func (s *mysqlComputeAllocationMembershipStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_allocation_memberships WHERE id = ?`, id)
	return err
}
