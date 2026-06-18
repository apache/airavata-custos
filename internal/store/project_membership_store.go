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

type mysqlProjectMembershipStore struct {
	db *sqlx.DB
}

// NewProjectMembershipStore returns a MySQL-backed ProjectMembershipStore.
func NewProjectMembershipStore(db *sqlx.DB) ProjectMembershipStore {
	return &mysqlProjectMembershipStore{db: db}
}

func (s *mysqlProjectMembershipStore) FindByPair(ctx context.Context, projectID, userID string) (*models.ProjectMembership, error) {
	var pm models.ProjectMembership
	err := s.db.GetContext(ctx, &pm,
		`SELECT `+projectMembershipColumns+`
		   FROM project_memberships
		  WHERE project_id = ? AND user_id = ?`, projectID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &pm, nil
}

func (s *mysqlProjectMembershipStore) FindByProject(ctx context.Context, projectID string) ([]models.ProjectMembership, error) {
	var rows []models.ProjectMembership
	err := s.db.SelectContext(ctx, &rows,
		`SELECT `+projectMembershipColumns+`
		   FROM project_memberships
		  WHERE project_id = ?
		  ORDER BY added_time`, projectID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// FindPIByProject returns the project's PI row when one exists.
func (s *mysqlProjectMembershipStore) FindPIByProject(ctx context.Context, projectID string) (*models.ProjectMembership, error) {
	var pm models.ProjectMembership
	err := s.db.GetContext(ctx, &pm,
		`SELECT `+projectMembershipColumns+`
		   FROM project_memberships
		  WHERE project_id = ? AND role = 'PI'`, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &pm, nil
}

func (s *mysqlProjectMembershipStore) Create(ctx context.Context, tx *sql.Tx, pm *models.ProjectMembership) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO project_memberships (project_id, user_id, role, added_time)
		 VALUES (?, ?, ?, ?)`,
		pm.ProjectID, pm.UserID, string(pm.Role), pm.AddedTime)
	return err
}

func (s *mysqlProjectMembershipStore) UpdateRole(ctx context.Context, tx *sql.Tx, projectID, userID string, role models.ProjectRole) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE project_memberships SET role = ? WHERE project_id = ? AND user_id = ?`,
		string(role), projectID, userID)
	return err
}

func (s *mysqlProjectMembershipStore) Delete(ctx context.Context, tx *sql.Tx, projectID, userID string) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM project_memberships WHERE project_id = ? AND user_id = ?`,
		projectID, userID)
	return err
}

// ReassignUser moves all project_memberships rows from fromUserID to toUserID.
// Rows that would collide on the survivor's PK (same project) are dropped
// from the retiring user first; remaining rows are then re-pointed.
func (s *mysqlProjectMembershipStore) ReassignUser(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM project_memberships
		  WHERE user_id = ?
		    AND project_id IN (SELECT project_id FROM (
		        SELECT project_id FROM project_memberships WHERE user_id = ?
		    ) AS s)`,
		fromUserID, toUserID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE project_memberships SET user_id = ? WHERE user_id = ?`,
		toUserID, fromUserID)
	return err
}
