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

const computeAllocationMembershipColumns = "id, compute_allocation_id, user_id, start_time, end_time, membership_status"

// MembershipWithUser is the result shape of the join-based list methods. Role
// is project-level and comes from project_memberships (COALESCE'd to MEMBER
// when no row exists). Joined user + allocation fields stay off the core
// ComputeAllocationMembership entity.
type MembershipWithUser struct {
	models.ComputeAllocationMembership
	Role           string
	DisplayName    string
	Email          string
	AllocationName string
}

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
		     (id, compute_allocation_id, user_id, start_time, end_time, membership_status)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		m.ID, m.ComputeAllocationID, m.UserID, m.StartTime, m.EndTime, string(m.MembershipStatus))
	return err
}

func (s *mysqlComputeAllocationMembershipStore) Update(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationMembership) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocation_memberships
		    SET compute_allocation_id = ?,
		        user_id               = ?,
		        start_time            = ?,
		        end_time              = ?,
		        membership_status     = ?
		  WHERE id = ?`,
		m.ComputeAllocationID, m.UserID, m.StartTime, m.EndTime, string(m.MembershipStatus), m.ID)
	return err
}

func (s *mysqlComputeAllocationMembershipStore) ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM compute_allocation_memberships
		 WHERE user_id = ?
		   AND compute_allocation_id IN (
		       SELECT compute_allocation_id FROM (
		           SELECT compute_allocation_id FROM compute_allocation_memberships WHERE user_id = ?
		       ) AS s
		   )`,
		fromUserID, toUserID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocation_memberships SET user_id = ? WHERE user_id = ?`,
		toUserID, fromUserID)
	return err
}

func (s *mysqlComputeAllocationMembershipStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_allocation_memberships WHERE id = ?`, id)
	return err
}

// FindByAllocationWithUser returns memberships for an allocation joined with
// users + project_memberships so each row carries display_name, email, and
// the project-level role (defaulted to MEMBER when no project_memberships
// row exists).
func (s *mysqlComputeAllocationMembershipStore) FindByAllocationWithUser(ctx context.Context, allocationID string) ([]MembershipWithUser, error) {
	type row struct {
		models.ComputeAllocationMembership
		Role      string `db:"role"`
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		UserEmail string `db:"user_email"`
	}
	var rows []row
	err := s.db.SelectContext(ctx, &rows,
		`SELECT m.id, m.compute_allocation_id, m.user_id, m.start_time, m.end_time,
		        m.membership_status,
		        COALESCE(pm.role, 'MEMBER') AS role,
		        u.first_name, u.last_name, u.email AS user_email
		   FROM compute_allocation_memberships m
		   JOIN compute_allocations a    ON a.id = m.compute_allocation_id
		   JOIN users u                  ON u.id = m.user_id
		   LEFT JOIN project_memberships pm
		         ON pm.project_id = a.project_id AND pm.user_id = m.user_id
		  WHERE m.compute_allocation_id = ?
		  ORDER BY m.start_time`, allocationID)
	if err != nil {
		return nil, err
	}
	out := make([]MembershipWithUser, 0, len(rows))
	for _, r := range rows {
		out = append(out, MembershipWithUser{
			ComputeAllocationMembership: r.ComputeAllocationMembership,
			Role:                        r.Role,
			DisplayName:                 displayName(r.FirstName, r.LastName, r.UserEmail),
			Email:                       r.UserEmail,
		})
	}
	return out, nil
}

// FindByProjectWithUser returns every membership across every allocation in
// the project, joined with the user, the owning allocation name, and the
// project-level role. Returned rows are ordered by user; callers aggregate it
// per-user (collapsing into the response's allocation list).
func (s *mysqlComputeAllocationMembershipStore) FindByProjectWithUser(ctx context.Context, projectID string) ([]MembershipWithUser, error) {
	type row struct {
		models.ComputeAllocationMembership
		Role           string `db:"role"`
		FirstName      string `db:"first_name"`
		LastName       string `db:"last_name"`
		UserEmail      string `db:"user_email"`
		AllocationName string `db:"allocation_name"`
	}
	var rows []row
	err := s.db.SelectContext(ctx, &rows,
		`SELECT m.id, m.compute_allocation_id, m.user_id, m.start_time, m.end_time,
		        m.membership_status,
		        COALESCE(pm.role, 'MEMBER') AS role,
		        u.first_name, u.last_name, u.email AS user_email,
		        ca.name AS allocation_name
		   FROM compute_allocation_memberships m
		   JOIN compute_allocations ca   ON ca.id = m.compute_allocation_id
		   JOIN users u                  ON u.id = m.user_id
		   LEFT JOIN project_memberships pm
		         ON pm.project_id = ca.project_id AND pm.user_id = m.user_id
		  WHERE ca.project_id = ?
		  ORDER BY u.email, ca.name`, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]MembershipWithUser, 0, len(rows))
	for _, r := range rows {
		out = append(out, MembershipWithUser{
			ComputeAllocationMembership: r.ComputeAllocationMembership,
			Role:                        r.Role,
			DisplayName:                 displayName(r.FirstName, r.LastName, r.UserEmail),
			Email:                       r.UserEmail,
			AllocationName:              r.AllocationName,
		})
	}
	return out, nil
}

func displayName(first, last, email string) string {
	full := ""
	if first != "" || last != "" {
		full = first
		if first != "" && last != "" {
			full += " "
		}
		full += last
	}
	if full == "" {
		return email
	}
	return full
}
