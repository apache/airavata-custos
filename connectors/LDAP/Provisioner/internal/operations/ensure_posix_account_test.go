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
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"

	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/client"
	"github.com/apache/airavata-custos/pkg/models"
)

// ---- fakes ------------------------------------------------------------

type fakeCore struct {
	user    *models.User
	userErr error

	// user_identities keyed by userID + source.
	identities    map[string][]models.UserIdentity
	listErr       error
	createUIErr   error
	createdIdents []models.UserIdentity

	auditEvents []*models.AuditEvent
}

func (f *fakeCore) GetUser(_ context.Context, _ string) (*models.User, error) {
	return f.user, f.userErr
}
func (f *fakeCore) ListUserIdentitiesForUser(_ context.Context, userID string) ([]models.UserIdentity, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.identities[userID], nil
}
func (f *fakeCore) CreateUserIdentity(_ context.Context, ui *models.UserIdentity) (*models.UserIdentity, error) {
	if f.createUIErr != nil {
		return nil, f.createUIErr
	}
	f.createdIdents = append(f.createdIdents, *ui)
	return ui, nil
}
func (f *fakeCore) CreateAuditEvent(_ context.Context, e *models.AuditEvent) (*models.AuditEvent, error) {
	f.auditEvents = append(f.auditEvents, e)
	return e, nil
}

// fakeConn mirrors the client-package fake so the orchestrator can drive
// a real *client.Client through a fake connection.
type fakeConn struct {
	bindErr, addErr, modifyErr, searchErr error
	searchResult                          *ldap.SearchResult
	// per-search-request response override, keyed by filter string.
	perFilterResults map[string]*ldap.SearchResult

	addCalls    []*ldap.AddRequest
	modifyCalls []*ldap.ModifyRequest
	searchCalls []*ldap.SearchRequest

	// nextAddErr lets tests fire a constraint violation once, then succeed.
	pendingAddErrs []error
}

