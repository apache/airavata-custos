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

package operations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"

	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/client"
	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/store"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// ---- fakes ------------------------------------------------------------

type fakeCore struct {
	user        *models.User
	userErr     error
	auditEvents []*models.AuditEvent
	provisioned []string
}

func (f *fakeCore) GetUser(_ context.Context, _ string) (*models.User, error) {
	return f.user, f.userErr
}
func (f *fakeCore) CreateAuditEvent(_ context.Context, e *models.AuditEvent) (*models.AuditEvent, error) {
	f.auditEvents = append(f.auditEvents, e)
	return e, nil
}
func (f *fakeCore) MarkComputeClusterUserProvisioned(_ context.Context, id string) error {
	f.provisioned = append(f.provisioned, id)
	return nil
}

// fakeAccounts keeps ldap_posix_accounts in memory with the store's
// numbering and username rules.
type fakeAccounts struct {
	rows map[string]*store.Account
}

func newAccounts(rows ...store.Account) *fakeAccounts {
	f := &fakeAccounts{rows: map[string]*store.Account{}}
	for _, r := range rows {
		f.rows[r.ComputeClusterUserID] = &r
	}
	return f
}

func (f *fakeAccounts) Get(_ context.Context, id string) (*store.Account, error) {
	if r, ok := f.rows[id]; ok {
		c := *r
		return &c, nil
	}
	return nil, nil
}
func (f *fakeAccounts) GetByUsername(_ context.Context, username string) (*store.Account, error) {
	for _, r := range f.rows {
		if r.Username == username {
			c := *r
			return &c, nil
		}
	}
	return nil, nil
}
func (f *fakeAccounts) next(floor int64) int64 {
	n := floor
	for _, r := range f.rows {
		n = max(n, r.UIDNumber+1)
	}
	return n
}
func (f *fakeAccounts) Create(_ context.Context, a store.Account, floor int64) (*store.Account, error) {
	a.UIDNumber = f.next(floor)
	a.GIDNumber = a.UIDNumber
	a.Status = store.StatusActive
	f.rows[a.ComputeClusterUserID] = &a
	c := a
	return &c, nil
}
func (f *fakeAccounts) Renumber(_ context.Context, id string, floor int64) (*store.Account, error) {
	n := f.next(floor)
	r := f.rows[id]
	r.UIDNumber, r.GIDNumber = n, n
	c := *r
	return &c, nil
}
func (f *fakeAccounts) SetStatus(_ context.Context, id, status string) error {
	f.rows[id].Status = status
	return nil
}

type fakeConn struct {
	addErr, modifyErr, delErr error
	// per-search-request response override, keyed by filter string.
	perFilterResults map[string]*ldap.SearchResult
	// Errors for successive Add calls, before falling back to addErr.
	pendingAddErrs []error

	addCalls    []*ldap.AddRequest
	modifyCalls []*ldap.ModifyRequest
	delCalls    []*ldap.DelRequest
}

func (f *fakeConn) Bind(_, _ string) error { return nil }
func (f *fakeConn) Add(r *ldap.AddRequest) error {
	f.addCalls = append(f.addCalls, r)
	if len(f.pendingAddErrs) > 0 {
		err := f.pendingAddErrs[0]
		f.pendingAddErrs = f.pendingAddErrs[1:]
		return err
	}
	return f.addErr
}
func (f *fakeConn) Modify(r *ldap.ModifyRequest) error {
	f.modifyCalls = append(f.modifyCalls, r)
	return f.modifyErr
}
func (f *fakeConn) Del(r *ldap.DelRequest) error {
	f.delCalls = append(f.delCalls, r)
	return f.delErr
}
func (f *fakeConn) Search(r *ldap.SearchRequest) (*ldap.SearchResult, error) {
	if res, ok := f.perFilterResults[r.Filter]; ok {
		return res, nil
	}
	return &ldap.SearchResult{}, nil
}
func (f *fakeConn) Close() error { return nil }

type fakeDialer struct {
	conn *fakeConn
}

func (d *fakeDialer) Dial(_ string, _ bool, _ time.Duration) (client.Connection, error) {
	return d.conn, nil
}

// ---- fixtures ---------------------------------------------------------

const (
	accountFilter = "(&(objectClass=posixAccount)(uid=alovelace))"
	accountDN     = "uid=alovelace,ou=people,dc=test"
)

func sampleCU() *models.ComputeClusterUser {
	return &models.ComputeClusterUser{
		ID:               "csu-1",
		ComputeClusterID: "cluster-1",
		UserID:           "user-1",
		LocalUsername:    "alovelace",
	}
}

