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
	"time"

	"github.com/apache/airavata-custos/allocations/domain/model"
)

type projectStore interface {
	FindByID(ctx context.Context, id string) (*model.Project, error)
	Save(ctx context.Context, tx *sql.Tx, p *model.Project) error
	Update(ctx context.Context, tx *sql.Tx, p *model.Project) error
}

type ProjectService struct {
	projects projectStore
}

func NewProjectService(projects projectStore) *ProjectService {
	return &ProjectService{projects: projects}
}

// CreateOrFindProject returns an existing project or creates a new active
// project with the given ID and grant number.
func (s *ProjectService) CreateOrFindProject(ctx context.Context, tx *sql.Tx, projectID, grantNumber string) (*model.Project, error) {
	existing, err := s.projects.FindByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("project_service: finding project %s: %w", projectID, err)
	}
	if existing != nil {
		return existing, nil
	}

	now := time.Now().UTC()
	p := &model.Project{
		ID:          projectID,
		GrantNumber: grantNumber,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.projects.Save(ctx, tx, p); err != nil {
		return nil, fmt.Errorf("project_service: saving project %s: %w", projectID, err)
	}

	slog.DebugContext(ctx, "created project", "project_id", projectID, "grant_number", grantNumber)
	return p, nil
}

// InactivateProject marks a project as inactive. Returns an error if the
// project does not exist.
func (s *ProjectService) InactivateProject(ctx context.Context, tx *sql.Tx, projectID string) error {
	p, err := s.projects.FindByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project_service: finding project %s: %w", projectID, err)
	}
	if p == nil {
		return fmt.Errorf("project %s not found", projectID)
	}

	p.IsActive = false
	if err := s.projects.Update(ctx, tx, p); err != nil {
		return fmt.Errorf("project_service: inactivating project %s: %w", projectID, err)
	}

	slog.DebugContext(ctx, "inactivated project", "project_id", projectID)
	return nil
}

// ReactivateProject marks a project as active. Returns an error if the
// project does not exist.
func (s *ProjectService) ReactivateProject(ctx context.Context, tx *sql.Tx, projectID string) error {
	p, err := s.projects.FindByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project_service: finding project %s: %w", projectID, err)
	}
	if p == nil {
		return fmt.Errorf("project %s not found", projectID)
	}

	p.IsActive = true
	if err := s.projects.Update(ctx, tx, p); err != nil {
		return fmt.Errorf("project_service: reactivating project %s: %w", projectID, err)
	}

	slog.DebugContext(ctx, "reactivated project", "project_id", projectID)
	return nil
}
