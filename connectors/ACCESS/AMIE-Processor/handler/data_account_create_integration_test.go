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

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/pkg/models"
)

const (
	dacExistingDN = "CN=User One Existing,O=Baseline Org,C=US"
	dacNewDN1     = "CN=User One New 1,O=Baseline Org,C=US"
	dacNewDN2     = "CN=User One New 2,O=Baseline Org,C=US"
)

// baseDataAccountCreateBody returns a fully populated data_account_create
// body. AMIE ships GlobalID and DnList in this packet, the handler reads
// GlobalID (not PersonID) to find the local user.
func baseDataAccountCreateBody() map[string]any {
	return map[string]any{
		"PersonID":  "site-user-001",
		"ProjectID": "TG-BL-001",
		"GlobalID":  "bl-user-001",
		"DnList": []any{
			dacExistingDN, // duplicate; INSERT IGNORE keeps the existing row
			dacNewDN1,
			dacNewDN2,
		},
	}
}

// seedUserWithIdentity creates an Organization + User + UserIdentity (source
// "access", externalID = userGlobalID) so the handler's lookup succeeds.
// Returns the created user ID.
func seedUserWithIdentity(t *testing.T, database *sqlx.DB, userGlobalID string) string {
	t.Helper()
	svc := newTestCoreService(database)
	ctx := context.Background()

	org, err := svc.CreateOrganization(ctx, &models.Organization{
		OriginatedID: "BASELINE",
		Name:         "Baseline Org",
	})
	if err != nil {
		t.Fatalf("seed organization: %v", err)
	}
	user, err := svc.CreateUser(ctx, &models.User{
		OrganizationID: org.ID,
		FirstName:      "User",
		LastName:       "One",
		Email:          "user.one@baseline.example.edu",
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if _, err := svc.CreateUserIdentity(ctx, &models.UserIdentity{
		UserID:     user.ID,
		Source:     "access",
		ExternalID: userGlobalID,
		Email:      "user.one@baseline.example.edu",
	}); err != nil {
		t.Fatalf("seed user identity: %v", err)
	}
	return user.ID
}

// newDataAccountCreateHandlerForTest assembles the handler with real wiring
// (real userDN store, fakeAmieClient).
func newDataAccountCreateHandlerForTest(database *sqlx.DB, amie *fakeAmieClient) *DataAccountCreateHandler {
	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	dnStore := store.NewUserDNStore(database)
	return NewDataAccountCreateHandler(svc, dnStore, amie, audit)
}

// TestDataAccountCreate_HappyPath verifies the happy path for a packet whose
// user already exists locally (the normal flow: data_account_create follows
// notify_account_create, which created the user).
//   - persist each DN in DnList against the user (additive, no pruning)
//   - reply inform_transaction_complete
//   - audit PERSIST_DNS and REPLY_SENT
func TestDataAccountCreate_HappyPath(t *testing.T) {
	database := setupTestDB(t)

	body := baseDataAccountCreateBody()
	pkt := insertPacket(t, database, "data_account_create", body)

	userID := seedUserWithIdentity(t, database, "bl-user-001")
	seedExistingDN(t, database, userID, dacExistingDN)

	amie := &fakeAmieClient{}
	h := newDataAccountCreateHandlerForTest(database, amie)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// Additive: existing DN preserved, two new DNs added, duplicate ignored.
	if got, want := countDNsForUser(t, database, userID), 3; got != want {
		t.Errorf("amie_user_dns for user: got %d, want %d", got, want)
	}
	if got := countRows(t, database, "amie_user_dns"); got != 3 {
		t.Errorf("amie_user_dns total: got %d, want 3 (1 existing + 2 new, dup ignored)", got)
	}

	if got := countAuditActions(t, database, pkt.ID, model.AuditPersistDNs); got != 1 {
		t.Errorf("audit PERSIST_DNS: got %d, want 1", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT: got %d, want 1", got)
	}

	// Reply contract: type must be inform_transaction_complete with the
	// required body fields (DetailCode, Message, StatusCode).
	if got, want := amie.lastReplyType(), "inform_transaction_complete"; got != want {
		t.Fatalf("reply type: got %q, want %q", got, want)
	}
	reply := amie.lastReplyBody()
	if reply == nil {
		t.Fatal("reply body missing")
	}
	if got, _ := reply["StatusCode"].(string); got != "Success" {
		t.Errorf("reply.StatusCode: got %q, want Success", got)
	}
	if _, ok := reply["DetailCode"]; !ok {
		t.Errorf("reply.DetailCode: missing; required by reply contract")
	}
	if _, ok := reply["Message"]; !ok {
		t.Errorf("reply.Message: missing; required by reply contract")
	}
}

// TestDataAccountCreate_UnknownUser asserts the handler's soft-skip path
// when no local user matches the packet's GlobalID.
func TestDataAccountCreate_UnknownUser(t *testing.T) {
	database := setupTestDB(t)

	body := baseDataAccountCreateBody()
	body["GlobalID"] = "bl-user-unknown"
	pkt := insertPacket(t, database, "data_account_create", body)

	amie := &fakeAmieClient{}
	h := newDataAccountCreateHandlerForTest(database, amie)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if got := countRows(t, database, "amie_user_dns"); got != 0 {
		t.Errorf("amie_user_dns on unknown user: got %d, want 0", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditPersistDNs); got != 0 {
		t.Errorf("PERSIST_DNS on unknown user: got %d, want 0", got)
	}
	if got, want := amie.lastReplyType(), "inform_transaction_complete"; got != want {
		t.Errorf("reply type: got %q, want %q", got, want)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT on unknown user: got %d, want 1", got)
	}
}

// TestDataAccountCreate_EmptyDnList asserts that when the packet carries no
// DNs, the handler skips PERSIST_DNS but still completes the transaction
// with an inform_transaction_complete reply. DnList is allowed but not
// required.
func TestDataAccountCreate_EmptyDnList(t *testing.T) {
	database := setupTestDB(t)

	body := baseDataAccountCreateBody()
	delete(body, "DnList")
	pkt := insertPacket(t, database, "data_account_create", body)

	userID := seedUserWithIdentity(t, database, "bl-user-001")
	seedExistingDN(t, database, userID, dacExistingDN)

	amie := &fakeAmieClient{}
	h := newDataAccountCreateHandlerForTest(database, amie)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if got, want := countDNsForUser(t, database, userID), 1; got != want {
		t.Errorf("amie_user_dns: got %d, want %d (existing row preserved)", got, want)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditPersistDNs); got != 0 {
		t.Errorf("PERSIST_DNS on empty DnList: got %d, want 0", got)
	}
	if got, want := amie.lastReplyType(), "inform_transaction_complete"; got != want {
		t.Errorf("reply type: got %q, want %q", got, want)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT: got %d, want 1", got)
	}
}

// TestDataAccountCreate_ReplyFailurePropagates asserts that a failed reply
// surfaces as an error to the caller and that REPLY_SENT is NOT audited.
// The transaction does not complete until the handler successfully delivers
// inform_transaction_complete.
func TestDataAccountCreate_ReplyFailurePropagates(t *testing.T) {
	database := setupTestDB(t)

	body := baseDataAccountCreateBody()
	pkt := insertPacket(t, database, "data_account_create", body)

	_ = seedUserWithIdentity(t, database, "bl-user-001")

	amie := &fakeAmieClient{FailWith: errors.New("simulated AMIE outage")}
	h := newDataAccountCreateHandlerForTest(database, amie)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error when reply fails, got nil")
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 0 {
		t.Errorf("REPLY_SENT on failed reply: got %d, want 0", got)
	}
	// See B11: DN writes + PERSIST_DNS share the handler's tx, so on reply
	// failure the tx rolls back.
	if got := countRows(t, database, "amie_user_dns"); got != 0 {
		t.Errorf("amie_user_dns after rollback: got %d, want 0", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditPersistDNs); got != 0 {
		t.Errorf("PERSIST_DNS after rollback: got %d, want 0", got)
	}
}
