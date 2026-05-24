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
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/pkg/models"
)

// seedUser inserts an organization, a user, a user_identity bound to (access,
// userGlobalID), and the supplied DNs into amie_user_dns. It returns the
// inserted user ID, user_identity ID, and DN row IDs so tests can assert
// downstream mutations relative to the seeded state.
func seedUser(t *testing.T, database *sqlx.DB, userGlobalID, firstName, lastName, email string, dns []string) (userID, identityID string, dnIDs []string) {
	t.Helper()
	orgID := uuid.NewString()
	if _, err := database.Exec(
		"INSERT INTO organizations (id, originated_id, name) VALUES (?, ?, ?)",
		orgID, "SEED-ORG", "Seed Org",
	); err != nil {
		t.Fatalf("seed organization: %v", err)
	}
	userID = uuid.NewString()
	if _, err := database.Exec(
		"INSERT INTO users (id, organization_id, first_name, last_name, middle_name, email, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, orgID, firstName, lastName, "", email, string(models.UserActive),
	); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	identityID = uuid.NewString()
	seedMeta, err := json.Marshal(map[string]any{
		"organization":    "Old Org",
		"org_code":        "OLD-ORG",
		"nsf_status_code": "PENDING",
	})
	if err != nil {
		t.Fatalf("marshal seed metadata: %v", err)
	}
	if _, err := database.Exec(
		"INSERT INTO user_identities (id, user_id, source, external_id, email, metadata) VALUES (?, ?, ?, ?, ?, ?)",
		identityID, userID, "access", userGlobalID, email, string(seedMeta),
	); err != nil {
		t.Fatalf("seed user_identity: %v", err)
	}
	for _, dn := range dns {
		dnID := uuid.NewString()
		if _, err := database.Exec(
			"INSERT INTO amie_user_dns (id, user_id, dn) VALUES (?, ?, ?)",
			dnID, userID, dn,
		); err != nil {
			t.Fatalf("seed amie_user_dns: %v", err)
		}
		dnIDs = append(dnIDs, dnID)
	}
	return userID, identityID, dnIDs
}

// listUserDNs reads the DN strings currently stored for a given user, sorted
// for stable comparison.
func listUserDNs(t *testing.T, database *sqlx.DB, userID string) []string {
	t.Helper()
	var dns []string
	if err := database.Select(&dns,
		"SELECT dn FROM amie_user_dns WHERE user_id = ? ORDER BY dn",
		userID,
	); err != nil {
		t.Fatalf("list amie_user_dns: %v", err)
	}
	return dns
}

// newRequestUserModifyHandlerForTest wires the handler with the package's
// real services + the supplied fake AMIE client and a freshly minted
// UserDNStore on the shared connection.
func newRequestUserModifyHandlerForTest(database *sqlx.DB, amie *fakeAmieClient) *RequestUserModifyHandler {
	return NewRequestUserModifyHandler(
		newTestCoreService(database),
		store.NewUserDNStore(database),
		amie,
		newTestAuditService(database),
	)
}

// baseUserModifyBody builds a baseline request_user_modify body with the
// fields the handler actually reads. Tests mutate the returned map.
func baseUserModifyBody(actionType, userGlobalID string) map[string]any {
	return map[string]any{
		"ActionType":   actionType,
		"UserGlobalID": userGlobalID,
	}
}

