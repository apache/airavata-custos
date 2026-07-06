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

package client

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/go-ldap/ldap/v3"
)

type fakeConn struct {
	bindErr, addErr, modifyErr, searchErr error
	searchResult                          *ldap.SearchResult
	// per-base-DN responses so a single fake can back both
	// posixAccount searches (BaseDN) and posixGroup searches
	// (GroupBaseDN).
	perBase map[string]*ldap.SearchResult

	// Sequential Add errors (first Add uses [0], second uses [1], etc.)
	// so tests can simulate a constraint violation followed by success.
	pendingAddErrs []error

	bindCalls   int
	addCalls    []*ldap.AddRequest
	modifyCalls []*ldap.ModifyRequest
	searchCalls []*ldap.SearchRequest
}

func (f *fakeConn) Bind(_, _ string) error { f.bindCalls++; return f.bindErr }
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
	if res, ok := f.perBase[r.BaseDN]; ok {
		return res, nil
	}
	if f.searchResult != nil {
		return f.searchResult, nil
	}
	return &ldap.SearchResult{}, nil
}
func (f *fakeConn) Close() error { return nil }

type fakeDialer struct {
	conn    *fakeConn
	dialErr error
	calls   int
}

func (d *fakeDialer) Dial(_ string, _ bool, _ time.Duration) (Connection, error) {
	d.calls++
	if d.dialErr != nil {
		return nil, d.dialErr
	}
	return d.conn, nil
}

func sampleConfig() Config {
	return Config{
		URL:          "ldaps://ldap.test:636",
		BindDN:       "cn=writer,dc=test",
		BindPassword: "secret",
		BaseDN:       "ou=people,dc=test",
		VerifySSL:    true,
		Timeout:      5 * time.Second,
	}
}

func sampleAccount() PosixAccount {
	return PosixAccount{
		UID:           "alovelace",
		UIDNumber:     50123,
		GIDNumber:     50123,
		GivenName:     "Ada",
		Surname:       "Lovelace",
		Mail:          "ada@example.edu",
		HomeDirectory: "/home/alovelace",
		LoginShell:    "/bin/bash",
	}
}

// ---- Config / basic Find/Add/Modify -----------------------------------

func TestNew_ValidatesConfig(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
	}{
		{"missing URL", Config{BindDN: "cn=x", BaseDN: "dc=test"}},
		{"missing BindDN", Config{URL: "ldap://x", BaseDN: "dc=test"}},
		{"missing BaseDN", Config{URL: "ldap://x", BindDN: "cn=x"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := New(tc.cfg); err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestFindPosixAccount_ReturnsEmptyWhenAbsent(t *testing.T) {
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: &fakeConn{}})
	dn, attrs, err := c.FindPosixAccount("alovelace")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if dn != "" || attrs != nil {
		t.Errorf("expected empty result, got dn=%q attrs=%v", dn, attrs)
	}
}

func TestFindPosixAccount_ReturnsEntryWhenPresent(t *testing.T) {
	entry := ldap.NewEntry("uid=alovelace,ou=people,dc=test", map[string][]string{
		"uid": {"alovelace"}, "uidNumber": {"50123"},
	})
	conn := &fakeConn{searchResult: &ldap.SearchResult{Entries: []*ldap.Entry{entry}}}
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: conn})

	dn, attrs, err := c.FindPosixAccount("alovelace")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if dn != "uid=alovelace,ou=people,dc=test" {
		t.Errorf("unexpected DN: %s", dn)
	}
	if got := attrs["uidNumber"]; len(got) != 1 || got[0] != "50123" {
		t.Errorf("unexpected uidNumber attr: %v", got)
	}
}

func TestFindPosixAccount_ErrorsOnDuplicateHits(t *testing.T) {
	e1 := ldap.NewEntry("uid=alovelace,ou=people,dc=test", map[string][]string{"uid": {"alovelace"}})
	e2 := ldap.NewEntry("uid=alovelace,ou=other,dc=test", map[string][]string{"uid": {"alovelace"}})
	conn := &fakeConn{searchResult: &ldap.SearchResult{Entries: []*ldap.Entry{e1, e2}}}
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: conn})

	if _, _, err := c.FindPosixAccount("alovelace"); err == nil {
		t.Fatal("expected error on duplicate results")
	}
}

