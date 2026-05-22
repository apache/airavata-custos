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

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// CreateProject persists a new project. If project.ID is empty a new UUID is
// generated. The referenced PI user must already exist.
func (s *Service) CreateProject(ctx context.Context, project *models.Project) (*models.Project, error) {
	if project == nil {
		return nil, fmt.Errorf("%w: project is nil", ErrInvalidInput)
	}
	if project.Title == "" {
		return nil, fmt.Errorf("%w: project title is required", ErrInvalidInput)
	}
	if project.ProjectPIID == "" {
		return nil, fmt.Errorf("%w: project_pi_id is required", ErrInvalidInput)
	}

	if pi, err := s.users.FindByID(ctx, project.ProjectPIID); err != nil {
		return nil, fmt.Errorf("verify project PI: %w", err)
	} else if pi == nil {
		return nil, fmt.Errorf("%w: PI user %q does not exist", ErrInvalidInput, project.ProjectPIID)
	}

	if project.OriginatedID != "" {
		if existing, err := s.projs.FindByOriginatedID(ctx, project.OriginatedID); err != nil {
			return nil, fmt.Errorf("lookup project by originated id: %w", err)
		} else if existing != nil {
			return nil, fmt.Errorf("%w: project with originated_id %q", ErrAlreadyExists, project.OriginatedID)
		}
	}

	if project.ID == "" {
		project.ID = newID()
	}
	if project.Status == "" {
		project.Status = models.ProjectActive
	}
	if err := validateProjectStatus(project.Status); err != nil {
		return nil, err
	}
	if project.CreatedTime.IsZero() {
		project.CreatedTime = nowUTC()
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.projs.Create(ctx, tx, project)
	}); err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	s.eventBus.Publish(events.ProjectCreateEvent, project)
	return project, nil
}

// GetProject retrieves a project by ID. Returns ErrNotFound when no project matches.
func (s *Service) GetProject(ctx context.Context, id string) (*models.Project, error) {
	p, err := s.projs.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	if p == nil {
		return nil, ErrNotFound
	}
	return p, nil
}

// GetProjectByOriginatedID retrieves a project by its external originated ID.
func (s *Service) GetProjectByOriginatedID(ctx context.Context, originatedID string) (*models.Project, error) {
	p, err := s.projs.FindByOriginatedID(ctx, originatedID)
	if err != nil {
		return nil, fmt.Errorf("get project by originated id: %w", err)
	}
	if p == nil {
		return nil, ErrNotFound
	}
	return p, nil
}

// ListProjectsByPI returns all projects whose PI matches the given user ID.
func (s *Service) ListProjectsByPI(ctx context.Context, piUserID string) ([]models.Project, error) {
	projects, err := s.projs.FindByPI(ctx, piUserID)
	if err != nil {
		return nil, fmt.Errorf("list projects by PI: %w", err)
	}
	return projects, nil
}

// UpdateProject persists changes to an existing project. Fields left
// blank/zero on the supplied record fall back to the stored value.
func (s *Service) UpdateProject(ctx context.Context, project *models.Project) error {
	if project == nil || project.ID == "" {
		return fmt.Errorf("%w: project id is required", ErrInvalidInput)
	}
	existing, err := s.projs.FindByID(ctx, project.ID)
	if err != nil {
		return fmt.Errorf("lookup project: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}
	if project.OriginatedID == "" {
		project.OriginatedID = existing.OriginatedID
	}
	if project.Title == "" {
		project.Title = existing.Title
	}
	if project.Origination == "" {
		project.Origination = existing.Origination
	}
	if project.ProjectPIID == "" {
		project.ProjectPIID = existing.ProjectPIID
	}
	if project.Status == "" {
		project.Status = existing.Status
	}
	if err := validateProjectStatus(project.Status); err != nil {
		return err
	}
	if project.CreatedTime.IsZero() {
		project.CreatedTime = existing.CreatedTime
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.projs.Update(ctx, tx, project)
	}); err != nil {
		return fmt.Errorf("update project: %w", err)
	}

	s.eventBus.Publish(events.ProjectUpdateEvent, project)
	return nil
}

// UpdateProjectStatus sets the lifecycle status of the project identified by
// id. Other fields are preserved.
func (s *Service) UpdateProjectStatus(ctx context.Context, id string, status models.ProjectStatus) (*models.Project, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: project id is required", ErrInvalidInput)
	}
	if status == "" {
		return nil, fmt.Errorf("%w: status is required", ErrInvalidInput)
	}
	if err := validateProjectStatus(status); err != nil {
		return nil, err
	}
	existing, err := s.projs.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("lookup project: %w", err)
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.projs.UpdateStatus(ctx, tx, id, status)
	}); err != nil {
		return nil, fmt.Errorf("update project status: %w", err)
	}
	existing.Status = status

	s.eventBus.Publish(events.ProjectUpdateEvent, existing)
	return existing, nil
}

// DeleteProject removes a project by ID.
func (s *Service) DeleteProject(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: project id is required", ErrInvalidInput)
	}
	project, err := s.projs.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup project: %w", err)
	}
	if project == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.projs.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete project: %w", err)
	}

	s.eventBus.Publish(events.ProjectDeleteEvent, project)
	return nil
}
