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
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

type pgProjectStore struct {
	db *sqlx.DB
}

// NewProjectStore returns a PostgreSQL-backed ProjectStore.
func NewProjectStore(db *sqlx.DB) ProjectStore {
	return &pgProjectStore{db: db}
}

const projectColumns = `id, originated_id, title, origination, project_pi_id, status, created_time`

// ProjectWithPI is FindByIDWithPI / ListWithPI row shape: the project plus
// the joined PI's display fields. Kept here so handlers can render the
// portal-facing payload from a single SQL round trip.
type ProjectWithPI struct {
	models.Project
	PIDisplayName string `db:"pi_display_name"`
	PIEmail       string `db:"pi_email"`
}

// projectWithPISelect joins users so a single query returns everything the
// portal needs to render a project row.
var projectWithPISelect = `SELECT ` + Qualify("p", projectColumns) + `,
        ` + DisplayNameSQL + ` AS pi_display_name,
        u.email AS pi_email
   FROM projects p
   JOIN users u ON u.id = p.project_pi_id`

// FindByIDWithPI returns the project joined with its PI's display fields, or
// nil if no project matches.
func (s *pgProjectStore) FindByIDWithPI(ctx context.Context, id string) (*ProjectWithPI, error) {
	var p ProjectWithPI
	err := s.db.GetContext(ctx, &p, projectWithPISelect+` WHERE p.id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *pgProjectStore) FindByID(ctx context.Context, id string) (*models.Project, error) {
	var p models.Project
	err := s.db.GetContext(ctx, &p,
		`SELECT `+projectColumns+` FROM projects WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *pgProjectStore) FindByOriginatedID(ctx context.Context, originatedID string) (*models.Project, error) {
	var p models.Project
	err := s.db.GetContext(ctx, &p,
		`SELECT `+projectColumns+` FROM projects WHERE originated_id = $1`, originatedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *pgProjectStore) Create(ctx context.Context, tx *sql.Tx, p *models.Project) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO projects (`+projectColumns+`)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		p.ID, p.OriginatedID, p.Title, p.Origination, p.ProjectPIID, p.Status, p.CreatedTime)
	return err
}

func (s *pgProjectStore) Update(ctx context.Context, tx *sql.Tx, p *models.Project) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE projects SET originated_id = $1, title = $2, origination = $3, project_pi_id = $4, status = $5
		 WHERE id = $6`,
		p.OriginatedID, p.Title, p.Origination, p.ProjectPIID, p.Status, p.ID)
	return err
}

func (s *pgProjectStore) UpdateStatus(ctx context.Context, tx *sql.Tx, id string, status models.ProjectStatus) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE projects SET status = $1 WHERE id = $2`,
		status, id)
	return err
}

func (s *pgProjectStore) ReassignPI(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE projects SET project_pi_id = $1 WHERE project_pi_id = $2`,
		toUserID, fromUserID)
	return err
}

func (s *pgProjectStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM projects WHERE id = $1`, id)
	return err
}

// ListWithPIForParticipant returns the projects the user has a role on, PI
// joined, newest first.
func (s *pgProjectStore) ListWithPIForParticipant(ctx context.Context, userID string) ([]ProjectWithPI, error) {
	var rows []ProjectWithPI
	err := s.db.SelectContext(ctx, &rows, projectWithPISelect+`
	  WHERE p.id IN (SELECT project_id FROM project_roles WHERE user_id = $1)
	  ORDER BY p.created_time DESC`, userID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *pgProjectStore) ListWithPI(ctx context.Context, f ProjectListFilter) ([]ProjectWithPI, int, error) {
	where := []string{}
	args := []any{}
	if f.PIID != "" {
		where = append(where, `p.project_pi_id = ?`)
		args = append(args, f.PIID)
	}
	if f.Status != "" {
		where = append(where, `p.status = ?`)
		args = append(args, f.Status)
	}
	if f.Query != "" {
		where = append(where, `(p.title ILIKE ? OR p.originated_id ILIKE ?)`)
		q := "%" + f.Query + "%"
		args = append(args, q, q)
	}
	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}
	var total int
	if err := s.db.GetContext(ctx, &total, s.db.Rebind(`SELECT COUNT(*) FROM projects p`+clause), args...); err != nil {
		return nil, 0, err
	}
	var rows []ProjectWithPI
	if err := s.db.SelectContext(ctx, &rows, s.db.Rebind(projectWithPISelect+clause+` ORDER BY p.created_time DESC LIMIT ? OFFSET ?`),
		append(args, PageLimit(f.Limit), max(f.Offset, 0))...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
