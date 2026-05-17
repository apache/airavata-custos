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

	"github.com/apache/airavata-custos/pkg/models"
)

// CreateComputeAllocationChangeRequestEvent records a new event against a
// change request. The referenced change request must exist. If Timestamp is
// zero the server's current UTC time is used. Event records are append-only
// and have no update operation.
func (s *Service) CreateComputeAllocationChangeRequestEvent(ctx context.Context, evt *models.ComputeAllocationChangeRequestEvent) (*models.ComputeAllocationChangeRequestEvent, error) {
	if evt == nil {
		return nil, fmt.Errorf("%w: compute allocation change request event is nil", ErrInvalidInput)
	}
	if evt.ComputeAllocationChangeRequestID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_change_request_id is required", ErrInvalidInput)
	}
	if evt.EventType == "" {
		return nil, fmt.Errorf("%w: event_type is required", ErrInvalidInput)
	}

	if cr, err := s.changeRequests.FindByID(ctx, evt.ComputeAllocationChangeRequestID); err != nil {
		return nil, fmt.Errorf("lookup compute allocation change request: %w", err)
	} else if cr == nil {
		return nil, fmt.Errorf("%w: compute allocation change request %q not found", ErrInvalidInput, evt.ComputeAllocationChangeRequestID)
	}

	if evt.ID == "" {
		evt.ID = newID()
	}
	if evt.Timestamp.IsZero() {
		evt.Timestamp = nowUTC()
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.changeEvents.Create(ctx, tx, evt)
	}); err != nil {
		return nil, fmt.Errorf("create compute allocation change request event: %w", err)
	}
	return evt, nil
}

// GetComputeAllocationChangeRequestEvent retrieves an event by its ID.
// Returns ErrNotFound when no event matches.
func (s *Service) GetComputeAllocationChangeRequestEvent(ctx context.Context, id string) (*models.ComputeAllocationChangeRequestEvent, error) {
	e, err := s.changeEvents.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get compute allocation change request event: %w", err)
	}
	if e == nil {
		return nil, ErrNotFound
	}
	return e, nil
}

// ListEventsForChangeRequest returns every event recorded against the given
// change request, ordered chronologically by timestamp.
func (s *Service) ListEventsForChangeRequest(ctx context.Context, changeRequestID string) ([]models.ComputeAllocationChangeRequestEvent, error) {
	if changeRequestID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_change_request_id is required", ErrInvalidInput)
	}
	rows, err := s.changeEvents.FindByChangeRequest(ctx, changeRequestID)
	if err != nil {
		return nil, fmt.Errorf("list events for change request: %w", err)
	}
	return rows, nil
}

// GetLatestEventForChangeRequest returns the most recent event for the given
// change request. Returns ErrNotFound when the change request has no events.
func (s *Service) GetLatestEventForChangeRequest(ctx context.Context, changeRequestID string) (*models.ComputeAllocationChangeRequestEvent, error) {
	if changeRequestID == "" {
		return nil, fmt.Errorf("%w: compute_allocation_change_request_id is required", ErrInvalidInput)
	}
	e, err := s.changeEvents.FindLatestByChangeRequest(ctx, changeRequestID)
	if err != nil {
		return nil, fmt.Errorf("get latest event for change request: %w", err)
	}
	if e == nil {
		return nil, ErrNotFound
	}
	return e, nil
}

// DeleteComputeAllocationChangeRequestEvent removes an event record by ID.
// This is intended for administrative cleanup of erroneous entries; the
// event log is otherwise append-only.
func (s *Service) DeleteComputeAllocationChangeRequestEvent(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: compute allocation change request event id is required", ErrInvalidInput)
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.changeEvents.Delete(ctx, tx, id)
	}); err != nil {
		return fmt.Errorf("delete compute allocation change request event: %w", err)
	}
	return nil
}
