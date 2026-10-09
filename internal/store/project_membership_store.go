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

const projectMembershipColumns = "project_id, user_id, role, added_time"

type pgProjectMembershipStore struct {
	db *sqlx.DB
}

// NewProjectMembershipStore returns a PostgreSQL-backed ProjectMembershipStore.
func NewProjectMembershipStore(db *sqlx.DB) ProjectMembershipStore {
	return &pgProjectMembershipStore{db: db}
}

func (s *pgProjectMembershipStore) FindByPair(ctx context.Context, projectID, userID string) (*models.ProjectMembership, error) {
	var pm models.ProjectMembership
	err := s.db.GetContext(ctx, &pm,
		`SELECT `+projectMembershipColumns+`
		   FROM project_roles
		  WHERE project_id = $1 AND user_id = $2 AND rank < 3
		  LIMIT 1`, projectID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &pm, nil
}

func (s *pgProjectMembershipStore) FindByProject(ctx context.Context, projectID string) ([]models.ProjectMembership, error) {
	var rows []models.ProjectMembership
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+projectMembershipColumns+`
		   FROM project_memberships
		  WHERE project_id = $1
		  ORDER BY added_time`, projectID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// IsParticipant reports whether the user has any role on the project.
func (s *pgProjectMembershipStore) IsParticipant(ctx context.Context, projectID, userID string) (bool, error) {
	var participant bool
	err := s.db.GetContext(ctx, &participant,
		`SELECT EXISTS (SELECT 1 FROM project_roles WHERE project_id = $1 AND user_id = $2)`,
		projectID, userID)
	if err != nil {
		return false, err
	}
	return participant, nil
}

func (s *pgProjectMembershipStore) Create(ctx context.Context, tx *sql.Tx, pm *models.ProjectMembership) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO project_memberships (project_id, user_id, role, added_time)
		 VALUES ($1, $2, $3, $4)`,
		pm.ProjectID, pm.UserID, string(pm.Role), pm.AddedTime)
	return err
}

func (s *pgProjectMembershipStore) UpdateRole(ctx context.Context, tx *sql.Tx, projectID, userID string, role models.ProjectRole) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE project_memberships SET role = $1 WHERE project_id = $2 AND user_id = $3`,
		string(role), projectID, userID)
	return err
}

func (s *pgProjectMembershipStore) Delete(ctx context.Context, tx *sql.Tx, projectID, userID string) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM project_memberships WHERE project_id = $1 AND user_id = $2`,
		projectID, userID)
	return err
}

// ReassignUser moves fromUserID's tags to toUserID, keeping the higher tag per
// project and none where toUserID is PI.
func (s *pgProjectMembershipStore) ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	// Rows deleted on non-PI projects are exactly fromUserID's to move; the upsert
	// never touches a deleted row. GREATEST keeps the higher tag: 'CO_PI' > 'ALLOCATION_MANAGER'.
	_, err := tx.ExecContext(ctx,
		`WITH d AS (DELETE FROM project_memberships pm USING projects p
		  WHERE pm.project_id = p.id AND (pm.user_id = $1 OR (pm.user_id = $2 AND p.project_pi_id = $2))
		  RETURNING pm.project_id, pm.role, pm.added_time, p.project_pi_id)
		 INSERT INTO project_memberships (project_id, user_id, role, added_time)
		 SELECT project_id, $2, role, added_time FROM d WHERE project_pi_id <> $2
		 ON CONFLICT (project_id, user_id) DO UPDATE SET role = GREATEST(project_memberships.role, EXCLUDED.role)`,
		fromUserID, toUserID)
	return err
}
