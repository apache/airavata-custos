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
	"strings"
	"testing"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
)

// baseRPCBody returns a fully populated request_project_create body with
// every required field present.
func baseRPCBody() map[string]any {
	return map[string]any{
		"AllocationType":        "new",
		"EndDate":               "2027-01-01",
		"GrantNumber":           "TG-BL-001",
		"PfosNumber":            "PFOS-1",
		"PiFirstName":           "Pat",
		"PiLastName":            "First",
		"PiOrganization":        "Baseline Org",
		"PiOrgCode":             "BASELINE",
		"StartDate":             "2026-01-01",
		"ResourceList":          []any{"compute.access-ci.org"},
		"RecordID":              "rec-bl-001",
		"ServiceUnitsAllocated": "1000",
		"PiGlobalID":            "bl-pi-001",
		"PiEmail":               "pat.first@baseline.example.edu",
	}
}

// TestRequestProjectCreate_HappyPath asserts the spec-required side effects
// of a first-delivery request_project_create with a complete body. The
// handler MUST:
//   - persist the PI as a local Person with a site-assigned PersonID
//   - persist a local Project with a site-assigned ProjectID, identified
//     long-term by GrantNumber
//   - persist an initial Allocation for the project carrying
//     ServiceUnitsAllocated / StartDate / EndDate
//   - reply notify_project_create with PiPersonID, PiRemoteSiteLogin, ProjectID
func TestRequestProjectCreate_HappyPath(t *testing.T) {
	database := setupTestDB(t)

	body := baseRPCBody()
	pkt := insertPacket(t, database, "request_project_create", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestProjectCreateHandler(svc, testClusterID, amie, audit)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	if got := countRows(t, database, "organizations"); got != 1 {
		t.Errorf("organizations: got %d, want 1", got)
	}
	if got := countRows(t, database, "users"); got != 1 {
		t.Errorf("users: got %d, want 1", got)
	}
	if got := countRows(t, database, "user_identities"); got != 1 {
		t.Errorf("user_identities: got %d, want 1", got)
	}
	if got := countRows(t, database, "projects"); got != 1 {
		t.Errorf("projects: got %d, want 1", got)
	}
	if got := countRows(t, database, "compute_allocations"); got != 1 {
		t.Errorf("compute_allocations: got %d, want 1", got)
	}

	var org struct {
		OriginatedID string `db:"originated_id"`
		Name         string `db:"name"`
	}
	if err := database.Get(&org, "SELECT originated_id, name FROM organizations LIMIT 1"); err != nil {
		t.Fatalf("read organization: %v", err)
	}
	if org.OriginatedID != "BASELINE" {
		t.Errorf("organization.originated_id: got %q, want %q", org.OriginatedID, "BASELINE")
	}
	if org.Name != "Baseline Org" {
		t.Errorf("organization.name: got %q, want %q", org.Name, "Baseline Org")
	}

	var user struct {
		ID        string `db:"id"`
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
		Status    string `db:"status"`
	}
	if err := database.Get(&user, "SELECT id, first_name, last_name, email, status FROM users LIMIT 1"); err != nil {
		t.Fatalf("read user: %v", err)
	}
	if user.FirstName != "Pat" || user.LastName != "First" {
		t.Errorf("user name: got %q %q, want Pat First", user.FirstName, user.LastName)
	}
	if user.Email != "pat.first@baseline.example.edu" {
		t.Errorf("user.email: got %q", user.Email)
	}
	if user.Status != "ACTIVE" {
		t.Errorf("user.status: got %q, want ACTIVE", user.Status)
	}

	var ident struct {
		Source     string  `db:"source"`
		ExternalID string  `db:"external_id"`
		Email      *string `db:"email"`
		UserID     string  `db:"user_id"`
	}
	if err := database.Get(&ident, "SELECT source, external_id, email, user_id FROM user_identities LIMIT 1"); err != nil {
		t.Fatalf("read user_identity: %v", err)
	}
	if ident.Source != "access" {
		t.Errorf("identity.source: got %q, want access", ident.Source)
	}
	if ident.ExternalID != "bl-pi-001" {
		t.Errorf("identity.external_id: got %q", ident.ExternalID)
	}
	if ident.UserID != user.ID {
		t.Errorf("identity.user_id: got %q, want %q", ident.UserID, user.ID)
	}

	var proj struct {
		ID           string `db:"id"`
		OriginatedID string `db:"originated_id"`
		ProjectPIID  string `db:"project_pi_id"`
		Status       string `db:"status"`
	}
	if err := database.Get(&proj, "SELECT id, originated_id, project_pi_id, status FROM projects LIMIT 1"); err != nil {
		t.Fatalf("read project: %v", err)
	}
	if proj.OriginatedID != "TG-BL-001" {
		t.Errorf("project.originated_id: got %q, want TG-BL-001", proj.OriginatedID)
	}
	if proj.ProjectPIID != user.ID {
		t.Errorf("project.project_pi_id: got %q, want %q", proj.ProjectPIID, user.ID)
	}
	if proj.Status != "ACTIVE" {
		t.Errorf("project.status: got %q, want ACTIVE", proj.Status)
	}

	var alloc struct {
		ProjectID        string `db:"project_id"`
		Name             string `db:"name"`
		InitialSUAmount  int64  `db:"initial_su_amount"`
		ComputeClusterID string `db:"compute_cluster_id"`
	}
	if err := database.Get(&alloc,
		"SELECT project_id, name, initial_su_amount, compute_cluster_id FROM compute_allocations LIMIT 1",
	); err != nil {
		t.Fatalf("read allocation: %v", err)
	}
	if alloc.ProjectID != proj.ID {
		t.Errorf("allocation.project_id: got %q, want %q", alloc.ProjectID, proj.ID)
	}
	if alloc.InitialSUAmount != 1000 {
		t.Errorf("allocation.initial_su_amount: got %d, want 1000", alloc.InitialSUAmount)
	}
	if alloc.ComputeClusterID != testClusterID {
		t.Errorf("allocation.compute_cluster_id: got %q, want %q", alloc.ComputeClusterID, testClusterID)
	}

	for _, action := range []model.AuditAction{
		model.AuditCreatePerson,
		model.AuditCreateAccount,
		model.AuditCreateProject,
		model.AuditCreateAllocation,
		model.AuditReplySent,
	} {
		if got := countAuditActions(t, database, pkt.ID, action); got != 1 {
			t.Errorf("audit %s: got %d, want 1", action, got)
		}
	}

	var piCU struct {
		LocalUsername string `db:"local_username"`
	}
	if err := database.Get(&piCU,
		"SELECT local_username FROM compute_cluster_users WHERE user_id = ? AND compute_cluster_id = ?",
		user.ID, testClusterID,
	); err != nil {
		t.Fatalf("read PI compute_cluster_user: %v", err)
	}
	if piCU.LocalUsername == "" {
		t.Errorf("PI compute_cluster_user.local_username: empty")
	}

	if got, want := amie.lastReplyType(), "notify_project_create"; got != want {
		t.Fatalf("reply type: got %q, want %q", got, want)
	}
	reply := amie.lastReplyBody()
	if reply == nil {
		t.Fatal("reply body missing")
	}
	if got, _ := reply["ProjectID"].(string); got != proj.ID {
		t.Errorf("reply.ProjectID: got %q, want %q", got, proj.ID)
	}
	if got, _ := reply["PiPersonID"].(string); got != user.ID {
		t.Errorf("reply.PiPersonID: got %q, want %q", got, user.ID)
	}
	if got, _ := reply["PiRemoteSiteLogin"].(string); got == "" {
		t.Errorf("reply.PiRemoteSiteLogin: empty; required value")
	} else if got != piCU.LocalUsername {
		t.Errorf("reply.PiRemoteSiteLogin: got %q, want %q", got, piCU.LocalUsername)
	}
}

// TestRequestProjectCreate_SupplementIsIdempotent asserts the protocol supplement
// rule: re-delivery of request_project_create against the same GrantNumber
// MUST NOT create a second Project or a second base Allocation. The grant
// event is recorded as a compute_allocation_diffs row instead.
func TestRequestProjectCreate_SupplementIsIdempotent(t *testing.T) {
	database := setupTestDB(t)

	first := baseRPCBody()
	firstPkt := insertPacket(t, database, "request_project_create", first)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestProjectCreateHandler(svc, testClusterID, amie, audit)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": firstPkt.Type, "body": first}, firstPkt, "")
	}); err != nil {
		t.Fatalf("first Handle: %v", err)
	}

	second := baseRPCBody()
	second["AllocationType"] = "supplement"
	second["RecordID"] = "rec-bl-001-supp"
	second["ServiceUnitsAllocated"] = "500"
	secondPkt := insertPacket(t, database, "request_project_create", second)

	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": secondPkt.Type, "body": second}, secondPkt, "")
	}); err != nil {
		t.Fatalf("second Handle: %v", err)
	}

	if got := countRows(t, database, "projects"); got != 1 {
		t.Errorf("projects after supplement: got %d, want 1 (no second project)", got)
	}
	if got := countRows(t, database, "compute_allocations"); got != 1 {
		t.Errorf("compute_allocations after supplement: got %d, want 1 (no second base allocation)", got)
	}
	if got := countRows(t, database, "compute_allocation_diffs"); got != 1 {
		t.Errorf("compute_allocation_diffs: got %d, want 1 (supplement records a diff)", got)
	}

	var diff struct {
		DiffType    string `db:"diff_type"`
		NewSUAmount int64  `db:"new_su_amount"`
	}
	if err := database.Get(&diff,
		"SELECT diff_type, new_su_amount FROM compute_allocation_diffs LIMIT 1",
	); err != nil {
		t.Fatalf("read diff: %v", err)
	}
	if diff.DiffType != "SUPPLEMENT" {
		t.Errorf("diff.diff_type: got %q, want SUPPLEMENT", diff.DiffType)
	}
	if diff.NewSUAmount != 500 {
		t.Errorf("diff.new_su_amount: got %d, want 500 (this packet's delta only, not cumulative)", diff.NewSUAmount)
	}
}