// TestRequestUserModify_ReplaceHappyPath: ActionType=replace with a full
// body. The handler MUST:
//   - update the User row's basic profile from the packet
//   - update the UserIdentity row's email + source-specific metadata
//     (organization, org_code, nsf_status_code)
//   - reconcile the user's DN list against the packet's DnList
//   - reply inform_transaction_complete with StatusCode=Success
func TestRequestUserModify_ReplaceHappyPath(t *testing.T) {
	database := setupTestDB(t)

	const (
		userGlobalID = "amie-user-1001"
		seedDN1      = "/C=US/O=Old/CN=Old User A"
		seedDN2      = "/C=US/O=Old/CN=Old User B"
		newDN        = "/C=US/O=New Org/CN=New User"
	)
	userID, identityID, _ := seedUser(t, database,
		userGlobalID, "Old", "Name", "old@example.edu",
		[]string{seedDN1, seedDN2},
	)

	body := baseUserModifyBody("replace", userGlobalID)
	body["UserFirstName"] = "New"
	body["UserLastName"] = "Name"
	body["UserEmail"] = "new@example.edu"
	body["UserOrganization"] = "New Org"
	body["UserOrgCode"] = "NEW-ORG"
	body["NsfStatusCode"] = "ACTIVE"
	// Replace semantics: presence of DnList means the local DN set must be
	// replaced with the packet's set. seedDN1 is retained and seedDN2 must
	// be dropped.
	body["DnList"] = []any{seedDN1, newDN}

	pkt := insertPacket(t, database, "request_user_modify", body)
	amie := &fakeAmieClient{}
	h := newRequestUserModifyHandlerForTest(database, amie)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// User row: first name, last name, email updated.
	var user struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
	}
	if err := database.Get(&user, "SELECT first_name, last_name, email FROM users WHERE id = ?", userID); err != nil {
		t.Fatalf("read user: %v", err)
	}
	if user.FirstName != "New" {
		t.Errorf("user.first_name: got %q, want New", user.FirstName)
	}
	if user.LastName != "Name" {
		t.Errorf("user.last_name: got %q, want Name", user.LastName)
	}
	if user.Email != "new@example.edu" {
		t.Errorf("user.email: got %q, want new@example.edu", user.Email)
	}

	// UserIdentity row: email + metadata updated.
	var ident struct {
		Email    *string `db:"email"`
		Metadata *string `db:"metadata"`
	}
	if err := database.Get(&ident, "SELECT email, metadata FROM user_identities WHERE id = ?", identityID); err != nil {
		t.Fatalf("read user_identity: %v", err)
	}
	if ident.Email == nil || *ident.Email != "new@example.edu" {
		got := "<nil>"
		if ident.Email != nil {
			got = *ident.Email
		}
		t.Errorf("user_identity.email: got %q, want new@example.edu", got)
	}
	if ident.Metadata == nil {
		t.Fatalf("user_identity.metadata: nil; want JSON with new org / org_code / nsf_status_code")
	}
	var meta map[string]any
	if err := json.Unmarshal([]byte(*ident.Metadata), &meta); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	if got, _ := meta["organization"].(string); got != "New Org" {
		t.Errorf("metadata.organization: got %q, want New Org", got)
	}
	if got, _ := meta["org_code"].(string); got != "NEW-ORG" {
		t.Errorf("metadata.org_code: got %q, want NEW-ORG", got)
	}
	if got, _ := meta["nsf_status_code"].(string); got != "ACTIVE" {
		t.Errorf("metadata.nsf_status_code: got %q, want ACTIVE", got)
	}

	// DN reconciliation: seedDN1 kept, seedDN2 removed, newDN added.
	got := listUserDNs(t, database, userID)
	want := []string{newDN, seedDN1}
	sort.Strings(want)
	if !equalStringSlices(got, want) {
		t.Errorf("DN set: got %v, want %v", got, want)
	}

	// Audits.
	if got := countAuditActions(t, database, pkt.ID, model.AuditUpdatePerson); got != 1 {
		t.Errorf("audit UPDATE_PERSON: got %d, want 1", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditPersistDNs); got != 1 {
		t.Errorf("audit PERSIST_DNS: got %d, want 1 (added + removed > 0)", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT: got %d, want 1", got)
	}

	// Reply: inform_transaction_complete with Success / DetailCode=1.
	if got, want := amie.lastReplyType(), "inform_transaction_complete"; got != want {
		t.Errorf("reply type: got %q, want %q", got, want)
	}
	reply := amie.lastReplyBody()
	if reply == nil {
		t.Fatal("reply body missing")
	}
	if got, _ := reply["StatusCode"].(string); got != "Success" {
		t.Errorf("reply.StatusCode: got %q, want Success", got)
	}
}

// TestRequestUserModify_ReplaceIdempotentSameDNs, re-handling the same
// replace packet must not produce duplicate DN rows or extra PERSIST_DNS audit
// activity beyond what each handle naturally emits. Replace is row-level
// convergent (every update sets the same target state).
func TestRequestUserModify_ReplaceIdempotentSameDNs(t *testing.T) {
	database := setupTestDB(t)

	const (
		userGlobalID = "amie-user-1002"
		dnA          = "/C=US/O=Org/CN=A"
		dnB          = "/C=US/O=Org/CN=B"
	)
	userID, _, _ := seedUser(t, database,
		userGlobalID, "First", "Last", "user1002@example.edu",
		[]string{dnA, dnB},
	)

	body := baseUserModifyBody("replace", userGlobalID)
	body["UserFirstName"] = "First"
	body["UserLastName"] = "Last"
	body["UserEmail"] = "user1002@example.edu"
	body["UserOrganization"] = "Org"
	body["UserOrgCode"] = "ORG"
	body["DnList"] = []any{dnA, dnB}

	pkt1 := insertPacket(t, database, "request_user_modify", body)
	amie := &fakeAmieClient{}
	h := newRequestUserModifyHandlerForTest(database, amie)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt1.Type, "body": body}, pkt1, "")
	}); err != nil {
		t.Fatalf("first Handle: %v", err)
	}

	pkt2 := insertPacket(t, database, "request_user_modify", body)
	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt2.Type, "body": body}, pkt2, "")
	}); err != nil {
		t.Fatalf("second Handle: %v", err)
	}

	// DN set unchanged: still exactly {dnA, dnB}.
	got := listUserDNs(t, database, userID)
	want := []string{dnA, dnB}
	sort.Strings(want)
	if !equalStringSlices(got, want) {
		t.Errorf("DN set after re-handle: got %v, want %v", got, want)
	}
	if total := countRows(t, database, "amie_user_dns"); total != 2 {
		t.Errorf("amie_user_dns total rows: got %d, want 2 (no duplicates)", total)
	}

	// PERSIST_DNS only emitted when added+removed > 0. The second handle is a
	// pure no-op on the DN set, so it must NOT emit a second PERSIST_DNS.
	if got := countAuditActions(t, database, pkt2.ID, model.AuditPersistDNs); got != 0 {
		t.Errorf("audit PERSIST_DNS on no-op replace: got %d, want 0", got)
	}

	// Each handle emits its own UPDATE_PERSON and REPLY_SENT, these are
	// per-packet, not deduped. We assert the second packet emitted them once
	// each (i.e. the handler ran end-to-end on the second delivery).
	if got := countAuditActions(t, database, pkt2.ID, model.AuditUpdatePerson); got != 1 {
		t.Errorf("audit UPDATE_PERSON on second handle: got %d, want 1", got)
	}
	if got := countAuditActions(t, database, pkt2.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT on second handle: got %d, want 1", got)
	}
}

