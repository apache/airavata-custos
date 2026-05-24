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
	"testing"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
)

// baseInformTransactionCompleteBody returns a well-formed ITC body. Required
// fields per the protocol: DetailCode, Message, StatusCode.
func baseInformTransactionCompleteBody() map[string]any {
	return map[string]any{
		"StatusCode": "Success",
		"DetailCode": "1",
		"Message":    "Transaction completed successfully.",
	}
}

// TestInformTransactionComplete_HappyPath verifies the ITC contract:
//   - the handler records the arrival via a TRANSACTION_COMPLETE audit row
//   - NO reply is sent (ITC is terminal; the only packet type with no reply)
//   - no domain mutations
func TestInformTransactionComplete_HappyPath(t *testing.T) {
	database := setupTestDB(t)

	body := baseInformTransactionCompleteBody()
	pkt := insertPacket(t, database, "inform_transaction_complete", body)

	audit := newTestAuditService(database)
	h := NewInformTransactionCompleteHandler(audit)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if got := countAuditActions(t, database, pkt.ID, model.AuditTransactionComplete); got != 1 {
		t.Errorf("audit TRANSACTION_COMPLETE: got %d, want 1", got)
	}
	// No reply expected. The handler doesn't take an AmieClient, that alone
	// proves the design, but assert REPLY_SENT is absent too for regression
	// safety in case the handler shape changes later.
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 0 {
		t.Errorf("REPLY_SENT on terminal ITC: got %d, want 0 (no reply allowed)", got)
	}

	// No domain mutations, assert no domain rows were created.
	for _, table := range []string{
		"users", "user_identities", "organizations",
		"projects", "compute_allocations", "amie_user_dns",
	} {
		if got := countRows(t, database, table); got != 0 {
			t.Errorf("%s after ITC: got %d, want 0 (ITC has no domain writes)", table, got)
		}
	}
}

// TestInformTransactionComplete_MissingBody asserts that a packet with no
// "body" key is rejected by the handler (`getBody` returns an error). A
// well-formed ITC requires DetailCode/Message/StatusCode; a packet without
// a body cannot satisfy them.
func TestInformTransactionComplete_MissingBody(t *testing.T) {
	database := setupTestDB(t)

	// Persist a packet whose RawJSON has no "body" key. We hand-build the
	// packetJSON map the same way: missing "body".
	body := baseInformTransactionCompleteBody()
	pkt := insertPacket(t, database, "inform_transaction_complete", body)

	audit := newTestAuditService(database)
	h := NewInformTransactionCompleteHandler(audit)

	packetJSON := map[string]any{"type": pkt.Type}
	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, packetJSON, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error for missing body, got nil")
	}

	// No audit row should be written when body validation fails.
	if got := countAuditActions(t, database, pkt.ID, model.AuditTransactionComplete); got != 0 {
		t.Errorf("TRANSACTION_COMPLETE on missing body: got %d, want 0", got)
	}
}
