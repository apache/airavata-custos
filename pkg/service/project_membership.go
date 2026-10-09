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

// EnsureProjectMembership sets the user's project role. The PI lives in
// projects.project_pi_id and outranks every role, so any role for the PI is
// already held; PI for anyone else returns ErrPIChange. MEMBER (or "") removes
// the role row; the user stays a member through compute_allocation_memberships.
func (s *Service) EnsureProjectMembership(ctx context.Context, projectID, userID, role string) error {
	if projectID == "" {
		return fmt.Errorf("%w: project_id is required", ErrInvalidInput)
	}
	if userID == "" {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}

	existing, err := s.projMemberships.FindByPair(ctx, projectID, userID)
	if err != nil {
		return fmt.Errorf("lookup project membership: %w", err)
	}
	pr := models.ProjectRole(role)
	switch {
	case existing != nil && existing.Role == models.ProjectRolePI:
		return nil
	case pr == models.ProjectRolePI:
		return fmt.Errorf("%w: project %q", ErrPIChange, projectID)
	}
	if role == "MEMBER" || role == "" {
		return s.inTx(ctx, func(tx *sql.Tx) error {
			return s.projMemberships.Delete(ctx, tx, projectID, userID)
		})
	}
	if pr != models.ProjectRoleCoPI && pr != models.ProjectRoleAllocationManager {
		return fmt.Errorf("%w: unknown project role %q", ErrInvalidInput, role)
	}

	return s.inTx(ctx, func(tx *sql.Tx) error {
		return s.projMemberships.Upsert(ctx, tx, &models.ProjectMembership{
			ProjectID: projectID, UserID: userID, Role: pr, AddedTime: nowUTC(),
		})
	})
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
