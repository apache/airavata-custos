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

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
)

// ListChangeRequests returns change requests filtered by the supplied
// criteria. Used by the portal's admin change-request queue.
func (s *Service) ListChangeRequests(ctx context.Context, f store.ChangeRequestListFilter) ([]models.ComputeAllocationChangeRequest, error) {
	rows, err := s.changeRequests.List(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list change requests: %w", err)
	}
	return rows, nil
}

// CreateComputeAllocationChangeRequest records a new change request against a
// compute allocation. The referenced allocation must exist and a requester
// must be supplied. ChangeStatus defaults to "PENDING" and Timestamp to the
// server's current UTC time. A "CREATED" event is appended to the change
// request's audit log in the same transaction.
func (s *Service) CreateComputeAllocationChangeRequest(ctx context.Context, req *models.ComputeAllocationChangeRequest) (*models.ComputeAllocationChangeRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: compute allocation change request is nil", ErrInvalidInput)
	}
	if req.ComputeAllocationID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_id is required", ErrInvalidInput)
	}
	if req.RequesterID == "" {
		return nil, fmt.Errorf("%w: requester_id is required", ErrInvalidInput)
	}

	if alloc, err := s.allocs.FindByID(ctx, req.ComputeAllocationID); err != nil {
		return nil, fmt.Errorf("lookup compute allocation: %w", err)
	} else if alloc == nil {
		return nil, fmt.Errorf("%w: compute allocation %q not found", ErrInvalidInput, req.ComputeAllocationID)
	}

	if req.ID == "" {
		req.ID = newID()
	}
	if req.ChangeStatus == "" {
		req.ChangeStatus = "PENDING"
	}
	if req.Timestamp.IsZero() {
		req.Timestamp = nowUTC()
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.changeRequests.Create(ctx, tx, req); err != nil {
			return err
		}
		return s.changeEvents.Create(ctx, tx, &models.ComputeAllocationChangeRequestEvent{
			ID:                               newID(),
			ComputeAllocationChangeRequestID: req.ID,
			EventType:                        "CREATED",
			Description:                      "Change request created",
			Timestamp:                        req.Timestamp,
		})
	}); err != nil {
		return nil, fmt.Errorf("create compute allocation change request: %w", err)
	}
	return req, nil
}

// GetComputeAllocationChangeRequest retrieves a change request by its ID.
// Returns ErrNotFound when no request matches.
func (s *Service) GetComputeAllocationChangeRequest(ctx context.Context, id string) (*models.ComputeAllocationChangeRequest, error) {
	c, err := s.changeRequests.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get compute allocation change request: %w", err)
	}
	if c == nil {
		return nil, ErrNotFound
	}
	return c, nil
}

// ListChangeRequestsForAllocation returns every change request ever recorded
// against the given allocation, ordered chronologically by timestamp.
func (s *Service) ListChangeRequestsForAllocation(ctx context.Context, allocationID string) ([]models.ComputeAllocationChangeRequest, error) {
	if allocationID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_id is required", ErrInvalidInput)
	}
	rows, err := s.changeRequests.FindByAllocation(ctx, allocationID)
	if err != nil {
		return nil, fmt.Errorf("list change requests for allocation: %w", err)
	}
	return rows, nil
}

// ListChangeRequestsByRequester returns every change request submitted by the
// given user, ordered chronologically by timestamp.
func (s *Service) ListChangeRequestsByRequester(ctx context.Context, requesterID string) ([]models.ComputeAllocationChangeRequest, error) {
	if requesterID == "" {
		return nil, fmt.Errorf("%w: requester_id is required", ErrInvalidInput)
	}
	rows, err := s.changeRequests.FindByRequester(ctx, requesterID)
	if err != nil {
		return nil, fmt.Errorf("list change requests by requester: %w", err)
	}
	return rows, nil
}

// UpdateComputeAllocationChangeRequest replaces mutable fields of an existing
// change request — typically used to transition change_status and record the
// approver. An audit event is appended in the same transaction: when
// change_status transitions to a new value the event_type is the new status
// (e.g. "APPROVED"), otherwise "UPDATED".
func (s *Service) UpdateComputeAllocationChangeRequest(ctx context.Context, req *models.ComputeAllocationChangeRequest) (*models.ComputeAllocationChangeRequest, error) {
	if req == nil || req.ID == "" {
		return nil, fmt.Errorf("%w: compute allocation change request id is required", ErrInvalidInput)
	}
	existing, err := s.changeRequests.FindByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("lookup compute allocation change request: %w", err)
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if req.ComputeAllocationID == "" {
		req.ComputeAllocationID = existing.ComputeAllocationID
	}
	if req.RequesterID == "" {
		req.RequesterID = existing.RequesterID
	}
	if req.ChangeStatus == "" {
		req.ChangeStatus = existing.ChangeStatus
	}
	if req.Timestamp.IsZero() {
		req.Timestamp = existing.Timestamp
	}

	eventType := "UPDATED"
	eventDescription := "Change request updated"
	if req.ChangeStatus != existing.ChangeStatus {
		eventType = req.ChangeStatus
		eventDescription = fmt.Sprintf("Change request transitioned from %s to %s", existing.ChangeStatus, req.ChangeStatus)
	}

	// When a change request transitions into APPROVED, materialise the
	// approved change as a ComputeAllocationDiff so the allocation's audit
	// log reflects the agreed-upon new SU amount and status.
	approved := req.ChangeStatus == "APPROVED" && existing.ChangeStatus != "APPROVED"

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.changeRequests.Update(ctx, tx, req); err != nil {
			return err
		}
		return s.changeEvents.Create(ctx, tx, &models.ComputeAllocationChangeRequestEvent{
			ID:                               newID(),
			ComputeAllocationChangeRequestID: req.ID,
			EventType:                        eventType,
			Description:                      eventDescription,
			Timestamp:                        nowUTC(),
		})
	}); err != nil {
		return nil, fmt.Errorf("update compute allocation change request: %w", err)
	}

	if approved {
		diffStatus := req.RequestedStatus
		if diffStatus == "" {
			diffStatus = models.ACTIVE
		}
		if _, err := s.CreateComputeAllocationDiff(ctx, &models.ComputeAllocationDiff{
			ComputeAllocationID: req.ComputeAllocationID,
			DiffType:            "CHANGE_REQUEST_APPROVED",
			NewSUAmount:         req.RequestedSUAmount,
			Status:              diffStatus,
			Description:         fmt.Sprintf("Applied approved change request %s", req.ID),
		}); err != nil {
			return nil, fmt.Errorf("record approved change request diff: %w", err)
		}
	}
	return req, nil
}

// DeleteComputeAllocationChangeRequest removes a change request by ID. A
// "DELETED" event is appended to the audit log first so the trail survives
// the parent record's removal.
func (s *Service) DeleteComputeAllocationChangeRequest(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: compute allocation change request id is required", ErrInvalidInput)
	}
	existing, err := s.changeRequests.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("lookup compute allocation change request: %w", err)
	}
	if existing == nil {
		return ErrNotFound
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.changeEvents.Create(ctx, tx, &models.ComputeAllocationChangeRequestEvent{
			ID:                               newID(),
			ComputeAllocationChangeRequestID: id,
			EventType:                        "DELETED",
			Description:                      "Change request deleted",
			Timestamp:                        nowUTC(),
		}); err != nil {
			return err
		}
		return s.changeRequests.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete compute allocation change request: %w", err)
	}
	return nil
}