func sampleUser() *models.User {
	return &models.User{
		ID:        "user-1",
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     "ada@example.edu",
	}
}

func storedAccount() store.Account {
	return store.Account{
		ComputeClusterUserID: "csu-1",
		UIDNumber:            50123,
		GIDNumber:            50123,
		Username:             "alovelace",
		HomeDirectory:        "/home/alovelace",
		LoginShell:           "/bin/bash",
		EntryDN:              accountDN,
		Status:               store.StatusActive,
	}
}

func existingEntry(uidNumber string) *ldap.SearchResult {
	return &ldap.SearchResult{Entries: []*ldap.Entry{ldap.NewEntry(accountDN, map[string][]string{
		"uid":       {"alovelace"},
		"uidNumber": {uidNumber},
	})}}
}

func testConfig() client.Config {
	return client.Config{
		URL:             "ldap://x",
		BindDN:          "cn=x",
		BaseDN:          "ou=people,dc=test",
		CustosClusterID: "cluster-1",
		HomedirPrefix:   "/home/",
		DefaultShell:    "/bin/bash",
		Timeout:         time.Second,
		MinUID:          50000,
	}
}

func newOrch(core *fakeCore, conn *fakeConn, accounts *fakeAccounts) *Orchestrator {
	c := client.NewWithDialer(testConfig(), &fakeDialer{conn: conn})
	return &Orchestrator{c: c, core: core, accounts: accounts, floor: 50000}
}

func newOrchWithGroups(core *fakeCore, conn *fakeConn, accounts *fakeAccounts) *Orchestrator {
	cfg := testConfig()
	cfg.GroupBaseDN = "ou=groups,dc=test"
	return &Orchestrator{c: client.NewWithDialer(cfg, &fakeDialer{conn: conn}), core: core, accounts: accounts, floor: 50000}
}

// ---- tests ------------------------------------------------------------

func TestEnsurePOSIXAccount_NewUser_StoresAndAdds(t *testing.T) {
	core, conn, accounts := &fakeCore{user: sampleUser()}, &fakeConn{}, newAccounts()

	if err := newOrch(core, conn, accounts).EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	row := accounts.rows["csu-1"]
	if row == nil || row.UIDNumber != 50000 || row.GIDNumber != 50000 || row.EntryDN != accountDN || row.LoginShell != "/bin/bash" {
		t.Fatalf("expected a row keyed on the cluster account; got %+v", row)
	}
	if len(conn.addCalls) != 1 {
		t.Fatalf("expected 1 Add call, got %d", len(conn.addCalls))
	}
	got := indexAttrs(conn.addCalls[0].Attributes)
	if got["uidNumber"][0] != "50000" || got["gidNumber"][0] != "50000" {
		t.Errorf("uidNumber/gidNumber: got %v/%v, want 50000", got["uidNumber"], got["gidNumber"])
	}
	if !contains(auditEventTypes(core.auditEvents), "LDAPAccountCreated") {
		t.Errorf("expected LDAPAccountCreated audit; got %v", auditEventTypes(core.auditEvents))
	}
	if len(core.provisioned) != 1 {
		t.Errorf("expected the cluster account to be marked provisioned")
	}
}

func TestEnsurePOSIXAccount_NumbersAboveExistingRows(t *testing.T) {
	other := storedAccount()
	other.ComputeClusterUserID, other.Username, other.Status = "csu-0", "bbabbage", store.StatusInactive
	accounts := newAccounts(other)

	if err := newOrch(&fakeCore{user: sampleUser()}, &fakeConn{}, accounts).EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// An inactive row keeps its number, so it is not handed out again.
	if got := accounts.rows["csu-1"].UIDNumber; got != 50124 {
		t.Errorf("uid_number: got %d, want 50124", got)
	}
}

func TestEnsurePOSIXAccount_HomeDirectoryHasSeparator(t *testing.T) {
	for _, prefix := range []string{"/home", "/home/"} {
		t.Run(prefix, func(t *testing.T) {
			cfg := testConfig()
			cfg.HomedirPrefix = prefix
			accounts := newAccounts()
			o := &Orchestrator{c: client.NewWithDialer(cfg, &fakeDialer{conn: &fakeConn{}}), core: &fakeCore{user: sampleUser()}, accounts: accounts, floor: 50000}
			if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := accounts.rows["csu-1"].HomeDirectory; got != "/home/alovelace" {
				t.Errorf("home_directory: got %q, want /home/alovelace", got)
			}
		})
	}
}

