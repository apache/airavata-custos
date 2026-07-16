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

package store

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/pkg/models"
)

var (
	sharedTestDB     *sqlx.DB
	sharedTestDBOnce sync.Once
	sharedTestDBErr  error
)

// setupAccessTestDB opens (once per process) and migrates the test DB pointed
// to by CORE_TEST_DATABASE_DSN or DATABASE_DSN, then truncates the access
// tables.
func setupAccessTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	dsn := os.Getenv("CORE_TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_DSN")
	}
	if dsn == "" {
		t.Skip("integration env not set: CORE_TEST_DATABASE_DSN or DATABASE_DSN required")
	}
	sharedTestDBOnce.Do(func() {
		database, err := db.Open(db.Config{
			DSN:          dsn,
			MaxOpenConns: 5,
			MaxIdleConns: 2,
		})
		if err != nil {
			sharedTestDBErr = err
			return
		}
		if err := db.MigrateEmbedded(database); err != nil {
			sharedTestDBErr = err
			return
		}
		sharedTestDB = database
	})
	if sharedTestDBErr != nil {
		t.Fatalf("setup db: %v", sharedTestDBErr)
	}
	for _, tbl := range []string{"access_request_events", "access_requests", "access_events"} {
		if _, err := sharedTestDB.Exec("DELETE FROM " + tbl); err != nil {
			t.Fatalf("clear %s: %v", tbl, err)
		}
	}
	return sharedTestDB
}

// seedAccessEvent inserts the org/user/project/allocation chain an
// access_events row depends on and returns the event code.
func seedAccessEvent(t *testing.T, database *sqlx.DB) string {
	t.Helper()
	orgID, userID, projID, allocID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	code := "EVT-" + orgID[:8]
	stmts := []struct {
		q    string
		args []any
	}{
		{"INSERT INTO organizations (id, originated_id, name) VALUES (?, ?, ?)",
			[]any{orgID, "ORG-" + orgID[:8], "Access Test Org"}},
		{"INSERT INTO users (id, organization_id, first_name, last_name, middle_name, email, status) VALUES (?, ?, 'Test', 'PI', '', ?, 'ACTIVE')",
			[]any{userID, orgID, userID[:8] + "@example.invalid"}},
		{"INSERT INTO projects (id, originated_id, title, origination, project_pi_id, status) VALUES (?, ?, 'Access Test', 'TEST', ?, 'ACTIVE')",
			[]any{projID, "PRJ-" + projID[:8], userID}},
		{"INSERT INTO compute_clusters (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE name = name",
			[]any{"access-test-cluster", "access-test-cluster"}},
		{"INSERT INTO compute_allocations (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time) VALUES (?, ?, ?, 'ACTIVE', 'access-test-cluster', 100, NOW(6), NOW(6))",
			[]any{allocID, projID, "alloc-" + allocID[:8]}},
		{"INSERT INTO access_events (code, compute_allocation_id, organization_id) VALUES (?, ?, ?)",
			[]any{code, allocID, orgID}},
	}
	for _, s := range stmts {
		if _, err := database.Exec(s.q, s.args...); err != nil {
			t.Fatalf("seed: %v (%s)", err, s.q)
		}
	}
	return code
}