func TestAddPosixAccount_WritesRequiredAttributes(t *testing.T) {
	conn := &fakeConn{}
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: conn})

	dn, err := c.AddPosixAccount(sampleAccount())
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if dn != "uid=alovelace,ou=people,dc=test" {
		t.Errorf("unexpected DN: %s", dn)
	}
	if len(conn.addCalls) != 1 {
		t.Fatalf("expected 1 Add call, got %d", len(conn.addCalls))
	}
	got := indexAttrs(conn.addCalls[0].Attributes)
	for _, k := range []string{"objectClass", "uid", "cn", "sn", "uidNumber", "gidNumber", "homeDirectory"} {
		if _, ok := got[k]; !ok {
			t.Errorf("expected attribute %s to be set", k)
		}
	}
}

func TestAddPosixAccount_RejectsMissingRequired(t *testing.T) {
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: &fakeConn{}})
	cases := []struct {
		name string
		mut  func(*PosixAccount)
	}{
		{"empty UID", func(a *PosixAccount) { a.UID = "" }},
		{"zero UIDNumber", func(a *PosixAccount) { a.UIDNumber = 0 }},
		{"zero GIDNumber", func(a *PosixAccount) { a.GIDNumber = 0 }},
		{"empty HomeDirectory", func(a *PosixAccount) { a.HomeDirectory = "" }},
		{"empty Surname", func(a *PosixAccount) { a.Surname = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := sampleAccount()
			tc.mut(&a)
			if _, err := c.AddPosixAccount(a); err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestModifyPosixAccount_ReplacesMutableAttributes(t *testing.T) {
	conn := &fakeConn{}
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: conn})

	if err := c.ModifyPosixAccount("uid=alovelace,ou=people,dc=test", sampleAccount()); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(conn.modifyCalls) != 1 {
		t.Fatalf("expected 1 Modify call, got %d", len(conn.modifyCalls))
	}
	replaced := map[string]bool{}
	for _, ch := range conn.modifyCalls[0].Changes {
		if ch.Operation == ldap.ReplaceAttribute {
			replaced[ch.Modification.Type] = true
		}
	}
	for _, k := range []string{"cn", "sn", "homeDirectory"} {
		if !replaced[k] {
			t.Errorf("expected %s to be Replace'd", k)
		}
	}
	if replaced["uidNumber"] {
		t.Error("uidNumber must not be modified by ModifyPosixAccount")
	}
}

// ---- Allocation and atomicity ----------------------------------------

func TestAllocateNextUID_ReturnsMinUIDWhenDirectoryEmpty(t *testing.T) {
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: &fakeConn{}})
	got, err := c.AllocateNextUID(50000)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got != 50000 {
		t.Errorf("got %d, want 50000", got)
	}
}

func TestAllocateNextUID_ReturnsMaxPlusOne(t *testing.T) {
	entries := []*ldap.Entry{
		ldap.NewEntry("uid=alice,ou=people,dc=test", map[string][]string{"uidNumber": {"50000"}}),
		ldap.NewEntry("uid=bob,ou=people,dc=test", map[string][]string{"uidNumber": {"50123"}}),
	}
	conn := &fakeConn{searchResult: &ldap.SearchResult{Entries: entries}}
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: conn})

	got, err := c.AllocateNextUID(50000)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got != 50124 {
		t.Errorf("got %d, want 50124", got)
	}
}

func TestAllocateNextUID_UsesMinUIDWhenMaxBelowFloor(t *testing.T) {
	entries := []*ldap.Entry{
		ldap.NewEntry("uid=oldsys,ou=people,dc=test", map[string][]string{"uidNumber": {"1234"}}),
	}
	conn := &fakeConn{searchResult: &ldap.SearchResult{Entries: entries}}
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: conn})

	got, _ := c.AllocateNextUID(50000)
	if got != 50000 {
		t.Errorf("got %d, want 50000 (floor should win)", got)
	}
}

func TestAllocateNextUID_IgnoresNonNumericValues(t *testing.T) {
	entries := []*ldap.Entry{
		ldap.NewEntry("uid=broken,ou=people,dc=test", map[string][]string{"uidNumber": {"not-a-number"}}),
		ldap.NewEntry("uid=alice,ou=people,dc=test", map[string][]string{"uidNumber": {"50100"}}),
	}
	conn := &fakeConn{searchResult: &ldap.SearchResult{Entries: entries}}
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: conn})

	got, _ := c.AllocateNextUID(50000)
	if got != 50101 {
		t.Errorf("got %d, want 50101", got)
	}
}