func TestEnsurePOSIXAccount_UserWithoutLastName(t *testing.T) {
	cases := []struct {
		name, first, wantSN, wantCN string
	}{
		{"first name only", "Ada", "Ada", "Ada"},
		{"no names", "", "alovelace", "alovelace"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			user := sampleUser()
			user.FirstName, user.LastName = tc.first, ""
			conn := &fakeConn{}
			if err := newOrch(&fakeCore{user: user}, conn, newAccounts()).EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
				t.Fatalf("a user without a last name must still be provisioned: %v", err)
			}
			got := indexAttrs(conn.addCalls[0].Attributes)
			if got["sn"][0] != tc.wantSN || got["cn"][0] != tc.wantCN {
				t.Errorf("sn=%v cn=%v, want sn=%q cn=%q", got["sn"], got["cn"], tc.wantSN, tc.wantCN)
			}
		})
	}
}

func TestEnsurePOSIXAccount_StoredAccount_ModifiesExisting(t *testing.T) {
	core := &fakeCore{user: sampleUser()}
	conn := &fakeConn{perFilterResults: map[string]*ldap.SearchResult{accountFilter: existingEntry("50123")}}
	accounts := newAccounts(storedAccount())

	if err := newOrch(core, conn, accounts).EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conn.modifyCalls) != 1 || len(conn.addCalls) != 0 {
		t.Errorf("expected 1 Modify and 0 Add; got %d and %d", len(conn.modifyCalls), len(conn.addCalls))
	}
	if len(accounts.rows) != 1 {
		t.Errorf("expected no new row; got %d", len(accounts.rows))
	}
	if !contains(auditEventTypes(core.auditEvents), "LDAPAccountUpdated") {
		t.Errorf("expected LDAPAccountUpdated audit; got %v", auditEventTypes(core.auditEvents))
	}
}

func TestEnsurePOSIXAccount_StoredAccount_ConflictsOnOtherUIDNumber(t *testing.T) {
	core := &fakeCore{user: sampleUser()}
	conn := &fakeConn{perFilterResults: map[string]*ldap.SearchResult{accountFilter: existingEntry("40000")}}

	err := newOrch(core, conn, newAccounts(storedAccount())).EnsurePOSIXAccount(context.Background(), sampleCU())
	if !errors.Is(err, errConflict) || !errors.Is(err, events.ErrPermanent) {
		t.Fatalf("expected a permanent conflict, got %v", err)
	}
	if len(conn.modifyCalls) != 0 {
		t.Error("an entry with another uidNumber must not be modified")
	}
	assertFailedStep(t, core, "verify_posix_account")
}

func TestEnsurePOSIXAccount_ExistingEntryWithoutRow_IsConflict(t *testing.T) {
	// The entry may belong to another person, so it must not be adopted
	// and overwritten.
	core := &fakeCore{user: sampleUser()}
	conn := &fakeConn{perFilterResults: map[string]*ldap.SearchResult{accountFilter: existingEntry("50777")}}
	accounts := newAccounts()

	err := newOrch(core, conn, accounts).EnsurePOSIXAccount(context.Background(), sampleCU())
	if !errors.Is(err, errConflict) || !errors.Is(err, events.ErrPermanent) {
		t.Fatalf("expected a permanent conflict, got %v", err)
	}
	if len(conn.modifyCalls) != 0 || len(conn.addCalls) != 0 || len(accounts.rows) != 0 {
		t.Error("nothing may be written for a conflicting username")
	}
	assertFailedStep(t, core, "claim_username")
}

func TestEnsurePOSIXAccount_UsernameHeldByAnotherRow_IsConflict(t *testing.T) {
	held := storedAccount()
	held.ComputeClusterUserID, held.Status = "csu-0", store.StatusInactive
	core, conn := &fakeCore{user: sampleUser()}, &fakeConn{}

	err := newOrch(core, conn, newAccounts(held)).EnsurePOSIXAccount(context.Background(), sampleCU())
	if !errors.Is(err, errConflict) || !errors.Is(err, events.ErrPermanent) {
		t.Fatalf("expected a permanent conflict, got %v", err)
	}
	if len(conn.addCalls) != 0 {
		t.Error("no entry may be written")
	}
	assertFailedStep(t, core, "claim_username")
}

