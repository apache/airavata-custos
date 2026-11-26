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

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

const computeClusterQoSColumns = `id, compute_cluster_id, name, priority, max_wall_minutes, max_jobs_per_user, max_tres_per_user, flags, is_default, slurm_pending_since`

type pgComputeClusterQoSStore struct {
	db *sqlx.DB
}

// NewComputeClusterQoSStore returns a PostgreSQL-backed ComputeClusterQoSStore.
func NewComputeClusterQoSStore(db *sqlx.DB) ComputeClusterQoSStore {
	return &pgComputeClusterQoSStore{db: db}
}

var pgTypes = pgtype.NewMap()

type qosRows interface {
	Scan(dest ...any) error
}

func scanComputeClusterQoS(row qosRows) (*models.ComputeClusterQoS, error) {
	var q models.ComputeClusterQoS
	err := row.Scan(&q.ID, &q.ComputeClusterID, &q.Name, &q.Priority, &q.MaxWallMinutes, &q.MaxJobsPerUser,
		&q.MaxTRESPerUser, pgTypes.SQLScanner(&q.Flags), &q.IsDefault, &q.SlurmPendingSince)
	if q.Flags == nil {
		q.Flags = []string{}
	}
	return &q, err
}

func (s *pgComputeClusterQoSStore) FindByIDTx(ctx context.Context, tx *sql.Tx, id string, lock RowLock) (*models.ComputeClusterQoS, error) {
	q, err := scanComputeClusterQoS(tx.QueryRowContext(ctx,
		`SELECT `+computeClusterQoSColumns+` FROM compute_cluster_qos WHERE id = $1 `+string(lock), id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return q, err
}

func (s *pgComputeClusterQoSStore) FindByCluster(ctx context.Context, clusterID, allocationID string) ([]models.ComputeClusterQoS, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+computeClusterQoSColumns+` FROM compute_cluster_qos WHERE compute_cluster_id = $1
		    AND ($2 = '' OR id IN (SELECT compute_cluster_qos_id FROM compute_allocation_qos WHERE compute_allocation_id = $2))
		  ORDER BY name`, clusterID, allocationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.ComputeClusterQoS{}
	for rows.Next() {
		q, err := scanComputeClusterQoS(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *q)
	}
	return list, rows.Err()
}

// clearOtherDefaults unsets the cluster's other default when q is the default.
func clearOtherDefaults(ctx context.Context, tx *sql.Tx, q *models.ComputeClusterQoS) error {
	if !q.IsDefault {
		return nil
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE compute_cluster_qos SET is_default = false WHERE compute_cluster_id = $1 AND is_default AND id <> $2`,
		q.ComputeClusterID, q.ID)
	return err
}

func (s *pgComputeClusterQoSStore) Create(ctx context.Context, tx *sql.Tx, q *models.ComputeClusterQoS) error {
	if err := clearOtherDefaults(ctx, tx, q); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO compute_cluster_qos (id, compute_cluster_id, name, priority, max_wall_minutes, max_jobs_per_user, max_tres_per_user, flags, is_default)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		q.ID, q.ComputeClusterID, q.Name, q.Priority, q.MaxWallMinutes, q.MaxJobsPerUser, q.MaxTRESPerUser, q.Flags, q.IsDefault)
	return err
}

func (s *pgComputeClusterQoSStore) Update(ctx context.Context, tx *sql.Tx, q *models.ComputeClusterQoS) (bool, error) {
	if err := clearOtherDefaults(ctx, tx, q); err != nil {
		return false, err
	}
	return execOnce(ctx, tx,
		`UPDATE compute_cluster_qos SET name = $3, priority = $4, max_wall_minutes = $5, max_jobs_per_user = $6,
                max_tres_per_user = $7, flags = $8, is_default = $9
          WHERE id = $1 AND compute_cluster_id = $2`,
		q.ID, q.ComputeClusterID, q.Name, q.Priority, q.MaxWallMinutes, q.MaxJobsPerUser, q.MaxTRESPerUser, q.Flags, q.IsDefault)
}

func (s *pgComputeClusterQoSStore) Delete(ctx context.Context, tx *sql.Tx, clusterID, id string) (bool, error) {
	return execOnce(ctx, tx, `DELETE FROM compute_cluster_qos WHERE id = $1 AND compute_cluster_id = $2`, id, clusterID)
}

func (s *pgComputeClusterQoSStore) ReplaceForAllocation(ctx context.Context, tx *sql.Tx, allocationID string, qosIDs []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM compute_allocation_qos WHERE compute_allocation_id = $1`, allocationID); err != nil {
		return err
	}
	for _, id := range qosIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO compute_allocation_qos (compute_allocation_id, compute_cluster_qos_id) VALUES ($1, $2)`, allocationID, id); err != nil {
			return err
		}
	}
	return nil
}
