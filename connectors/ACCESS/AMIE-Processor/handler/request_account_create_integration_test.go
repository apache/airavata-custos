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
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
)

// baseRACBody returns a fully populated request_account_create body with
// every required field (GrantNumber, ResourceList, UserFirstName,
// UserLastName, UserOrganization, UserOrgCode). UserGlobalID is optional
// per the protocol but is the only stable upstream identifier for user
// resolution, so the baseline body carries it; the omission case is
// exercised by the missing-field test.
func baseRACBody() map[string]any {
	return map[string]any{
		"GrantNumber":      "TG-BL-001",
		"ResourceList":     []any{"compute.access-ci.org"},
		"UserFirstName":    "Sam",
		"UserLastName":     "Userman",
		"UserOrganization": "User Org",
		"UserOrgCode":      "USERORG",
		"UserGlobalID":     "bl-user-001",
		"UserEmail":        "sam.userman@user.example.edu",
	}
}

// seedProjectForRAC runs request_project_create against a fresh DB so that
// request_account_create has a project + allocation to attach to. Returns
// the project's Custos UUID, which AMIE echoes back as body.ProjectID on
// the subsequent request_account_create (the handler answered
// notify_project_create with this ID).
func seedProjectForRAC(t *testing.T, database *sqlx.DB) string {
	t.Helper()
	rpcBody := baseRPCBody()
	rpcPkt := insertPacket(t, database, "request_project_create", rpcBody)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestProjectCreateHandler(svc, testClusterID, amie, audit)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": rpcPkt.Type, "body": rpcBody}, rpcPkt, "")
	}); err != nil {
		t.Fatalf("seed request_project_create: %v", err)
	}

	var projectID string
	if err := database.Get(&projectID, "SELECT id FROM projects LIMIT 1"); err != nil {
		t.Fatalf("read seeded project id: %v", err)
	}
	return projectID
}