func (f *fakeConn) Bind(_, _ string) error { return f.bindErr }
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
func (f *fakeConn) Search(r *ldap.SearchRequest) (*ldap.SearchResult, error) {
	f.searchCalls = append(f.searchCalls, r)
	if f.searchErr != nil {
		return nil, f.searchErr
	}
	if res, ok := f.perFilterResults[r.Filter]; ok {
		return res, nil
	}
	if f.searchResult != nil {
		return f.searchResult, nil
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

// fakeUIDAllocator is an in-memory monotonic counter used in place of
// the DB-backed store.UIDSequence in unit tests. Starts at `next` and
// hands out next, next+1, next+2 ... on successive calls. Errors can
// be programmed via `err` (returned once, then cleared).
type fakeUIDAllocator struct {
	next  int64
	err   error
	calls int
}

func (f *fakeUIDAllocator) Allocate(_ context.Context, _ string) (int64, error) {
	f.calls++
	if f.err != nil {
		e := f.err
		f.err = nil
		return 0, e
	}
	uid := f.next
	f.next++
	return uid, nil
}

func newAllocator(start int64) *fakeUIDAllocator {
	return &fakeUIDAllocator{next: start}
}

// ---- fixtures ---------------------------------------------------------

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

const testClusterID = "cluster-1"

// testIdentitySource is what o.identitySource() returns when
// CustosClusterID = testClusterID. Callers that pre-populate the fake
// cache must use this exact string.
var testIdentitySource = "ldap:" + testClusterID

// newOrch builds an orchestrator with a default allocator starting at
// MinUID (50000). Tests that need to observe or override allocator
// state use newOrchWithAllocator.
func newOrch(core CoreService, conn *fakeConn) *Orchestrator {
	return newOrchWithAllocator(core, conn, newAllocator(50000))
}

func newOrchWithAllocator(core CoreService, conn *fakeConn, uids *fakeUIDAllocator) *Orchestrator {
	c := client.NewWithDialer(client.Config{
		URL:             "ldap://x",
		BindDN:          "cn=x",
		BaseDN:          "ou=people,dc=test",
		CustosClusterID: testClusterID,
		HomedirPrefix:   "/home/",
		DefaultShell:    "/bin/bash",
		Timeout:         time.Second,
		MinUID:          50000,
	}, &fakeDialer{conn: conn})
	return &Orchestrator{c: c, core: core, uids: uids}
}

func newOrchWithGroups(core CoreService, conn *fakeConn) *Orchestrator {
	c := client.NewWithDialer(client.Config{
		URL:             "ldap://x",
		BindDN:          "cn=x",
		BaseDN:          "ou=people,dc=test",
		GroupBaseDN:     "ou=groups,dc=test",
		CustosClusterID: testClusterID,
		HomedirPrefix:   "/home/",
		DefaultShell:    "/bin/bash",
		Timeout:         time.Second,
		MinUID:          50000,
	}, &fakeDialer{conn: conn})
	return &Orchestrator{c: c, core: core, uids: newAllocator(50000)}
}

// ---- tests ------------------------------------------------------------

func TestEnsurePOSIXAccount_NewUser_AllocatesAndAdds(t *testing.T) {
	core := &fakeCore{user: sampleUser()}
	conn := &fakeConn{} // empty search → allocator returns MinUID
	o := newOrch(core, conn)

	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conn.addCalls) != 1 {
		t.Fatalf("expected 1 Add call, got %d", len(conn.addCalls))
	}
	got := indexAttrs(conn.addCalls[0].Attributes)
	if got["uidNumber"][0] != "50000" {
		t.Errorf("uidNumber: got %q, want 50000", got["uidNumber"][0])
	}
	if got["gidNumber"][0] != "50000" {
		t.Errorf("gidNumber must mirror uidNumber; got %q, want 50000", got["gidNumber"][0])
	}
	if len(core.createdIdents) != 1 {
		t.Fatalf("expected 1 UserIdentity write, got %d", len(core.createdIdents))
	}
	if core.createdIdents[0].Source != testIdentitySource {
		t.Errorf("cache source: got %q, want %q", core.createdIdents[0].Source, testIdentitySource)
	}
	if core.createdIdents[0].ExternalID != "50000" {
		t.Errorf("cache external_id: got %q, want 50000", core.createdIdents[0].ExternalID)
	}
	auditTypes := auditEventTypes(core.auditEvents)
	if !contains(auditTypes, "LDAPAccountCreated") {
		t.Errorf("expected LDAPAccountCreated audit; got %v", auditTypes)
	}
}

func TestEnsurePOSIXAccount_CachedUID_ModifiesExisting(t *testing.T) {
	core := &fakeCore{
		user: sampleUser(),
		identities: map[string][]models.UserIdentity{
			"user-1": {{UserID: "user-1", Source: testIdentitySource, ExternalID: "50123"}},
		},
	}
	entry := ldap.NewEntry("uid=alovelace,ou=people,dc=test", map[string][]string{
		"uid":       {"alovelace"},
		"uidNumber": {"50123"},
	})
	conn := &fakeConn{
		perFilterResults: map[string]*ldap.SearchResult{
			"(&(objectClass=posixAccount)(uid=alovelace))": {Entries: []*ldap.Entry{entry}},
		},
	}
	o := newOrch(core, conn)

	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conn.modifyCalls) != 1 {
		t.Errorf("expected 1 Modify call, got %d", len(conn.modifyCalls))
	}
	if len(conn.addCalls) != 0 {
		t.Errorf("expected 0 Add calls when entry exists, got %d", len(conn.addCalls))
	}
	// Should NOT write another cache row since we already had one.
	if len(core.createdIdents) != 0 {
		t.Errorf("expected 0 new cache rows on cache hit; got %d", len(core.createdIdents))
	}
	auditTypes := auditEventTypes(core.auditEvents)
	if !contains(auditTypes, "LDAPAccountUpdated") {
		t.Errorf("expected LDAPAccountUpdated audit; got %v", auditTypes)
	}
}

