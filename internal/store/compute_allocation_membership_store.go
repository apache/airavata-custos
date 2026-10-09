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
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

const computeAllocationMembershipColumns = "id, compute_allocation_id, user_id, start_time, end_time, membership_status"

// MembershipWithUser is the result shape of the join-based list methods. Role
// is the user's project-wide role (MEMBER without one). Joined user fields stay
// off the core ComputeAllocationMembership entity.
type MembershipWithUser struct {
	models.ComputeAllocationMembership
	Role        string `db:"role"`
	DisplayName string `db:"display_name"`
	Email       string `db:"email"`
	// The member's account on the allocation's cluster; empty and nil without one.
	LocalUsername string     `db:"local_username"`
	ProvisionedAt *time.Time `db:"provisioned_at"`
}

// displayNameSQL is the user's full name, or their email when the name is empty.
const displayNameSQL = `COALESCE(NULLIF(TRIM(u.first_name || ' ' || u.last_name), ''), u.email) AS display_name`

type pgComputeAllocationMembershipStore struct {
	db *sqlx.DB
}

// NewComputeAllocationMembershipStore returns a PostgreSQL-backed
// ComputeAllocationMembershipStore.
func NewComputeAllocationMembershipStore(db *sqlx.DB) ComputeAllocationMembershipStore {
	return &pgComputeAllocationMembershipStore{db: db}
}

func (s *pgComputeAllocationMembershipStore) FindByID(ctx context.Context, id string) (*models.ComputeAllocationMembership, error) {
	var m models.ComputeAllocationMembership
	err := s.db.GetContext(ctx, &m,
		`SELECT `+computeAllocationMembershipColumns+` FROM compute_allocation_memberships WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (s *pgComputeAllocationMembershipStore) FindByPair(ctx context.Context, allocationID, userID string) (*models.ComputeAllocationMembership, error) {
	var m models.ComputeAllocationMembership
	err := s.db.GetContext(ctx, &m,
		`SELECT `+computeAllocationMembershipColumns+`
		 FROM compute_allocation_memberships
		 WHERE compute_allocation_id = $1 AND user_id = $2`, allocationID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (s *pgComputeAllocationMembershipStore) FindByAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationMembership, error) {
	var rows []models.ComputeAllocationMembership
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationMembershipColumns+`
		 FROM compute_allocation_memberships
		 WHERE compute_allocation_id = $1
		 ORDER BY start_time`, allocationID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgComputeAllocationMembershipStore) FindByUser(ctx context.Context, userID string) ([]models.ComputeAllocationMembership, error) {
	var rows []models.ComputeAllocationMembership
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+computeAllocationMembershipColumns+`
		 FROM compute_allocation_memberships
		 WHERE user_id = $1
		 ORDER BY start_time`, userID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgComputeAllocationMembershipStore) Create(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationMembership) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_allocation_memberships
		     (id, compute_allocation_id, user_id, start_time, end_time, membership_status)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		m.ID, m.ComputeAllocationID, m.UserID, m.StartTime, m.EndTime, string(m.MembershipStatus))
	return err
}

func (s *pgComputeAllocationMembershipStore) Update(ctx context.Context, tx *sql.Tx, m *models.ComputeAllocationMembership) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocation_memberships
		    SET compute_allocation_id = $1,
		        user_id               = $2,
		        start_time            = $3,
		        end_time              = $4,
		        membership_status     = $5
		  WHERE id = $6`,
		m.ComputeAllocationID, m.UserID, m.StartTime, m.EndTime, string(m.MembershipStatus), m.ID)
	return err
}

func (s *pgComputeAllocationMembershipStore) ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM compute_allocation_memberships
		 WHERE user_id = $1
		   AND compute_allocation_id IN (
		       SELECT compute_allocation_id FROM (
		           SELECT compute_allocation_id FROM compute_allocation_memberships WHERE user_id = $2
		       ) AS s
		   )`,
		fromUserID, toUserID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_allocation_memberships SET user_id = $1 WHERE user_id = $2`,
		toUserID, fromUserID)
	return err
}

func (s *pgComputeAllocationMembershipStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_allocation_memberships WHERE id = $1`, id)
	return err
}

// FindByAllocationWithUser returns memberships for an allocation joined with
// users so each row carries display_name, email, and the user's project-wide
// role (MEMBER without one).
func (s *pgComputeAllocationMembershipStore) FindByAllocationWithUser(ctx context.Context, allocationID string) ([]MembershipWithUser, error) {
	rows := []MembershipWithUser{}
	err := s.db.SelectContext(ctx, &rows,
		`SELECT m.id, m.compute_allocation_id, m.user_id, m.start_time, m.end_time, m.membership_status,
		        COALESCE(r.role, 'MEMBER') AS role, `+displayNameSQL+`, u.email,
		        COALESCE(cu.local_username, '') AS local_username, cu.provisioned_at
		   FROM compute_allocation_memberships m
		   JOIN compute_allocations a ON a.id = m.compute_allocation_id
		   JOIN users u ON u.id = m.user_id
		   LEFT JOIN project_roles r
		          ON r.project_id = a.project_id AND r.user_id = m.user_id AND r.compute_allocation_id IS NULL
		   LEFT JOIN compute_cluster_users cu
		          ON cu.compute_cluster_id = a.compute_cluster_id AND cu.user_id = m.user_id
		  WHERE m.compute_allocation_id = $1
		  ORDER BY m.start_time`, allocationID)
	return rows, err
}

// ProjectMember is one user on a project: role holders and holders of a
// membership on any of its allocations. Status is ACTIVE when any membership
// is active or the user holds only a role.
type ProjectMember struct {
	ID          string                   `json:"id" db:"id" binding:"required"` // The user's id.
	ProjectID   string                   `json:"project_id" db:"project_id" binding:"required"`
	UserID      string                   `json:"user_id" db:"user_id" binding:"required"`
	Email       string                   `json:"email" db:"email" binding:"required"`
	DisplayName string                   `json:"display_name" db:"display_name" binding:"required"`
	Role        string                   `json:"role" db:"role" binding:"required"`
	Status      string                   `json:"status" db:"status" binding:"required"`
	AddedTime   time.Time                `json:"added_time" db:"added_time" binding:"required"`
	Allocations ProjectMemberAllocations `json:"allocations" db:"allocations" binding:"required"`
}

// ProjectMemberAllocations scans the JSON array of a member's allocations.
type ProjectMemberAllocations []ProjectMemberAllocationRef

type ProjectMemberAllocationRef struct {
	ID   string `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
	Role string `json:"role" binding:"required"`
}

func (a *ProjectMemberAllocations) Scan(src any) error { return json.Unmarshal(src.([]byte), a) }

// FindByProjectWithUser returns the project's members ordered by email; the
// full join adds role holders without a membership.
func (s *pgComputeAllocationMembershipStore) FindByProjectWithUser(ctx context.Context, projectID string) ([]ProjectMember, error) {
	rows := []ProjectMember{}
	err := s.db.SelectContext(ctx, &rows,
		`SELECT u.id, $1 AS project_id, u.id AS user_id, u.email, `+displayNameSQL+`,
		        COALESCE(r.role, 'MEMBER') AS role,
		        CASE WHEN bool_or(m.membership_status = 'ACTIVE') IS FALSE THEN MIN(m.membership_status) ELSE 'ACTIVE' END AS status,
		        LEAST(MIN(m.start_time), r.added_time) AS added_time,
		        COALESCE(json_agg(json_build_object('id', ca.id, 'name', ca.name, 'role', COALESCE(r.role, 'MEMBER')) ORDER BY ca.name)
		          FILTER (WHERE ca.id IS NOT NULL), '[]') AS allocations
		   FROM compute_allocation_memberships m
		   JOIN compute_allocations ca ON ca.id = m.compute_allocation_id AND ca.project_id = $1
		   FULL JOIN (SELECT user_id, role, added_time FROM project_roles
		               WHERE project_id = $1 AND compute_allocation_id IS NULL) r USING (user_id)
		   JOIN users u ON u.id = user_id
		  GROUP BY u.id, r.role, r.added_time
		  ORDER BY u.email`, projectID)
	return rows, err
}
