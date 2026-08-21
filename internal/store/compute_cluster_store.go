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

type pgComputeClusterStore struct {
	db *sqlx.DB
}

// NewComputeClusterStore returns a PostgreSQL-backed ComputeClusterStore.
func NewComputeClusterStore(db *sqlx.DB) ComputeClusterStore {
	return &pgComputeClusterStore{db: db}
}

func (s *pgComputeClusterStore) FindByID(ctx context.Context, id string) (*models.ComputeCluster, error) {
	var c models.ComputeCluster
	err := s.db.GetContext(ctx, &c,
		`SELECT id, name FROM compute_clusters WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *pgComputeClusterStore) FindByName(ctx context.Context, name string) (*models.ComputeCluster, error) {
	var c models.ComputeCluster
	err := s.db.GetContext(ctx, &c,
		`SELECT id, name FROM compute_clusters WHERE name = $1`, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (s *pgComputeClusterStore) List(ctx context.Context) ([]models.ComputeCluster, error) {
	var clusters []models.ComputeCluster
	err := s.db.SelectContext(ctx, &clusters,
		`SELECT id, name FROM compute_clusters ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return clusters, nil
}

func (s *pgComputeClusterStore) Create(ctx context.Context, tx *sql.Tx, c *models.ComputeCluster) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_clusters (id, name) VALUES ($1, $2)`,
		c.ID, c.Name)
	return err
}

func (s *pgComputeClusterStore) Update(ctx context.Context, tx *sql.Tx, c *models.ComputeCluster) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_clusters SET name = $1 WHERE id = $2`,
		c.Name, c.ID)
	return err
}

func (s *pgComputeClusterStore) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM compute_clusters WHERE id = $1`, id)
	return err
}
