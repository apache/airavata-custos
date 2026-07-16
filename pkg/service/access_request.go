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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
)

// defaultAccessRequestLifetime is how long an approved trial account lasts
// when the approver does not pick an explicit expiry.
const defaultAccessRequestLifetime = 30 * 24 * time.Hour

// GetAccessEvent retrieves an access event by its code. Returns ErrNotFound
// when no event matches.
func (s *Service) GetAccessEvent(ctx context.Context, code string) (*models.AccessEvent, error) {
	if code == "" {
		return nil, fmt.Errorf("%w: event code is required", ErrInvalidInput)
	}
	ev, err := s.accessEvents.FindByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("get access event: %w", err)
	}
	if ev == nil {
		return nil, fmt.Errorf("%w: access event %q", ErrNotFound, code)
	}
	return ev, nil
}

// CreateAccessRequest records a self-service access request identified by a
// verified token subject. The event code must resolve and only one PENDING
// request per subject is allowed. A "CREATED" event is appended to the
// request's audit log in the same transaction.
func (s *Service) CreateAccessRequest(ctx context.Context, req *models.AccessRequest) (*models.AccessRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: access request is nil", ErrInvalidInput)
	}
	if req.OIDCSub == "" {
		return nil, fmt.Errorf("%w: oidc_sub is required", ErrInvalidInput)
	}
	if req.Email == "" {
		return nil, fmt.Errorf("%w: email is required", ErrInvalidInput)
	}
	if req.Institution == "" {
		return nil, fmt.Errorf("%w: institution is required", ErrInvalidInput)
	}
	if req.EventCode == "" {
		return nil, fmt.Errorf("%w: event_code is required", ErrInvalidInput)
	}

	if ev, err := s.accessEvents.FindByCode(ctx, req.EventCode); err != nil {
		return nil, fmt.Errorf("lookup access event: %w", err)
	} else if ev == nil {
		return nil, fmt.Errorf("%w: access event %q", ErrNotFound, req.EventCode)
	}

	if pending, err := s.accessRequests.HasPendingBySub(ctx, req.OIDCSub); err != nil {
		return nil, fmt.Errorf("check pending access request: %w", err)
	} else if pending {
		return nil, fmt.Errorf("%w: a pending access request already exists for this subject", ErrAlreadyExists)
	}

	if req.ID == "" {
		req.ID = newID()
	}
	if req.Status == "" {
		req.Status = models.AccessRequestPending
	}
	if req.Timestamp.IsZero() {
		req.Timestamp = nowUTC()
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.accessRequests.Create(ctx, tx, req); err != nil {
			return err
		}
		return s.accessRequestEvents.Create(ctx, tx, &models.AccessRequestEvent{
			ID:              newID(),
			AccessRequestID: req.ID,
			EventType:       models.AccessRequestEventCreated,
			Description:     "Access request created",
			Timestamp:       req.Timestamp,
		})
	}); err != nil {
		return nil, fmt.Errorf("create access request: %w", err)
	}
	return req, nil
}

// GetLatestAccessRequestBySub returns the newest access request submitted by
// the given token subject. Returns ErrNotFound when the subject has none.
func (s *Service) GetLatestAccessRequestBySub(ctx context.Context, sub string) (*models.AccessRequest, error) {
	if sub == "" {
		return nil, fmt.Errorf("%w: oidc_sub is required", ErrInvalidInput)
	}
	req, err := s.accessRequests.FindLatestBySub(ctx, sub)
	if err != nil {
		return nil, fmt.Errorf("get latest access request: %w", err)
	}
	if req == nil {
		return nil, ErrNotFound
	}
	return req, nil
}

// ListAccessRequests returns access requests filtered by the supplied criteria.
func (s *Service) ListAccessRequests(ctx context.Context, f store.AccessRequestListFilter) ([]models.AccessRequest, error) {
	rows, err := s.accessRequests.List(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list access requests: %w", err)
	}
	return rows, nil
}

