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

type pgDeletionRequestStore struct {
	db *sqlx.DB
}

func NewDeletionRequestStore(db *sqlx.DB) DeletionRequestStore {
	return &pgDeletionRequestStore{db: db}
}

func (s *pgDeletionRequestStore) Create(ctx context.Context, tx *sql.Tx, r *models.DeletionRequest) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO deletion_requests (entity_type, entity_id, requested_by, requested_at)
		 VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`,
		r.EntityType, r.EntityID, r.RequestedBy, r.RequestedAt)
	return err
}

func (s *pgDeletionRequestStore) Find(ctx context.Context, entityType, entityID string) (*models.DeletionRequest, error) {
	var r models.DeletionRequest
	err := s.db.GetContext(ctx, &r,
		`SELECT entity_type, entity_id, requested_by, requested_at FROM deletion_requests WHERE entity_type = $1 AND entity_id = $2`,
		entityType, entityID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &r, err
}

// List skips a request whose account or allocation is already gone.
func (s *pgDeletionRequestStore) List(ctx context.Context) ([]models.DeletionRequest, error) {
	rows := []models.DeletionRequest{}
	err := s.db.SelectContext(ctx, &rows,
		`SELECT d.entity_type, d.entity_id, d.requested_by, d.requested_at,
		        COALESCE(cu.local_username, a.name) AS name, c.name AS cluster_name
		   FROM deletion_requests d
		   LEFT JOIN compute_cluster_users cu ON d.entity_type = 'compute_cluster_user' AND cu.id = d.entity_id
		   LEFT JOIN compute_allocations a    ON d.entity_type = 'compute_allocation' AND a.id = d.entity_id
		   JOIN compute_clusters c            ON c.id = COALESCE(cu.compute_cluster_id, a.compute_cluster_id)
		  ORDER BY d.requested_at`)
	return rows, err
}

func (s *pgDeletionRequestStore) Delete(ctx context.Context, tx *sql.Tx, entityType, entityID string) error {
	_, err := tx.ExecContext(ctx, `DELETE FROM deletion_requests WHERE entity_type = $1 AND entity_id = $2`, entityType, entityID)
	return err
}