func TestEnsurePOSIXAccount_AdoptsExistingLDAPEntry(t *testing.T) {
	// No cache, but the entry exists in LDAP (out-of-band or prior run
	// that failed to cache). Orchestrator reads the uidNumber, caches
	// it, and Modifies to sync attributes.
	core := &fakeCore{user: sampleUser()}
	entry := ldap.NewEntry("uid=alovelace,ou=people,dc=test", map[string][]string{
		"uid":       {"alovelace"},
		"uidNumber": {"50777"},
	})
	conn := &fakeConn{
		perFilterResults: map[string]*ldap.SearchResult{
			"(&(objectClass=posixAccount)(uid=alovelace))": {Entries: []*ldap.Entry{entry}},
		},
	}
	o := newOrch(core, conn)

	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(core.createdIdents) != 1 || core.createdIdents[0].ExternalID != "50777" {
		t.Errorf("expected cache write for existing uid 50777; got %+v", core.createdIdents)
	}
	if len(conn.modifyCalls) != 1 {
		t.Errorf("expected 1 Modify call, got %d", len(conn.modifyCalls))
	}
}

func TestEnsurePOSIXAccount_RetriesOnConstraintViolation(t *testing.T) {
	core := &fakeCore{user: sampleUser()}
	// Two Add attempts: first fails with constraint violation, second
	// succeeds. Search always returns empty → allocator would return
	// MinUID (50000) each time, but the retry loop still counts.
	conn := &fakeConn{
		pendingAddErrs: []error{
			&ldap.Error{ResultCode: ldap.LDAPResultConstraintViolation},
		},
	}
	o := newOrch(core, conn)

	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conn.addCalls) != 2 {
		t.Errorf("expected 2 Add attempts (one collision + retry), got %d", len(conn.addCalls))
	}
	if len(core.createdIdents) != 1 {
		t.Errorf("expected cache write after successful add; got %d", len(core.createdIdents))
	}
}

func TestEnsurePOSIXAccount_AdoptsOnConcurrentEntryAlreadyExists(t *testing.T) {
	// Multi-instance Custos: another instance wrote the account moments
	// before we tried. Our Add returns EntryAlreadyExists (LDAP code
	// 68). We must re-Find and adopt, NOT emit LDAPProvisioningFailed.
	//
	// Timeline:
	//   1. Initial Find → empty (no cache, no entry)
	//   2. Allocator scan → empty
	//   3. Add → EntryAlreadyExists (other instance beat us)
	//   4. Adoption Find → the racing entry (this is the key step)
	//   5. Modify (via ensureEntry) → success
	//   6. Audit: LDAPAccountUpdated (not Failed)
	core := &fakeCore{user: sampleUser()}
	existing := ldap.NewEntry("uid=alovelace,ou=people,dc=test", map[string][]string{
		"uid":       {"alovelace"},
		"uidNumber": {"50888"},
	})
	inner := &fakeConn{
		pendingAddErrs: []error{
			&ldap.Error{ResultCode: ldap.LDAPResultEntryAlreadyExists},
		},
	}
	// afterAddConn returns `existing` for posixAccount searches only
	// AFTER Add has been called. This models the LDAP server state
	// transition when a concurrent writer commits between our Find and
	// our Add.
	wrapped := &afterAddConn{inner: inner, existing: existing}
	c := client.NewWithDialer(client.Config{
		URL:           "ldap://x",
		BindDN:        "cn=x",
		BaseDN:        "ou=people,dc=test",
		HomedirPrefix: "/home/",
		DefaultShell:  "/bin/bash",
		Timeout:       time.Second,
		MinUID:        50000,
	}, &wrappedDialer{conn: wrapped})
	o := &Orchestrator{c: c, core: core, uids: newAllocator(50000)}

	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	auditTypes := auditEventTypes(core.auditEvents)
	if contains(auditTypes, "LDAPProvisioningFailed") {
		t.Errorf("concurrent EntryAlreadyExists must NOT emit LDAPProvisioningFailed; got %v", auditTypes)
	}
	if !contains(auditTypes, "LDAPAccountUpdated") {
		t.Errorf("expected LDAPAccountUpdated after adoption; got %v", auditTypes)
	}
	if len(core.createdIdents) != 1 || core.createdIdents[0].ExternalID != "50888" {
		t.Errorf("expected cache write with adopted uid 50888; got %+v", core.createdIdents)
	}
}