// TestRequestProjectCreate_MissingGrantNumber asserts the handler rejects a
// packet missing a the protocol-required field, without mutating the DB.
func TestRequestProjectCreate_MissingGrantNumber(t *testing.T) {
	database := setupTestDB(t)

	body := baseRPCBody()
	delete(body, "GrantNumber")
	pkt := insertPacket(t, database, "request_project_create", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestProjectCreateHandler(svc, testClusterID, amie, audit)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error for missing GrantNumber, got nil")
	}
	if !strings.Contains(err.Error(), "GrantNumber") {
		t.Errorf("error should mention GrantNumber, got: %v", err)
	}
	if got := countRows(t, database, "projects"); got != 0 {
		t.Errorf("projects on validation failure: got %d, want 0", got)
	}
	if got := countRows(t, database, "users"); got != 0 {
		t.Errorf("users on validation failure: got %d, want 0", got)
	}
	if len(amie.Replies) != 0 {
		t.Errorf("amie replies on validation failure: got %d, want 0", len(amie.Replies))
	}
}

// TestRequestProjectCreate_ReplyFailurePropagates asserts the handler
// surfaces a ReplyToPacket failure to the caller so the processor can decide
// to retry. The protocol does not explicitly mandate atomicity on reply
// failure, so this test does not check rollback shape. (Design gap: core
// service writes commit in their own txns, only audit + reply share the
// handler's tx, tracked separately.)
func TestRequestProjectCreate_ReplyFailurePropagates(t *testing.T) {
	database := setupTestDB(t)

	body := baseRPCBody()
	pkt := insertPacket(t, database, "request_project_create", body)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{FailWith: errors.New("simulated AMIE outage")}
	h := NewRequestProjectCreateHandler(svc, testClusterID, amie, audit)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if err == nil {
		t.Fatal("expected error when reply fails, got nil")
	}
	if !strings.Contains(err.Error(), "reply") && !strings.Contains(err.Error(), "AMIE") {
		t.Errorf("error should reference reply path, got: %v", err)
	}
	// REPLY_SENT audit row must NOT exist when the reply failed.
	if got := countAuditActions(t, database, pkt.ID, model.AuditReplySent); got != 0 {
		t.Errorf("audit REPLY_SENT on failed reply: got %d, want 0", got)
	}
}

// silence unused-import false positives when only some tests reference these.
var _ = model.AuditCreatePerson
