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
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/pkg/models"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

// seededProject groups the rows seeded by seedActiveProject so tests can read
// back the expected IDs after the handler runs.
type seededProject struct {
	OrgID        string
	PIUserID     string
	OtherUserID  string
	ProjectID    string
	AllocationID string
	PIMembership string
	OtherMember  string
}

// seedActiveProject builds the pre-existing state every request_project_*
// activate/inactivate test starts from: one organization, a PI user, a
// non-PI user, an ACTIVE project, one ACTIVE allocation under it, and an
// ACTIVE membership for each user on that allocation.
//
// Tests exercising the protocol's PI-only-reactivate asymmetry need both members to
// observe that the non-PI membership is NOT touched on reactivate. The
// inactivate test reuses the same shape because the protocol says ALL accounts on the
// project must be inactivated.
func seedActiveProject(t *testing.T, database *sqlx.DB, svc *coreservice.Service) seededProject {
	t.Helper()
	ctx := context.Background()

	org, err := svc.CreateOrganization(ctx, &models.Organization{
		OriginatedID: "SEED-ORG",
		Name:         "Seed Org",
	})
	if err != nil {
		t.Fatalf("seed organization: %v", err)
	}

	pi, err := svc.CreateUser(ctx, &models.User{
		OrganizationID: org.ID,
		FirstName:      "Seed",
		LastName:       "PI",
		Email:          "seed.pi@example.edu",
	})
	if err != nil {
		t.Fatalf("seed PI user: %v", err)
	}
	if _, err := svc.CreateUserIdentity(ctx, &models.UserIdentity{
		UserID:     pi.ID,
		Source:     amieIdentitySource,
		ExternalID: "seed-pi-global",
		Email:      pi.Email,
	}); err != nil {
		t.Fatalf("seed PI identity: %v", err)
	}

	other, err := svc.CreateUser(ctx, &models.User{
		OrganizationID: org.ID,
		FirstName:      "Seed",
		LastName:       "Member",
		Email:          "seed.member@example.edu",
	})
	if err != nil {
		t.Fatalf("seed non-PI user: %v", err)
	}
	if _, err := svc.CreateUserIdentity(ctx, &models.UserIdentity{
		UserID:     other.ID,
		Source:     amieIdentitySource,
		ExternalID: "seed-member-global",
		Email:      other.Email,
	}); err != nil {
		t.Fatalf("seed non-PI identity: %v", err)
	}

	project, err := svc.CreateProject(ctx, &models.Project{
		OriginatedID: "TG-SEED-001",
		Title:        "TG-SEED-001",
		Origination:  amieIdentitySource,
		ProjectPIID:  pi.ID,
	})
	if err != nil {
		t.Fatalf("seed project: %v", err)
	}

	alloc, err := svc.CreateComputeAllocation(ctx, &models.ComputeAllocation{
		ProjectID:        project.ID,
		Name:             "TG-SEED-001",
		ComputeClusterID: testClusterID,
		InitialSUAmount:  1000,
		StartTime:        time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndTime:          time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("seed allocation: %v", err)
	}

	piMem, err := svc.CreateComputeAllocationMembership(ctx, &models.ComputeAllocationMembership{
		ComputeAllocationID: alloc.ID,
		UserID:              pi.ID,
		StartTime:           alloc.StartTime,
		EndTime:             alloc.EndTime,
	})
	if err != nil {
		t.Fatalf("seed PI membership: %v", err)
	}
	otherMem, err := svc.CreateComputeAllocationMembership(ctx, &models.ComputeAllocationMembership{
		ComputeAllocationID: alloc.ID,
		UserID:              other.ID,
		StartTime:           alloc.StartTime,
		EndTime:             alloc.EndTime,
	})
	if err != nil {
		t.Fatalf("seed non-PI membership: %v", err)
	}

	return seededProject{
		OrgID:        org.ID,
		PIUserID:     pi.ID,
		OtherUserID:  other.ID,
		ProjectID:    project.ID,
		AllocationID: alloc.ID,
		PIMembership: piMem.ID,
		OtherMember:  otherMem.ID,
	}
}

// baseRPInactivateBody returns a fully populated request_project_inactivate
// body with every required field. ProjectID is filled in by each test from
// the seeded project.
func baseRPInactivateBody(projectID string) map[string]any {
	return map[string]any{
		"ProjectID":    projectID,
		"ResourceList": []any{"compute.access-ci.org"},
		"Comment":      "End of allocation period",
	}
}

func membershipStatus(t *testing.T, database *sqlx.DB, id string) models.AllocationStatus {
	t.Helper()
	var status string
	if err := database.Get(&status,
		"SELECT membership_status FROM compute_allocation_memberships WHERE id = ?", id,
	); err != nil {
		t.Fatalf("read membership %s: %v", id, err)
	}
	return models.AllocationStatus(status)
}

// TestRequestProjectInactivate_HappyPath asserts the protocol contract:
//
//   - the Project flips to INACTIVE.
//   - every ComputeAllocation under the project flips to INACTIVE (the
//     connector models the protocol's "suspended or expired" as INACTIVE).
//   - every membership under the project flips to INACTIVE.
//   - the handler replies notify_project_inactivate and emits the
//     INACTIVATE_PROJECT / INACTIVATE_MEMBERSHIP / REPLY_SENT audit rows.
func TestRequestProjectInactivate_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestCoreService(database)
	seed := seedActiveProject(t, database, svc)

	body := baseRPInactivateBody(seed.ProjectID)
	pkt := insertPacket(t, database, "request_project_inactivate", body)

	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestProjectInactivateHandler(svc, amie, audit)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// Project must be INACTIVE.
	proj, err := svc.GetProject(context.Background(), seed.ProjectID)
	if err != nil {
		t.Fatalf("read project: %v", err)
	}
	if proj.Status != models.ProjectInactive {
		t.Errorf("project.status: got %q, want %q", proj.Status, models.ProjectInactive)
	}

	// Allocation must be INACTIVE.
	alloc, err := svc.GetComputeAllocation(context.Background(), seed.AllocationID)
	if err != nil {
		t.Fatalf("read allocation: %v", err)
	}
	if alloc.Status != models.INACTIVE {
		t.Errorf("allocation.status: got %q, want %q", alloc.Status, models.INACTIVE)
	}

	// Both memberships (PI and non-PI) must be INACTIVE, the protocol says ALL accounts
	// on the project are inactivated.
	if got := membershipStatus(t, database, seed.PIMembership); got != models.INACTIVE {
		t.Errorf("PI membership status: got %q, want %q", got, models.INACTIVE)
	}
	if got := membershipStatus(t, database, seed.OtherMember); got != models.INACTIVE {
		t.Errorf("non-PI membership status: got %q, want %q", got, models.INACTIVE)
	}

	// A status-change diff row is recorded for the allocation.
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
	if diff.Status != string(models.INACTIVE) {
		t.Errorf("diff.status: got %q, want %q", diff.Status, models.INACTIVE)
	}

	// Audit trail: INACTIVATE_PROJECT, INACTIVATE_MEMBERSHIP (x2), REPLY_SENT.
	if got := countAuditActions(t, database, pkt.ID, model.AuditInactivateProject); got != 1 {
		t.Errorf("audit INACTIVATE_PROJECT: got %d, want 1", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditInactivateMembership); got != 2 {
		t.Errorf("audit INACTIVATE_MEMBERSHIP: got %d, want 2 (PI + non-PI)", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT: got %d, want 1", got)
	}

	// Reply: type and ProjectID.
	if got, want := amie.lastReplyType(), "notify_project_inactivate"; got != want {
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

// TestRequestProjectInactivate_UnknownProject asserts the handler's current
// soft-skip behavior: when the ProjectID does not resolve to a local row,
// the handler logs a warning, still sends the reply, and returns success.
//
// Spec context: the protocol is silent on the unknown-project case. The connector's
// "soft skip" behavior is a known design concern (it may mask real bugs in
// production). Mirroring the existing handler shape until the contract is
// clarified upstream.
func TestRequestProjectInactivate_UnknownProject(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestCoreService(database)

	body := baseRPInactivateBody("does-not-exist-uuid")
	pkt := insertPacket(t, database, "request_project_inactivate", body)

	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestProjectInactivateHandler(svc, amie, audit)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error for unknown project: %v (B7 in bugs tracking documents this soft skip)", err)
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

	// No INACTIVATE_PROJECT audit, only REPLY_SENT.
	if got := countAuditActions(t, database, pkt.ID, model.AuditInactivateProject); got != 0 {
		t.Errorf("audit INACTIVATE_PROJECT on unknown project: got %d, want 0", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT on unknown project: got %d, want 1", got)
	}

	// Reply still sent.
	if got, want := amie.lastReplyType(), "notify_project_inactivate"; got != want {
		t.Fatalf("reply type: got %q, want %q", got, want)
	}
}

// TestRequestProjectInactivate_MissingProjectID asserts the handler rejects
// a packet missing the protocol-required ProjectID field, without mutating the DB.
func TestRequestProjectInactivate_MissingProjectID(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestCoreService(database)
	seed := seedActiveProject(t, database, svc)

	body := baseRPInactivateBody(seed.ProjectID)
	delete(body, "ProjectID")
	pkt := insertPacket(t, database, "request_project_inactivate", body)

	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestProjectInactivateHandler(svc, amie, audit)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error for missing ProjectID, got nil")
	}
	if !strings.Contains(err.Error(), "ProjectID") {
		t.Errorf("error should mention ProjectID, got: %v", err)
	}

	// Seeded project must remain ACTIVE, no mutations on validation failure.
	proj, err := svc.GetProject(context.Background(), seed.ProjectID)
	if err != nil {
		t.Fatalf("read seeded project: %v", err)
	}
	if proj.Status != models.ProjectActive {
		t.Errorf("project.status after validation failure: got %q, want %q (ACTIVE)", proj.Status, models.ProjectActive)
	}
	alloc, err := svc.GetComputeAllocation(context.Background(), seed.AllocationID)
	if err != nil {
		t.Fatalf("read seeded allocation: %v", err)
	}
	if alloc.Status != models.ACTIVE {
		t.Errorf("allocation.status after validation failure: got %q, want %q", alloc.Status, models.ACTIVE)
	}
	if got := membershipStatus(t, database, seed.PIMembership); got != models.ACTIVE {
		t.Errorf("PI membership after validation failure: got %q, want %q", got, models.ACTIVE)
	}

	if len(amie.Replies) != 0 {
		t.Errorf("amie replies on validation failure: got %d, want 0", len(amie.Replies))
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditInactivateProject); got != 0 {
		t.Errorf("audit INACTIVATE_PROJECT on validation failure: got %d, want 0", got)
	}
}

// TestRequestProjectInactivate_ReplyFailurePropagates asserts the handler
// surfaces a ReplyToPacket failure to the caller. REPLY_SENT must NOT exist
// when the reply failed. The protocol does not mandate atomicity on reply
// failure, so this test does not check rollback shape.
func TestRequestProjectInactivate_ReplyFailurePropagates(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestCoreService(database)
	seed := seedActiveProject(t, database, svc)

	body := baseRPInactivateBody(seed.ProjectID)
	pkt := insertPacket(t, database, "request_project_inactivate", body)

	audit := newTestAuditService(database)
	amie := &fakeAmieClient{FailWith: errors.New("simulated AMIE outage")}
	h := NewRequestProjectInactivateHandler(svc, amie, audit)

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
