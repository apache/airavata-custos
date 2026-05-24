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
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/pkg/models"
)

// mergeTestUsers is the pair of pre-existing users the merge tests operate on.
// Each user has its own organization, an AMIE user_identity (the GlobalID the
// merge packet keys on), a posix compute_cluster_user, an amie_user_dns row,
// and a compute_allocation_membership against a shared project + allocation.
// The retiring user is also the PI of an additional retiring-owned project so
// the spec's "projects with project_pi_id=retiring reassigned to surviving"
// invariant can be exercised.
type mergeTestUsers struct {
	survivingID      string
	retiringID       string
	survivingGlobal  string
	retiringGlobal   string
	survivingDN      string
	retiringDN       string
	survivingProject string // project where survivor is PI
	retiringProject  string // project where retiring is PI; should move to survivor
	allocationID     string // allocation under survivingProject
}

// seedMergeUsers inserts two ACTIVE users with the per-user state listed on
// mergeTestUsers above, plus one compute_allocation under the surviving user's
// project that both users have a membership on. Direct DB inserts are used
// (rather than going through the AMIE handler chain) so the merge tests can
// assert post-merge state cleanly: the seed only writes what the test needs.
func seedMergeUsers(t *testing.T, database *sqlx.DB) mergeTestUsers {
	t.Helper()
	ctx := context.Background()

	out := mergeTestUsers{
		survivingID:      uuid.NewString(),
		retiringID:       uuid.NewString(),
		survivingGlobal:  "bl-survivor-001",
		retiringGlobal:   "bl-retiring-001",
		survivingDN:      "/C=US/O=Test/CN=Pat Survivor",
		retiringDN:       "/C=US/O=Test/CN=Pat Retiring",
		survivingProject: uuid.NewString(),
		retiringProject:  uuid.NewString(),
		allocationID:     uuid.NewString(),
	}
	survivingOrg := uuid.NewString()
	retiringOrg := uuid.NewString()
	survivingMembership := uuid.NewString()
	retiringMembership := uuid.NewString()
	survivingClusterUser := uuid.NewString()
	retiringClusterUser := uuid.NewString()
	survivingIdentity := uuid.NewString()
	retiringIdentity := uuid.NewString()

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seed begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	exec := func(query string, args ...any) {
		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			t.Fatalf("seed exec %q: %v", query, err)
		}
	}

	exec("INSERT INTO organizations (id, originated_id, name) VALUES (?, ?, ?)",
		survivingOrg, "SURV", "Surviving Org")
	exec("INSERT INTO organizations (id, originated_id, name) VALUES (?, ?, ?)",
		retiringOrg, "RETR", "Retiring Org")

	exec("INSERT INTO users (id, organization_id, first_name, last_name, middle_name, email, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		out.survivingID, survivingOrg, "Pat", "Survivor", "", "pat.survivor@example.edu", string(models.UserActive))
	exec("INSERT INTO users (id, organization_id, first_name, last_name, middle_name, email, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		out.retiringID, retiringOrg, "Pat", "Retiring", "", "pat.retiring@example.edu", string(models.UserActive))

	exec("INSERT INTO user_identities (id, user_id, source, external_id) VALUES (?, ?, ?, ?)",
		survivingIdentity, out.survivingID, amieIdentitySource, out.survivingGlobal)
	exec("INSERT INTO user_identities (id, user_id, source, external_id) VALUES (?, ?, ?, ?)",
		retiringIdentity, out.retiringID, amieIdentitySource, out.retiringGlobal)

	exec("INSERT INTO compute_cluster_users (id, compute_cluster_id, user_id, local_username) VALUES (?, ?, ?, ?)",
		survivingClusterUser, testClusterID, out.survivingID, "psurvivor")
	exec("INSERT INTO compute_cluster_users (id, compute_cluster_id, user_id, local_username) VALUES (?, ?, ?, ?)",
		retiringClusterUser, testClusterID, out.retiringID, "pretiring")

	exec("INSERT INTO amie_user_dns (id, user_id, dn) VALUES (?, ?, ?)",
		uuid.NewString(), out.survivingID, out.survivingDN)
	exec("INSERT INTO amie_user_dns (id, user_id, dn) VALUES (?, ?, ?)",
		uuid.NewString(), out.retiringID, out.retiringDN)

	// Two projects: one PI'd by each user. The retiring user's project should
	// be reassigned to the surviving user after the merge per the
	// service.MergeUsers contract (projs.ReassignPI).
	exec("INSERT INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES (?, ?, ?, ?, ?, ?)",
		out.survivingProject, "TG-SURV-001", "Surviving Project", "ACCESS", out.survivingID, string(models.ProjectActive))
	exec("INSERT INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES (?, ?, ?, ?, ?, ?)",
		out.retiringProject, "TG-RETR-001", "Retiring Project", "ACCESS", out.retiringID, string(models.ProjectActive))

	// One allocation under the surviving project, with both users as members.
	// After the merge the retiring user's membership row must move to the
	// surviving user.
	startTime := time.Now().UTC()
	endTime := startTime.Add(365 * 24 * time.Hour)
	exec(`INSERT INTO compute_allocations
	          (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time)
	      VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		out.allocationID, out.survivingProject, "TG-SURV-001-alloc",
		string(models.ACTIVE), testClusterID, int64(1000), startTime, endTime)

	exec(`INSERT INTO compute_allocation_memberships
	          (id, compute_allocation_id, user_id, start_time, end_time, membership_status)
	      VALUES (?, ?, ?, ?, ?, ?)`,
		survivingMembership, out.allocationID, out.survivingID, startTime, endTime, string(models.ACTIVE))
	exec(`INSERT INTO compute_allocation_memberships
	          (id, compute_allocation_id, user_id, start_time, end_time, membership_status)
	      VALUES (?, ?, ?, ?, ?, ?)`,
		retiringMembership, out.allocationID, out.retiringID, startTime, endTime, string(models.ACTIVE))

	if err := tx.Commit(); err != nil {
		t.Fatalf("seed commit: %v", err)
	}
	return out
}

// baseRPMBody returns a fully populated request_person_merge body with
// every required field: KeepGlobalID, KeepPersonID, DeleteGlobalID,
// DeletePersonID.
func baseRPMBody(seed mergeTestUsers) map[string]any {
	return map[string]any{
		"KeepGlobalID":   seed.survivingGlobal,
		"KeepPersonID":   seed.survivingID,
		"DeleteGlobalID": seed.retiringGlobal,
		"DeletePersonID": seed.retiringID,
	}
}

// TestRequestPersonMerge_HappyPath asserts the spec-required side effects of
// a first-delivery request_person_merge with a complete body, against a DB
// pre-seeded with both users + their per-user state.
//
// The handler MUST:
//   - flip the retiring user's status to MERGED (soft-delete)
//   - reassign every user_identity row from retiring -> surviving
//   - reassign every amie_user_dns row from retiring -> surviving
//   - reassign every compute_allocation_membership from retiring -> surviving
//   - reassign every project with project_pi_id=retiring to surviving
//   - write MERGE_PERSONS + REPLY_SENT audit rows for the packet
//   - reply `inform_transaction_complete`
func TestRequestPersonMerge_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	seed := seedMergeUsers(t, database)

	body := baseRPMBody(seed)
	pkt := insertPacket(t, database, "request_person_merge", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	userDNStore := store.NewUserDNStore(database)
	h := NewRequestPersonMergeHandler(svc, userDNStore, amie, audit)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// Retiring user's status must be MERGED (soft-delete). The surviving user
	// stays ACTIVE.
	var retiringStatus string
	if err := database.Get(&retiringStatus, "SELECT status FROM users WHERE id = ?", seed.retiringID); err != nil {
		t.Fatalf("read retiring user status: %v", err)
	}
	if retiringStatus != string(models.UserMerged) {
		t.Errorf("retiring user.status: got %q, want %q", retiringStatus, string(models.UserMerged))
	}
	var survivingStatus string
	if err := database.Get(&survivingStatus, "SELECT status FROM users WHERE id = ?", seed.survivingID); err != nil {
		t.Fatalf("read surviving user status: %v", err)
	}
	if survivingStatus != string(models.UserActive) {
		t.Errorf("surviving user.status: got %q, want %q", survivingStatus, string(models.UserActive))
	}

	// user_identities: every row originally on the retiring user must now point
	// at the surviving user.
	var identitiesOnRetiring int
	if err := database.Get(&identitiesOnRetiring,
		"SELECT COUNT(*) FROM user_identities WHERE user_id = ?", seed.retiringID); err != nil {
		t.Fatalf("count identities on retiring: %v", err)
	}
	if identitiesOnRetiring != 0 {
		t.Errorf("user_identities on retiring after merge: got %d, want 0", identitiesOnRetiring)
	}
	var identitiesOnSurviving int
	if err := database.Get(&identitiesOnSurviving,
		"SELECT COUNT(*) FROM user_identities WHERE user_id = ?", seed.survivingID); err != nil {
		t.Fatalf("count identities on surviving: %v", err)
	}
	// Survivor had 1 (KeepGlobalID), retiring had 1 (DeleteGlobalID) → 2 after.
	if identitiesOnSurviving != 2 {
		t.Errorf("user_identities on surviving after merge: got %d, want 2", identitiesOnSurviving)
	}

	// amie_user_dns: every DN row originally on the retiring user must now
	// point at the surviving user (and the surviving user's original DN row
	// must still be present).
	var dnsOnRetiring int
	if err := database.Get(&dnsOnRetiring,
		"SELECT COUNT(*) FROM amie_user_dns WHERE user_id = ?", seed.retiringID); err != nil {
		t.Fatalf("count dns on retiring: %v", err)
	}
	if dnsOnRetiring != 0 {
		t.Errorf("amie_user_dns on retiring after merge: got %d, want 0", dnsOnRetiring)
	}
	var dnsOnSurviving int
	if err := database.Get(&dnsOnSurviving,
		"SELECT COUNT(*) FROM amie_user_dns WHERE user_id = ?", seed.survivingID); err != nil {
		t.Fatalf("count dns on surviving: %v", err)
	}
	if dnsOnSurviving != 2 {
		t.Errorf("amie_user_dns on surviving after merge: got %d, want 2", dnsOnSurviving)
	}

	// compute_allocation_memberships: retiring's membership row reassigned to
	// surviving. UNIQUE(allocation_id, user_id) means the duplicate row is
	// dropped, but the survivor's existing row remains, so the surviving user
	// ends with exactly one membership on that allocation.
	var membershipsOnRetiring int
	if err := database.Get(&membershipsOnRetiring,
		"SELECT COUNT(*) FROM compute_allocation_memberships WHERE user_id = ?", seed.retiringID); err != nil {
		t.Fatalf("count memberships on retiring: %v", err)
	}
	if membershipsOnRetiring != 0 {
		t.Errorf("compute_allocation_memberships on retiring after merge: got %d, want 0", membershipsOnRetiring)
	}
	var membershipsOnSurviving int
	if err := database.Get(&membershipsOnSurviving,
		"SELECT COUNT(*) FROM compute_allocation_memberships WHERE user_id = ?", seed.survivingID); err != nil {
		t.Fatalf("count memberships on surviving: %v", err)
	}
	if membershipsOnSurviving < 1 {
		t.Errorf("compute_allocation_memberships on surviving after merge: got %d, want >= 1", membershipsOnSurviving)
	}

	// projects: any project with project_pi_id=retiring must move to surviving.
	// (Asserted explicitly because the domain-writes doc's row 6 says "FKs
	// follow the surviving person automatically", service.MergeUsers
	// implements that via projs.ReassignPI.)
	var retiringProjectPI string
	if err := database.Get(&retiringProjectPI,
		"SELECT project_pi_id FROM projects WHERE id = ?", seed.retiringProject); err != nil {
		t.Fatalf("read retiring project PI: %v", err)
	}
	if retiringProjectPI != seed.survivingID {
		t.Errorf("project_pi_id of retiring's project: got %q, want %q (surviving user)",
			retiringProjectPI, seed.survivingID)
	}

	// Audit: MERGE_PERSONS and REPLY_SENT, exactly one each.
	if got := countAuditActions(t, database, pkt.ID, model.AuditMergePersons); got != 1 {
		t.Errorf("audit MERGE_PERSONS: got %d, want 1", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT: got %d, want 1", got)
	}

	// Reply: inform_transaction_complete.
	if got, want := amie.lastReplyType(), "inform_transaction_complete"; got != want {
		t.Fatalf("reply type: got %q, want %q", got, want)
	}
	reply := amie.lastReplyBody()
	if reply == nil {
		t.Fatal("reply body missing")
	}
	if got, _ := reply["StatusCode"].(string); got != "Success" {
		t.Errorf("reply.StatusCode: got %q, want %q", got, "Success")
	}
}

// TestRequestPersonMerge_ReplaySafeAfterMerge asserts that re-delivery of a
// `request_person_merge` packet after a successful merge is a spec-correct
// no-op success. The protocol is silent on idempotency, but the contract
// for any incoming packet is that the handler must reply ITC. The
// handler's existing replay guard keys on retiring user.status=MERGED and
// short-circuits to the reply path without re-doing the merge. We assert:
//   - second Handle returns nil (success)
//   - state did not change between the two calls (no extra merge writes)
//   - the second packet's audit log records REPLY_SENT but NOT a second
//     MERGE_PERSONS (the merge already happened on the first call)
//   - the surviving user got two replies in total (one per delivery)
func TestRequestPersonMerge_ReplaySafeAfterMerge(t *testing.T) {
	database := setupTestDB(t)
	seed := seedMergeUsers(t, database)

	body := baseRPMBody(seed)
	firstPkt := insertPacket(t, database, "request_person_merge", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	userDNStore := store.NewUserDNStore(database)
	h := NewRequestPersonMergeHandler(svc, userDNStore, amie, audit)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": firstPkt.Type, "body": body}, firstPkt, "")
	}); err != nil {
		t.Fatalf("first Handle: %v", err)
	}

	// Snapshot state after first merge.
	identitiesOnSurviving := func() int {
		var n int
		if err := database.Get(&n,
			"SELECT COUNT(*) FROM user_identities WHERE user_id = ?", seed.survivingID); err != nil {
			t.Fatalf("count identities: %v", err)
		}
		return n
	}
	dnsOnSurviving := func() int {
		var n int
		if err := database.Get(&n,
			"SELECT COUNT(*) FROM amie_user_dns WHERE user_id = ?", seed.survivingID); err != nil {
			t.Fatalf("count dns: %v", err)
		}
		return n
	}
	identitiesBefore := identitiesOnSurviving()
	dnsBefore := dnsOnSurviving()

	// Re-deliver: SAME body, distinct packet row (AMIE may re-send the same
	// transaction after an connector restart before ACK).
	secondPkt := insertPacket(t, database, "request_person_merge", body)
	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": secondPkt.Type, "body": body}, secondPkt, "")
	}); err != nil {
		t.Fatalf("second Handle: %v", err)
	}

	// State must not have shifted on replay: no second merge writes.
	if got := identitiesOnSurviving(); got != identitiesBefore {
		t.Errorf("user_identities on surviving after replay: got %d, want %d", got, identitiesBefore)
	}
	if got := dnsOnSurviving(); got != dnsBefore {
		t.Errorf("amie_user_dns on surviving after replay: got %d, want %d", got, dnsBefore)
	}

	// First packet has MERGE_PERSONS + REPLY_SENT. Second packet has REPLY_SENT
	// only (the replay short-circuits past the merge audit row).
	if got := countAuditActions(t, database, firstPkt.ID, model.AuditMergePersons); got != 1 {
		t.Errorf("first packet MERGE_PERSONS audit: got %d, want 1", got)
	}
	if got := countAuditActions(t, database, firstPkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("first packet REPLY_SENT audit: got %d, want 1", got)
	}
	if got := countAuditActions(t, database, secondPkt.ID, model.AuditMergePersons); got != 0 {
		t.Errorf("second packet MERGE_PERSONS audit on replay: got %d, want 0 (already merged)", got)
	}
	if got := countAuditActions(t, database, secondPkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("second packet REPLY_SENT audit on replay: got %d, want 1", got)
	}

	// Each delivery sends its own reply back to AMIE.
	if len(amie.Replies) != 2 {
		t.Errorf("amie replies after two deliveries: got %d, want 2", len(amie.Replies))
	}
}

// TestRequestPersonMerge_SameKeepAndDelete asserts the handler rejects a
// packet where KeepGlobalID == DeleteGlobalID. Per the protocol the two
// fields identify two distinct persons (the survivor and the retiring
// duplicate).
// Merging a user with themselves is nonsensical and would silently no-op
// without flipping status; the handler must return an error and not write
// any audit rows.
func TestRequestPersonMerge_SameKeepAndDelete(t *testing.T) {
	database := setupTestDB(t)
	seed := seedMergeUsers(t, database)

	body := baseRPMBody(seed)
	// Force keep == delete on both GlobalID and PersonID. The handler's check
	// keys on GlobalID equality so that's the dispositive equality, but matching
	// the PersonIDs too keeps the body internally consistent.
	body["DeleteGlobalID"] = body["KeepGlobalID"]
	body["DeletePersonID"] = body["KeepPersonID"]
	pkt := insertPacket(t, database, "request_person_merge", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	userDNStore := store.NewUserDNStore(database)
	h := NewRequestPersonMergeHandler(svc, userDNStore, amie, audit)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error for KeepGlobalID == DeleteGlobalID, got nil")
	}

	// No state change: surviving user keeps every identity/DN row, status
	// untouched, audit log empty for this packet.
	var survivingStatus string
	if err := database.Get(&survivingStatus, "SELECT status FROM users WHERE id = ?", seed.survivingID); err != nil {
		t.Fatalf("read surviving user status: %v", err)
	}
	if survivingStatus != string(models.UserActive) {
		t.Errorf("surviving user.status: got %q, want %q (untouched)", survivingStatus, string(models.UserActive))
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditMergePersons); got != 0 {
		t.Errorf("audit MERGE_PERSONS on validation failure: got %d, want 0", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 0 {
		t.Errorf("audit REPLY_SENT on validation failure: got %d, want 0", got)
	}
	if len(amie.Replies) != 0 {
		t.Errorf("amie replies on validation failure: got %d, want 0", len(amie.Replies))
	}
}

// TestRequestPersonMerge_UnknownRetiringOrSurviving asserts the handler errors
// when a body GlobalID does not resolve to an existing user_identity row.
//
// Spec note: the protocol is explicitly silent on this case
// (the protocol example does not pre-check existence, and the ACCESS narrative doc
// does not describe the transaction body). The handler's design choice, and
// the spec-defensible one, is to return an error so the processor retries.
// Reply-with-ITC-Success would acknowledge a merge that did not happen and
// confuse AMIE's record-keeping. Reply-with-failure is not in the documented
// reply set for this packet type (the only documented reply is ITC success).
//
// We assert: (1) the handler returns an error, (2) no audit rows are written
// for the packet, (3) no reply is sent to AMIE.
func TestRequestPersonMerge_UnknownRetiringOrSurviving(t *testing.T) {
	database := setupTestDB(t)
	seed := seedMergeUsers(t, database)

	body := baseRPMBody(seed)
	// Point DeleteGlobalID at an external_id that does not exist on any
	// user_identity row. Use a syntactically-plausible AMIE GlobalID.
	body["DeleteGlobalID"] = "bl-nonexistent-999"
	pkt := insertPacket(t, database, "request_person_merge", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	userDNStore := store.NewUserDNStore(database)
	h := NewRequestPersonMergeHandler(svc, userDNStore, amie, audit)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error for unknown DeleteGlobalID, got nil")
	}

	// State unchanged.
	var retiringStatus string
	if err := database.Get(&retiringStatus, "SELECT status FROM users WHERE id = ?", seed.retiringID); err != nil {
		t.Fatalf("read retiring user status: %v", err)
	}
	if retiringStatus != string(models.UserActive) {
		t.Errorf("retiring user.status: got %q, want %q (must not flip on lookup failure)",
			retiringStatus, string(models.UserActive))
	}

	// No audit rows for this packet, the handler bailed before the merge or
	// the reply was attempted.
	if got := countAuditActions(t, database, pkt.ID, model.AuditMergePersons); got != 0 {
		t.Errorf("audit MERGE_PERSONS on unknown user: got %d, want 0", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 0 {
		t.Errorf("audit REPLY_SENT on unknown user: got %d, want 0", got)
	}
	if len(amie.Replies) != 0 {
		t.Errorf("amie replies on unknown user: got %d, want 0 (must retry, not ack)", len(amie.Replies))
	}
}

// TestRequestPersonMerge_ReplyFailurePropagates asserts the handler surfaces
// a ReplyToPacket failure to the caller so the processor can decide to retry.
// The REPLY_SENT audit row must NOT be written when the reply did not actually
// go out.
//
// As with the analogous test in request_project_create_integration_test.go:
// the merge itself goes through service.MergeUsers, which opens its own
// transaction. So the surviving/retiring state changes commit independently
// of the handler tx; we do not assert rollback shape here, only the (1)
// error propagation and (2) audit invariant.
func TestRequestPersonMerge_ReplyFailurePropagates(t *testing.T) {
	database := setupTestDB(t)
	seed := seedMergeUsers(t, database)

	body := baseRPMBody(seed)
	pkt := insertPacket(t, database, "request_person_merge", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{FailWith: errors.New("simulated AMIE outage")}
	userDNStore := store.NewUserDNStore(database)
	h := NewRequestPersonMergeHandler(svc, userDNStore, amie, audit)

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
var _ = model.AuditMergePersons