// afterAddConn wraps a fakeConn and switches its posixAccount search
// results to return `existing` only after Add has been called at least
// once. Models a concurrent writer committing between our Find and Add.
type afterAddConn struct {
	inner    *fakeConn
	existing *ldap.Entry
	addSeen  bool
}

func (t *afterAddConn) Bind(u, p string) error { return t.inner.Bind(u, p) }
func (t *afterAddConn) Add(r *ldap.AddRequest) error {
	err := t.inner.Add(r)
	t.addSeen = true
	return err
}
func (t *afterAddConn) Modify(r *ldap.ModifyRequest) error { return t.inner.Modify(r) }
func (t *afterAddConn) Search(r *ldap.SearchRequest) (*ldap.SearchResult, error) {
	if t.addSeen && r.Filter == "(&(objectClass=posixAccount)(uid=alovelace))" {
		t.inner.searchCalls = append(t.inner.searchCalls, r)
		return &ldap.SearchResult{Entries: []*ldap.Entry{t.existing}}, nil
	}
	return t.inner.Search(r)
}
func (t *afterAddConn) Close() error { return t.inner.Close() }

type wrappedDialer struct{ conn *afterAddConn }

func (d *wrappedDialer) Dial(_ string, _ bool, _ time.Duration) (client.Connection, error) {
	return d.conn, nil
}

func TestEnsurePOSIXAccount_GivesUpAfterMaxRetries(t *testing.T) {
	core := &fakeCore{user: sampleUser()}
	conn := &fakeConn{
		pendingAddErrs: []error{
			&ldap.Error{ResultCode: ldap.LDAPResultConstraintViolation},
			&ldap.Error{ResultCode: ldap.LDAPResultConstraintViolation},
			&ldap.Error{ResultCode: ldap.LDAPResultConstraintViolation},
		},
	}
	o := newOrch(core, conn)

	err := o.EnsurePOSIXAccount(context.Background(), sampleCU())
	if err == nil {
		t.Fatal("expected error after max retries exhausted")
	}
	if len(conn.addCalls) != maxAllocRetries {
		t.Errorf("expected %d Add attempts, got %d", maxAllocRetries, len(conn.addCalls))
	}
	auditTypes := auditEventTypes(core.auditEvents)
	if !contains(auditTypes, "LDAPProvisioningFailed") {
		t.Errorf("expected LDAPProvisioningFailed audit; got %v", auditTypes)
	}
}

func TestEnsurePOSIXAccount_PropagatesNonRetryableAddError(t *testing.T) {
	core := &fakeCore{user: sampleUser()}
	conn := &fakeConn{
		pendingAddErrs: []error{errors.New("network down")},
	}
	o := newOrch(core, conn)

	err := o.EnsurePOSIXAccount(context.Background(), sampleCU())
	if err == nil {
		t.Fatal("expected error to propagate")
	}
	if len(conn.addCalls) != 1 {
		t.Errorf("expected 1 Add attempt (no retry on non-constraint), got %d", len(conn.addCalls))
	}
}

func TestEnsurePOSIXAccount_RejectsUnsafeLocalUsername(t *testing.T) {
	// DN metacharacters or non-POSIX characters must be rejected before
	// any LDAP call — the RDN concatenation elsewhere is not escaped
	// and relies on this validation upstream.
	cases := []struct {
		name string
		uid  string
	}{
		{"comma", "ada,evil"},
		{"plus", "ada+evil"},
		{"equals", "ada=evil"},
		{"space", "ada evil"},
		{"backslash", `ada\evil`},
		{"uppercase", "Alovelace"},
		{"leading digit", "1ada"},
		{"empty-ish", "  "},
		{"over 32 chars", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			core := &fakeCore{user: sampleUser()}
			conn := &fakeConn{}
			o := newOrch(core, conn)

			cu := sampleCU()
			cu.LocalUsername = tc.uid
			if err := o.EnsurePOSIXAccount(context.Background(), cu); err == nil {
				t.Fatal("expected validation error")
			}
			if len(conn.addCalls) != 0 {
				t.Errorf("no LDAP writes should occur on invalid username; got %d Add calls", len(conn.addCalls))
			}
		})
	}
}

