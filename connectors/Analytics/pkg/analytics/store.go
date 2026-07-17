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

package analytics

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// ProjectRow is a project the caller is involved with plus the
// caller's governance role on it. Role is null when the caller has only
// allocation membership (a plain member).
type ProjectRow struct {
	ProjectID string         `db:"id"`
	Title     string         `db:"title"`
	Role      sql.NullString `db:"role"`
}

// AllocationRow is one allocation with its consumed credits, tagged
// with its project so callers can group across projects in one query.
type AllocationRow struct {
	ProjectID       string    `db:"project_id"`
	ID              string    `db:"id"`
	Name            string    `db:"name"`
	Status          string    `db:"status"`
	InitialSUAmount int64     `db:"initial_su_amount"`
	UsedSUAmount    float64   `db:"used_su_amount"`
	StartTime       time.Time `db:"start_time"`
	EndTime         time.Time `db:"end_time"`
}

// DailyRow is credits consumed on one day against one resource.
type DailyRow struct {
	Day        time.Time `db:"day"`
	ResourceID string    `db:"resource_id"`
	Credits    float64   `db:"credits"`
}

// ResourceRow aggregates one resource's consumption across an
// allocation, including the caller's own slice.
type ResourceRow struct {
	ResourceID   string  `db:"id"`
	Name         string  `db:"name"`
	ResourceType string  `db:"resource_type"`
	Used         float64 `db:"used"`
	UsedNative   float64 `db:"used_native"`
	UsedByCaller float64 `db:"used_by_caller"`
}

// MemberRow is one member's consumption against an allocation.
type MemberRow struct {
	UserID string  `db:"user_id"`
	Name   string  `db:"name"`
	Used   float64 `db:"used"`
}

// JobRow is one usage record (a job's charge) against an allocation.
type JobRow struct {
	ID             string    `db:"id"`
	JobID          string    `db:"job_id"`
	CalculatedTime time.Time `db:"calculated_time"`
	UserID         string    `db:"user_id"`
	UserName       string    `db:"user_name"`
	ResourceID     string    `db:"resource_id"`
	ResourceName   string    `db:"resource_name"`
	ResourceType   string    `db:"resource_type"`
	UsedRawAmount  float64   `db:"used_raw_amount"`
	UsedSUAmount   float64   `db:"used_su_amount"`
}

// Store aggregates usage for the analytics endpoints. SUs round to whole
// credits because used_su_amount is a DOUBLE.
type Store interface {
	// ProjectsForUser returns the projects the user touches through a governance
	// role or an active allocation membership. Role is null for a plain member.
	ProjectsForUser(ctx context.Context, userID string) ([]ProjectRow, error)
	// AllocationsForProjects returns allocations the caller may see in the given
	// projects: all allocations where they hold a governance role, plus any
	// allocation they are an active member of. Empty projectIDs returns no rows.
	AllocationsForProjects(ctx context.Context, projectIDs []string, userID string) ([]AllocationRow, error)
	// TotalUsed returns the rounded total credits consumed against an allocation.
	TotalUsed(ctx context.Context, allocationID string) (float64, error)
	// DailyUsage returns per-day, per-resource credits for one allocation.
	DailyUsage(ctx context.Context, allocationID string) ([]DailyRow, error)
	// ResourceUsage returns per-resource credits (and the caller's slice) for
	// one allocation, over resources that carry usage.
	ResourceUsage(ctx context.Context, allocationID, callerID string) ([]ResourceRow, error)
	// MemberUsage returns per-member credits for one allocation, ranked by
	// consumption. Callers gate this behind a role check.
	MemberUsage(ctx context.Context, allocationID string) ([]MemberRow, error)
	// Jobs returns a page of usage records (newest first) for one allocation
	// and the total count. A non-nil userID restricts to that user's records.
	Jobs(ctx context.Context, allocationID string, userID *string, limit, offset int) ([]JobRow, int, error)
	// ProjectRole returns the caller's governance role on a project, or an empty
	// string when they hold none.
	ProjectRole(ctx context.Context, projectID, userID string) (string, error)
	// IsActiveMember reports whether the user holds an active membership on the
	// allocation.
	IsActiveMember(ctx context.Context, allocationID, userID string) (bool, error)
}

type mysqlStore struct {
	db *sqlx.DB
}

// NewStore returns a MySQL-backed analytics Store.
func NewStore(db *sqlx.DB) Store {
	return &mysqlStore{db: db}
}

