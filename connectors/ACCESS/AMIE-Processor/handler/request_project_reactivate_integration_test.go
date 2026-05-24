//go:build integration

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

package handler

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/pkg/models"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

// inactivateSeeded flips the seeded project, allocation, and memberships to
// INACTIVE so a request_project_reactivate test starts from the
// "project previously inactivated" precondition stated in the protocol line 451.
func inactivateSeeded(t *testing.T, database *sqlx.DB, svc *coreservice.Service, seed seededProject) {
	t.Helper()
	ctx := context.Background()
	if _, err := svc.UpdateProjectStatus(ctx, seed.ProjectID, models.ProjectInactive); err != nil {
		t.Fatalf("inactivate seeded project: %v", err)
	}
	alloc, err := svc.GetComputeAllocation(ctx, seed.AllocationID)
	if err != nil {
		t.Fatalf("read seeded allocation: %v", err)
	}
	alloc.Status = models.INACTIVE
	if err := svc.UpdateComputeAllocation(ctx, alloc); err != nil {
		t.Fatalf("inactivate seeded allocation: %v", err)
	}
	if _, err := svc.UpdateMembershipStatus(ctx, seed.PIMembership, models.INACTIVE); err != nil {
		t.Fatalf("inactivate PI membership: %v", err)
	}
	if _, err := svc.UpdateMembershipStatus(ctx, seed.OtherMember, models.INACTIVE); err != nil {
		t.Fatalf("inactivate non-PI membership: %v", err)
	}
}

// baseRPReactivateBody returns a fully populated request_project_reactivate
// body. ProjectID and ResourceList are required; PersonID is allowed and
// carries the PI.
func baseRPReactivateBody(projectID, piGlobalID string) map[string]any {
	return map[string]any{
		"ProjectID":    projectID,
		"ResourceList": []any{"compute.access-ci.org"},
		"PersonID":     piGlobalID,
		"Comment":      "Reactivation requested",
	}
}

