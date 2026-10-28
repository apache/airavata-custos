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
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/amietest"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	amieservice "github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/service"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/internal/db/dbtest"
	corestore "github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/internal/tracing/tracingtest"
	"github.com/apache/airavata-custos/pkg/events"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

const testClusterID = amietest.ClusterID

func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	database := amietest.OpenDB(t, amietest.Tables...)
	amietest.SeedCluster(t, database)
	return database
}

func newTestCoreService(database *sqlx.DB) *coreservice.Service {
	bus, err := events.New(context.Background(), database)
	if err != nil {
		panic("event bus: " + err.Error())
	}
	return coreservice.New(database, bus)
}

func newTestAuditService(database *sqlx.DB) *amieservice.AuditService {
	return amieservice.NewAuditService(corestore.NewAuditEventStore(database), store.NewAuditExtrasStore(database))
}

type fakeReply struct {
	PacketRecID int64
	Body        map[string]any
}

// fakeAmieClient captures replies in-process. Set FailWith to simulate a
// failed reply send.
type fakeAmieClient struct {
	mu       sync.Mutex
	Replies  []fakeReply
	FailWith error
}

func (f *fakeAmieClient) ReplyToPacket(_ context.Context, packetRecID int64, reply map[string]any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.FailWith != nil {
		return f.FailWith
	}
	f.Replies = append(f.Replies, fakeReply{PacketRecID: packetRecID, Body: reply})
	return nil
}

func (f *fakeAmieClient) lastReplyBody() map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Replies) == 0 {
		return nil
	}
	last := f.Replies[len(f.Replies)-1]
	body, _ := last.Body["body"].(map[string]any)
	return body
}

func (f *fakeAmieClient) lastReplyType() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.Replies) == 0 {
		return ""
	}
	last := f.Replies[len(f.Replies)-1]
	t, _ := last.Body["type"].(string)
	return t
}

// heldReply returns the type and body of the reply a handler kept on the packet.
func heldReply(t *testing.T, pkt *model.Packet) (string, map[string]any) {
	t.Helper()
	if pkt.Status != model.PacketStatusWaitingApproval || pkt.HeldReply == nil {
		t.Fatalf("expected the packet to hold its reply, got status %s", pkt.Status)
	}
	var reply map[string]any
	if err := json.Unmarshal([]byte(*pkt.HeldReply), &reply); err != nil {
		t.Fatalf("decode held reply: %v", err)
	}
	typ, _ := reply["type"].(string)
	body, _ := reply["body"].(map[string]any)
	return typ, body
}

// approvePendingAccounts approves every pending cluster account, as an admin would.
func approvePendingAccounts(t *testing.T, database *sqlx.DB, svc *coreservice.Service) {
	t.Helper()
	var ids []string
	if err := database.Select(&ids, "SELECT id FROM compute_cluster_users WHERE approval_status = 'PENDING'"); err != nil {
		t.Fatalf("list pending cluster accounts: %v", err)
	}
	var reviewer string
	if err := database.Get(&reviewer, "SELECT id FROM users LIMIT 1"); err != nil {
		t.Fatalf("pick a reviewer: %v", err)
	}
	for _, id := range ids {
		if _, err := svc.ApproveComputeClusterUser(tracingtest.Context(), id, reviewer); err != nil {
			t.Fatalf("approve %s: %v", id, err)
		}
	}
}

func insertPacket(t *testing.T, database *sqlx.DB, packetType string, body map[string]any) *model.Packet {
	t.Helper()
	raw, err := json.Marshal(map[string]any{"type": packetType, "body": body})
	if err != nil {
		t.Fatalf("marshal packet: %v", err)
	}
	pkt := &model.Packet{
		ID:         uuid.NewString(),
		AmieID:     time.Now().UnixNano(),
		Type:       packetType,
		Status:     model.PacketStatusNew,
		RawJSON:    string(raw),
		ReceivedAt: time.Now().UTC(),
	}
	tx, err := database.BeginTxx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := store.NewPacketStore(database).Save(context.Background(), tx.Tx, pkt); err != nil {
		_ = tx.Rollback()
		t.Fatalf("save packet: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit packet: %v", err)
	}
	return pkt
}

func runHandlerInTx(t *testing.T, database *sqlx.DB, fn func(ctx context.Context, tx *sql.Tx) error) error {
	t.Helper()
	ctx := context.Background()
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

var countRows = dbtest.CountRows

func countAuditActions(t *testing.T, database *sqlx.DB, packetID string, action model.AuditAction) int {
	t.Helper()
	var n int
	if err := database.Get(&n,
		`SELECT COUNT(*) FROM audit_events ae
		 JOIN amie_audit_extras x ON x.audit_event_id = ae.id
		 WHERE x.packet_id = $1 AND ae.event_type = $2`,
		packetID, string(action),
	); err != nil {
		t.Fatalf("count audit %s: %v", action, err)
	}
	return n
}
