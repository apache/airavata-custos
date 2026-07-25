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
	"time"

	"github.com/apache/airavata-custos/pkg/models"
)

// DefaultAccessCheckStuckAfter is how long a check keeps failing before it
// stops reading as a transient retry and is reported as stuck.
const DefaultAccessCheckStuckAfter = time.Hour

// AccessStatus is a user's checks on one allocation plus their milestone
// history, newest first. LocalUsername is the account the user signs in
// with on the allocation's cluster, empty until one is assigned.
type AccessStatus struct {
	Checks        []models.AccessCheck
	Events        []models.AccessCheckEvent
	LocalUsername string
}

// MemberAccessStatus is one allocation member's checks, for the manager view.
type MemberAccessStatus struct {
	UserID        string
	DisplayName   string
	Email         string
	LocalUsername string
	Checks        []models.AccessCheck
}

// SetAccessCheckStuckAfter overrides the failing-to-stuck threshold.
func (s *Service) SetAccessCheckStuckAfter(d time.Duration) {
	if d > 0 {
		s.accessCheckStuckAfter = d
	}
}

// AccessCheckStuckAfter returns the failing-to-stuck threshold.
func (s *Service) AccessCheckStuckAfter() time.Duration {
	if s.accessCheckStuckAfter > 0 {
		return s.accessCheckStuckAfter
	}
	return DefaultAccessCheckStuckAfter
}

// RecordAccessCheckResult upserts one probe result and writes milestone
// events on state changes only. A check that has never passed stays PENDING
// (still setting up); failures only start counting once the check has been
// OK at least once. infrastructure marks a failure as the probe's inability
// to reach the external system: it shows as a calm retry but never
// escalates to stuck, and it never creates a member's first row (the
// synthesized state is a better answer than "we could not look").
func (s *Service) RecordAccessCheckResult(ctx context.Context, allocationID, userID string, checkType models.AccessCheckType, ok bool, detail string, infrastructure bool) error {
	existing, err := s.accessChecks.FindByTarget(ctx, allocationID, userID, checkType)
	if err != nil {
		return fmt.Errorf("find access check: %w", err)
	}
	now := nowUTC()

	return s.inTx(ctx, func(tx *sql.Tx) error {
		if existing == nil {
			if !ok && infrastructure {
				return nil
			}
			c := &models.AccessCheck{
				ID:                  newID(),
				ComputeAllocationID: allocationID,
				UserID:              userID,
				CheckType:           checkType,
				Status:              models.AccessCheckPending,
				Detail:              detail,
				LastCheckedAt:       now,
			}
			if ok {
				c.Status = models.AccessCheckOK
				c.LastOKAt = &now
			}
			if err := s.accessChecks.Create(ctx, tx, c); err != nil {
				return fmt.Errorf("create access check: %w", err)
			}
			// STARTED strictly precedes ONLINE so same-instant rows sort correctly.
			if err := s.writeCheckEvent(ctx, tx, c.ID, models.AccessCheckEventStarted, now.Add(-time.Microsecond)); err != nil {
				return err
			}
			if ok {
				return s.writeCheckEvent(ctx, tx, c.ID, models.AccessCheckEventOnline, now)
			}
			return nil
		}

		prev := existing.Status
		existing.Detail = detail
		existing.LastCheckedAt = now
		existing.Infrastructure = !ok && infrastructure
		switch {
		case ok:
			existing.LastOKAt = &now
			existing.FailingSince = nil
			existing.Status = models.AccessCheckOK
		case prev == models.AccessCheckOK:
			existing.Status = models.AccessCheckFailing
			existing.FailingSince = &now
		}
		if err := s.accessChecks.Update(ctx, tx, existing); err != nil {
			return fmt.Errorf("update access check: %w", err)
		}

		switch {
		case ok && prev == models.AccessCheckPending:
			return s.writeCheckEvent(ctx, tx, existing.ID, models.AccessCheckEventOnline, now)
		case ok && prev == models.AccessCheckFailing:
			return s.writeCheckEvent(ctx, tx, existing.ID, models.AccessCheckEventRecovered, now)
		case !ok && prev == models.AccessCheckOK:
			return s.writeCheckEvent(ctx, tx, existing.ID, models.AccessCheckEventFailed, now)
		case !ok && prev == models.AccessCheckFailing:
			return s.maybeWriteStuckEvent(ctx, tx, existing, now)
		}
		return nil
	})
}

func (s *Service) writeCheckEvent(ctx context.Context, tx *sql.Tx, checkID, eventType string, at time.Time) error {
	err := s.accessChecks.CreateEvent(ctx, tx, &models.AccessCheckEvent{
		ID: newID(), AccessCheckID: checkID, EventType: eventType, OccurredAt: at,
	})
	if err != nil {
		return fmt.Errorf("write %s event: %w", eventType, err)
	}
	return nil
}

