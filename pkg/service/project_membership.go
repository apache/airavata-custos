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

	"github.com/apache/airavata-custos/pkg/models"
)

// EnsureProjectMembership upserts the (project, user, role) row enforcing the
// PI-can't-change rule and the special-role conflict policy:
//   - same user already holds the requested role: no-op success.
//   - PI for a different user when a PI already exists: returns ErrPIChange.
//   - existing role is PI: returns ErrPIChange (PI cannot be downgraded).
//   - otherwise: upsert (latest non-PI role wins).
//
// MEMBER calls collapse to a Delete: the user is "demoted" to derived-only
// membership via compute_allocation_memberships.
func (s *Service) EnsureProjectMembership(ctx context.Context, projectID, userID, role string) error {
	if projectID == "" {
		return fmt.Errorf("%w: project_id is required", ErrInvalidInput)
	}
	if userID == "" {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}

	if role == "MEMBER" || role == "" {
		existing, err := s.projMemberships.FindByPair(ctx, projectID, userID)
		if err != nil {
			return fmt.Errorf("lookup project membership: %w", err)
		}
		if existing == nil {
			return nil
		}
		if existing.Role == models.ProjectRolePI {
			return fmt.Errorf("%w: cannot demote PI of project %q", ErrPIChange, projectID)
		}
		return s.inTx(ctx, func(tx *sql.Tx) error {
			return s.projMemberships.Delete(ctx, tx, projectID, userID)
		})
	}

	pr, ok := normalizeProjectRole(role)
	if !ok {
		return fmt.Errorf("%w: unknown project role %q", ErrInvalidInput, role)
	}

	existing, err := s.projMemberships.FindByPair(ctx, projectID, userID)
	if err != nil {
		return fmt.Errorf("lookup project membership: %w", err)
	}

	if pr == models.ProjectRolePI {
		currentPI, err := s.projMemberships.FindPIByProject(ctx, projectID)
		if err != nil {
			return fmt.Errorf("lookup PI: %w", err)
		}
		if currentPI != nil && currentPI.UserID != userID {
			return fmt.Errorf("%w: project %q already has PI %q",
				ErrPIChange, projectID, currentPI.UserID)
		}
	}

	if existing != nil {
		if existing.Role == models.ProjectRolePI && pr != models.ProjectRolePI {
			return fmt.Errorf("%w: cannot downgrade PI of project %q", ErrPIChange, projectID)
		}
		if existing.Role == pr {
			return nil
		}
		return s.inTx(ctx, func(tx *sql.Tx) error {
			return s.projMemberships.UpdateRole(ctx, tx, projectID, userID, pr)
		})
	}

	pm := &models.ProjectMembership{
		ProjectID: projectID,
		UserID:    userID,
		Role:      pr,
		AddedTime: nowUTC(),
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		return s.projMemberships.Create(ctx, tx, pm)
	})
}

// ProjectRoleForUser returns the user's governance role on a project, or an
// empty role when they hold none.
func (s *Service) ProjectRoleForUser(ctx context.Context, projectID, userID string) (models.ProjectRole, error) {
	pm, err := s.projMemberships.FindByPair(ctx, projectID, userID)
	if err != nil {
		return "", fmt.Errorf("lookup project role: %w", err)
	}
	if pm == nil {
		return "", nil
	}
	return pm.Role, nil
}

// ListProjectMemberships returns every project_memberships row for the project.
func (s *Service) ListProjectMemberships(ctx context.Context, projectID string) ([]models.ProjectMembership, error) {
	if projectID == "" {
		return nil, fmt.Errorf("%w: project_id is required", ErrInvalidInput)
	}
	rows, err := s.projMemberships.FindByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project memberships: %w", err)
	}
	return rows, nil
}

func normalizeProjectRole(raw string) (models.ProjectRole, bool) {
	switch raw {
	case "PI":
		return models.ProjectRolePI, true
	case "CO_PI":
		return models.ProjectRoleCoPI, true
	case "ALLOCATION_MANAGER":
		return models.ProjectRoleAllocationManager, true
	}
	return "", false
}
