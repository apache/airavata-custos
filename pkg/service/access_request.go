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
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/apache/airavata-custos/internal/email"
	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/posix"
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
	if req.DesiredUsername != "" && !posix.ValidChosen(req.DesiredUsername) {
		return nil, fmt.Errorf("%w: desired_username must be 1-32 characters, start with a lowercase letter", ErrInvalidInput)
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

	if s.autoApproveEmails[strings.ToLower(strings.TrimSpace(req.Email))] {
		if approved := s.autoApproveRequest(ctx, req); approved != nil {
			return approved, nil
		}
	}
	return req, nil
}

// autoApproveRequest approves a just-created request on behalf of the configured
// approver with a far-future expiry, so known team members skip the manual queue.
// Returns nil (leaving the request PENDING for a manual approve) if the approver
// is unknown or provisioning fails.
func (s *Service) autoApproveRequest(ctx context.Context, req *models.AccessRequest) *models.AccessRequest {
	approver, err := s.users.FindByEmail(ctx, s.autoApproveApprover)
	if err != nil || approver == nil {
		slog.Warn("auto-approve skipped: approver not found", "approver", s.autoApproveApprover, "email", req.Email)
		return nil
	}
	// Membership end_time is TIMESTAMP(6), capped at 2038-01-19, so this near-max
	// is the practical "no expiry" for team members.
	exp := time.Date(2038, 1, 1, 0, 0, 0, 0, time.UTC)
	approved, err := s.ApproveAccessRequest(ctx, req.ID, approver.ID, &exp)
	if err != nil {
		slog.Warn("auto-approve failed, request left pending", "email", req.Email, "error", err)
		return nil
	}
	slog.Info("auto-approved access request", "email", req.Email)
	return approved
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

	user, username, err := s.provisionAccessRequest(ctx, req, exp)
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

	// Best effort and off the request path: the approval stands even when
	// the notification cannot be delivered.
	go s.sendAccountReadyEmail(req, username, exp)

	return req, nil
}

func (s *Service) sendAccountReadyEmail(req *models.AccessRequest, username string, expiresAt time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	first, _ := splitName(req.Name)
	err := s.mailer.SendAccountReady(ctx, req.Email, email.AccountReadyData{
		Username:  username,
		FirstName: first,
		ExpiresOn: expiresAt.Format("January 2, 2006"),
	})
	if err != nil {
		slog.Error("account-ready email failed", "request_id", req.ID, "error", err)
		s.appendAccessRequestEvent(ctx, req.ID, models.AccessRequestEventFailed, fmt.Sprintf("Account-ready email failed: %v", err))
	}
}