func TestAllocateNextUID_DefaultsToDefaultMinUIDWhenZero(t *testing.T) {
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: &fakeConn{}})
	got, _ := c.AllocateNextUID(0)
	if got != DefaultMinUID {
		t.Errorf("got %d, want DefaultMinUID (%d)", got, DefaultMinUID)
	}
}

func TestAllocateAndAddPosixAccount_AllocatesAndWrites(t *testing.T) {
	conn := &fakeConn{}
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: conn})

	uid, dn, err := c.AllocateAndAddPosixAccount(50000, func(uid int64) PosixAccount {
		a := sampleAccount()
		a.UIDNumber = uid
		a.GIDNumber = uid
		return a
	})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if uid != 50000 {
		t.Errorf("uid: got %d, want 50000", uid)
	}
	if dn != "uid=alovelace,ou=people,dc=test" {
		t.Errorf("dn: got %s", dn)
	}
	if len(conn.searchCalls) != 1 {
		t.Errorf("expected 1 Search (allocator scan), got %d", len(conn.searchCalls))
	}
	if len(conn.addCalls) != 1 {
		t.Errorf("expected 1 Add, got %d", len(conn.addCalls))
	}
}

func TestAllocateAndAddPosixAccount_ReturnsAttemptedUIDOnAddError(t *testing.T) {
	conn := &fakeConn{
		pendingAddErrs: []error{&ldap.Error{ResultCode: ldap.LDAPResultConstraintViolation}},
	}
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: conn})

	uid, _, err := c.AllocateAndAddPosixAccount(50000, func(uid int64) PosixAccount {
		a := sampleAccount()
		a.UIDNumber = uid
		a.GIDNumber = uid
		return a
	})
	if err == nil {
		t.Fatal("expected error to propagate")
	}
	if !IsConstraintViolation(err) {
		t.Errorf("expected IsConstraintViolation, got %v", err)
	}
	// Attempted uid returned so the caller can bump the floor for the
	// next retry.
	if uid != 50000 {
		t.Errorf("attempted uid: got %d, want 50000", uid)
	}
}

// ---- posixGroup ------------------------------------------------------

func groupConfig() Config {
	c := sampleConfig()
	c.GroupBaseDN = "ou=groups,dc=test"
	return c
}

func TestFindPosixGroup_ErrorsWhenGroupBaseDNMissing(t *testing.T) {
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: &fakeConn{}})
	if _, err := c.FindPosixGroup("alovelace"); err == nil {
		t.Fatal("expected error when GroupBaseDN not configured")
	}
}

func TestFindPosixGroup_ReturnsEmptyWhenAbsent(t *testing.T) {
	c := NewWithDialer(groupConfig(), &fakeDialer{conn: &fakeConn{}})
	dn, err := c.FindPosixGroup("alovelace")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if dn != "" {
		t.Errorf("expected empty, got %s", dn)
	}
}

func TestFindPosixGroup_ReturnsDNWhenPresent(t *testing.T) {
	entry := ldap.NewEntry("cn=alovelace,ou=groups,dc=test", map[string][]string{
		"cn": {"alovelace"}, "gidNumber": {"50123"},
	})
	conn := &fakeConn{perBase: map[string]*ldap.SearchResult{
		"ou=groups,dc=test": {Entries: []*ldap.Entry{entry}},
	}}
	c := NewWithDialer(groupConfig(), &fakeDialer{conn: conn})

	dn, err := c.FindPosixGroup("alovelace")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if dn != "cn=alovelace,ou=groups,dc=test" {
		t.Errorf("got %s", dn)
	}
}

func TestAddPosixGroup_ErrorsWhenGroupBaseDNMissing(t *testing.T) {
	c := NewWithDialer(sampleConfig(), &fakeDialer{conn: &fakeConn{}})
	if _, err := c.AddPosixGroup("alovelace", 50123); err == nil {
		t.Fatal("expected error when GroupBaseDN not configured")
	}
}

func TestAddPosixGroup_WritesRequiredAttrs(t *testing.T) {
	conn := &fakeConn{}
	c := NewWithDialer(groupConfig(), &fakeDialer{conn: conn})

	dn, err := c.AddPosixGroup("alovelace", 50123)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if dn != "cn=alovelace,ou=groups,dc=test" {
		t.Errorf("got %s", dn)
	}
	if len(conn.addCalls) != 1 {
		t.Fatalf("expected 1 Add call, got %d", len(conn.addCalls))
	}
	got := indexAttrs(conn.addCalls[0].Attributes)
	if got["gidNumber"][0] != "50123" {
		t.Errorf("gidNumber: got %s", got["gidNumber"][0])
	}
	if got["cn"][0] != "alovelace" {
		t.Errorf("cn: got %s", got["cn"][0])
	}
	if !contains(got["objectClass"], "posixGroup") {
		t.Errorf("objectClass missing posixGroup: %v", got["objectClass"])
	}
}