func TestEnsurePOSIXAccount_RenumbersOnConstraintViolation(t *testing.T) {
	conn := &fakeConn{pendingAddErrs: []error{&ldap.Error{ResultCode: ldap.LDAPResultConstraintViolation}}}
	accounts := newAccounts()

	if err := newOrch(&fakeCore{user: sampleUser()}, conn, accounts).EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conn.addCalls) != 2 {
		t.Fatalf("expected 2 Add attempts, got %d", len(conn.addCalls))
	}
	if got := indexAttrs(conn.addCalls[1].Attributes)["uidNumber"][0]; got != "50001" {
		t.Errorf("retry uidNumber: got %s, want 50001", got)
	}
	if got := accounts.rows["csu-1"].UIDNumber; got != 50001 {
		t.Errorf("stored uid_number: got %d, want 50001", got)
	}
}

func TestEnsurePOSIXAccount_GivesUpAfterMaxRetries(t *testing.T) {
	core := &fakeCore{user: sampleUser()}
	conn := &fakeConn{addErr: &ldap.Error{ResultCode: ldap.LDAPResultConstraintViolation}}

	if err := newOrch(core, conn, newAccounts()).EnsurePOSIXAccount(context.Background(), sampleCU()); err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if len(conn.addCalls) != maxAllocRetries+1 {
		t.Errorf("expected %d Add attempts, got %d", maxAllocRetries+1, len(conn.addCalls))
	}
	assertFailedStep(t, core, "add_posix_account")
}

func TestEnsurePOSIXAccount_PropagatesNonRetryableAddError(t *testing.T) {
	core := &fakeCore{user: sampleUser()}
	conn := &fakeConn{addErr: &ldap.Error{ResultCode: ldap.LDAPResultInsufficientAccessRights}}

	if err := newOrch(core, conn, newAccounts()).EnsurePOSIXAccount(context.Background(), sampleCU()); err == nil {
		t.Fatal("expected error")
	}
	if len(conn.addCalls) != 1 {
		t.Errorf("expected no retry, got %d Add calls", len(conn.addCalls))
	}
	assertFailedStep(t, core, "add_posix_account")
}

func TestEnsurePOSIXAccount_RejectsUnsafeLocalUsername(t *testing.T) {
	for _, name := range []string{"", "Alice", "a,b", "a=b", "a b", "1abc", strings.Repeat("a", 33)} {
		t.Run(name, func(t *testing.T) {
			core, conn := &fakeCore{user: sampleUser()}, &fakeConn{}
			cu := sampleCU()
			cu.LocalUsername = name
			if err := newOrch(core, conn, newAccounts()).EnsurePOSIXAccount(context.Background(), cu); !errors.Is(err, events.ErrPermanent) {
				t.Fatal("expected error")
			}
			if len(conn.addCalls) != 0 {
				t.Error("nothing may be written for an unsafe username")
			}
			assertFailedStep(t, core, "validate_local_username")
		})
	}
}

func TestEnsurePOSIXAccount_FailsOnUserLookupFailure(t *testing.T) {
	core := &fakeCore{userErr: errors.New("db down")}
	if err := newOrch(core, &fakeConn{}, newAccounts()).EnsurePOSIXAccount(context.Background(), sampleCU()); err == nil {
		t.Fatal("expected error")
	}
	assertFailedStep(t, core, "get_custos_user")
}

func TestEnsurePOSIXAccount_CreatesPrimaryGroupWhenGroupBaseDNSet(t *testing.T) {
	core, conn := &fakeCore{user: sampleUser()}, &fakeConn{}

	if err := newOrchWithGroups(core, conn, newAccounts()).EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conn.addCalls) != 2 {
		t.Fatalf("expected account and group Adds, got %d", len(conn.addCalls))
	}
	group := conn.addCalls[1]
	if group.DN != "cn=alovelace,ou=groups,dc=test" || indexAttrs(group.Attributes)["gidNumber"][0] != "50000" {
		t.Errorf("unexpected group: %s %v", group.DN, indexAttrs(group.Attributes))
	}
}

func TestEnsurePOSIXAccount_ToleratesGroupAlreadyExistsError(t *testing.T) {
	conn := &fakeConn{pendingAddErrs: []error{nil, &ldap.Error{ResultCode: ldap.LDAPResultEntryAlreadyExists}}}
	if err := newOrchWithGroups(&fakeCore{user: sampleUser()}, conn, newAccounts()).EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("a concurrent group add should be treated as success: %v", err)
	}
}

func TestDeactivatePOSIXAccount_DeletesEntryAndKeepsNumber(t *testing.T) {
	core, conn := &fakeCore{}, &fakeConn{}
	accounts := newAccounts(storedAccount())

	if err := newOrch(core, conn, accounts).DeactivatePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conn.delCalls) != 1 || conn.delCalls[0].DN != accountDN {
		t.Fatalf("expected the account entry to be deleted; got %+v", conn.delCalls)
	}
	if row := accounts.rows["csu-1"]; row.Status != store.StatusInactive || row.UIDNumber != 50123 {
		t.Errorf("the row must stay, inactive, with its number; got %+v", row)
	}
	if !contains(auditEventTypes(core.auditEvents), "LDAPAccountDeactivated") {
		t.Errorf("expected LDAPAccountDeactivated audit; got %v", auditEventTypes(core.auditEvents))
	}
}