// ListAccessRequestDecisionEvents returns the APPROVED/DENIED audit rows for
// the given access request ids, ordered by timestamp ascending.
func (s *Service) ListAccessRequestDecisionEvents(ctx context.Context, requestIDs []string) ([]models.AccessRequestEvent, error) {
	rows, err := s.accessRequestEvents.FindDecisionEventsByRequestIDs(ctx, requestIDs)
	if err != nil {
		return nil, fmt.Errorf("list access request decision events: %w", err)
	}
	return rows, nil
}

// ApproveAccessRequest transitions a PENDING request to APPROVED. Account
// provisioning runs first; if any step fails the request stays PENDING with
// a FAILED event row and the error is returned, so a later re-approve can
// pick up where the partial run stopped. On success the status transition
// and an APPROVED event row commit in one transaction.
func (s *Service) ApproveAccessRequest(ctx context.Context, id, approverID string, expiresAt *time.Time) (*models.AccessRequest, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: access request id is required", ErrInvalidInput)
	}
	if approverID == "" {
		return nil, fmt.Errorf("%w: approver id is required", ErrInvalidInput)
	}
	req, err := s.accessRequests.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("lookup access request: %w", err)
	}
	if req == nil {
		return nil, ErrNotFound
	}
	if req.Status != models.AccessRequestPending {
		return nil, fmt.Errorf("%w: access request is %s, only PENDING can be approved", ErrInvalidInput, req.Status)
	}

	exp := nowUTC().Add(defaultAccessRequestLifetime)
	if expiresAt != nil {
		exp = expiresAt.UTC()
	}

	user, err := s.provisionAccessRequest(ctx, req, exp)
	if err != nil {
		s.appendAccessRequestEvent(ctx, req.ID, models.AccessRequestEventFailed, fmt.Sprintf("Approval failed: %v", err))
		return nil, err
	}

	req.Status = models.AccessRequestApproved
	req.ApproverID = approverID
	req.ExpiresAt = &exp
	req.CreatedUserID = user.ID
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.accessRequests.Update(ctx, tx, req); err != nil {
			return err
		}
		return s.accessRequestEvents.Create(ctx, tx, &models.AccessRequestEvent{
			ID:              newID(),
			AccessRequestID: req.ID,
			EventType:       models.AccessRequestEventApproved,
			Description:     "Access request approved",
			Timestamp:       nowUTC(),
		})
	}); err != nil {
		return nil, fmt.Errorf("approve access request: %w", err)
	}
	return req, nil
}

