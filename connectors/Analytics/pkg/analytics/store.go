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

	"github.com/apache/airavata-custos/internal/store"
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
	ProjectID string `db:"project_id"`
	Allocation
}

// DailyRow is credits consumed on one day against one resource.
type DailyRow struct {
	Day        time.Time `db:"day"`
	ResourceID string    `db:"resource_id"`
	Credits    float64   `db:"credits"`
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

// Store aggregates usage for the analytics endpoints.
type Store interface {
	// ProjectsForUser returns the projects the user touches through a governance
	// role or an active allocation membership. Role is null for a plain member.
	ProjectsForUser(ctx context.Context, userID string) ([]ProjectRow, error)
	// AllocationsByID returns the given allocations with their consumed credits, oldest first.
	AllocationsByID(ctx context.Context, ids []string) ([]AllocationRow, error)
	// DailyUsage returns per-day, per-resource credits for one allocation.
	DailyUsage(ctx context.Context, allocationID string) ([]DailyRow, error)
	// ResourceUsage returns per-resource credits (and the caller's slice) for
	// one allocation, over resources that carry usage.
	ResourceUsage(ctx context.Context, allocationID, callerID string) ([]UsageResource, error)
	// MemberUsage returns per-member credits for one allocation, ranked by
	// consumption. Callers gate this behind a role check.
	MemberUsage(ctx context.Context, allocationID string) ([]UsageMember, error)
	// Jobs returns a page of usage records (newest first) for one allocation
	// and the total count. A non-nil userID restricts to that user's records.
	Jobs(ctx context.Context, allocationID string, userID *string, limit, offset int) ([]JobRow, int, error)
}

type pgStore struct {
	db *sqlx.DB
}

// NewStore returns a PostgreSQL-backed analytics Store.
func NewStore(db *sqlx.DB) Store {
	return &pgStore{db: db}
}

func (s *pgStore) ProjectsForUser(ctx context.Context, userID string) ([]ProjectRow, error) {
	var rows []ProjectRow
	err := s.db.SelectContext(ctx, &rows,
		`SELECT p.id, p.title, r.role
		   FROM projects p
		   LEFT JOIN project_roles r ON r.project_id = p.id AND r.user_id = $1 AND r.compute_allocation_id IS NULL
		  WHERE p.id IN (SELECT project_id FROM project_roles WHERE user_id = $1)
		  ORDER BY p.title`, userID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgStore) AllocationsByID(ctx context.Context, ids []string) ([]AllocationRow, error) {
	var rows []AllocationRow
	err := s.db.SelectContext(ctx, &rows,
		`SELECT ca.project_id, ca.id, ca.name, ca.initial_su_amount, ca.end_time,
		        COALESCE(SUM(u.used_su_amount), 0) AS used_su_amount
		   FROM compute_allocations ca
		   LEFT JOIN compute_allocation_usages u
		     ON u.compute_allocation_id = ca.id
		  WHERE ca.id = ANY($1)
		  GROUP BY ca.id
		  ORDER BY ca.start_time`, ids)
	return rows, err
}

func (s *pgStore) DailyUsage(ctx context.Context, allocationID string) ([]DailyRow, error) {
	var rows []DailyRow
	err := s.db.SelectContext(ctx, &rows,
		`SELECT DATE(u.calculated_time) AS day,
		        u.compute_allocation_resource_id AS resource_id,
		        SUM(u.used_su_amount) AS credits
		   FROM compute_allocation_usages u
		  WHERE u.compute_allocation_id = $1
		  GROUP BY day, resource_id
		  ORDER BY day`, allocationID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgStore) ResourceUsage(ctx context.Context, allocationID, callerID string) ([]UsageResource, error) {
	rows := []UsageResource{}
	err := s.db.SelectContext(ctx, &rows,
		`SELECT r.id, r.name, r.resource_type,
		        SUM(u.used_su_amount) AS used,
		        SUM(u.used_raw_amount) AS used_native,
		        SUM(CASE WHEN u.user_id = $1 THEN u.used_su_amount ELSE 0 END) AS used_by_caller
		   FROM compute_allocation_usages u
		   JOIN compute_allocation_resources r
		     ON r.id = u.compute_allocation_resource_id
		  WHERE u.compute_allocation_id = $2
		  GROUP BY r.id
		  ORDER BY r.name`, callerID, allocationID)
	return rows, err
}

func (s *pgStore) Jobs(ctx context.Context, allocationID string, userID *string, limit, offset int) ([]JobRow, int, error) {
	const where = `WHERE u.compute_allocation_id = $1 AND u.user_id = COALESCE($2, u.user_id)`
	var total int
	if err := s.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM compute_allocation_usages u `+where, allocationID, userID); err != nil {
		return nil, 0, err
	}

	var rows []JobRow
	err := s.db.SelectContext(ctx, &rows,
		`SELECT u.id, u.job_id, u.calculated_time, u.user_id,
		        TRIM(CONCAT(usr.first_name, ' ', usr.last_name)) AS user_name, -- not DisplayNameSQL: usages carry no user FK, so usr may be missing
		        u.compute_allocation_resource_id AS resource_id,
		        COALESCE(r.name, '') AS resource_name,
		        COALESCE(r.resource_type, '') AS resource_type,
		        u.used_raw_amount, u.used_su_amount
		   FROM compute_allocation_usages u
		   LEFT JOIN users usr ON usr.id = u.user_id
		   LEFT JOIN compute_allocation_resources r ON r.id = u.compute_allocation_resource_id
		  `+where+`
		  ORDER BY u.calculated_time DESC, u.id
		  LIMIT $3 OFFSET $4`, allocationID, userID, limit, offset)
	return rows, total, err
}

func (s *pgStore) MemberUsage(ctx context.Context, allocationID string) ([]UsageMember, error) {
	rows := []UsageMember{}
	err := s.db.SelectContext(ctx, &rows,
		`SELECT u.id AS user_id, `+store.DisplayNameSQL+` AS name,
		        SUM(x.used_su_amount) AS used
		   FROM compute_allocation_usages x
		   JOIN users u ON u.id = x.user_id
		  WHERE x.compute_allocation_id = $1
		  GROUP BY u.id
		  ORDER BY used DESC`, allocationID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
