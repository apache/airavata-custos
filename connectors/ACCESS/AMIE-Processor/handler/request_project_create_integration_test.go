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
	"github.com/apache/airavata-custos/pkg/models"
	coreservice "github.com/apache/airavata-custos/pkg/service"
	"github.com/jmoiron/sqlx"
)

// baseRPCBody returns a fully populated request_project_create body with
// every required field present.
func baseRPCBody() map[string]any {
	return map[string]any{
		"AllocationType":        "new",
		"EndDate":               "2027-01-01",
		"GrantNumber":           "TG-BL-001",
		"ProjectTitle":          "Baseline project",
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
	h := NewRequestProjectCreateHandler(svc, testClusterID, amie, audit, nil)

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

	var title string
	if err := database.Get(&title, "SELECT title FROM projects LIMIT 1"); err != nil || title != "Baseline project" {
		t.Errorf("project title: got %q, %v; want %q", title, err, "Baseline project")
	}
	if got := countRows(t, database, "compute_allocations"); got != 1 {
		t.Errorf("compute_allocations: got %d, want 1", got)
	}
	// Unregistered packet resources create no mapping and do not fail the packet.
	if got := countRows(t, database, "compute_allocation_resource_mappings"); got != 0 {
		t.Errorf("compute_allocation_resource_mappings: got %d, want 0 (resource not in catalog)", got)
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
	if user.Status != "PENDING" {
		t.Errorf("user.status: got %q, want PENDING", user.Status)
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
		ProjectID        string  `db:"project_id"`
		Name             string  `db:"name"`
		InitialSUAmount  float64 `db:"initial_su_amount"`
		ComputeClusterID string  `db:"compute_cluster_id"`
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
		t.Errorf("allocation.initial_su_amount: got %g, want 1000", alloc.InitialSUAmount)
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
		"SELECT local_username FROM compute_cluster_users WHERE user_id = $1 AND compute_cluster_id = $2",
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

// Make sure a supplement for a known grant adds its SUs to the existing
// allocation and records a history row, without a second project or allocation.
func TestRequestProjectCreate_SupplementIsIdempotent(t *testing.T) {
	database := setupTestDB(t)

	first := baseRPCBody()
	firstPkt := insertPacket(t, database, "request_project_create", first)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestProjectCreateHandler(svc, testClusterID, amie, audit, nil)

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
		DiffType    string  `db:"diff_type"`
		NewSUAmount float64 `db:"new_su_amount"`
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
		t.Errorf("diff.new_su_amount: got %g, want 500 (this packet's delta only, not cumulative)", diff.NewSUAmount)
	}
	if got := allocationAfter(t, database); got.InitialSUAmount != 1500 {
		t.Errorf("initial_su_amount: got %g, want 1500", got.InitialSUAmount)
	}
}

// deliverTwice runs the base project packet, then a second one for the same grant with the given fields changed.
func deliverTwice(t *testing.T, changes map[string]any) (*sqlx.DB, *coreservice.Service) {
	t.Helper()
	database := setupTestDB(t)
	svc := newTestCoreService(database)
	h := NewRequestProjectCreateHandler(svc, testClusterID, &fakeAmieClient{}, newTestAuditService(database), nil)

	second := baseRPCBody()
	for k, v := range changes {
		second[k] = v
	}
	for _, body := range []map[string]any{baseRPCBody(), second} {
		pkt := insertPacket(t, database, "request_project_create", body)
		if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
			return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
		}); err != nil {
			t.Fatalf("Handle: %v", err)
		}
	}
	return database, svc
}

func allocationAfter(t *testing.T, database *sqlx.DB) models.ComputeAllocation {
	t.Helper()
	var a models.ComputeAllocation
	if err := database.Get(&a,
		"SELECT id, initial_su_amount, start_time, end_time FROM compute_allocations LIMIT 1",
	); err != nil {
		t.Fatalf("read allocation: %v", err)
	}
	return a
}

// Make sure a renewal replaces the SUs and the dates, so usage from the
// earlier period stops counting and usage up to the new end date counts exactly.
func TestRequestProjectCreate_RenewalReplacesGrant(t *testing.T) {
	database, svc := deliverTwice(t, map[string]any{
		"AllocationType":        "renewal",
		"RecordID":              "rec-bl-001-renew",
		"ServiceUnitsAllocated": "2000",
		"StartDate":             "2027-01-01",
		"EndDate":               "2028-01-01",
	})

	got := allocationAfter(t, database)
	if got.InitialSUAmount != 2000 {
		t.Errorf("initial_su_amount: got %g, want 2000", got.InitialSUAmount)
	}
	if got.StartTime.UTC().Format("2006-01-02") != "2027-01-01" || got.EndTime.UTC().Format("2006-01-02") != "2028-01-01" {
		t.Errorf("dates: got %s to %s, want 2027-01-01 to 2028-01-01", got.StartTime, got.EndTime)
	}

	if _, err := database.Exec(
		`INSERT INTO compute_allocation_usages (id, compute_allocation_id, used_su_amount, calculated_time)
		 VALUES ('old-usage', $1, 800, '2026-06-01'), ('new-usage', $1, 0.25, '2028-01-01 23:00:00+00')`, got.ID); err != nil {
		t.Fatalf("insert usage: %v", err)
	}
	used, err := svc.GetTotalSUUsageForAllocation(context.Background(), got.ID)
	if err != nil {
		t.Fatalf("usage: %v", err)
	}
	if used != 0.25 {
		t.Errorf("usage after renewal: got %g, want 0.25", used)
	}
}

// Make sure an extension moves only the end date and leaves the SUs alone.
func TestRequestProjectCreate_ExtensionMovesEndDate(t *testing.T) {
	database, _ := deliverTwice(t, map[string]any{
		"AllocationType":        "extension",
		"RecordID":              "rec-bl-001-ext",
		"ServiceUnitsAllocated": "0",
		"EndDate":               "2027-06-01",
	})

	got := allocationAfter(t, database)
	if got.InitialSUAmount != 1000 {
		t.Errorf("initial_su_amount: got %g, want 1000", got.InitialSUAmount)
	}
	if got.EndTime.UTC().Format("2006-01-02") != "2027-06-01" {
		t.Errorf("end date: got %s, want 2027-06-01", got.EndTime)
	}
}

// Make sure a project packet for an inactivated project sets the project, its
// allocation and the PI's membership back to active.
func TestRequestProjectCreate_ReactivatesInactiveProject(t *testing.T) {
	database := setupTestDB(t)
	seedProjectForRAC(t, database)
	for _, q := range []string{
		"UPDATE projects SET status = 'INACTIVE'",
		"UPDATE compute_allocations SET status = 'INACTIVE'",
		"UPDATE compute_allocation_memberships SET membership_status = 'INACTIVE'",
	} {
		if _, err := database.Exec(q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	svc := newTestCoreService(database)
	h := NewRequestProjectCreateHandler(svc, testClusterID, &fakeAmieClient{}, newTestAuditService(database), nil)
	body := baseRPCBody()
	pkt := insertPacket(t, database, "request_project_create", body)
	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	var statuses []string
	if err := database.Select(&statuses,
		`SELECT status FROM projects
		 UNION ALL SELECT status FROM compute_allocations
		 UNION ALL SELECT membership_status FROM compute_allocation_memberships`); err != nil {
		t.Fatalf("read statuses: %v", err)
	}
	for _, s := range statuses {
		if s != "ACTIVE" {
			t.Errorf("statuses: got %v, want all ACTIVE", statuses)
			break
		}
	}
}

// Make sure a negative transfer deducts its exact amount, fraction included,
// from the existing allocation.
func TestRequestProjectCreate_NegativeTransferDeducts(t *testing.T) {
	database, _ := deliverTwice(t, map[string]any{
		"AllocationType":        "transfer",
		"RecordID":              "rec-bl-001-out",
		"ServiceUnitsAllocated": "-300.5",
	})

	if got := allocationAfter(t, database); got.InitialSUAmount != 699.5 {
		t.Errorf("initial_su_amount: got %g, want 699.5", got.InitialSUAmount)
	}
}

// Make sure a project packet whose PI has no email is rejected as invalid and creates nothing.
func TestRequestProjectCreate_MissingPIEmailIsInvalid(t *testing.T) {
	database := setupTestDB(t)
	body := baseRPCBody()
	delete(body, "PiEmail")
	pkt := insertPacket(t, database, "request_project_create", body)
	h := NewRequestProjectCreateHandler(newTestCoreService(database), testClusterID, &fakeAmieClient{}, newTestAuditService(database), nil)

	err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	})
	if !errors.Is(err, model.ErrInvalidPacket) {
		t.Fatalf("error: got %v, want an invalid packet error", err)
	}
	if got := countRows(t, database, "organizations"); got != 0 {
		t.Errorf("organizations: got %d, want 0", got)
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
	h := NewRequestProjectCreateHandler(svc, testClusterID, amie, audit, nil)

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
	h := NewRequestProjectCreateHandler(svc, testClusterID, amie, audit, nil)

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

// TestRequestProjectCreate_AttachesRegisteredResources asserts a catalog-registered
// packet resource is mapped with no caps and not duplicated on re-delivery.
func TestRequestProjectCreate_AttachesRegisteredResources(t *testing.T) {
	database := setupTestDB(t)

	svc := newTestCoreService(database)
	audit := newTestAuditService(database)
	amie := &fakeAmieClient{}
	h := NewRequestProjectCreateHandler(svc, testClusterID, amie, audit, nil)

	resource, err := svc.CreateComputeAllocationResource(context.Background(), &models.ComputeAllocationResource{
		Name:             "compute.access-ci.org",
		ResourceType:     "CPU_HOURS",
		ResourceAmount:   1000,
		ComputeClusterID: testClusterID,
	})
	if err != nil {
		t.Fatalf("seed catalog resource: %v", err)
	}

	body := baseRPCBody()
	pkt := insertPacket(t, database, "request_project_create", body)
	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var mapping struct {
		ComputeAllocationResourceID string `db:"compute_allocation_resource_id"`
		ResourceAmount              int64  `db:"resource_amount"`
		ResourceTime                int64  `db:"resource_time"`
	}
	if err := database.Get(&mapping,
		"SELECT compute_allocation_resource_id, resource_amount, resource_time FROM compute_allocation_resource_mappings LIMIT 1",
	); err != nil {
		t.Fatalf("read mapping: %v", err)
	}
	if mapping.ComputeAllocationResourceID != resource.ID {
		t.Errorf("mapping.resource_id: got %q, want %q", mapping.ComputeAllocationResourceID, resource.ID)
	}
	// The packet carries no native amounts; the SU budget is the limit.
	if mapping.ResourceAmount != 0 || mapping.ResourceTime != 0 {
		t.Errorf("mapping caps: got (%d,%d), want (0,0)", mapping.ResourceAmount, mapping.ResourceTime)
	}
	if got := countAuditActions(t, database, pkt.ID, model.AuditAttachResource); got != 1 {
		t.Errorf("audit ATTACH_RESOURCE: got %d, want 1", got)
	}

	// Re-delivery (supplement) must not duplicate the mapping.
	second := baseRPCBody()
	second["AllocationType"] = "supplement"
	secondPkt := insertPacket(t, database, "request_project_create", second)
	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": secondPkt.Type, "body": second}, secondPkt, "")
	}); err != nil {
		t.Fatalf("second Handle: %v", err)
	}
	if got := countRows(t, database, "compute_allocation_resource_mappings"); got != 1 {
		t.Errorf("mappings after re-delivery: got %d, want 1", got)
	}
	if got := countAuditActions(t, database, secondPkt.ID, model.AuditAttachResource); got != 0 {
		t.Errorf("audit ATTACH_RESOURCE on re-delivery: got %d, want 0", got)
	}
}

// Make sure an alias with two catalog resources attaches the allocation to both.
func TestRequestProjectCreate_AttachesEveryAliasedResource(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestCoreService(database)
	h := NewRequestProjectCreateHandler(svc, testClusterID, &fakeAmieClient{}, newTestAuditService(database), map[string][]string{
		"compute.access-ci.org": {"gpu-b200", "gpu-rtxpro"},
	})
	for _, name := range []string{"gpu-b200", "gpu-rtxpro"} {
		if _, err := svc.CreateComputeAllocationResource(context.Background(), &models.ComputeAllocationResource{
			Name: name, ResourceType: "gres/gpu", ResourceAmount: 1, ComputeClusterID: testClusterID,
		}); err != nil {
			t.Fatalf("seed catalog resource %s: %v", name, err)
		}
	}

	body := baseRPCBody()
	pkt := insertPacket(t, database, "request_project_create", body)
	if err := runHandlerInTx(t, database, func(ctx context.Context, tx *sql.Tx) error {
		return h.Handle(ctx, tx, map[string]any{"type": pkt.Type, "body": body}, pkt, "")
	}); err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}

	var attached []string
	if err := database.Select(&attached,
		`SELECT r.name FROM compute_allocation_resource_mappings m
		 JOIN compute_allocation_resources r ON r.id = m.compute_allocation_resource_id ORDER BY r.name`); err != nil {
		t.Fatalf("read mappings: %v", err)
	}
	if strings.Join(attached, ",") != "gpu-b200,gpu-rtxpro" {
		t.Errorf("attached resources: got %v, want gpu-b200 and gpu-rtxpro", attached)
	}
}

// silence unused-import false positives when only some tests reference these.
var _ = model.AuditCreatePerson