// TestRequestUserModify_ReplaceWithDifferentDNs, replace with a DnList that
// REMOVES some existing DNs and ADDS new ones. Replace semantics treat DnList
// as the desired set: drop locals absent from the packet, add packet entries
// not present locally. Caveat: a partial upstream DnList would wipe local
// DNs, so this is only safe when upstream sends complete lists.
func TestRequestUserModify_ReplaceWithDifferentDNs(t *testing.T) {
	database := setupTestDB(t)

	const (
		userGlobalID = "amie-user-1003"
		keepDN       = "/C=US/O=Org/CN=Keep"
		dropDN       = "/C=US/O=Org/CN=Drop"
		addDN1       = "/C=US/O=Org/CN=Add1"
		addDN2       = "/C=US/O=Org/CN=Add2"
	)
	userID, _, _ := seedUser(t, database,
		userGlobalID, "Stable", "Person", "user1003@example.edu",
		[]string{keepDN, dropDN},
	)

	body := baseUserModifyBody("replace", userGlobalID)
	body["UserFirstName"] = "Stable"
	body["UserLastName"] = "Person"
	body["UserEmail"] = "user1003@example.edu"
	body["UserOrganization"] = "Org"
	body["UserOrgCode"] = "ORG"
	body["DnList"] = []any{keepDN, addDN1, addDN2}

	pkt := insertPacket(t, database, "request_user_modify", body)
	amie := &fakeAmieClient{}
	h := newRequestUserModifyHandlerForTest(database, amie)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	got := listUserDNs(t, database, userID)
	want := []string{addDN1, addDN2, keepDN}
	sort.Strings(want)
	if !equalStringSlices(got, want) {
		t.Errorf("DN set after destructive sync: got %v, want %v", got, want)
	}
	// dropDN must be gone.
	for _, dn := range got {
		if dn == dropDN {
			t.Errorf("DN %q should have been removed by destructive sync", dropDN)
		}
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditPersistDNs); got != 1 {
		t.Errorf("audit PERSIST_DNS: got %d, want 1 (DN delta non-empty)", got)
	}
}

