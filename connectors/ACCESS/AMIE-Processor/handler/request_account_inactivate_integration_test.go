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

// seededMembership bundles every row seedMembership inserts so individual
// tests can pin assertions to the right primary keys.
type seededMembership struct {
	Org         *models.Organization
	User        *models.User
	Project     *models.Project
	Allocation  *models.ComputeAllocation
	Membership  *models.ComputeAllocationMembership
	ClusterUser *models.ComputeClusterUser
}

// seedMembership creates an ACTIVE user + project + allocation + ACTIVE
// membership ready for the inactivate/reactivate handlers to flip. The user
// also gets a compute_cluster_users row so tests can assert the handlers do
// NOT touch cluster-side identity (the protocol requires that "the PersonID
// and the related login on the resource MUST be retained").
func seedMembership(t *testing.T, database *sqlx.DB) *seededMembership {
	t.Helper()
	svc := newTestCoreService(database)
	ctx := context.Background()

	org, err := svc.CreateOrganization(ctx, &models.Organization{
		OriginatedID: "SEED-ORG",
		Name:         "Seed Org",
	})
	if err != nil {
		t.Fatalf("seed organization: %v", err)
	}
	user, err := svc.CreateUser(ctx, &models.User{
		OrganizationID: org.ID,
		FirstName:      "Seed",
		LastName:       "User",
		Email:          "seed.user@example.edu",
		Status:         models.UserActive,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	project, err := svc.CreateProject(ctx, &models.Project{
		OriginatedID: "TG-SEED-001",
		Title:        "Seed Project",
		Origination:  "ACCESS",
		ProjectPIID:  user.ID,
		Status:       models.ProjectActive,
	})
	if err != nil {
		t.Fatalf("seed project: %v", err)
	}
	now := time.Now().UTC()
	alloc, err := svc.CreateComputeAllocation(ctx, &models.ComputeAllocation{
		ProjectID:        project.ID,
		Name:             "Seed Allocation",
		Status:           models.ACTIVE,
		ComputeClusterID: testClusterID,
		InitialSUAmount:  1000,
		StartTime:        now,
		EndTime:          now.Add(365 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("seed allocation: %v", err)
	}
	membership, err := svc.CreateComputeAllocationMembership(ctx, &models.ComputeAllocationMembership{
		ComputeAllocationID: alloc.ID,
		UserID:              user.ID,
		StartTime:           now,
		EndTime:             now.Add(365 * 24 * time.Hour),
		MembershipStatus:    models.ACTIVE,
	})
	if err != nil {
		t.Fatalf("seed membership: %v", err)
	}
	clusterUser, err := svc.CreateComputeClusterUser(ctx, &models.ComputeClusterUser{
		ComputeClusterID: testClusterID,
		UserID:           user.ID,
		LocalUsername:    "seedlogin",
	})
	if err != nil {
		t.Fatalf("seed cluster user: %v", err)
	}
	return &seededMembership{
		Org:         org,
		User:        user,
		Project:     project,
		Allocation:  alloc,
		Membership:  membership,
		ClusterUser: clusterUser,
	}
}

// baseRAIBody returns a fully populated request_account_inactivate body. Per
// the protocol every required field is present. PersonID and ProjectID are the
// site-local IDs returned by the original notify_account_create /
// notify_project_create transaction.
func baseRAIBody(personID, projectID string) map[string]any {
	return map[string]any{
		"PersonID":     personID,
		"ProjectID":    projectID,
		"ResourceList": []any{"compute.access-ci.org"},
		"Comment":      "test inactivation",
	}
}

func getMembershipStatus(t *testing.T, database *sqlx.DB, membershipID string) string {
	t.Helper()
	var status string
	if err := database.Get(&status,
		"SELECT membership_status FROM compute_allocation_memberships WHERE id = ?", membershipID,
	); err != nil {
		t.Fatalf("read membership %s: %v", membershipID, err)
	}
	return status
}

func getUserStatus(t *testing.T, database *sqlx.DB, userID string) string {
	t.Helper()
	var status string
	if err := database.Get(&status, "SELECT status FROM users WHERE id = ?", userID); err != nil {
		t.Fatalf("read user %s: %v", userID, err)
	}
	return status
}

// TestRequestAccountInactivate_HappyPath asserts the protocol mandate: the handler MUST
// flip the membership for (PersonID, ProjectID) to INACTIVE. The mutation
// is scoped to `compute_allocation_memberships` only. ACCESS doc makes the
// negative assertion explicit: "the PersonID (as well as the related login
// on the resource) must be retained for this user", so `users.status` and
// `compute_cluster_users` must NOT change.
func TestRequestAccountInactivate_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	seed := seedMembership(t, database)

	body := baseRAIBody(seed.User.ID, seed.Project.ID)
	pkt := insertPacket(t, database, "request_account_inactivate", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestAccountInactivateHandler(svc, amie, audit)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// the protocol + domain-writes: the membership for this user on this project flips
	// to INACTIVE.
	if got := getMembershipStatus(t, database, seed.Membership.ID); got != string(models.INACTIVE) {
		t.Errorf("membership.status: got %q, want %q", got, models.INACTIVE)
	}

	// Account inactivation does NOT imply user inactivation; the PersonID +
	// the resource-side login must be retained.
	if got := getUserStatus(t, database, seed.User.ID); got != string(models.UserActive) {
		t.Errorf("user.status: got %q, want %q (must NOT flip)", got, models.UserActive)
	}
	if got := countRows(t, database, "compute_cluster_users"); got != 1 {
		t.Errorf("compute_cluster_users: got %d, want 1 (resource login must be retained)", got)
	}

	// Audit trail: one INACTIVATE_MEMBERSHIP per flipped row, plus the reply.
	if got := countAuditActions(t, database, pkt.ID, model.AuditInactivateMembership); got != 1 {
		t.Errorf("audit INACTIVATE_MEMBERSHIP: got %d, want 1", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT: got %d, want 1", got)
	}

	// Reply contract: notify_account_inactivate.
	if got, want := amie.lastReplyType(), "notify_account_inactivate"; got != want {
		t.Fatalf("reply type: got %q, want %q", got, want)
	}
	reply := amie.lastReplyBody()
	if reply == nil {
		t.Fatal("reply body missing")
	}
	// The NAI reply allows PersonID, ProjectID, ResourceList to be inferred
	// from the request rather than echoed. The connector currently echoes
	// PersonID + ProjectID; both should match what we sent.
	if got, _ := reply["PersonID"].(string); got != seed.User.ID {
		t.Errorf("reply.PersonID: got %q, want %q", got, seed.User.ID)
	}
	if got, _ := reply["ProjectID"].(string); got != seed.Project.ID {
		t.Errorf("reply.ProjectID: got %q, want %q", got, seed.Project.ID)
	}
}

// TestRequestAccountInactivate_UnknownProjectOrUser asserts the soft-skip
// behavior. The protocol only mandates the inactivation action and the reply; it does
// not mandate that the handler to error when the (PersonID, ProjectID) pair is
// unknown. The handler is expected to send the reply and make no mutations.
// This matches `flipUserMemberships` returning an empty slice for unknown
// project.
func TestRequestAccountInactivate_UnknownProjectOrUser(t *testing.T) {
	database := setupTestDB(t)
	// Seed a membership so we can prove the handler doesn't touch unrelated
	// rows when the packet refers to something else.
	seed := seedMembership(t, database)

	body := baseRAIBody("nope-person", "nope-project")
	pkt := insertPacket(t, database, "request_account_inactivate", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestAccountInactivateHandler(svc, amie, audit)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle on unknown project/user must soft-skip, got error: %v", err)
	}

	// Seeded membership must remain ACTIVE, handler must not touch it.
	if got := getMembershipStatus(t, database, seed.Membership.ID); got != string(models.ACTIVE) {
		t.Errorf("seeded membership.status: got %q, want %q (unrelated packet must not flip it)", got, models.ACTIVE)
	}
	// No INACTIVATE_MEMBERSHIP audit (nothing was flipped).
	if got := countAuditActions(t, database, pkt.ID, model.AuditInactivateMembership); got != 0 {
		t.Errorf("audit INACTIVATE_MEMBERSHIP: got %d, want 0", got)
	}
	// Reply still sent (the handler completes the transaction).
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT: got %d, want 1", got)
	}
	if got, want := amie.lastReplyType(), "notify_account_inactivate"; got != want {
		t.Errorf("reply type: got %q, want %q", got, want)
	}
}

// TestRequestAccountInactivate_MissingRequiredField asserts the handler
// rejects a packet missing a required field, with no mutations. PersonID is
// one of the three required body fields.
//
// NOTE: the protocol also requires ResourceList, but the current handler does not
// validate it. We exercise the PersonID path here since the handler enforces it.
func TestRequestAccountInactivate_MissingRequiredField(t *testing.T) {
	database := setupTestDB(t)
	seed := seedMembership(t, database)

	body := baseRAIBody(seed.User.ID, seed.Project.ID)
	delete(body, "PersonID")
	pkt := insertPacket(t, database, "request_account_inactivate", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestAccountInactivateHandler(svc, amie, audit)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error for missing PersonID, got nil")
	}
	if !strings.Contains(err.Error(), "PersonID") {
		t.Errorf("error should mention PersonID, got: %v", err)
	}
	// No mutations: seeded membership must still be ACTIVE.
	if got := getMembershipStatus(t, database, seed.Membership.ID); got != string(models.ACTIVE) {
		t.Errorf("seeded membership.status after validation failure: got %q, want %q", got, models.ACTIVE)
	}
	if len(amie.Replies) != 0 {
		t.Errorf("amie replies on validation failure: got %d, want 0", len(amie.Replies))
	}
}

// TestRequestAccountInactivate_ReplyFailurePropagates asserts the handler
// surfaces a ReplyToPacket failure to the caller so the processor can decide
// to retry. The REPLY_SENT audit row must be absent when the reply failed.
// (Note the broader atomicity gap: the membership flip is already committed
// because UpdateMembershipStatus opens its own tx; this test only asserts
// the error propagation + audit shape.)
func TestRequestAccountInactivate_ReplyFailurePropagates(t *testing.T) {
	database := setupTestDB(t)
	seed := seedMembership(t, database)

	body := baseRAIBody(seed.User.ID, seed.Project.ID)
	pkt := insertPacket(t, database, "request_account_inactivate", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{FailWith: errors.New("simulated AMIE outage")}
	h := NewRequestAccountInactivateHandler(svc, amie, audit)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error when reply fails, got nil")
	}
	if !strings.Contains(err.Error(), "reply") && !strings.Contains(err.Error(), "AMIE") {
		t.Errorf("error should reference reply path, got: %v", err)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 0 {
		t.Errorf("audit REPLY_SENT on failed reply: got %d, want 0", got)
	}
}

// silence unused-import false positives when only some tests reference these.
var _ = coreservice.ErrNotFound
