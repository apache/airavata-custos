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
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

const (
	dpcExistingDN = "CN=Pat First Existing,O=Baseline Org,C=US"
	dpcNewDN1     = "CN=Pat First New 1,O=Baseline Org,C=US"
	dpcNewDN2     = "CN=Pat First New 2,O=Baseline Org,C=US"
)

// baseDataProjectCreateBody returns a fully populated data_project_create
// body. AMIE ships GlobalID and DnList in this packet; the handler reads
// GlobalID (not PersonID) to find the local user.
func baseDataProjectCreateBody() map[string]any {
	return map[string]any{
		"PersonID":  "site-pi-001",
		"ProjectID": "TG-BL-001",
		"GlobalID":  "bl-pi-001",
		"DnList": []any{
			dpcExistingDN, // duplicate of seeded row; INSERT IGNORE keeps one
			dpcNewDN1,
			dpcNewDN2,
		},
	}
}

// seedPIWithIdentity creates an Organization + User + UserIdentity (source
// "access", externalID = piGlobalID) so the handler's lookup succeeds.
// Returns the created user ID.
func seedPIWithIdentity(t *testing.T, database *sqlx.DB, piGlobalID string) string {
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
		FirstName:      "Pat",
		LastName:       "First",
		Email:          "pat.first@baseline.example.edu",
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if _, err := svc.CreateUserIdentity(ctx, &models.UserIdentity{
		UserID:     user.ID,
		Source:     "access",
		ExternalID: piGlobalID,
		Email:      "pat.first@baseline.example.edu",
	}); err != nil {
		t.Fatalf("seed user identity: %v", err)
	}
	return user.ID
}

// seedExistingDN inserts an amie_user_dns row outside any handler-managed tx
// so we can later assert that the handler preserved it.
func seedExistingDN(t *testing.T, database *sqlx.DB, userID, dn string) {
	t.Helper()
	tx, err := database.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx for DN seed: %v", err)
	}
	if err := store.NewUserDNStore(database).Add(
		context.Background(), tx, &model.UserDN{UserID: userID, DN: dn},
	); err != nil {
		_ = tx.Rollback()
		t.Fatalf("seed existing DN: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit DN seed: %v", err)
	}
}

// countDNsForUser is a convenience over the user_dn store for assertions.
func countDNsForUser(t *testing.T, database *sqlx.DB, userID string) int {
	t.Helper()
	dns, err := store.NewUserDNStore(database).ListByUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("list DNs: %v", err)
	}
	return len(dns)
}

// newDataProjectCreateHandlerForTest assembles the handler with real wiring
// (real userDN store, fakeAmieClient).
func newDataProjectCreateHandlerForTest(database *sqlx.DB, amie *fakeAmieClient) (*DataProjectCreateHandler, *coreservice.Service) {
	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	dnStore := store.NewUserDNStore(database)
	return NewDataProjectCreateHandler(svc, dnStore, amie, audit), svc
}