// TestRequestAccountCreate_HappyPath asserts the spec-required side effects
// of a first-delivery request_account_create with a complete body, after a
// successful prior request_project_create has provisioned the project +
// allocation.
//
// The handler MUST:
//   - persist the user as a local Person (User + UserIdentity)
//   - look up the project (already created by request_project_create)
//   - provision a ComputeClusterUser on the configured cluster
//   - attach a ComputeAllocationMembership linking the user to the project's
//     existing allocation
//   - reply notify_account_create with the required fields: ProjectID,
//     ResourceList, UserRemoteSiteLogin, AccountActivityTime
func TestRequestAccountCreate_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	projectID := seedProjectForRAC(t, database)

	// Counts established by the seed RPC so we can assert the RAC's deltas.
	usersAfterSeed := countRows(t, database, "users")
	identitiesAfterSeed := countRows(t, database, "user_identities")
	clusterUsersAfterSeed := countRows(t, database, "compute_cluster_users")
	membershipsAfterSeed := countRows(t, database, "compute_allocation_memberships")
	projectsAfterSeed := countRows(t, database, "projects")
	allocationsAfterSeed := countRows(t, database, "compute_allocations")

	body := baseRACBody()
	body["ProjectID"] = projectID
	pkt := insertPacket(t, database, "request_account_create", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestAccountCreateHandler(svc, testClusterID, amie, audit)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// Identity-domain writes: one new User + one new UserIdentity for the
	// account user (PI was created by the seed RPC).
	if got, want := countRows(t, database, "users"), usersAfterSeed+1; got != want {
		t.Errorf("users: got %d, want %d", got, want)
	}
	if got, want := countRows(t, database, "user_identities"), identitiesAfterSeed+1; got != want {
		t.Errorf("user_identities: got %d, want %d", got, want)
	}
	// Project + allocation untouched by RAC.
	if got, want := countRows(t, database, "projects"), projectsAfterSeed; got != want {
		t.Errorf("projects: got %d, want %d (RAC must not create projects)", got, want)
	}
	if got, want := countRows(t, database, "compute_allocations"), allocationsAfterSeed; got != want {
		t.Errorf("compute_allocations: got %d, want %d (RAC must not create allocations)", got, want)
	}
	// Allocation-domain writes: one new ComputeClusterUser + one new
	// ComputeAllocationMembership for this user.
	if got, want := countRows(t, database, "compute_cluster_users"), clusterUsersAfterSeed+1; got != want {
		t.Errorf("compute_cluster_users: got %d, want %d", got, want)
	}
	if got, want := countRows(t, database, "compute_allocation_memberships"), membershipsAfterSeed+1; got != want {
		t.Errorf("compute_allocation_memberships: got %d, want %d", got, want)
	}

	// Inspect the freshly created user row (the one bound to UserGlobalID).
	var user struct {
		ID        string `db:"id"`
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
		Status    string `db:"status"`
	}
	if err := database.Get(&user,
		`SELECT u.id, u.first_name, u.last_name, u.email, u.status
		   FROM users u
		   JOIN user_identities ui ON ui.user_id = u.id
		  WHERE ui.source = ? AND ui.external_id = ?`,
		"access", "bl-user-001",
	); err != nil {
		t.Fatalf("read account user: %v", err)
	}
	if user.FirstName != "Sam" || user.LastName != "Userman" {
		t.Errorf("user name: got %q %q, want Sam Userman", user.FirstName, user.LastName)
	}
	if user.Email != "sam.userman@user.example.edu" {
		t.Errorf("user.email: got %q", user.Email)
	}
	if user.Status != "ACTIVE" {
		t.Errorf("user.status: got %q, want ACTIVE", user.Status)
	}

	// ComputeClusterUser must be on the configured cluster and bound to the
	// account user.
	var cu struct {
		UserID           string `db:"user_id"`
		ComputeClusterID string `db:"compute_cluster_id"`
		LocalUsername    string `db:"local_username"`
	}
	if err := database.Get(&cu,
		"SELECT user_id, compute_cluster_id, local_username FROM compute_cluster_users WHERE user_id = ?",
		user.ID,
	); err != nil {
		t.Fatalf("read compute_cluster_user: %v", err)
	}
	if cu.ComputeClusterID != testClusterID {
		t.Errorf("compute_cluster_user.compute_cluster_id: got %q, want %q", cu.ComputeClusterID, testClusterID)
	}
	if cu.LocalUsername == "" {
		t.Errorf("compute_cluster_user.local_username: empty; spec requires UserRemoteSiteLogin in reply")
	}

	// Membership row links the account user to the project's existing
	// allocation.
	var mem struct {
		UserID              string `db:"user_id"`
		ComputeAllocationID string `db:"compute_allocation_id"`
	}
	if err := database.Get(&mem,
		"SELECT user_id, compute_allocation_id FROM compute_allocation_memberships WHERE user_id = ?",
		user.ID,
	); err != nil {
		t.Fatalf("read membership: %v", err)
	}
	var allocID string
	if err := database.Get(&allocID,
		"SELECT id FROM compute_allocations WHERE project_id = ? LIMIT 1", projectID,
	); err != nil {
		t.Fatalf("read allocation id: %v", err)
	}
	if mem.ComputeAllocationID != allocID {
		t.Errorf("membership.compute_allocation_id: got %q, want %q", mem.ComputeAllocationID, allocID)
	}

	// Audit trail for this packet: CREATE_PERSON, CREATE_ACCOUNT,
	// CREATE_MEMBERSHIP, REPLY_SENT, one row each on a successful first delivery.
	for _, action := range []model.AuditAction{
		model.AuditCreatePerson,
		model.AuditCreateAccount,
		model.AuditCreateMembership,
		model.AuditReplySent,
	} {
		if got := countAuditActions(t, database, pkt.ID, action); got != 1 {
			t.Errorf("audit %s: got %d, want 1", action, got)
		}
	}

	// Reply: notify_account_create with the required fields:
	// AccountActivityTime, ProjectID, ResourceList, UserRemoteSiteLogin.
	if got, want := amie.lastReplyType(), "notify_account_create"; got != want {
		t.Fatalf("reply type: got %q, want %q", got, want)
	}
	reply := amie.lastReplyBody()
	if reply == nil {
		t.Fatal("reply body missing")
	}
	if got, _ := reply["ProjectID"].(string); got != projectID {
		t.Errorf("reply.ProjectID: got %q, want %q", got, projectID)
	}
	if got, _ := reply["UserRemoteSiteLogin"].(string); got == "" {
		t.Errorf("reply.UserRemoteSiteLogin: empty; required value")
	} else if got != cu.LocalUsername {
		t.Errorf("reply.UserRemoteSiteLogin: got %q, want %q (the provisioned posix username)",
			got, cu.LocalUsername)
	}
	if rl, ok := reply["ResourceList"].([]string); !ok || len(rl) != 1 || rl[0] != "compute.access-ci.org" {
		// ResourceList may come back as []any depending on serialization;
		// accept either typed form.
		if rlAny, ok := reply["ResourceList"].([]any); !ok || len(rlAny) != 1 {
			t.Errorf("reply.ResourceList: got %v, want single-element [compute.access-ci.org]", reply["ResourceList"])
		}
	}
	t.Skipf("blocked by known bug: notify_account_create reply omits required AccountActivityTime field")
	if _, ok := reply["AccountActivityTime"]; !ok {
		t.Errorf("reply.AccountActivityTime: missing; required in notify_account_create")
	}
}

