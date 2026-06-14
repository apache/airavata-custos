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
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	amiedb "github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/db"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	amieservice "github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/service"
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/internal/db"
	corestore "github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/events"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

const testClusterID = "00000000-0000-0000-0000-000000000001"

var (
	sharedDB     *sqlx.DB
	sharedDBOnce sync.Once
	sharedDBErr  error
)

func isLocalAMIEConfigAvailable() bool {
	for _, key := range []string{
		"DATABASE_DSN",
		"AMIE_BASE_URL",
		"AMIE_SITE_CODE",
		"AMIE_API_KEY",
		"AMIE_CLUSTER_ID",
	} {
		if os.Getenv(key) == "" {
			return false
		}
	}
	return true
}

// setupTestDB opens the DB once per process; subsequent calls re-truncate
// and re-seed but reuse the same connection.
func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	if !isLocalAMIEConfigAvailable() {
		t.Skip("integration env not set; run via scripts/run-amie-integration-tests.sh")
	}
	sharedDBOnce.Do(func() {
		database, err := db.Open(db.Config{
			DSN:          os.Getenv("DATABASE_DSN"),
			MaxOpenConns: 5,
			MaxIdleConns: 2,
		})
		if err != nil {
			sharedDBErr = err
			return
		}
		if err := db.MigrateEmbedded(database); err != nil {
			sharedDBErr = err
			return
		}
		if err := db.MigrateConnectorFS(database, amiedb.MigrationFS(), "migrations", "amie"); err != nil {
			sharedDBErr = err
			return
		}
		sharedDB = database
	})
	if sharedDBErr != nil {
		t.Fatalf("setup db: %v", sharedDBErr)
	}
	truncateAll(t, sharedDB)
	seedDefaultCluster(t, sharedDB)
	return sharedDB
}

// truncateAll disables FK checks so we don't have to order the table list.
func truncateAll(t *testing.T, database *sqlx.DB) {
	t.Helper()
	tables := []string{
		"amie_audit_extras",
		"audit_events",
		"amie_processing_errors",
		"amie_processing_events",
		"amie_packets",
		"amie_user_dns",
		"compute_allocation_membership_resource_overrides",
		"compute_allocation_usages",
		"compute_allocation_memberships",
		"compute_allocation_change_request_events",
		"compute_allocation_change_requests",
		"compute_allocation_diffs",
		"compute_allocation_resource_rates",
		"compute_allocation_resource_mappings",
		"compute_allocation_resources",
		"compute_allocations",
		"compute_cluster_users",
		"compute_clusters",
		"projects",
		"user_identities",
		"users",
		"organizations",
	}
	if _, err := database.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		t.Fatalf("disable FK: %v", err)
	}
	for _, tbl := range tables {
		if _, err := database.Exec("TRUNCATE TABLE " + tbl); err != nil {
			t.Fatalf("truncate %s: %v", tbl, err)
		}
	}
	if _, err := database.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		t.Fatalf("re-enable FK: %v", err)
	}
}

// seedDefaultCluster inserts the row AMIE_CLUSTER_ID points at. Handlers
// refuse to run without it, and FKs on compute_allocations /
// compute_cluster_users require it.
func seedDefaultCluster(t *testing.T, database *sqlx.DB) {
	t.Helper()
	if _, err := database.Exec(
		"INSERT INTO compute_clusters (id, name) VALUES (?, ?)",
		testClusterID, "default-cluster",
	); err != nil {
		t.Fatalf("seed cluster: %v", err)
	}
}

func newTestCoreService(database *sqlx.DB) *coreservice.Service {
	return coreservice.New(database, events.New())
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

func countRows(t *testing.T, database *sqlx.DB, table string) int {
	t.Helper()
	var n int
	if err := database.Get(&n, "SELECT COUNT(*) FROM "+table); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func countAuditActions(t *testing.T, database *sqlx.DB, packetID string, action model.AuditAction) int {
	t.Helper()
	var n int
	if err := database.Get(&n,
		`SELECT COUNT(*) FROM audit_events ae
		 JOIN amie_audit_extras x ON x.audit_event_id = ae.id
		 WHERE x.packet_id = ? AND ae.event_type = ?`,
		packetID, string(action),
	); err != nil {
		t.Fatalf("count audit %s: %v", action, err)
	}
	return n
}