// TestDataProjectCreate_HappyPath verifies the happy path for a packet
// whose PI already exists locally (the normal flow: data_project_create
// follows notify_project_create, which created the PI). The handler MUST:
//   - persist each DN in DnList against the PI user (additive, no pruning)
//   - reply inform_transaction_complete
//   - audit PERSIST_DNS and REPLY_SENT
func TestDataProjectCreate_HappyPath(t *testing.T) {
	database := setupTestDB(t)

	body := baseDataProjectCreateBody()
	pkt := insertPacket(t, database, "data_project_create", body)

	userID := seedPIWithIdentity(t, database, "bl-pi-001")
	seedExistingDN(t, database, userID, dpcExistingDN)

	amie := &fakeAmieClient{}
	h, _ := newDataProjectCreateHandlerForTest(database, amie)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// Additive write: the existing DN must still be present; the two new DNs
	// must be added. Duplicate of existing DN is a silent no-op (INSERT IGNORE).
	if got, want := countDNsForUser(t, database, userID), 3; got != want {
		t.Errorf("amie_user_dns for PI: got %d, want %d", got, want)
	}
	if got := countRows(t, database, "amie_user_dns"); got != 3 {
		t.Errorf("amie_user_dns total: got %d, want 3 (1 existing + 2 new, dup ignored)", got)
	}

	// PERSIST_DNS audit row: handler emits exactly one summary row for the
	// batch (see data_project_create.go).
	if got := countAuditActions(t, database, pkt.ID, model.AuditPersistDNs); got != 1 {
		t.Errorf("audit PERSIST_DNS: got %d, want 1", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT: got %d, want 1", got)
	}

	// Reply contract: type must be inform_transaction_complete with the
	// required fields (DetailCode, Message, StatusCode).
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

// TestDataProjectCreate_UnknownUser asserts the handler's soft-skip path
// when no local user matches the packet's GlobalID. The reply MUST still
// be sent (the handler closes the transaction regardless), but no DN rows and
// no PERSIST_DNS audit row.
func TestDataProjectCreate_UnknownUser(t *testing.T) {
	database := setupTestDB(t)

	body := baseDataProjectCreateBody()
	body["GlobalID"] = "bl-pi-unknown"
	pkt := insertPacket(t, database, "data_project_create", body)

	amie := &fakeAmieClient{}
	h, _ := newDataProjectCreateHandlerForTest(database, amie)

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
	// Reply contract still holds: the handler closes the transaction.
	if got, want := amie.lastReplyType(), "inform_transaction_complete"; got != want {
		t.Errorf("reply type: got %q, want %q", got, want)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 1 {
		t.Errorf("audit REPLY_SENT on unknown user: got %d, want 1", got)
	}
}

// TestDataProjectCreate_EmptyDnList asserts that when the packet carries no
// DNs (DnList absent or empty), the handler skips PERSIST_DNS but still
// completes the transaction with an inform_transaction_complete reply.
// DnList is allowed but not required; missing/empty is valid.
func TestDataProjectCreate_EmptyDnList(t *testing.T) {
	database := setupTestDB(t)

	body := baseDataProjectCreateBody()
	delete(body, "DnList")
	pkt := insertPacket(t, database, "data_project_create", body)

	userID := seedPIWithIdentity(t, database, "bl-pi-001")
	seedExistingDN(t, database, userID, dpcExistingDN)

	amie := &fakeAmieClient{}
	h, _ := newDataProjectCreateHandlerForTest(database, amie)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	// "Additive, no pruning": the existing DN must NOT be touched when DnList
	// is empty.
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

// TestDataProjectCreate_ReplyFailurePropagates asserts that a failed reply
// surfaces as an error to the caller and that REPLY_SENT is NOT audited.
// The transaction does not complete until the handler successfully delivers
// inform_transaction_complete.
func TestDataProjectCreate_ReplyFailurePropagates(t *testing.T) {
	database := setupTestDB(t)

	body := baseDataProjectCreateBody()
	pkt := insertPacket(t, database, "data_project_create", body)

	_ = seedPIWithIdentity(t, database, "bl-pi-001")

	amie := &fakeAmieClient{FailWith: errors.New("simulated AMIE outage")}
	h, _ := newDataProjectCreateHandlerForTest(database, amie)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error when reply fails, got nil")
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 0 {
		t.Errorf("REPLY_SENT on failed reply: got %d, want 0", got)
	}
	// DN writes and PERSIST_DNS share the handler's tx, so on reply failure
	// the whole tx rolls back and neither row should exist.
	if got := countRows(t, database, "amie_user_dns"); got != 0 {
		t.Errorf("amie_user_dns after rollback: got %d, want 0", got)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditPersistDNs); got != 0 {
		t.Errorf("PERSIST_DNS after rollback: got %d, want 0", got)
	}
}