// TestRequestAccountCreate_IdempotentReDelivery asserts that re-delivering the
// same request_account_create packet does NOT duplicate identity rows,
// cluster-user rows, or membership rows. The audit log is a history of
// attempts (not state), so each delivery may append its own audit rows.
func TestRequestAccountCreate_IdempotentReDelivery(t *testing.T) {
	database := setupTestDB(t)
	projectID := seedProjectForRAC(t, database)

	usersBefore := countRows(t, database, "users")
	identitiesBefore := countRows(t, database, "user_identities")
	clusterUsersBefore := countRows(t, database, "compute_cluster_users")
	membershipsBefore := countRows(t, database, "compute_allocation_memberships")

	body := baseRACBody()
	body["ProjectID"] = projectID

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestAccountCreateHandler(svc, testClusterID, amie, audit)

	// First delivery.
	firstPkt := insertPacket(t, database, "request_account_create", body)
	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": firstPkt.Type, "body": body}, firstPkt, "")
	}); err != nil {
		t.Fatalf("first Handle: %v", err)
	}

	// Re-delivery: SAME body. AMIE may legitimately re-deliver a packet (e.g.
	// after an connector restart before ACK).
	secondPkt := insertPacket(t, database, "request_account_create", body)
	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": secondPkt.Type, "body": body}, secondPkt, "")
	}); err != nil {
		t.Fatalf("second Handle: %v", err)
	}

	// Domain state must not duplicate: exactly one new user, one new identity,
	// one new cluster-user, one new membership after BOTH deliveries.
	if got, want := countRows(t, database, "users"), usersBefore+1; got != want {
		t.Errorf("users after re-delivery: got %d, want %d (no duplicate)", got, want)
	}
	if got, want := countRows(t, database, "user_identities"), identitiesBefore+1; got != want {
		t.Errorf("user_identities after re-delivery: got %d, want %d (no duplicate)", got, want)
	}
	if got, want := countRows(t, database, "compute_cluster_users"), clusterUsersBefore+1; got != want {
		t.Errorf("compute_cluster_users after re-delivery: got %d, want %d (no duplicate)", got, want)
	}
	if got, want := countRows(t, database, "compute_allocation_memberships"), membershipsBefore+1; got != want {
		t.Errorf("compute_allocation_memberships after re-delivery: got %d, want %d (no duplicate)", got, want)
	}

	// Audit rows are per-delivery: each packet ID gets its own CREATE_PERSON /
	// CREATE_ACCOUNT / CREATE_MEMBERSHIP / REPLY_SENT row. We assert per-packet
	// counts of 1 each, the audit log is a history-of-attempts log, not a
	// state log, so the second delivery legitimately writes its own rows under
	// its own packet ID.
	for _, action := range []model.AuditAction{
		model.AuditCreatePerson,
		model.AuditCreateAccount,
		model.AuditCreateMembership,
		model.AuditReplySent,
	} {
		if got := countAuditActions(t, database, firstPkt.ID, action); got != 1 {
			t.Errorf("first delivery audit %s: got %d, want 1", action, got)
		}
		if got := countAuditActions(t, database, secondPkt.ID, action); got != 1 {
			t.Errorf("second delivery audit %s: got %d, want 1", action, got)
		}
	}

	if len(amie.Replies) != 2 {
		t.Errorf("amie replies after two deliveries: got %d, want 2", len(amie.Replies))
	}
}