func TestDeactivatePOSIXAccount_DeletesPrimaryGroup(t *testing.T) {
	const groupDN = "cn=alovelace,ou=groups,dc=test"
	conn := &fakeConn{perFilterResults: map[string]*ldap.SearchResult{
		"(&(objectClass=posixGroup)(cn=alovelace))": {Entries: []*ldap.Entry{ldap.NewEntry(groupDN, nil)}},
	}}
	core := &fakeCore{}

	if err := newOrchWithGroups(core, conn, newAccounts(storedAccount())).DeactivatePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conn.delCalls) != 2 || conn.delCalls[1].DN != groupDN {
		t.Fatalf("expected the account and group deleted; got %+v", conn.delCalls)
	}
	if !contains(auditEventTypes(core.auditEvents), "LDAPGroupDeleted") {
		t.Errorf("expected LDAPGroupDeleted audit; got %v", auditEventTypes(core.auditEvents))
	}
}

func TestDeactivatePOSIXAccount_NoRowIsNoop(t *testing.T) {
	core, conn := &fakeCore{}, &fakeConn{}
	if err := newOrch(core, conn, newAccounts()).DeactivatePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conn.delCalls) != 0 {
		t.Error("an account the connector never created needs nothing done")
	}
}

func TestDeactivatePOSIXAccount_DeleteFailureKeepsRowActive(t *testing.T) {
	core := &fakeCore{}
	conn := &fakeConn{delErr: &ldap.Error{ResultCode: ldap.LDAPResultInsufficientAccessRights}}
	accounts := newAccounts(storedAccount())

	if err := newOrch(core, conn, accounts).DeactivatePOSIXAccount(context.Background(), sampleCU()); err == nil {
		t.Fatal("expected an error so the bus retries")
	}
	if accounts.rows["csu-1"].Status != store.StatusActive {
		t.Error("the row must stay active while the entry is still in the directory")
	}
	assertFailedStep(t, core, "delete_posix_account")
}

func TestEnsurePOSIXAccount_SkipsDeactivatedAccount(t *testing.T) {
	// An approval still being retried when the account is deleted must not
	// bring the entry back.
	inactive := storedAccount()
	inactive.Status = store.StatusInactive
	core, conn, accounts := &fakeCore{user: sampleUser()}, &fakeConn{}, newAccounts(inactive)

	if err := newOrch(core, conn, accounts).EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conn.addCalls) != 0 || len(conn.modifyCalls) != 0 {
		t.Errorf("nothing may be written; got %d Add and %d Modify", len(conn.addCalls), len(conn.modifyCalls))
	}
	if accounts.rows["csu-1"].Status != store.StatusInactive || len(core.provisioned) != 0 {
		t.Error("the account must stay deactivated and not be marked provisioned")
	}
}

func TestEnsurePOSIXAccount_LogLinesCarryCorrelationID(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	conn := &fakeConn{pendingAddErrs: []error{&ldap.Error{ResultCode: ldap.LDAPResultConstraintViolation}}}
	if err := newOrchWithGroups(&fakeCore{user: sampleUser()}, conn, newAccounts()).EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected several log lines; got %q", buf.String())
	}
	for _, line := range lines {
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("bad log line %q: %v", line, err)
		}
		if rec["correlation_id"] != "csu-1" {
			t.Errorf("log line without correlation_id: %s", line)
		}
	}
}

// ---- helpers ----------------------------------------------------------

func assertFailedStep(t *testing.T, core *fakeCore, step string) {
	t.Helper()
	for _, e := range core.auditEvents {
		if e.EventType == "LDAPProvisioningFailed" && strings.HasPrefix(e.Details, "step="+step+" ") {
			return
		}
	}
	t.Errorf("expected LDAPProvisioningFailed with step=%s; got %v", step, auditEventTypes(core.auditEvents))
}

func indexAttrs(attrs []ldap.Attribute) map[string][]string {
	out := make(map[string][]string, len(attrs))
	for _, a := range attrs {
		out[a.Type] = a.Vals
	}
	return out
}

func auditEventTypes(events []*models.AuditEvent) []string {
	out := make([]string, len(events))
	for i, e := range events {
		out[i] = e.EventType
	}
	return out
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