// provisionAccessRequest creates the user, identity binding, cluster user,
// and allocation membership for an approval. Every step tolerates leftovers
// from an earlier partial run so a failed approval can be retried.
func (s *Service) provisionAccessRequest(ctx context.Context, req *models.AccessRequest, expiresAt time.Time) (*models.User, error) {
	ev, err := s.accessEvents.FindByCode(ctx, req.EventCode)
	if err != nil {
		return nil, fmt.Errorf("lookup access event: %w", err)
	}
	if ev == nil {
		return nil, fmt.Errorf("%w: access event %q", ErrNotFound, req.EventCode)
	}
	alloc, err := s.allocs.FindByID(ctx, ev.ComputeAllocationID)
	if err != nil {
		return nil, fmt.Errorf("lookup compute allocation: %w", err)
	}
	if alloc == nil {
		return nil, fmt.Errorf("%w: compute allocation %q not found", ErrInvalidInput, ev.ComputeAllocationID)
	}
	if alloc.Status != models.ACTIVE {
		return nil, fmt.Errorf("%w: compute allocation %q is %s, must be ACTIVE", ErrInvalidInput, alloc.ID, alloc.Status)
	}

	first, last := splitName(req.Name)
	user, err := s.CreateUser(ctx, &models.User{
		OrganizationID: ev.OrganizationID,
		FirstName:      first,
		LastName:       last,
		Email:          req.Email,
		Status:         models.UserActive,
	})
	if errors.Is(err, ErrAlreadyExists) {
		user, err = s.users.FindByEmail(ctx, req.Email)
		if err == nil && user == nil {
			err = fmt.Errorf("user with email %q vanished during approval", req.Email)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	if _, err := s.CreateUserIdentity(ctx, &models.UserIdentity{
		UserID:     user.ID,
		Source:     identitySourceOIDC,
		ExternalID: req.OIDCSub,
		OIDCSub:    req.OIDCSub,
		Email:      req.Email,
	}); err != nil && !errors.Is(err, ErrAlreadyExists) {
		return nil, fmt.Errorf("create user identity: %w", err)
	}

	// Check the pair first: the allocator treats a pair duplicate as a
	// username collision and would burn its retry budget on the same pair.
	if existing, err := s.clusterUsers.FindByPair(ctx, alloc.ComputeClusterID, user.ID); err != nil {
		return nil, fmt.Errorf("lookup compute cluster user: %w", err)
	} else if existing == nil {
		if _, err := s.AllocateComputeClusterUser(ctx, user, alloc.ComputeClusterID); err != nil {
			return nil, fmt.Errorf("allocate compute cluster user: %w", err)
		}
	}

	if _, err := s.CreateComputeAllocationMembership(ctx, &models.ComputeAllocationMembership{
		ComputeAllocationID: alloc.ID,
		UserID:              user.ID,
		StartTime:           nowUTC(),
		EndTime:             expiresAt,
		MembershipStatus:    models.ACTIVE,
	}); err != nil && !errors.Is(err, ErrAlreadyExists) {
		return nil, fmt.Errorf("create compute allocation membership: %w", err)
	}
	return user, nil
}

// DenyAccessRequest transitions a PENDING request to DENIED, recording the
// reason and the approver. A DENIED event row is appended in-tx.
func (s *Service) DenyAccessRequest(ctx context.Context, id, approverID, denyReason string) (*models.AccessRequest, error) {
	if id == "" {
		return nil, fmt.Errorf("%w: access request id is required", ErrInvalidInput)
	}
	if approverID == "" {
		return nil, fmt.Errorf("%w: approver id is required", ErrInvalidInput)
	}
	req, err := s.accessRequests.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("lookup access request: %w", err)
	}
	if req == nil {
		return nil, ErrNotFound
	}
	if req.Status != models.AccessRequestPending {
		return nil, fmt.Errorf("%w: access request is %s, only PENDING can be denied", ErrInvalidInput, req.Status)
	}

	req.Status = models.AccessRequestDenied
	req.ApproverID = approverID
	req.DenyReason = denyReason
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.accessRequests.Update(ctx, tx, req); err != nil {
			return err
		}
		return s.accessRequestEvents.Create(ctx, tx, &models.AccessRequestEvent{
			ID:              newID(),
			AccessRequestID: req.ID,
			EventType:       models.AccessRequestEventDenied,
			Description:     "Access request denied",
			Timestamp:       nowUTC(),
		})
	}); err != nil {
		return nil, fmt.Errorf("deny access request: %w", err)
	}
	return req, nil
}

// appendAccessRequestEvent writes a best-effort audit row outside the main
// flow; a failure here must not mask the error being reported.
func (s *Service) appendAccessRequestEvent(ctx context.Context, requestID, eventType, description string) {
	_ = s.inTx(ctx, func(tx *sql.Tx) error {
		return s.accessRequestEvents.Create(ctx, tx, &models.AccessRequestEvent{
			ID:              newID(),
			AccessRequestID: requestID,
			EventType:       eventType,
			Description:     description,
			Timestamp:       nowUTC(),
		})
	})
}

// splitName maps a display name onto first/last: first token, then the rest.
func splitName(full string) (first, last string) {
	parts := strings.Fields(full)
	switch len(parts) {
	case 0:
		return "", ""
	case 1:
		return parts[0], ""
	default:
		return parts[0], strings.Join(parts[1:], " ")
	}
}
