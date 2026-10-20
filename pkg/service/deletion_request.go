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
	"encoding/json"
	"fmt"

	"github.com/apache/airavata-custos/pkg/models"
)

// Deleting a cluster account or an allocation takes a request and then an admin's approval, which runs the delete and
// so the connectors' removals. A denial drops the request.
const (
	DeletionOfClusterUser = "compute_cluster_user"
	DeletionOfAllocation  = "compute_allocation"
)

// RequestDeletion records a request to delete the cluster account or allocation.
func (s *Service) RequestDeletion(ctx context.Context, entityType, entityID, requestedBy string) error {
	var err error
	var found bool
	switch entityType {
	case DeletionOfClusterUser:
		var cu *models.ComputeClusterUser
		cu, err = s.clusterUsers.FindByID(ctx, entityID)
		found = cu != nil
	case DeletionOfAllocation:
		var a *models.ComputeAllocation
		a, err = s.allocs.FindByID(ctx, entityID)
		found = a != nil
	default:
		return fmt.Errorf("%w: cannot request deletion of %q", ErrInvalidInput, entityType)
	}
	if err != nil {
		return err
	}
	if !found {
		return ErrNotFound
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		r := &models.DeletionRequest{EntityType: entityType, EntityID: entityID, RequestedBy: requestedBy, RequestedAt: nowUTC()}
		if err := s.deletions.Create(ctx, tx, r); err != nil {
			return err
		}
		return s.writeDeletionAuditTx(ctx, tx, "DELETION_REQUESTED", entityType, entityID, requestedBy)
	})
}

func (s *Service) ListDeletionRequests(ctx context.Context) ([]models.DeletionRequest, error) {
	return s.deletions.List(ctx)
}

// ApproveDeletion runs the requested delete, then closes the request.
func (s *Service) ApproveDeletion(ctx context.Context, entityType, entityID, approverID string) error {
	if err := s.pendingDeletion(ctx, entityType, entityID); err != nil {
		return err
	}
	del := s.DeleteComputeAllocation
	if entityType == DeletionOfClusterUser {
		del = s.DeleteComputeClusterUser
	}
	if err := del(ctx, entityID); err != nil {
		return err
	}
	return s.closeDeletion(ctx, "DELETION_APPROVED", entityType, entityID, approverID)
}

// DenyDeletion drops the request and keeps the cluster account or allocation.
func (s *Service) DenyDeletion(ctx context.Context, entityType, entityID, reviewerID string) error {
	if err := s.pendingDeletion(ctx, entityType, entityID); err != nil {
		return err
	}
	return s.closeDeletion(ctx, "DELETION_DENIED", entityType, entityID, reviewerID)
}

func (s *Service) pendingDeletion(ctx context.Context, entityType, entityID string) error {
	r, err := s.deletions.Find(ctx, entityType, entityID)
	if err == nil && r == nil {
		return ErrNotFound
	}
	return err
}

func (s *Service) closeDeletion(ctx context.Context, auditType, entityType, entityID, actorID string) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.deletions.Delete(ctx, tx, entityType, entityID); err != nil {
			return err
		}
		return s.writeDeletionAuditTx(ctx, tx, auditType, entityType, entityID, actorID)
	})
}

func (s *Service) writeDeletionAuditTx(ctx context.Context, tx *sql.Tx, eventType, entityType, entityID, actorID string) error {
	details, err := json.Marshal(map[string]string{"actor_id": actorID})
	if err != nil {
		return err
	}
	return s.auditEvents.Create(ctx, tx, &models.AuditEvent{
		ID: newID(), EventType: eventType, EventTime: nowUTC(), EntityID: entityID, EntityType: entityType, Details: string(details),
	})
}