func TestEnsurePOSIXAccount_DLQsOnMissingLocalUsername(t *testing.T) {
	core := &fakeCore{user: sampleUser()}
	o := newOrch(core, &fakeConn{})

	cu := sampleCU()
	cu.LocalUsername = ""
	if err := o.EnsurePOSIXAccount(context.Background(), cu); err == nil {
		t.Fatal("expected error for empty local_username")
	}
	if len(core.auditEvents) != 1 || core.auditEvents[0].EventType != "LDAPProvisioningFailed" {
		t.Errorf("expected LDAPProvisioningFailed audit; got %+v", core.auditEvents)
	}
}

func TestEnsurePOSIXAccount_DLQsOnUserLookupFailure(t *testing.T) {
	core := &fakeCore{userErr: errors.New("db unreachable")}
	o := newOrch(core, &fakeConn{})

	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err == nil {
		t.Fatal("expected error on user lookup failure")
	}
}

func TestEnsurePOSIXAccount_CreatesPrimaryGroupWhenGroupBaseDNSet(t *testing.T) {
	core := &fakeCore{user: sampleUser()}
	conn := &fakeConn{}
	o := newOrchWithGroups(core, conn)

	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(conn.addCalls) != 2 {
		t.Fatalf("expected 2 Add calls (account + group), got %d", len(conn.addCalls))
	}
	// One Add's DN under BaseDN, the other under GroupBaseDN.
	var accountDN, groupDN string
	for _, r := range conn.addCalls {
		if r.DN[:4] == "uid=" {
			accountDN = r.DN
		} else if r.DN[:3] == "cn=" {
			groupDN = r.DN
		}
	}
	if accountDN != "uid=alovelace,ou=people,dc=test" {
		t.Errorf("account DN: got %q", accountDN)
	}
	if groupDN != "cn=alovelace,ou=groups,dc=test" {
		t.Errorf("group DN: got %q", groupDN)
	}
	auditTypes := auditEventTypes(core.auditEvents)
	if !contains(auditTypes, "LDAPGroupCreated") {
		t.Errorf("expected LDAPGroupCreated audit; got %v", auditTypes)
	}
}

func TestEnsurePOSIXAccount_SkipsGroupWhenGroupBaseDNEmpty(t *testing.T) {
	core := &fakeCore{user: sampleUser()}
	conn := &fakeConn{}
	o := newOrch(core, conn) // GroupBaseDN unset

	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(conn.addCalls) != 1 {
		t.Errorf("expected 1 Add (account only), got %d", len(conn.addCalls))
	}
}

func TestEnsurePOSIXAccount_SkipsGroupAddWhenExists(t *testing.T) {
	// Group already present in LDAP → Find returns it, Add is skipped.
	core := &fakeCore{user: sampleUser()}
	groupEntry := ldap.NewEntry("cn=alovelace,ou=groups,dc=test", map[string][]string{
		"cn": {"alovelace"}, "gidNumber": {"50000"},
	})
	conn := &fakeConn{
		perFilterResults: map[string]*ldap.SearchResult{
			"(&(objectClass=posixGroup)(cn=alovelace))": {Entries: []*ldap.Entry{groupEntry}},
		},
	}
	o := newOrchWithGroups(core, conn)

	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	// Only the posixAccount Add should have happened.
	if len(conn.addCalls) != 1 {
		t.Fatalf("expected 1 Add (account only, group already exists), got %d", len(conn.addCalls))
	}
	if conn.addCalls[0].DN[:4] != "uid=" {
		t.Errorf("expected posixAccount Add, got DN %q", conn.addCalls[0].DN)
	}
}