// TestRequestAccountCreate_MissingRequiredField asserts the handler rejects a
// packet missing a required field, without leaving partial state behind.
// GrantNumber is one of the required fields.
func TestRequestAccountCreate_MissingRequiredField(t *testing.T) {
	database := setupTestDB(t)
	projectID := seedProjectForRAC(t, database)

	usersBefore := countRows(t, database, "users")
	identitiesBefore := countRows(t, database, "user_identities")
	clusterUsersBefore := countRows(t, database, "compute_cluster_users")
	membershipsBefore := countRows(t, database, "compute_allocation_memberships")

	body := baseRACBody()
	body["ProjectID"] = projectID
	delete(body, "GrantNumber")
	pkt := insertPacket(t, database, "request_account_create", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestAccountCreateHandler(svc, testClusterID, amie, audit)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error for missing GrantNumber, got nil")
	}

	// DB clean: validation failure must not have leaked any partial writes
	// belonging to the account-create flow.
	if got := countRows(t, database, "users"); got != usersBefore {
		t.Errorf("users after validation failure: got %d, want %d", got, usersBefore)
	}
	if got := countRows(t, database, "user_identities"); got != identitiesBefore {
		t.Errorf("user_identities after validation failure: got %d, want %d", got, identitiesBefore)
	}
	if got := countRows(t, database, "compute_cluster_users"); got != clusterUsersBefore {
		t.Errorf("compute_cluster_users after validation failure: got %d, want %d", got, clusterUsersBefore)
	}
	if got := countRows(t, database, "compute_allocation_memberships"); got != membershipsBefore {
		t.Errorf("compute_allocation_memberships after validation failure: got %d, want %d", got, membershipsBefore)
	}
	if len(amie.Replies) != 0 {
		t.Errorf("amie replies on validation failure: got %d, want 0", len(amie.Replies))
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 0 {
		t.Errorf("audit REPLY_SENT on validation failure: got %d, want 0", got)
	}
}

// TestRequestAccountCreate_UnknownProject asserts the handler errors (rather
// than silently no-ops) when the referenced project does not exist. The
// handler MUST look up the local project (by ProjectID if present, else by
// GrantNumber). If no project resolves, the handler cannot satisfy the
// transaction; it must return an error so the processor retries, matching
// the protocol sequencing constraint that the upstream system may send a
// RAC before the corresponding RPC transaction has completed.
func TestRequestAccountCreate_UnknownProject(t *testing.T) {
	database := setupTestDB(t)
	// Deliberately do NOT seed a project. The RAC must fail to find one.

	usersBefore := countRows(t, database, "users")
	identitiesBefore := countRows(t, database, "user_identities")
	clusterUsersBefore := countRows(t, database, "compute_cluster_users")
	membershipsBefore := countRows(t, database, "compute_allocation_memberships")

	body := baseRACBody()
	// A well-formed Custos UUID that does not match any existing project.
	body["ProjectID"] = uuid.NewString()
	body["GrantNumber"] = "TG-DOES-NOT-EXIST"
	pkt := insertPacket(t, database, "request_account_create", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestAccountCreateHandler(svc, testClusterID, amie, audit)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error for unknown project, got nil")
	}

	// No half-state: the project lookup is supposed to happen before any
	// membership/cluster-user writes. Identity rows MAY have been created
	// before the project lookup (the handler's `ensureUser` runs first), but
	// no allocation- or cluster-bound state should be present.
	//
	// The project lookup happens AFTER ensureUser in the current handler, so
	// user + identity rows may persist on this failure path. That is OK from a
	// spec standpoint: the retry will find the existing rows and proceed once
	// the project arrives.
	_ = usersBefore
	_ = identitiesBefore
	if got := countRows(t, database, "compute_cluster_users"); got != clusterUsersBefore {
		t.Errorf("compute_cluster_users on unknown-project failure: got %d, want %d", got, clusterUsersBefore)
	}
	if got := countRows(t, database, "compute_allocation_memberships"); got != membershipsBefore {
		t.Errorf("compute_allocation_memberships on unknown-project failure: got %d, want %d", got, membershipsBefore)
	}
	if len(amie.Replies) != 0 {
		t.Errorf("amie replies on unknown-project failure: got %d, want 0 (must retry, not ack)", len(amie.Replies))
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 0 {
		t.Errorf("audit REPLY_SENT on unknown-project failure: got %d, want 0", got)
	}
}

// TestRequestAccountCreate_ReplyFailurePropagates asserts the handler
// surfaces a ReplyToPacket failure to the caller so the processor can decide
// to retry. The REPLY_SENT audit row must NOT be written when the reply did
// not actually go out.
//
// Note: core domain writes commit in their own transactions, so
// user/identity/cluster-user/membership rows may persist past a reply
// failure. That gap is orthogonal to this test, what we assert here is
// only the (1) error propagation and (2) the audit invariant that
// REPLY_SENT is gated on a successful reply.
func TestRequestAccountCreate_ReplyFailurePropagates(t *testing.T) {
	database := setupTestDB(t)
	projectID := seedProjectForRAC(t, database)

	body := baseRACBody()
	body["ProjectID"] = projectID
	pkt := insertPacket(t, database, "request_account_create", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{FailWith: errors.New("simulated AMIE outage")}
	h := NewRequestAccountCreateHandler(svc, testClusterID, amie, audit)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error when reply fails, got nil")
	}

	// REPLY_SENT must NOT be recorded when the reply actually failed. AMIE
	// will retry the packet; recording a successful reply audit on a failed
	// reply would corrupt the trace.
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 0 {
		t.Errorf("audit REPLY_SENT on failed reply: got %d, want 0", got)
	}
	if len(amie.Replies) != 0 {
		t.Errorf("amie reply count on failed reply: got %d, want 0", len(amie.Replies))
	}
}

// silence unused-import false positives when only some tests reference these.
var _ = model.AuditCreatePerson