func (s *mysqlStore) ProjectsForUser(ctx context.Context, userID string) ([]ProjectRow, error) {
	var rows []ProjectRow
	err := s.db.SelectContext(ctx, &rows,
		`SELECT p.id, p.title, pm.role
		   FROM projects p
		   LEFT JOIN project_memberships pm
		     ON pm.project_id = p.id AND pm.user_id = ?
		  WHERE p.id IN (
		        SELECT project_id FROM project_memberships WHERE user_id = ?
		        UNION
		        SELECT ca.project_id
		          FROM compute_allocation_memberships cam
		          JOIN compute_allocations ca ON ca.id = cam.compute_allocation_id
		         WHERE cam.user_id = ? AND cam.membership_status = 'ACTIVE'
		  )
		  ORDER BY p.title`, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlStore) AllocationsForProjects(ctx context.Context, projectIDs []string, userID string) ([]AllocationRow, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	// Members see only allocations they actively belong to, so the caller is
	// never offered an allocation whose usage-summary would 404.
	query, args, err := sqlx.In(
		`SELECT ca.project_id, ca.id, ca.name, ca.status,
		        ca.initial_su_amount, ca.start_time, ca.end_time,
		        COALESCE(SUM(u.used_su_amount), 0) AS used_su_amount
		   FROM compute_allocations ca
		   LEFT JOIN compute_allocation_usages u
		     ON u.compute_allocation_id = ca.id
		  WHERE ca.project_id IN (?)
		    AND (
		        EXISTS (SELECT 1 FROM project_memberships pm
		                 WHERE pm.project_id = ca.project_id AND pm.user_id = ?
		                   AND pm.role IN ('PI', 'CO_PI', 'ALLOCATION_MANAGER'))
		        OR EXISTS (SELECT 1 FROM compute_allocation_memberships cam
		                    WHERE cam.compute_allocation_id = ca.id AND cam.user_id = ?
		                      AND cam.membership_status = 'ACTIVE')
		    )
		  GROUP BY ca.project_id, ca.id, ca.name, ca.status,
		           ca.initial_su_amount, ca.start_time, ca.end_time
		  ORDER BY ca.start_time`, projectIDs, userID, userID)
	if err != nil {
		return nil, err
	}
	var rows []AllocationRow
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(query), args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlStore) TotalUsed(ctx context.Context, allocationID string) (float64, error) {
	var total float64
	err := s.db.GetContext(ctx, &total,
		`SELECT COALESCE(SUM(used_su_amount), 0)
		   FROM compute_allocation_usages
		  WHERE compute_allocation_id = ?`, allocationID)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (s *mysqlStore) DailyUsage(ctx context.Context, allocationID string) ([]DailyRow, error) {
	var rows []DailyRow
	err := s.db.SelectContext(ctx, &rows,
		`SELECT DATE(u.calculated_time) AS day,
		        u.compute_allocation_resource_id AS resource_id,
		        SUM(u.used_su_amount) AS credits
		   FROM compute_allocation_usages u
		  WHERE u.compute_allocation_id = ?
		  GROUP BY day, resource_id
		  ORDER BY day`, allocationID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlStore) ResourceUsage(ctx context.Context, allocationID, callerID string) ([]ResourceRow, error) {
	var rows []ResourceRow
	err := s.db.SelectContext(ctx, &rows,
		`SELECT r.id, r.name, r.resource_type,
		        SUM(u.used_su_amount) AS used,
		        COALESCE(SUM(u.used_raw_amount), 0) AS used_native,
		        SUM(CASE WHEN u.user_id = ? THEN u.used_su_amount ELSE 0 END) AS used_by_caller
		   FROM compute_allocation_usages u
		   JOIN compute_allocation_resources r
		     ON r.id = u.compute_allocation_resource_id
		  WHERE u.compute_allocation_id = ?
		  GROUP BY r.id, r.name, r.resource_type
		  ORDER BY r.name`, callerID, allocationID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlStore) Jobs(ctx context.Context, allocationID string, userID *string, limit, offset int) ([]JobRow, int, error) {
	where := "u.compute_allocation_id = ?"
	args := []any{allocationID}
	if userID != nil {
		where += " AND u.user_id = ?"
		args = append(args, *userID)
	}

	var total int
	if err := s.db.GetContext(ctx, &total,
		"SELECT COUNT(*) FROM compute_allocation_usages u WHERE "+where, args...); err != nil {
		return nil, 0, err
	}

	var rows []JobRow
	err := s.db.SelectContext(ctx, &rows,
		`SELECT u.id, u.job_id, u.calculated_time, u.user_id,
		        TRIM(CONCAT(COALESCE(usr.first_name, ''), ' ', COALESCE(usr.last_name, ''))) AS user_name,
		        u.compute_allocation_resource_id AS resource_id,
		        COALESCE(r.name, '') AS resource_name,
		        COALESCE(r.resource_type, '') AS resource_type,
		        u.used_raw_amount,
		        u.used_su_amount AS used_su_amount
		   FROM compute_allocation_usages u
		   LEFT JOIN users usr ON usr.id = u.user_id
		   LEFT JOIN compute_allocation_resources r ON r.id = u.compute_allocation_resource_id
		  WHERE `+where+`
		  ORDER BY u.calculated_time DESC, u.id
		  LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (s *mysqlStore) MemberUsage(ctx context.Context, allocationID string) ([]MemberRow, error) {
	var rows []MemberRow
	err := s.db.SelectContext(ctx, &rows,
		`SELECT u.user_id,
		        COALESCE(NULLIF(TRIM(CONCAT(COALESCE(usr.first_name, ''), ' ', COALESCE(usr.last_name, ''))), ''), usr.email) AS name,
		        SUM(u.used_su_amount) AS used
		   FROM compute_allocation_usages u
		   JOIN users usr ON usr.id = u.user_id
		  WHERE u.compute_allocation_id = ?
		  GROUP BY u.user_id, name
		  ORDER BY used DESC`, allocationID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *mysqlStore) ProjectRole(ctx context.Context, projectID, userID string) (string, error) {
	var role string
	err := s.db.GetContext(ctx, &role,
		`SELECT role FROM project_memberships WHERE project_id = ? AND user_id = ?`,
		projectID, userID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return role, nil
}

func (s *mysqlStore) IsActiveMember(ctx context.Context, allocationID, userID string) (bool, error) {
	var exists bool
	err := s.db.GetContext(ctx, &exists,
		`SELECT EXISTS(
		    SELECT 1 FROM compute_allocation_memberships
		     WHERE compute_allocation_id = ? AND user_id = ?
		       AND membership_status = 'ACTIVE')`,
		allocationID, userID)
	if err != nil {
		return false, err
	}
	return exists, nil
}