func TestEnsurePOSIXAccount_ToleratesGroupAlreadyExistsError(t *testing.T) {
	// Find returns empty (race), then Add fails with EntryAlreadyExists
	// because another writer created the group between our Find and Add.
	// Must be treated as success, not as a provisioning failure.
	core := &fakeCore{user: sampleUser()}
	conn := &fakeConn{
		// First Add succeeds (account), second Add is the group and
		// fails with already-exists.
		pendingAddErrs: []error{
			nil,
			&ldap.Error{ResultCode: ldap.LDAPResultEntryAlreadyExists},
		},
	}
	o := newOrchWithGroups(core, conn)

	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	// Failed group Add should NOT produce a ProvisioningFailed audit.
	for _, e := range core.auditEvents {
		if e.EventType == "LDAPProvisioningFailed" {
			t.Errorf("no LDAPProvisioningFailed expected on concurrent group creation; got %+v", e)
		}
	}
}

// P1-1: ensureEntry (the "we have a uid already" path) must adopt on
// EntryAlreadyExists too, not just the fresh-allocation path.
func TestEnsurePOSIXAccount_EnsureEntryAdoptsOnConcurrentAdd(t *testing.T) {
	// Setup: cached uid exists (jumps into ensureEntry), Find returns
	// empty (entry deleted since last provisioning), Add fails with
	// EntryAlreadyExists (a concurrent instance re-created it). The
	// orchestrator must NOT DLQ; it must adopt and Modify.
	existing := ldap.NewEntry("uid=alovelace,ou=people,dc=test", map[string][]string{
		"uid":       {"alovelace"},
		"uidNumber": {"50123"},
	})
	inner := &fakeConn{
		pendingAddErrs: []error{
			&ldap.Error{ResultCode: ldap.LDAPResultEntryAlreadyExists},
		},
	}
	wrapped := &afterAddConn{inner: inner, existing: existing}
	core := &fakeCore{
		user: sampleUser(),
		identities: map[string][]models.UserIdentity{
			"user-1": {{UserID: "user-1", Source: testIdentitySource, ExternalID: "50123"}},
		},
	}
	c := client.NewWithDialer(client.Config{
		URL:             "ldap://x",
		BindDN:          "cn=x",
		BaseDN:          "ou=people,dc=test",
		CustosClusterID: testClusterID,
		HomedirPrefix:   "/home/",
		DefaultShell:    "/bin/bash",
		Timeout:         time.Second,
		MinUID:          50000,
	}, &wrappedDialer{conn: wrapped})
	o := &Orchestrator{c: c, core: core, uids: newAllocator(50000)}

	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	auditTypes := auditEventTypes(core.auditEvents)
	if contains(auditTypes, "LDAPProvisioningFailed") {
		t.Errorf("EnsureEntry EntryAlreadyExists must not DLQ; got audits %v", auditTypes)
	}
	if !contains(auditTypes, "LDAPAccountUpdated") {
		t.Errorf("expected LDAPAccountUpdated after adoption; got %v", auditTypes)
	}
}

// P1-2: cache scoped per cluster. A row cached with the OTHER cluster's
// source string must not be treated as a cache hit — a Custos deployment
// servicing two clusters keeps each cluster's uidNumber independent.
func TestEnsurePOSIXAccount_IgnoresOtherClustersCacheEntry(t *testing.T) {
	core := &fakeCore{
		user: sampleUser(),
		identities: map[string][]models.UserIdentity{
			// A row for the SAME user but a DIFFERENT cluster.
			"user-1": {{
				UserID:     "user-1",
				Source:     "ldap:some-other-cluster",
				ExternalID: "50999",
			}},
		},
	}
	conn := &fakeConn{}
	o := newOrch(core, conn) // this connector serves testClusterID

	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	// Because the other cluster's cache row was correctly ignored, we
	// allocated fresh (uid=50000) instead of reusing 50999.
	if len(conn.addCalls) != 1 {
		t.Fatalf("expected 1 Add, got %d", len(conn.addCalls))
	}
	written := indexAttrs(conn.addCalls[0].Attributes)
	if written["uidNumber"][0] != "50000" {
		t.Errorf("expected fresh allocation uid=50000, got %s (leaked from other cluster?)", written["uidNumber"][0])
	}
	// Cache row was written with the CURRENT cluster's source, not the
	// other cluster's.
	if len(core.createdIdents) != 1 {
		t.Fatalf("expected 1 cache write, got %d", len(core.createdIdents))
	}
	if core.createdIdents[0].Source != testIdentitySource {
		t.Errorf("cache source: got %q, want %q", core.createdIdents[0].Source, testIdentitySource)
	}
}

