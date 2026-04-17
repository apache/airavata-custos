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

package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/apache/airavata-custos/allocations/access-amie/model"
	"github.com/google/uuid"
)

type membershipStore interface {
	FindByProjectAndAccount(ctx context.Context, projectID, accountID string) (*model.ProjectMembership, error)
	FindByProject(ctx context.Context, projectID string) ([]model.ProjectMembership, error)
	FindByProjectAndRole(ctx context.Context, projectID, role string) ([]model.ProjectMembership, error)
	FindByProjectAndPerson(ctx context.Context, projectID, personID string) ([]model.ProjectMembership, error)
	Save(ctx context.Context, tx *sql.Tx, m *model.ProjectMembership) error
	Update(ctx context.Context, tx *sql.Tx, m *model.ProjectMembership) error
}

type membershipProjectStore interface {
	FindByID(ctx context.Context, id string) (*model.Project, error)
}

type membershipAccountStore interface {
	FindByPerson(ctx context.Context, personID string) ([]model.ClusterAccount, error)
}

type ProjectMembershipService struct {
	memberships membershipStore
	projects    membershipProjectStore
	accounts    membershipAccountStore
}

func NewProjectMembershipService(memberships membershipStore, projects membershipProjectStore, accounts membershipAccountStore) *ProjectMembershipService {
	return &ProjectMembershipService{
		memberships: memberships,
		projects:    projects,
		accounts:    accounts,
	}
}

// CreateMembership creates or reactivates a project membership for the given
// cluster account. If an active membership already exists it is returned
// unchanged.
func (s *ProjectMembershipService) CreateMembership(ctx context.Context, tx *sql.Tx, projectID, clusterAccountID, role string) (*model.ProjectMembership, error) {
	existing, err := s.memberships.FindByProjectAndAccount(ctx, projectID, clusterAccountID)
	if err != nil {
		return nil, fmt.Errorf("membership_service: finding membership for project %s account %s: %w", projectID, clusterAccountID, err)
	}

	if existing != nil {
		if existing.IsActive {
			return existing, nil
		}
		// Reactivate an inactive membership.
		existing.IsActive = true
		existing.Role = &role
		if err := s.memberships.Update(ctx, tx, existing); err != nil {
			return nil, fmt.Errorf("membership_service: reactivating membership %s: %w", existing.ID, err)
		}
		slog.DebugContext(ctx, "reactivated existing membership", "membership_id", existing.ID)
		return existing, nil
	}

	m := &model.ProjectMembership{
		ID:               uuid.NewString(),
		ProjectID:        projectID,
		ClusterAccountID: clusterAccountID,
		Role:             &role,
		IsActive:         true,
	}

	if err := s.memberships.Save(ctx, tx, m); err != nil {
		return nil, fmt.Errorf("membership_service: saving membership for project %s account %s: %w", projectID, clusterAccountID, err)
	}

	slog.DebugContext(ctx, "created membership", "membership_id", m.ID, "project_id", projectID, "account_id", clusterAccountID, "role", role)
	return m, nil
}

// InactivateMembershipsByPersonAndProject sets all memberships for a given
// person in a project to inactive. Returns the number of memberships affected.
func (s *ProjectMembershipService) InactivateMembershipsByPersonAndProject(ctx context.Context, tx *sql.Tx, projectID, personID string) (int, error) {
	memberships, err := s.memberships.FindByProjectAndPerson(ctx, projectID, personID)
	if err != nil {
		return 0, fmt.Errorf("membership_service: finding memberships for project %s person %s: %w", projectID, personID, err)
	}

	count := 0
	for i := range memberships {
		memberships[i].IsActive = false
		if err := s.memberships.Update(ctx, tx, &memberships[i]); err != nil {
			return count, fmt.Errorf("membership_service: inactivating membership %s: %w", memberships[i].ID, err)
		}
		count++
	}

	slog.DebugContext(ctx, "inactivated memberships by person and project", "project_id", projectID, "person_id", personID, "count", count)
	return count, nil
}

// InactivateAllForProject sets all memberships in a project to inactive.
func (s *ProjectMembershipService) InactivateAllForProject(ctx context.Context, tx *sql.Tx, projectID string) error {
	memberships, err := s.memberships.FindByProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("membership_service: finding memberships for project %s: %w", projectID, err)
	}

	for i := range memberships {
		memberships[i].IsActive = false
		if err := s.memberships.Update(ctx, tx, &memberships[i]); err != nil {
			return fmt.Errorf("membership_service: inactivating membership %s: %w", memberships[i].ID, err)
		}
	}

	slog.DebugContext(ctx, "inactivated all memberships for project", "project_id", projectID, "count", len(memberships))
	return nil
}

// ReactivatePiMembership reactivates all memberships with the "PI" role in
// the given project.
func (s *ProjectMembershipService) ReactivatePiMembership(ctx context.Context, tx *sql.Tx, projectID string) error {
	memberships, err := s.memberships.FindByProjectAndRole(ctx, projectID, "PI")
	if err != nil {
		return fmt.Errorf("membership_service: finding PI memberships for project %s: %w", projectID, err)
	}

	for i := range memberships {
		memberships[i].IsActive = true
		if err := s.memberships.Update(ctx, tx, &memberships[i]); err != nil {
			return fmt.Errorf("membership_service: reactivating PI membership %s: %w", memberships[i].ID, err)
		}
	}

	slog.DebugContext(ctx, "reactivated PI memberships for project", "project_id", projectID, "count", len(memberships))
	return nil
}

// ReactivateMembershipsByPersonAndProject reactivates all memberships for a
// given person in a project. Returns the number of memberships affected.
func (s *ProjectMembershipService) ReactivateMembershipsByPersonAndProject(ctx context.Context, tx *sql.Tx, projectID, personID string) (int, error) {
	memberships, err := s.memberships.FindByProjectAndPerson(ctx, projectID, personID)
	if err != nil {
		return 0, fmt.Errorf("membership_service: finding memberships for project %s person %s: %w", projectID, personID, err)
	}

	count := 0
	for i := range memberships {
		memberships[i].IsActive = true
		if err := s.memberships.Update(ctx, tx, &memberships[i]); err != nil {
			return count, fmt.Errorf("membership_service: reactivating membership %s: %w", memberships[i].ID, err)
		}
		count++
	}

	slog.DebugContext(ctx, "reactivated memberships by person and project", "project_id", projectID, "person_id", personID, "count", count)
	return count, nil
}
