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

	"github.com/apache/airavata-custos/allocations/domain/model"
	"github.com/jmoiron/sqlx"
)

type mariaDBMembershipStore struct {
	db *sqlx.DB
}

func NewMembershipStore(db *sqlx.DB) MembershipStore {
	return &mariaDBMembershipStore{db: db}
}

func (s *mariaDBMembershipStore) FindByProjectAndAccount(ctx context.Context, projectID, accountID string) (*model.ProjectMembership, error) {
	var m model.ProjectMembership
	err := s.db.GetContext(ctx, &m,
		`SELECT id, project_id, cluster_account_id, role, is_active, created_at, updated_at
		 FROM project_memberships WHERE project_id = ? AND cluster_account_id = ?`,
		projectID, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (s *mariaDBMembershipStore) FindByProject(ctx context.Context, projectID string) ([]model.ProjectMembership, error) {
	var results []model.ProjectMembership
	err := s.db.SelectContext(ctx, &results,
		`SELECT id, project_id, cluster_account_id, role, is_active, created_at, updated_at
		 FROM project_memberships WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *mariaDBMembershipStore) FindByProjectAndRole(ctx context.Context, projectID, role string) ([]model.ProjectMembership, error) {
	var results []model.ProjectMembership
	err := s.db.SelectContext(ctx, &results,
		`SELECT id, project_id, cluster_account_id, role, is_active, created_at, updated_at
		 FROM project_memberships WHERE project_id = ? AND role = ?`,
		projectID, role)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *mariaDBMembershipStore) FindByProjectAndPerson(ctx context.Context, projectID, personID string) ([]model.ProjectMembership, error) {
	var results []model.ProjectMembership
	err := s.db.SelectContext(ctx, &results,
		`SELECT pm.id, pm.project_id, pm.cluster_account_id, pm.role, pm.is_active, pm.created_at, pm.updated_at
		 FROM project_memberships pm
		 JOIN cluster_accounts ca ON pm.cluster_account_id = ca.id
		 WHERE pm.project_id = ? AND ca.person_id = ?`,
		projectID, personID)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (s *mariaDBMembershipStore) Save(ctx context.Context, tx *sql.Tx, m *model.ProjectMembership) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO project_memberships (id, project_id, cluster_account_id, role, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.ProjectID, m.ClusterAccountID, m.Role, m.IsActive, m.CreatedAt, m.UpdatedAt)
	return err
}

func (s *mariaDBMembershipStore) Update(ctx context.Context, tx *sql.Tx, m *model.ProjectMembership) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE project_memberships SET role = ?, is_active = ?, updated_at = ?
		 WHERE id = ?`,
		m.Role, m.IsActive, m.UpdatedAt, m.ID)
	return err
}