// provisionAccessRequest creates the user, identity binding, cluster user,
// and allocation membership for an approval. Every step tolerates leftovers
// from an earlier partial run so a failed approval can be retried.
func (s *Service) provisionAccessRequest(ctx context.Context, req *models.AccessRequest, expiresAt time.Time) (*models.User, string, error) {
	ev, err := s.accessEvents.FindByCode(ctx, req.EventCode)
	if err != nil {
		return nil, "", fmt.Errorf("lookup access event: %w", err)
	}
	if ev == nil {
		return nil, "", fmt.Errorf("%w: access event %q", ErrNotFound, req.EventCode)
	}
	alloc, err := s.allocs.FindByID(ctx, ev.ComputeAllocationID)
	if err != nil {
		return nil, "", fmt.Errorf("lookup compute allocation: %w", err)
	}
	if alloc == nil {
		return nil, "", fmt.Errorf("%w: compute allocation %q not found", ErrInvalidInput, ev.ComputeAllocationID)
	}
	if alloc.Status != models.ACTIVE {
		return nil, "", fmt.Errorf("%w: compute allocation %q is %s, must be ACTIVE", ErrInvalidInput, alloc.ID, alloc.Status)
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
		return nil, "", fmt.Errorf("create user: %w", err)
	}

	if _, err := s.CreateUserIdentity(ctx, &models.UserIdentity{
		UserID:     user.ID,
		Source:     identitySourceOIDC,
		ExternalID: req.OIDCSub,
		OIDCSub:    req.OIDCSub,
		Email:      req.Email,
	}); err != nil && !errors.Is(err, ErrAlreadyExists) {
		return nil, "", fmt.Errorf("create user identity: %w", err)
	}

	// Check the pair first: the allocator treats a pair duplicate as a
	// username collision and would burn its retry budget on the same pair.
	username := ""
	if existing, err := s.clusterUsers.FindByPair(ctx, alloc.ComputeClusterID, user.ID); err != nil {
		return nil, "", fmt.Errorf("lookup compute cluster user: %w", err)
	} else if existing != nil {
		username = existing.LocalUsername
	} else if req.DesiredUsername != "" {
		// The requester picked this login and it was free when they asked; a
		// duplicate here means someone took it since, so the approval fails and
		// stays PENDING for a retry with a different pick.
		ccu, err := s.CreateComputeClusterUser(ctx, &models.ComputeClusterUser{
			ComputeClusterID: alloc.ComputeClusterID,
			UserID:           user.ID,
			LocalUsername:    req.DesiredUsername,
		})
		if err != nil {
			return nil, "", fmt.Errorf("create cluster user with chosen username %q: %w", req.DesiredUsername, err)
		}
		username = ccu.LocalUsername
	} else {
		ccu, err := s.AllocateComputeClusterUser(ctx, user, alloc.ComputeClusterID)
		if err != nil {
			return nil, "", fmt.Errorf("allocate compute cluster user: %w", err)
		}
		username = ccu.LocalUsername
	}

	if _, err := s.CreateComputeAllocationMembership(ctx, &models.ComputeAllocationMembership{
		ComputeAllocationID: alloc.ID,
		UserID:              user.ID,
		StartTime:           nowUTC(),
		EndTime:             expiresAt,
		MembershipStatus:    models.ACTIVE,
	}); err != nil && !errors.Is(err, ErrAlreadyExists) {
		return nil, "", fmt.Errorf("create compute allocation membership: %w", err)
	}
	return user, username, nil
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

// UsernameCheck reports on a requested cluster login: a guaranteed-free
// suggestion plus the standing of the candidate the requester typed.
type UsernameCheck struct {
	Suggestion string
	Valid      bool
	Available  bool
}

// CheckAccessRequestUsername resolves the event's cluster, builds an available
// suggestion from the requester's identity, and reports whether the candidate
// login (when supplied) is well-formed and free on that cluster.
func (s *Service) CheckAccessRequestUsername(ctx context.Context, eventCode, candidate, name, email string) (*UsernameCheck, error) {
	if eventCode == "" {
		return nil, fmt.Errorf("%w: event code is required", ErrInvalidInput)
	}
	ev, err := s.accessEvents.FindByCode(ctx, eventCode)
	if err != nil {
		return nil, fmt.Errorf("lookup access event: %w", err)
	}
	if ev == nil {
		return nil, fmt.Errorf("%w: access event %q", ErrNotFound, eventCode)
	}
	alloc, err := s.allocs.FindByID(ctx, ev.ComputeAllocationID)
	if err != nil {
		return nil, fmt.Errorf("lookup compute allocation: %w", err)
	}
	if alloc == nil {
		return nil, fmt.Errorf("%w: compute allocation %q not found", ErrNotFound, ev.ComputeAllocationID)
	}

	suggestion, err := s.suggestClusterUsername(ctx, alloc.ComputeClusterID, name, email)
	if err != nil {
		return nil, err
	}
	res := &UsernameCheck{Suggestion: suggestion, Valid: true, Available: true}
	if candidate != "" {
		res.Valid = posix.ValidChosen(candidate)
		if !res.Valid {
			res.Available = false
			return res, nil
		}
		free, err := s.clusterUsernameFree(ctx, alloc.ComputeClusterID, candidate)
		if err != nil {
			return nil, err
		}
		res.Available = free
	}
	return res, nil
}

// clusterUsernameFree reports whether localUsername is unclaimed on the cluster.
func (s *Service) clusterUsernameFree(ctx context.Context, clusterID, localUsername string) (bool, error) {
	_, err := s.GetComputeClusterUserByClusterAndLocalUsername(ctx, clusterID, localUsername)
	if errors.Is(err, ErrNotFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

// suggestClusterUsername builds the generated-style base from the requester's
// name (falling back to the email local part) and appends the lowest free
// numeric suffix.
func (s *Service) suggestClusterUsername(ctx context.Context, clusterID, name, email string) (string, error) {
	first, last := splitName(name)
	base, _, err := posix.BuildBase(&models.User{FirstName: first, LastName: last}, posix.Prefix())
	if err != nil {
		local := posix.Normalize(emailLocalPart(email))
		if local == "" {
			local = "user"
		}
		base = posix.Prefix() + "-" + local
	}
	for n := 0; n < posix.MaxCollisionSuffix; n++ {
		candidate := base
		if n > 0 {
			candidate = base + strconv.Itoa(n+1)
		}
		free, err := s.clusterUsernameFree(ctx, clusterID, candidate)
		if err != nil {
			return "", err
		}
		if free {
			return candidate, nil
		}
	}
	return base, nil
}

// emailLocalPart returns the portion of an address before the "@".
func emailLocalPart(email string) string {
	if i := strings.IndexByte(email, '@'); i > 0 {
		return email[:i]
	}
	return email
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