// P1-3: a cache write failure after a successful LDAP add must NOT
// silently pass. It also must not DLQ (LDAP entry is in place; system
// self-heals on next event). The test verifies the flow completes AND
// that no LDAPProvisioningFailed fires — the escalated log is the
// operator-visible signal.
func TestEnsurePOSIXAccount_CacheWriteFailurePreservesProvisioning(t *testing.T) {
	core := &fakeCore{
		user:        sampleUser(),
		createUIErr: errors.New("simulated DB write failure"),
	}
	conn := &fakeConn{}
	o := newOrch(core, conn)

	// Provisioning must succeed even though cache write fails.
	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("cache write failure must not fail the flow: %v", err)
	}
	if len(conn.addCalls) != 1 {
		t.Errorf("expected LDAP Add to have succeeded before cache failure; got %d", len(conn.addCalls))
	}
	auditTypes := auditEventTypes(core.auditEvents)
	if contains(auditTypes, "LDAPProvisioningFailed") {
		t.Errorf("cache write failure must not emit LDAPProvisioningFailed; got %v", auditTypes)
	}
	if !contains(auditTypes, "LDAPAccountCreated") {
		t.Errorf("expected LDAPAccountCreated audit; got %v", auditTypes)
	}
}

// P2-1 (via monotonic counter): consecutive constraint violations
// must retry with different uidNumbers, not the same value. Under the
// current design the counter increments on every Allocate() call, so
// each retry naturally receives the next value in the sequence — no
// "floor bumping" logic is needed.
func TestEnsurePOSIXAccount_RetriesReceiveDistinctUIDs(t *testing.T) {
	core := &fakeCore{user: sampleUser()}
	conn := &fakeConn{
		pendingAddErrs: []error{
			&ldap.Error{ResultCode: ldap.LDAPResultConstraintViolation},
			&ldap.Error{ResultCode: ldap.LDAPResultConstraintViolation},
			nil, // third attempt succeeds
		},
	}
	uids := newAllocator(50000)
	o := newOrchWithAllocator(core, conn, uids)

	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	if len(conn.addCalls) != 3 {
		t.Fatalf("expected 3 Add attempts, got %d", len(conn.addCalls))
	}
	if uids.calls != 3 {
		t.Errorf("expected 3 Allocate calls (one per Add attempt), got %d", uids.calls)
	}
	seen := map[string]bool{}
	for i, r := range conn.addCalls {
		attrs := indexAttrs(r.Attributes)
		u := attrs["uidNumber"][0]
		if seen[u] {
			t.Errorf("attempt %d reused uidNumber=%s; each retry must pull a fresh counter value", i+1, u)
		}
		seen[u] = true
	}
}

func TestEnsurePOSIXAccount_IgnoresCorruptCacheRow(t *testing.T) {
	// A non-integer external_id in user_identities(source="ldap") must
	// not crash provisioning — treat it as un-cached and allocate fresh.
	core := &fakeCore{
		user: sampleUser(),
		identities: map[string][]models.UserIdentity{
			"user-1": {{UserID: "user-1", Source: testIdentitySource, ExternalID: "not-a-number"}},
		},
	}
	conn := &fakeConn{}
	o := newOrch(core, conn)

	if err := o.EnsurePOSIXAccount(context.Background(), sampleCU()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conn.addCalls) != 1 {
		t.Errorf("expected fresh Add after ignoring corrupt cache row; got %d", len(conn.addCalls))
	}
}

// ---- helpers ----------------------------------------------------------

func indexAttrs(attrs []ldap.Attribute) map[string][]string {
	m := make(map[string][]string, len(attrs))
	for _, a := range attrs {
		m[a.Type] = a.Vals
	}
	return m
}

func auditEventTypes(events []*models.AuditEvent) []string {
	out := make([]string, 0, len(events))
	for _, e := range events {
		out = append(out, e.EventType)
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

// silence unused-import warnings when the test file compiles with strconv
// used only in fixtures.
var _ = strconv.Itoa
