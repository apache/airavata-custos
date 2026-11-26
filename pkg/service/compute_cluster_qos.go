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
	"slices"
	"strconv"
	"strings"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
)

// ListComputeClusterQoS returns the cluster's QoS definitions, or those the allocation grants when allocationID is set, ordered by name.
func (s *Service) ListComputeClusterQoS(ctx context.Context, clusterID, allocationID string) ([]models.ComputeClusterQoS, error) {
	if _, err := s.GetComputeCluster(ctx, clusterID); err != nil {
		return nil, err
	}
	list, err := s.clusterQoS.FindByCluster(ctx, clusterID, allocationID)
	if err != nil {
		return nil, fmt.Errorf("list compute cluster qos: %w", err)
	}
	return list, nil
}

// CreateComputeClusterQoS stores a QoS on the cluster, pending a SLURM write.
func (s *Service) CreateComputeClusterQoS(ctx context.Context, q *models.ComputeClusterQoS) (*models.ComputeClusterQoS, error) {
	if err := validateComputeClusterQoS(q); err != nil {
		return nil, err
	}
	if _, err := s.GetComputeCluster(ctx, q.ComputeClusterID); err != nil {
		return nil, err
	}
	q.ID = newID()
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.clusterQoS.Create(ctx, tx, q); err != nil {
			return err
		}
		return s.markQoSChangePending(ctx, tx, nil, q)
	}); err != nil {
		if isQoSNameDuplicate(err) {
			return nil, fmt.Errorf("%w: qos %q on cluster %q", ErrAlreadyExists, q.Name, q.ComputeClusterID)
		}
		return nil, fmt.Errorf("create compute cluster qos: %w", err)
	}
	return q, nil
}

// UpdateComputeClusterQoS replaces the cluster QoS's limits, pending a SLURM write.
func (s *Service) UpdateComputeClusterQoS(ctx context.Context, q *models.ComputeClusterQoS) (*models.ComputeClusterQoS, error) {
	if err := validateComputeClusterQoS(q); err != nil {
		return nil, err
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		// Locks out a concurrent default change until the write.
		before, err := s.clusterQoS.FindByIDTx(ctx, tx, q.ID, store.UpdateLock)
		if err != nil {
			return err
		}
		if found, err := s.clusterQoS.Update(ctx, tx, q); err != nil {
			return err
		} else if !found {
			return ErrNotFound
		}
		return s.markQoSChangePending(ctx, tx, before, q)
	}); err != nil {
		if isQoSNameDuplicate(err) {
			return nil, fmt.Errorf("%w: qos %q on cluster %q", ErrAlreadyExists, q.Name, q.ComputeClusterID)
		}
		return nil, fmt.Errorf("update compute cluster qos: %w", err)
	}
	return q, nil
}

// markQoSChangePending marks the QoS, plus the cluster's allocations on a default change or its grantees on a rename.
func (s *Service) markQoSChangePending(ctx context.Context, tx *sql.Tx, before, after *models.ComputeClusterQoS) error {
	if !slurmMarkWanted(ctx) {
		return nil
	}
	if err := s.markSlurmPendingTx(ctx, tx, SlurmPendingQoS, after.ID); err != nil {
		return err
	}
	wasDefault := before != nil && before.IsDefault
	renamed := before != nil && before.Name != after.Name
	switch {
	case wasDefault != after.IsDefault || (renamed && after.IsDefault):
		_, err := store.MarkSlurmPending(ctx, tx, SlurmPendingAllocation, `compute_cluster_id = $1`, after.ComputeClusterID)
		return err
	case renamed:
		_, err := store.MarkSlurmPending(ctx, tx, SlurmPendingAllocation,
			`id IN (SELECT compute_allocation_id FROM compute_allocation_qos WHERE compute_cluster_qos_id = $1)`, after.ID)
		return err
	}
	return nil
}

// DeleteComputeClusterQoS removes a QoS; ErrInUse while it is the default or granted.
func (s *Service) DeleteComputeClusterQoS(ctx context.Context, clusterID, id string) error {
	var found bool
	if err := s.inTx(ctx, func(tx *sql.Tx) (err error) {
		if q, err := s.clusterQoS.FindByIDTx(ctx, tx, id, store.UpdateLock); err != nil {
			return err
		} else if q != nil && q.ComputeClusterID == clusterID && q.IsDefault {
			return fmt.Errorf("%w: qos %q is the cluster default", ErrInUse, id)
		}
		found, err = s.clusterQoS.Delete(ctx, tx, clusterID, id)
		return err
	}); err != nil {
		if strings.Contains(err.Error(), "fk_compute_allocation_qos_qos") {
			return fmt.Errorf("%w: qos %q is granted by an allocation", ErrInUse, id)
		}
		return fmt.Errorf("delete compute cluster qos: %w", err)
	}
	if !found {
		return ErrNotFound
	}
	return nil
}

// SetComputeAllocationQoS makes the named cluster QoS the allocation's only QoS, pending a SLURM write.
func (s *Service) SetComputeAllocationQoS(ctx context.Context, allocationID string, qosNames []string) error {
	alloc, err := s.GetComputeAllocation(ctx, allocationID)
	if err != nil {
		return err
	}
	onCluster, err := s.clusterQoS.FindByCluster(ctx, alloc.ComputeClusterID, "")
	if err != nil {
		return fmt.Errorf("list compute cluster qos: %w", err)
	}
	idByName := make(map[string]string, len(onCluster))
	for _, q := range onCluster {
		idByName[q.Name] = q.ID
	}
	qosIDs := make([]string, 0, len(qosNames))
	for _, name := range qosNames {
		id, ok := idByName[name]
		if !ok {
			return fmt.Errorf("%w: qos %q is not on the allocation's cluster", ErrInvalidInput, name)
		}
		qosIDs = append(qosIDs, id)
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.clusterQoS.ReplaceForAllocation(ctx, tx, allocationID, slices.Compact(slices.Sorted(slices.Values(qosIDs)))); err != nil {
			return err
		}
		return s.markSlurmPendingTx(ctx, tx, SlurmPendingAllocation, allocationID)
	}); err != nil {
		return fmt.Errorf("set compute allocation qos: %w", err)
	}
	return nil
}

func isQoSNameDuplicate(err error) bool {
	return strings.Contains(err.Error(), "uq_compute_cluster_qos_name")
}

func validateComputeClusterQoS(q *models.ComputeClusterQoS) error {
	switch {
	case q == nil:
		return fmt.Errorf("%w: qos is nil", ErrInvalidInput)
	case q.ComputeClusterID == "":
		return fmt.Errorf("%w: compute_cluster_id is required", ErrInvalidInput)
	case q.Name == "":
		return fmt.Errorf("%w: qos name is required", ErrInvalidInput)
	case !validTRES(q.MaxTRESPerUser):
		return fmt.Errorf("%w: max_tres_per_user %q is not a TRES string such as cpu=4,gres/gpu=2", ErrInvalidInput, q.MaxTRESPerUser)
	}
	if q.Flags == nil {
		q.Flags = []string{}
	}
	return nil
}

// validTRES holds for an empty or comma-separated SLURM TRES string.
func validTRES(s string) bool {
	if strings.TrimSpace(s) == "" {
		return true
	}
	for _, part := range strings.Split(s, ",") {
		key, count, ok := strings.Cut(part, "=")
		if _, err := strconv.ParseFloat(strings.TrimSpace(count), 64); !ok || strings.TrimSpace(key) == "" || err != nil {
			return false
		}
	}
	return true
}