func inTestTx(t *testing.T, database *sqlx.DB, fn func(tx *sql.Tx) error) {
	t.Helper()
	tx, err := database.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		t.Fatalf("tx fn: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func newAccessRequest(sub, code string, ts time.Time) *models.AccessRequest {
	return &models.AccessRequest{
		ID:          uuid.NewString(),
		OIDCSub:     sub,
		Email:       sub + "@example.invalid",
		Name:        "Test Requester",
		Institution: "Test University",
		EventCode:   code,
		Reason:      "workshop",
		Status:      models.AccessRequestPending,
		Timestamp:   ts,
	}
}

func TestAccessEventStoreFindByCode(t *testing.T) {
	database := setupAccessTestDB(t)
	code := seedAccessEvent(t, database)
	s := NewAccessEventStore(database)
	ctx := context.Background()

	got, err := s.FindByCode(ctx, code)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got == nil || got.Code != code || got.ComputeAllocationID == "" || got.OrganizationID == "" {
		t.Fatalf("unexpected event: %+v", got)
	}

	missing, err := s.FindByCode(ctx, "NO-SUCH-CODE")
	if err != nil {
		t.Fatalf("find missing: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for unknown code, got %+v", missing)
	}
}

func TestAccessRequestStoreCreateAndGet(t *testing.T) {
	database := setupAccessTestDB(t)
	code := seedAccessEvent(t, database)
	s := NewAccessRequestStore(database)
	ctx := context.Background()

	r := newAccessRequest("sub-create", code, time.Now().UTC().Truncate(time.Microsecond))
	inTestTx(t, database, func(tx *sql.Tx) error { return s.Create(ctx, tx, r) })

	got, err := s.FindByID(ctx, r.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got == nil {
		t.Fatal("request not found")
	}
	if got.OIDCSub != r.OIDCSub || got.EventCode != code || got.Status != models.AccessRequestPending ||
		got.Reason != "workshop" || got.ApproverID != "" || got.DenyReason != "" ||
		got.ExpiresAt != nil || got.CreatedUserID != "" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}

	missing, err := s.FindByID(ctx, uuid.NewString())
	if err != nil {
		t.Fatalf("find missing: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for unknown id, got %+v", missing)
	}
}

func TestAccessRequestStoreFindLatestBySubPicksNewest(t *testing.T) {
	database := setupAccessTestDB(t)
	code := seedAccessEvent(t, database)
	s := NewAccessRequestStore(database)
	ctx := context.Background()

	base := time.Now().UTC().Truncate(time.Microsecond)
	old := newAccessRequest("sub-latest", code, base.Add(-time.Hour))
	old.Status = models.AccessRequestDenied
	newest := newAccessRequest("sub-latest", code, base)
	inTestTx(t, database, func(tx *sql.Tx) error {
		if err := s.Create(ctx, tx, old); err != nil {
			return err
		}
		return s.Create(ctx, tx, newest)
	})

	got, err := s.FindLatestBySub(ctx, "sub-latest")
	if err != nil {
		t.Fatalf("latest: %v", err)
	}
	if got == nil || got.ID != newest.ID {
		t.Fatalf("latest = %+v, want id %s", got, newest.ID)
	}

	none, err := s.FindLatestBySub(ctx, "sub-unknown")
	if err != nil {
		t.Fatalf("latest unknown: %v", err)
	}
	if none != nil {
		t.Fatalf("expected nil for unknown sub, got %+v", none)
	}
}

func TestAccessRequestStoreListFilters(t *testing.T) {
	database := setupAccessTestDB(t)
	codeA := seedAccessEvent(t, database)
	codeB := seedAccessEvent(t, database)
	s := NewAccessRequestStore(database)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	pendingA := newAccessRequest("sub-a", codeA, now)
	deniedA := newAccessRequest("sub-b", codeA, now)
	deniedA.Status = models.AccessRequestDenied
	pendingB := newAccessRequest("sub-c", codeB, now)
	inTestTx(t, database, func(tx *sql.Tx) error {
		for _, r := range []*models.AccessRequest{pendingA, deniedA, pendingB} {
			if err := s.Create(ctx, tx, r); err != nil {
				return err
			}
		}
		return nil
	})

	all, err := s.List(ctx, AccessRequestListFilter{})
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("list all = %d, want 3", len(all))
	}

	pending, err := s.List(ctx, AccessRequestListFilter{Status: string(models.AccessRequestPending)})
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("list pending = %d, want 2", len(pending))
	}

	byEvent, err := s.List(ctx, AccessRequestListFilter{EventCode: codeB})
	if err != nil {
		t.Fatalf("list by event: %v", err)
	}
	if len(byEvent) != 1 || byEvent[0].ID != pendingB.ID {
		t.Fatalf("list by event = %+v, want only %s", byEvent, pendingB.ID)
	}

	both, err := s.List(ctx, AccessRequestListFilter{Status: string(models.AccessRequestDenied), EventCode: codeA})
	if err != nil {
		t.Fatalf("list both: %v", err)
	}
	if len(both) != 1 || both[0].ID != deniedA.ID {
		t.Fatalf("list both = %+v, want only %s", both, deniedA.ID)
	}
}

func TestAccessRequestStoreHasPendingBySub(t *testing.T) {
	database := setupAccessTestDB(t)
	code := seedAccessEvent(t, database)
	s := NewAccessRequestStore(database)
	ctx := context.Background()

	r := newAccessRequest("sub-pending", code, time.Now().UTC())
	inTestTx(t, database, func(tx *sql.Tx) error { return s.Create(ctx, tx, r) })

	has, err := s.HasPendingBySub(ctx, "sub-pending")
	if err != nil {
		t.Fatalf("has pending: %v", err)
	}
	if !has {
		t.Fatal("expected pending request")
	}

	r.Status = models.AccessRequestDenied
	r.DenyReason = "not eligible"
	inTestTx(t, database, func(tx *sql.Tx) error { return s.Update(ctx, tx, r) })

	has, err = s.HasPendingBySub(ctx, "sub-pending")
	if err != nil {
		t.Fatalf("has pending after deny: %v", err)
	}
	if has {
		t.Fatal("expected no pending request after deny")
	}

	got, err := s.FindByID(ctx, r.ID)
	if err != nil {
		t.Fatalf("find after update: %v", err)
	}
	if got.Status != models.AccessRequestDenied || got.DenyReason != "not eligible" {
		t.Fatalf("update did not stick: %+v", got)
	}
}

func TestAccessRequestEventStoreAppendAndGet(t *testing.T) {
	database := setupAccessTestDB(t)
	code := seedAccessEvent(t, database)
	reqStore := NewAccessRequestStore(database)
	evStore := NewAccessRequestEventStore(database)
	ctx := context.Background()

	r := newAccessRequest("sub-events", code, time.Now().UTC())
	base := time.Now().UTC().Truncate(time.Microsecond)
	created := &models.AccessRequestEvent{
		ID:              uuid.NewString(),
		AccessRequestID: r.ID,
		EventType:       models.AccessRequestEventCreated,
		Description:     "request created",
		Timestamp:       base,
	}
	approved := &models.AccessRequestEvent{
		ID:              uuid.NewString(),
		AccessRequestID: r.ID,
		EventType:       models.AccessRequestEventApproved,
		Timestamp:       base.Add(time.Minute),
	}
	inTestTx(t, database, func(tx *sql.Tx) error {
		if err := reqStore.Create(ctx, tx, r); err != nil {
			return err
		}
		if err := evStore.Create(ctx, tx, created); err != nil {
			return err
		}
		return evStore.Create(ctx, tx, approved)
	})

	got, err := evStore.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find event: %v", err)
	}
	if got == nil || got.EventType != models.AccessRequestEventCreated || got.Description != "request created" {
		t.Fatalf("event round-trip mismatch: %+v", got)
	}

	events, err := evStore.FindByRequest(ctx, r.ID)
	if err != nil {
		t.Fatalf("find by request: %v", err)
	}
	if len(events) != 2 || events[0].ID != created.ID || events[1].ID != approved.ID {
		t.Fatalf("events = %+v, want [created, approved]", events)
	}
}