// TestRequestUserModify_DeleteHappyPath: ActionType=delete with a DnList of
// DNs to inactivate. The protocol's guidance: "If delete: reads
// inactive_dn_list = packet.DnList. Inactivate the specified DNs for the
// user." The delete action affects DN rows only (no other identity entity).
// The named DNs must be removed; unnamed DNs and the User / UserIdentity
// rows must remain intact.
func TestRequestUserModify_DeleteHappyPath(t *testing.T) {
	database := setupTestDB(t)

	const (
		userGlobalID = "amie-user-1004"
		keepDN1      = "/C=US/O=Org/CN=Keep1"
		keepDN2      = "/C=US/O=Org/CN=Keep2"
		deleteDN     = "/C=US/O=Org/CN=Delete"
	)
	const (
		origFirst = "Tobe"
		origLast  = "Preserved"
		origEmail = "preserved@example.edu"
	)
	userID, identityID, _ := seedUser(t, database,
		userGlobalID, origFirst, origLast, origEmail,
		[]string{keepDN1, keepDN2, deleteDN},
	)

	body := baseUserModifyBody("delete", userGlobalID)
	body["DnList"] = []any{deleteDN}

	pkt := insertPacket(t, database, "request_user_modify", body)
	amie := &fakeAmieClient{}
	h := newRequestUserModifyHandlerForTest(database, amie)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// DNs: keepDN1 and keepDN2 retained, deleteDN removed.
	got := listUserDNs(t, database, userID)
	want := []string{keepDN1, keepDN2}
	sort.Strings(want)
	if !equalStringSlices(got, want) {
		t.Errorf("DN set after delete: got %v, want %v", got, want)
	}

	// User row untouched.
	var user struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
		Status    string `db:"status"`
	}
	if err := database.Get(&user, "SELECT first_name, last_name, email, status FROM users WHERE id = ?", userID); err != nil {
		t.Fatalf("read user: %v", err)
	}
	if user.FirstName != origFirst || user.LastName != origLast || user.Email != origEmail {
		t.Errorf("user row mutated: got %+v, want first=%q last=%q email=%q", user, origFirst, origLast, origEmail)
	}
	if user.Status != string(models.UserActive) {
		t.Errorf("user.status: got %q, want ACTIVE (delete on DNs must not change user lifecycle)", user.Status)
	}

	// UserIdentity row untouched (no DELETE_PERSON path for DN-only delete).
	if got := countRows(t, database, "user_identities"); got != 1 {
		t.Errorf("user_identities: got %d, want 1", got)
	}
	var identUserID string
	if err := database.Get(&identUserID, "SELECT user_id FROM user_identities WHERE id = ?", identityID); err != nil {
		t.Fatalf("read user_identity: %v", err)
	}
	if identUserID != userID {
		t.Errorf("user_identity.user_id rebound: got %q, want %q", identUserID, userID)
	}

	// Audits: PERSIST_DNS once (one DN removed), REPLY_SENT once,
	// no UPDATE_PERSON (delete path does not touch the user row).
	if got := countAuditActions(t, database, pkt.ID, model.AuditPersistDNs); got != 1 {
		t.Errorf("audit PERSIST_DNS: got %d, want 1", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditUpdatePerson); got != 0 {
		t.Errorf("audit UPDATE_PERSON on delete: got %d, want 0 (delete only affects DNs)", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT: got %d, want 1", got)
	}
}

// TestRequestUserModify_UnknownUser, packet refers to a UserGlobalID with no
// matching user_identity. The handler always replies ITC StatusCode='Success'
// regardless of branch: the connector may receive packets for linked person
// IDs that this site does not know, so soft-skip with success reply is the
// correct response. No state mutations expected.
func TestRequestUserModify_UnknownUser(t *testing.T) {
	database := setupTestDB(t)

	body := baseUserModifyBody("replace", "amie-user-not-here")
	body["UserFirstName"] = "Ghost"
	body["UserLastName"] = "User"
	body["UserEmail"] = "ghost@example.edu"
	body["DnList"] = []any{"/C=US/O=Ghost/CN=Ghost"}

	pkt := insertPacket(t, database, "request_user_modify", body)
	amie := &fakeAmieClient{}
	h := newRequestUserModifyHandlerForTest(database, amie)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error on unknown user (spec expects soft-skip): %v", err)
	}

	if got := countRows(t, database, "users"); got != 0 {
		t.Errorf("users: got %d, want 0 (no user_identity → no user creation)", got)
	}
	if got := countRows(t, database, "user_identities"); got != 0 {
		t.Errorf("user_identities: got %d, want 0", got)
	}
	if got := countRows(t, database, "amie_user_dns"); got != 0 {
		t.Errorf("amie_user_dns: got %d, want 0", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditUpdatePerson); got != 0 {
		t.Errorf("audit UPDATE_PERSON: got %d, want 0", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditPersistDNs); got != 0 {
		t.Errorf("audit PERSIST_DNS: got %d, want 0", got)
	}
	// Reply still goes out (soft-skip per the protocol's implementation guidance).
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT: got %d, want 1 (soft-skip still replies success)", got)
	}
	if got, want := amie.lastReplyType(), "inform_transaction_complete"; got != want {
		t.Errorf("reply type: got %q, want %q", got, want)
	}
}

// TestRequestUserModify_UnsupportedActionType: ActionType `add` has no
// defined semantics (only `replace` and `delete` are documented). The
// handler MUST refuse the packet and not mutate state. Sending "add"
// here also doubles as a guard against arbitrary action strings.
func TestRequestUserModify_UnsupportedActionType(t *testing.T) {
	database := setupTestDB(t)

	const userGlobalID = "amie-user-1005"
	userID, _, _ := seedUser(t, database,
		userGlobalID, "Stay", "Same", "user1005@example.edu",
		[]string{"/C=US/O=Org/CN=Stay"},
	)

	body := baseUserModifyBody("add", userGlobalID)
	body["UserFirstName"] = "Mutated"
	body["UserEmail"] = "mutated@example.edu"
	body["DnList"] = []any{"/C=US/O=Org/CN=New"}

	pkt := insertPacket(t, database, "request_user_modify", body)
	amie := &fakeAmieClient{}
	h := newRequestUserModifyHandlerForTest(database, amie)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error for unsupported ActionType, got nil")
	}
	if !strings.Contains(err.Error(), "ActionType") && !strings.Contains(err.Error(), "action") {
		t.Errorf("error should mention ActionType, got: %v", err)
	}

	// No state mutations.
	var user struct {
		FirstName string `db:"first_name"`
		Email     string `db:"email"`
	}
	if err := database.Get(&user, "SELECT first_name, email FROM users WHERE id = ?", userID); err != nil {
		t.Fatalf("read user: %v", err)
	}
	if user.FirstName != "Stay" || user.Email != "user1005@example.edu" {
		t.Errorf("user mutated on rejected action: got %+v", user)
	}
	if got := listUserDNs(t, database, userID); !equalStringSlices(got, []string{"/C=US/O=Org/CN=Stay"}) {
		t.Errorf("DN set mutated on rejected action: got %v", got)
	}
	if len(amie.Replies) != 0 {
		t.Errorf("amie replies on rejected action: got %d, want 0", len(amie.Replies))
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 0 {
		t.Errorf("audit REPLY_SENT on rejected action: got %d, want 0", got)
	}
}

// TestRequestUserModify_ReplyFailurePropagates, when the reply send fails
// the handler must surface the error so the processor can retry. The
// REPLY_SENT audit row must NOT exist (it is written only after a
// successful reply within the handler's tx; that tx rolls back on error).
// (Some core writes commit in their own txns; this test asserts only what
// the protocol reply contract requires.)
func TestRequestUserModify_ReplyFailurePropagates(t *testing.T) {
	database := setupTestDB(t)

	const (
		userGlobalID = "amie-user-1006"
		dn           = "/C=US/O=Org/CN=Reply Test"
	)
	seedUser(t, database,
		userGlobalID, "Reply", "Test", "user1006@example.edu",
		[]string{dn},
	)

	body := baseUserModifyBody("replace", userGlobalID)
	body["UserFirstName"] = "Reply"
	body["UserLastName"] = "Test"
	body["UserEmail"] = "user1006@example.edu"
	body["UserOrganization"] = "Org"
	body["UserOrgCode"] = "ORG"
	body["DnList"] = []any{dn}

	pkt := insertPacket(t, database, "request_user_modify", body)
	amie := &fakeAmieClient{FailWith: errors.New("simulated AMIE outage")}
	h := newRequestUserModifyHandlerForTest(database, amie)

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

// equalStringSlices reports whether two string slices contain the same
// elements in the same order. Tests sort their `want` before calling.
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
