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

type mysqlProjectStore struct {
	db *sqlx.DB
}

// NewProjectStore returns a MySQL-backed ProjectStore.
func NewProjectStore(db *sqlx.DB) ProjectStore {
	return &mysqlProjectStore{db: db}
}

const projectColumns = `id, originated_id, title, origination, project_pi_id, status, created_time`

// ProjectWithPI is FindByIDWithPI / ListWithPI row shape: the project plus
// the joined PI's display fields. Kept here so handlers can render the
// portal-facing payload from a single SQL round trip.
type ProjectWithPI struct {
	models.Project
	PIFirstName string `db:"pi_first_name"`
	PILastName  string `db:"pi_last_name"`
	PIEmail     string `db:"pi_email"`
}

// projectWithPISelect joins users so a single query returns everything the
// portal needs to render a project row.
const projectWithPISelect = `SELECT p.id, p.originated_id, p.title, p.origination, p.project_pi_id,
        p.status, p.created_time,
        u.first_name AS pi_first_name, u.last_name AS pi_last_name, u.email AS pi_email
   FROM projects p
   LEFT JOIN users u ON u.id = p.project_pi_id`

// FindByIDWithPI returns the project joined with its PI's display fields, or
// nil if no project matches.
func (s *mysqlProjectStore) FindByIDWithPI(ctx context.Context, id string) (*ProjectWithPI, error) {
	var p ProjectWithPI
	err := s.db.GetContext(ctx, &p, projectWithPISelect+` WHERE p.id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *mysqlProjectStore) FindByID(ctx context.Context, id string) (*models.Project, error) {
	var p models.Project
	err := s.db.GetContext(ctx, &p,
		`SELECT `+projectColumns+` FROM projects WHERE id = ?`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *mysqlProjectStore) FindByOriginatedID(ctx context.Context, originatedID string) (*models.Project, error) {
	var p models.Project
	err := s.db.GetContext(ctx, &p,
		`SELECT `+projectColumns+` FROM projects WHERE originated_id = ?`, originatedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (s *mysqlProjectStore) FindByPI(ctx context.Context, piUserID string) ([]models.Project, error) {
	var projects []models.Project
	err := s.db.SelectContext(ctx, &projects,
		`SELECT `+projectColumns+` FROM projects WHERE project_pi_id = ?`, piUserID)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (s *mysqlProjectStore) Create(ctx context.Context, tx *sql.Tx, p *models.Project) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO projects (id, originated_id, title, origination, project_pi_id, status, created_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.OriginatedID, p.Title, p.Origination, p.ProjectPIID, p.Status, p.CreatedTime)
	return err
}

func (s *mysqlProjectStore) Update(ctx context.Context, tx *sql.Tx, p *models.Project) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE projects SET originated_id = ?, title = ?, origination = ?, project_pi_id = ?, status = ?
		 WHERE id = ?`,
		p.OriginatedID, p.Title, p.Origination, p.ProjectPIID, p.Status, p.ID)
	return err
}

func (s *mysqlProjectStore) UpdateStatus(ctx context.Context, tx *sql.Tx, id string, status models.ProjectStatus) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE projects SET status = ? WHERE id = ?`,
		status, id)
	return err
}

func (s *mysqlProjectStore) ReassignPI(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE projects SET project_pi_id = ? WHERE project_pi_id = ?`,
		toUserID, fromUserID)
	return err
}

func (s *mysqlProjectStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	return err
}

func (s *mysqlProjectStore) List(ctx context.Context, f ProjectListFilter) ([]models.Project, int, error) {
	where := []string{}
	args := []any{}
	if f.PIID != "" {
		where = append(where, `project_pi_id = ?`)
		args = append(args, f.PIID)
	}
	if f.Status != "" {
		where = append(where, `status = ?`)
		args = append(args, f.Status)
	}
	if f.Query != "" {
		where = append(where, `(title LIKE ? OR originated_id LIKE ?)`)
		q := "%" + f.Query + "%"
		args = append(args, q, q)
	}
	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}
	var total int
	if err := s.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM projects`+clause, args...); err != nil {
		return nil, 0, err
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	query := `SELECT ` + projectColumns + ` FROM projects` + clause +
		` ORDER BY created_time DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	var rows []models.Project
	if err := s.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// ListWithPI is List joined with the PI user. Replaces per-row GetUser calls
// the handler used to fan out across the result set.
func (s *mysqlProjectStore) ListWithPI(ctx context.Context, f ProjectListFilter) ([]ProjectWithPI, int, error) {
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
		where = append(where, `(p.title LIKE ? OR p.originated_id LIKE ?)`)
		q := "%" + f.Query + "%"
		args = append(args, q, q)
	}
	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}
	var total int
	if err := s.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM projects p`+clause, args...); err != nil {
		return nil, 0, err
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	query := projectWithPISelect + clause + ` ORDER BY p.created_time DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)
	var rows []ProjectWithPI
	if err := s.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