// TestRequestProjectReactivate_HappyPath asserts the protocol contract:
//
//   - the Project flips back to ACTIVE.
//   - every ComputeAllocation under the project flips to ACTIVE.
//   - only the PI's membership reactivates; other members stay INACTIVE
//     .
//   - the handler replies notify_project_reactivate and emits the
//     REACTIVATE_PROJECT / REACTIVATE_MEMBERSHIP / REPLY_SENT audit rows.
func TestRequestProjectReactivate_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestCoreService(database)
	seed := seedActiveProject(t, database, svc)
	inactivateSeeded(t, database, svc, seed)

	body := baseRPReactivateBody(seed.ProjectID, "seed-pi-global")
	pkt := insertPacket(t, database, "request_project_reactivate", body)

	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestProjectReactivateHandler(svc, amie, audit)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// Project must be ACTIVE again.
	proj, err := svc.GetProject(context.Background(), seed.ProjectID)
	if err != nil {
		t.Fatalf("read project: %v", err)
	}
	if proj.Status != models.ProjectActive {
		t.Errorf("project.status: got %q, want %q", proj.Status, models.ProjectActive)
	}

	// Allocation must be ACTIVE.
	alloc, err := svc.GetComputeAllocation(context.Background(), seed.AllocationID)
	if err != nil {
		t.Fatalf("read allocation: %v", err)
	}
	if alloc.Status != models.ACTIVE {
		t.Errorf("allocation.status: got %q, want %q", alloc.Status, models.ACTIVE)
	}

	// the protocol asymmetry: PI membership ACTIVE, non-PI membership still INACTIVE.
	if got := membershipStatus(t, database, seed.PIMembership); got != models.ACTIVE {
		t.Errorf("PI membership status: got %q, want %q", got, models.ACTIVE)
	}
	if got := membershipStatus(t, database, seed.OtherMember); got != models.INACTIVE {
		t.Errorf("non-PI membership status: got %q, want %q", got, models.INACTIVE)
	}

	// One status-change diff for the allocation.
	if got := countRows(t, database, "compute_allocation_diffs"); got != 1 {
		t.Errorf("compute_allocation_diffs: got %d, want 1", got)
	}
	var diff struct {
		DiffType string `db:"diff_type"`
		Status   string `db:"status"`
	}
	if err := database.Get(&diff,
		"SELECT diff_type, status FROM compute_allocation_diffs LIMIT 1",
	); err != nil {
		t.Fatalf("read diff: %v", err)
	}
	if diff.DiffType != "ALLOCATION_STATUS_CHANGE" {
		t.Errorf("diff.diff_type: got %q, want ALLOCATION_STATUS_CHANGE", diff.DiffType)
	}
	if diff.Status != string(models.ACTIVE) {
		t.Errorf("diff.status: got %q, want %q", diff.Status, models.ACTIVE)
	}

	// Audit trail: REACTIVATE_PROJECT, REACTIVATE_MEMBERSHIP (1, PI only),
	// REPLY_SENT. No INACTIVATE_MEMBERSHIP rows from the non-PI member.
	if got := countAuditActions(t, database, pkt.ID, model.AuditReactivateProject); got != 1 {
		t.Errorf("audit REACTIVATE_PROJECT: got %d, want 1", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReactivateMembership); got != 1 {
		t.Errorf("audit REACTIVATE_MEMBERSHIP: got %d, want 1 (PI only, the protocol)", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT: got %d, want 1", got)
	}

	// Reply: type and ProjectID.
	if got, want := amie.lastReplyType(), "notify_project_reactivate"; got != want {
		t.Fatalf("reply type: got %q, want %q", got, want)
	}
	reply := amie.lastReplyBody()
	if reply == nil {
		t.Fatal("reply body missing")
	}
	if got, _ := reply["ProjectID"].(string); got != seed.ProjectID {
		t.Errorf("reply.ProjectID: got %q, want %q", got, seed.ProjectID)
	}
}

// TestRequestProjectReactivate_UnknownProject asserts the handler's current
// soft-skip behavior when the ProjectID does not resolve to a local row.
//
// Spec context: the protocol is silent on the unknown-project case. The connector's
// "soft skip" behavior is a known design concern (may mask real bugs in
// production).
func TestRequestProjectReactivate_UnknownProject(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestCoreService(database)

	body := baseRPReactivateBody("does-not-exist-uuid", "seed-pi-global")
	pkt := insertPacket(t, database, "request_project_reactivate", body)

	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestProjectReactivateHandler(svc, amie, audit)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error for unknown project: %v (B7 documents this soft skip)", err)
	}

	// No state mutations: nothing to mutate.
	if got := countRows(t, database, "projects"); got != 0 {
		t.Errorf("projects: got %d, want 0", got)
	}
	if got := countRows(t, database, "compute_allocations"); got != 0 {
		t.Errorf("compute_allocations: got %d, want 0", got)
	}
	if got := countRows(t, database, "compute_allocation_diffs"); got != 0 {
		t.Errorf("compute_allocation_diffs: got %d, want 0", got)
	}

	if got := countAuditActions(t, database, pkt.ID, model.AuditReactivateProject); got != 0 {
		t.Errorf("audit REACTIVATE_PROJECT on unknown project: got %d, want 0", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT on unknown project: got %d, want 1", got)
	}

	if got, want := amie.lastReplyType(), "notify_project_reactivate"; got != want {
		t.Fatalf("reply type: got %q, want %q", got, want)
	}
}

// TestRequestProjectReactivate_MissingProjectID asserts the handler rejects
// a packet missing the protocol-required ProjectID field, without mutating the DB.
func TestRequestProjectReactivate_MissingProjectID(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestCoreService(database)
	seed := seedActiveProject(t, database, svc)
	inactivateSeeded(t, database, svc, seed)

	body := baseRPReactivateBody(seed.ProjectID, "seed-pi-global")
	delete(body, "ProjectID")
	pkt := insertPacket(t, database, "request_project_reactivate", body)

	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestProjectReactivateHandler(svc, amie, audit)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error for missing ProjectID, got nil")
	}
	if !strings.Contains(err.Error(), "ProjectID") {
		t.Errorf("error should mention ProjectID, got: %v", err)
	}

	// Seeded project must remain INACTIVE, no mutations on validation failure.
	proj, err := svc.GetProject(context.Background(), seed.ProjectID)
	if err != nil {
		t.Fatalf("read seeded project: %v", err)
	}
	if proj.Status != models.ProjectInactive {
		t.Errorf("project.status after validation failure: got %q, want %q (INACTIVE)", proj.Status, models.ProjectInactive)
	}
	alloc, err := svc.GetComputeAllocation(context.Background(), seed.AllocationID)
	if err != nil {
		t.Fatalf("read seeded allocation: %v", err)
	}
	if alloc.Status != models.INACTIVE {
		t.Errorf("allocation.status after validation failure: got %q, want %q", alloc.Status, models.INACTIVE)
	}
	if got := membershipStatus(t, database, seed.PIMembership); got != models.INACTIVE {
		t.Errorf("PI membership after validation failure: got %q, want %q", got, models.INACTIVE)
	}

	if len(amie.Replies) != 0 {
		t.Errorf("amie replies on validation failure: got %d, want 0", len(amie.Replies))
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReactivateProject); got != 0 {
		t.Errorf("audit REACTIVATE_PROJECT on validation failure: got %d, want 0", got)
	}
}

// TestRequestProjectReactivate_ReplyFailurePropagates asserts the handler
// surfaces a ReplyToPacket failure to the caller. REPLY_SENT must NOT exist
// when the reply failed. (Core writes commit in their own txns, so this
// test does not check rollback shape.)
func TestRequestProjectReactivate_ReplyFailurePropagates(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestCoreService(database)
	seed := seedActiveProject(t, database, svc)
	inactivateSeeded(t, database, svc, seed)

	body := baseRPReactivateBody(seed.ProjectID, "seed-pi-global")
	pkt := insertPacket(t, database, "request_project_reactivate", body)

	audit := newTestAuditService(database)
	amie := &fakeAmieClient{FailWith: errors.New("simulated AMIE outage")}
	h := NewRequestProjectReactivateHandler(svc, amie, audit)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error when reply fails, got nil")
	}
	if !strings.Contains(err.Error(), "reply") && !strings.Contains(err.Error(), "AMIE") {
		t.Errorf("error should reference reply path, got: %v", err)
	}

	// REPLY_SENT audit row must NOT exist when the reply failed.
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 0 {
		t.Errorf("audit REPLY_SENT on failed reply: got %d, want 0", got)
	}
}
