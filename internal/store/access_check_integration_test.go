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
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

func setupAccessCheckTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	database := setupAccessTestDB(t)
	for _, tbl := range []string{"access_check_events", "access_checks"} {
		if _, err := database.Exec("DELETE FROM " + tbl); err != nil {
			t.Fatalf("clear %s: %v", tbl, err)
		}
	}
	return database
}

// seedCheckTarget creates the org/user/project/cluster/allocation rows the
// access_checks foreign keys need, returning (allocationID, userID).
func seedCheckTarget(t *testing.T, database *sqlx.DB) (string, string) {
	t.Helper()
	ctx := context.Background()
	orgID, userID := uuid.NewString(), uuid.NewString()
	projectID, clusterID, allocID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	mustExec := func(q string, args ...any) {
		t.Helper()
		if _, err := database.ExecContext(ctx, q, args...); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	mustExec(`INSERT INTO organizations (id, originated_id, name) VALUES (?, ?, ?)`,
		orgID, "orig-"+orgID[:8], "org-"+orgID[:8])
	mustExec(`INSERT INTO users (id, organization_id, first_name, last_name, email, status)
	          VALUES (?, ?, 'Check', 'User', ?, 'ACTIVE')`, userID, orgID, userID[:8]+"@example.edu")
	mustExec(`INSERT INTO compute_clusters (id, name) VALUES (?, ?)`, clusterID, "cl-"+clusterID[:8])
	mustExec(`INSERT INTO projects (id, originated_id, title, origination, project_pi_id, status)
	          VALUES (?, ?, ?, 'TEST', ?, 'ACTIVE')`, projectID, "orig-"+projectID[:8], "proj-"+projectID[:8], userID)
	mustExec(`INSERT INTO compute_allocations
	              (id, project_id, name, status, compute_cluster_id, initial_su_amount, start_time, end_time)
	          VALUES (?, ?, ?, 'ACTIVE', ?, 1000, NOW(6), DATE_ADD(NOW(6), INTERVAL 30 DAY))`,
		allocID, projectID, "alloc-"+allocID[:8], clusterID)
	return allocID, userID
}

func createCheck(t *testing.T, database *sqlx.DB, s AccessCheckStore, c *models.AccessCheck) {
	t.Helper()
	tx, err := database.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := s.Create(context.Background(), tx, c); err != nil {
		tx.Rollback()
		t.Fatalf("create check: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func TestAccessCheckStore_CreateFindUpdate(t *testing.T) {
	database := setupAccessCheckTestDB(t)
	s := NewAccessCheckStore(database)
	ctx := context.Background()
	allocID, userID := seedCheckTarget(t, database)

	now := time.Now().UTC().Truncate(time.Microsecond)
	c := &models.AccessCheck{
		ID: uuid.NewString(), ComputeAllocationID: allocID, UserID: userID,
		CheckType: models.AccessCheckJobSubmission, Status: models.AccessCheckPending,
		LastCheckedAt: now,
	}
	createCheck(t, database, s, c)

	got, err := s.FindByTarget(ctx, allocID, userID, models.AccessCheckJobSubmission)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got == nil || got.Status != models.AccessCheckPending || got.FailingSince != nil {
		t.Fatalf("find: got %+v, want PENDING with nil failing_since", got)
	}

	okAt := now.Add(time.Minute)
	got.Status = models.AccessCheckOK
	got.LastCheckedAt = okAt
	got.LastOKAt = &okAt
	tx, _ := database.Begin()
	if err := s.Update(ctx, tx, got); err != nil {
		t.Fatalf("update: %v", err)
	}
	tx.Commit()

	got2, err := s.FindByTarget(ctx, allocID, userID, models.AccessCheckJobSubmission)
	if err != nil || got2 == nil {
		t.Fatalf("refind: %v", err)
	}
	if got2.Status != models.AccessCheckOK || got2.LastOKAt == nil {
		t.Fatalf("after update: got %+v, want OK with last_ok_at", got2)
	}

	missing, err := s.FindByTarget(ctx, allocID, userID, models.AccessCheckSignIn)
	if err != nil || missing != nil {
		t.Fatalf("missing check type: got %+v err %v, want nil,nil", missing, err)
	}
}

func TestAccessCheckStore_UniqueTarget(t *testing.T) {
	database := setupAccessCheckTestDB(t)
	s := NewAccessCheckStore(database)
	allocID, userID := seedCheckTarget(t, database)

	now := time.Now().UTC()
	first := &models.AccessCheck{
		ID: uuid.NewString(), ComputeAllocationID: allocID, UserID: userID,
		CheckType: models.AccessCheckSignIn, Status: models.AccessCheckOK, LastCheckedAt: now,
	}
	createCheck(t, database, s, first)

	dup := &models.AccessCheck{
		ID: uuid.NewString(), ComputeAllocationID: allocID, UserID: userID,
		CheckType: models.AccessCheckSignIn, Status: models.AccessCheckOK, LastCheckedAt: now,
	}
	tx, _ := database.Begin()
	err := s.Create(context.Background(), tx, dup)
	tx.Rollback()
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "duplicate") {
		t.Fatalf("duplicate target: got %v, want duplicate key error", err)
	}
}

func TestAccessCheckStore_EventsNewestFirstAcrossChecks(t *testing.T) {
	database := setupAccessCheckTestDB(t)
	s := NewAccessCheckStore(database)
	ctx := context.Background()
	allocID, userID := seedCheckTarget(t, database)

	base := time.Now().UTC().Truncate(time.Microsecond)
	signIn := &models.AccessCheck{
		ID: uuid.NewString(), ComputeAllocationID: allocID, UserID: userID,
		CheckType: models.AccessCheckSignIn, Status: models.AccessCheckOK, LastCheckedAt: base,
	}
	jobs := &models.AccessCheck{
		ID: uuid.NewString(), ComputeAllocationID: allocID, UserID: userID,
		CheckType: models.AccessCheckJobSubmission, Status: models.AccessCheckOK, LastCheckedAt: base,
	}
	createCheck(t, database, s, signIn)
	createCheck(t, database, s, jobs)

	events := []models.AccessCheckEvent{
		{ID: uuid.NewString(), AccessCheckID: signIn.ID, EventType: models.AccessCheckEventStarted, OccurredAt: base},
		{ID: uuid.NewString(), AccessCheckID: signIn.ID, EventType: models.AccessCheckEventOnline, OccurredAt: base.Add(time.Minute)},
		{ID: uuid.NewString(), AccessCheckID: jobs.ID, EventType: models.AccessCheckEventOnline, OccurredAt: base.Add(2 * time.Minute)},
	}
	tx, _ := database.Begin()
	for i := range events {
		if err := s.CreateEvent(ctx, tx, &events[i]); err != nil {
			t.Fatalf("create event: %v", err)
		}
	}
	tx.Commit()

	got, err := s.FindEventsByChecks(ctx, []string{signIn.ID, jobs.ID})
	if err != nil {
		t.Fatalf("find events: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("events: got %d, want 3", len(got))
	}
	if got[0].EventType != models.AccessCheckEventOnline || got[0].AccessCheckID != jobs.ID {
		t.Errorf("newest first: got %+v", got[0])
	}
	if got[2].EventType != models.AccessCheckEventStarted {
		t.Errorf("oldest last: got %+v", got[2])
	}

	none, err := s.FindEventsByChecks(ctx, nil)
	if err != nil || none != nil {
		t.Fatalf("empty ids: got %v,%v want nil,nil", none, err)
	}
}
