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
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/internal/tracing/tracingtest"
	"github.com/apache/airavata-custos/pkg/models"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

// holdAccountReply runs request_account_create for a new person and saves the held packet, the way the processor would.
func holdAccountReply(t *testing.T, database *sqlx.DB, svc *coreservice.Service) *model.Packet {
	t.Helper()
	body := baseRACBody()
	body["ProjectID"] = seedProjectForRAC(t, database)
	pkt := insertPacket(t, database, "request_account_create", body)
	h := NewRequestAccountCreateHandler(svc, testClusterID, &fakeAmieClient{}, newTestAuditService(database))
	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		if err := h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, ""); err != nil {
			return err
		}
		return store.NewPacketStore(database).Update(ctx, tx, pkt)
	}); err != nil {
		t.Fatalf("hold reply: %v", err)
	}
	return pkt
}

func reviewedAccount(t *testing.T, svc *coreservice.Service, id string) models.ComputeClusterUser {
	t.Helper()
	cu, err := svc.GetComputeClusterUser(context.Background(), id)
	if err != nil {
		t.Fatalf("get cluster account: %v", err)
	}
	return *cu
}

func packetStatus(t *testing.T, database *sqlx.DB, id string) model.PacketStatus {
	t.Helper()
	var status model.PacketStatus
	if err := database.Get(&status, "SELECT status FROM amie_packets WHERE id = $1", id); err != nil {
		t.Fatalf("read packet status: %v", err)
	}
	return status
}

// Make sure approving the cluster account sends the held reply to ACCESS once, even
// when the bus delivers the approve event a second time.
func TestHeldReplies_ApprovalSendsHeldReplyOnce(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestCoreService(database)
	pkt := holdAccountReply(t, database, svc)
	approvePendingAccounts(t, database, svc)

	amie := &fakeAmieClient{}
	held := NewHeldReplies(database, store.NewPacketStore(database), amie, newTestAuditService(database))
	account := reviewedAccount(t, svc, *pkt.HeldFor)
	ctx, delivery := tracing.Start(context.Background(), "delivery")
	for range 2 {
		if err := held.approved(ctx, account); err != nil {
			t.Fatalf("approved: %v", err)
		}
	}
	var chained int
	if err := database.Get(&chained, "SELECT COUNT(*) FROM audit_events WHERE event_type = 'REPLY_SENT' AND parent_span_id = $1", delivery.SpanContext().SpanID().String()); err != nil || chained != 1 {
		t.Errorf("REPLY_SENT rows under the delivery span: %d (%v), want 1", chained, err)
	}

	if len(amie.Replies) != 1 || amie.lastReplyType() != "notify_account_create" {
		t.Fatalf("expected one notify_account_create, got %+v", amie.Replies)
	}
	if got := packetStatus(t, database, pkt.ID); got != model.PacketStatusDecoded {
		t.Errorf("packet status: got %s, want %s", got, model.PacketStatusDecoded)
	}
	var stillHeld int
	if err := database.Get(&stillHeld, "SELECT COUNT(*) FROM amie_packets WHERE id = $1 AND (held_reply IS NOT NULL OR held_for IS NOT NULL)", pkt.ID); err != nil {
		t.Fatalf("read held fields: %v", err)
	}
	if stillHeld != 0 {
		t.Errorf("expected the held reply to be cleared once it is sent")
	}
}

// Make sure a held packet that is processed again after the account is
// approved gets its reply and no longer carries the old held reply. This is
// what happens to a held packet after a person merge.
func TestHeldReplies_ProcessedAgainAfterApprovalClearsHeldReply(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestCoreService(database)
	pkt := holdAccountReply(t, database, svc)
	approvePendingAccounts(t, database, svc)

	var packetJSON map[string]any
	if err := json.Unmarshal([]byte(pkt.RawJSON), &packetJSON); err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	amie := &fakeAmieClient{}
	h := NewRequestAccountCreateHandler(svc, testClusterID, amie, newTestAuditService(database))
	pkt.Status = model.PacketStatusDecoded // the processor resets it before the handler runs
	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, packetJSON, pkt, "")
	}); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if amie.lastReplyType() != "notify_account_create" {
		t.Fatalf("reply type: got %q, want notify_account_create", amie.lastReplyType())
	}
	if pkt.HeldReply != nil || pkt.HeldFor != nil {
		t.Errorf("expected the old held reply to be cleared, got held_for %v", pkt.HeldFor)
	}
}

// Make sure denying the cluster account ends the transaction on the ACCESS side with a
// failure that carries the admin's reason.
func TestHeldReplies_DenialSendsFailureWithReason(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestCoreService(database)
	pkt := holdAccountReply(t, database, svc)
	var reviewer string
	if err := database.Get(&reviewer, "SELECT id FROM users LIMIT 1"); err != nil {
		t.Fatalf("pick a reviewer: %v", err)
	}
	if _, err := svc.DenyComputeClusterUser(tracingtest.Context(), *pkt.HeldFor, reviewer, "not on the collaborator list"); err != nil {
		t.Fatalf("deny: %v", err)
	}

	amie := &fakeAmieClient{}
	held := NewHeldReplies(database, store.NewPacketStore(database), amie, newTestAuditService(database))
	if err := held.denied(context.Background(), reviewedAccount(t, svc, *pkt.HeldFor)); err != nil {
		t.Fatalf("denied: %v", err)
	}

	if amie.lastReplyType() != "inform_transaction_complete" {
		t.Fatalf("reply type: got %q, want inform_transaction_complete", amie.lastReplyType())
	}
	body := amie.lastReplyBody()
	if body["StatusCode"] != "Failure" || body["Message"] != "not on the collaborator list" {
		t.Errorf("expected a failure with the reason, got %v", body)
	}
	if got := packetStatus(t, database, pkt.ID); got != model.PacketStatusRefused {
		t.Errorf("packet status: got %s, want %s", got, model.PacketStatusRefused)
	}
}

// Make sure a failed send leaves the packet waiting and returns the error, so
// the bus retries the delivery.
func TestHeldReplies_FailedSendKeepsPacketWaiting(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestCoreService(database)
	pkt := holdAccountReply(t, database, svc)
	approvePendingAccounts(t, database, svc)

	amie := &fakeAmieClient{FailWith: errors.New("simulated AMIE outage")}
	held := NewHeldReplies(database, store.NewPacketStore(database), amie, newTestAuditService(database))
	if err := held.approved(context.Background(), reviewedAccount(t, svc, *pkt.HeldFor)); err == nil {
		t.Fatal("expected an error when the reply fails")
	}
	if got := packetStatus(t, database, pkt.ID); got != model.PacketStatusWaitingApproval {
		t.Errorf("packet status: got %s, want %s", got, model.PacketStatusWaitingApproval)
	}
}