// maybeWriteStuckEvent marks the failure episode stuck once it outlives the
// threshold. One STUCK per episode: events after failing_since are checked so
// probe restarts and repeated failures do not repeat it. Infrastructure
// failures never earn a STUCK: the outage is systemic, not the member's.
func (s *Service) maybeWriteStuckEvent(ctx context.Context, tx *sql.Tx, c *models.AccessCheck, now time.Time) error {
	if c.Infrastructure || c.FailingSince == nil || now.Sub(*c.FailingSince) < s.AccessCheckStuckAfter() {
		return nil
	}
	events, err := s.accessChecks.FindEventsByChecks(ctx, []string{c.ID})
	if err != nil {
		return fmt.Errorf("find check events: %w", err)
	}
	for _, e := range events {
		if e.EventType == models.AccessCheckEventStuck && !e.OccurredAt.Before(*c.FailingSince) {
			return nil
		}
	}
	return s.writeCheckEvent(ctx, tx, c.ID, models.AccessCheckEventStuck, now)
}

// AccessStatusForUser returns the user's checks on the allocation. Checks a
// probe has not written yet are synthesized so the caller always sees both:
// sign-in derives from the cluster account's provisioned timestamp, job
// submission defaults to still setting up.
func (s *Service) AccessStatusForUser(ctx context.Context, allocationID, userID string) (*AccessStatus, error) {
	alloc, err := s.allocs.FindByID(ctx, allocationID)
	if err != nil {
		return nil, err
	}
	if alloc == nil {
		return nil, ErrNotFound
	}
	checks, err := s.accessChecks.FindByUserAndAllocation(ctx, allocationID, userID)
	if err != nil {
		return nil, err
	}
	cu := s.clusterUserFor(ctx, alloc, userID)
	checks, realIDs := fillDefaultChecks(allocationID, userID, cu, checks)
	events, err := s.accessChecks.FindEventsByChecks(ctx, realIDs)
	if err != nil {
		return nil, err
	}
	status := &AccessStatus{Checks: checks, Events: events}
	if cu != nil {
		status.LocalUsername = cu.LocalUsername
	}
	return status, nil
}

// AccessStatusForAllocation returns every ACTIVE member's checks, for the
// manager grid.
func (s *Service) AccessStatusForAllocation(ctx context.Context, allocationID string) ([]MemberAccessStatus, error) {
	alloc, err := s.allocs.FindByID(ctx, allocationID)
	if err != nil {
		return nil, err
	}
	if alloc == nil {
		return nil, ErrNotFound
	}
	members, err := s.memberships.FindByAllocationWithUser(ctx, allocationID)
	if err != nil {
		return nil, err
	}
	checks, err := s.accessChecks.FindByAllocation(ctx, allocationID)
	if err != nil {
		return nil, err
	}
	byUser := make(map[string][]models.AccessCheck)
	for _, c := range checks {
		byUser[c.UserID] = append(byUser[c.UserID], c)
	}
	out := make([]MemberAccessStatus, 0, len(members))
	for _, m := range members {
		if m.MembershipStatus != models.ACTIVE {
			continue
		}
		cu := s.clusterUserFor(ctx, alloc, m.UserID)
		mc, _ := fillDefaultChecks(allocationID, m.UserID, cu, byUser[m.UserID])
		row := MemberAccessStatus{
			UserID:      m.UserID,
			DisplayName: m.DisplayName,
			Email:       m.Email,
			Checks:      mc,
		}
		if cu != nil {
			row.LocalUsername = cu.LocalUsername
		}
		out = append(out, row)
	}
	return out, nil
}

// clusterUserFor returns the user's account on the allocation's cluster, or
// nil when none is assigned yet (or the lookup fails, which reads the same).
func (s *Service) clusterUserFor(ctx context.Context, alloc *models.ComputeAllocation, userID string) *models.ComputeClusterUser {
	cu, err := s.clusterUsers.FindByPair(ctx, alloc.ComputeClusterID, userID)
	if err != nil {
		return nil
	}
	return cu
}

// fillDefaultChecks appends synthetic checks (empty ID, no history) for the
// types no probe has reported yet, and returns the real check ids.
func fillDefaultChecks(allocationID, userID string, cu *models.ComputeClusterUser, checks []models.AccessCheck) ([]models.AccessCheck, []string) {
	have := make(map[models.AccessCheckType]bool, len(checks))
	realIDs := make([]string, 0, len(checks))
	for _, c := range checks {
		have[c.CheckType] = true
		realIDs = append(realIDs, c.ID)
	}
	now := nowUTC()
	if !have[models.AccessCheckSignIn] {
		c := models.AccessCheck{
			ComputeAllocationID: allocationID, UserID: userID,
			CheckType: models.AccessCheckSignIn, Status: models.AccessCheckPending,
			LastCheckedAt: now,
		}
		if cu != nil && cu.ProvisionedAt != nil {
			c.Status = models.AccessCheckOK
			c.LastCheckedAt = *cu.ProvisionedAt
			c.LastOKAt = cu.ProvisionedAt
		}
		checks = append(checks, c)
	}
	if !have[models.AccessCheckJobSubmission] {
		checks = append(checks, models.AccessCheck{
			ComputeAllocationID: allocationID, UserID: userID,
			CheckType: models.AccessCheckJobSubmission, Status: models.AccessCheckPending,
			LastCheckedAt: now,
		})
	}
	return checks, realIDs
}