func TestAddPosixGroup_RejectsZeroGID(t *testing.T) {
	c := NewWithDialer(groupConfig(), &fakeDialer{conn: &fakeConn{}})
	if _, err := c.AddPosixGroup("alovelace", 0); err == nil {
		t.Fatal("expected error for zero gidNumber")
	}
}

// ---- error helpers ---------------------------------------------------

func TestIsConstraintViolation(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"unrelated", errors.New("io"), false},
		{"constraint", &ldap.Error{ResultCode: ldap.LDAPResultConstraintViolation}, true},
		{"attr or value exists", &ldap.Error{ResultCode: ldap.LDAPResultAttributeOrValueExists}, true},
		{"entry already exists", &ldap.Error{ResultCode: ldap.LDAPResultEntryAlreadyExists}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsConstraintViolation(tc.err); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsAlreadyExists(t *testing.T) {
	if !IsAlreadyExists(&ldap.Error{ResultCode: ldap.LDAPResultEntryAlreadyExists}) {
		t.Error("expected true for EntryAlreadyExists")
	}
	if IsAlreadyExists(&ldap.Error{ResultCode: ldap.LDAPResultConstraintViolation}) {
		t.Error("constraint violation is not the same as already exists")
	}
}

// ---- connection reuse ------------------------------------------------

func TestSearchTimeout_ClampsSubSecondToOne(t *testing.T) {
	// go-ldap treats TimeLimit=0 as "no time limit" — the opposite of
	// what a caller who configured 500ms wants. The client must clamp
	// to at least 1s so a sub-second Timeout doesn't silently disable
	// the server-side time limit.
	cfg := sampleConfig()
	cfg.Timeout = 500 * time.Millisecond
	conn := &fakeConn{}
	c := NewWithDialer(cfg, &fakeDialer{conn: conn})

	_, _, _ = c.FindPosixAccount("alice")

	if len(conn.searchCalls) != 1 {
		t.Fatalf("expected 1 Search, got %d", len(conn.searchCalls))
	}
	if got := conn.searchCalls[0].TimeLimit; got < 1 {
		t.Errorf("TimeLimit must be clamped to >=1 (0 would mean unlimited on the server); got %d", got)
	}
}

func TestSearchTimeout_UsesConfiguredSecondsWhenAboveOne(t *testing.T) {
	cfg := sampleConfig()
	cfg.Timeout = 7 * time.Second
	conn := &fakeConn{}
	c := NewWithDialer(cfg, &fakeDialer{conn: conn})

	_, _, _ = c.FindPosixAccount("alice")

	if got := conn.searchCalls[0].TimeLimit; got != 7 {
		t.Errorf("TimeLimit: got %d, want 7", got)
	}
}

func TestClient_ReusesConnection(t *testing.T) {
	conn := &fakeConn{}
	d := &fakeDialer{conn: conn}
	c := NewWithDialer(sampleConfig(), d)

	_, _, _ = c.FindPosixAccount("alovelace")
	_, _, _ = c.FindPosixAccount("bob")

	if d.calls != 1 {
		t.Errorf("expected 1 Dial (reuse); got %d", d.calls)
	}
	if conn.bindCalls != 1 {
		t.Errorf("expected 1 Bind (reuse); got %d", conn.bindCalls)
	}
}

func TestClient_ReconnectsAfterError(t *testing.T) {
	conn := &fakeConn{addErr: errors.New("io error")}
	d := &fakeDialer{conn: conn}
	c := NewWithDialer(sampleConfig(), d)

	_, _ = c.AddPosixAccount(sampleAccount())
	conn.addErr = nil
	_, _ = c.AddPosixAccount(sampleAccount())

	if d.calls != 2 {
		t.Errorf("expected 2 Dial (reconnect after error); got %d", d.calls)
	}
}

// ---- helpers ---------------------------------------------------------

func indexAttrs(attrs []ldap.Attribute) map[string][]string {
	m := make(map[string][]string, len(attrs))
	for _, a := range attrs {
		m[a.Type] = a.Vals
	}
	return m
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

// silence unused-strconv warning
var _ = strconv.Itoa
