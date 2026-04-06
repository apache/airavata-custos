// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validation

import (
	"errors"
	"testing"

	"github.com/go-ldap/ldap/v3"
)

// mockLDAPConn implements LDAPConnection for testing.
type mockLDAPConn struct {
	bindErr   error
	searchRes *ldap.SearchResult
	searchErr error
}

func (m *mockLDAPConn) Bind(username, password string) error { return m.bindErr }
func (m *mockLDAPConn) Search(req *ldap.SearchRequest) (*ldap.SearchResult, error) {
	return m.searchRes, m.searchErr
}
func (m *mockLDAPConn) Close() error { return nil }

// mockLDAPConnector implements LDAPConnector for testing.
type mockLDAPConnector struct {
	conn    LDAPConnection
	connErr error
}

func (m *mockLDAPConnector) Connect(url string, verifySSL bool) (LDAPConnection, error) {
	return m.conn, m.connErr
}

func baseLDAPConfig() LDAPConfig {
	return LDAPConfig{
		URL:               "ldaps://ldap.test:636",
		BindDN:            "cn=reader,dc=test",
		BindPassword:      "secret",
		BaseDN:            "ou=people,dc=test",
		SearchFilter:      "(&(objectClass=posixAccount)(voPersonExternalID=%s))",
		UsernameAttribute: "uid",
		VerifySSL:         true,
	}
}

// ldapEntry creates an LDAP entry with a uid attribute for testing.
func ldapEntry(dn, uid string) *ldap.Entry {
	return ldap.NewEntry(dn, map[string][]string{
		"uid": {uid},
	})
}

func TestLDAPValidator_SubjectResolvesToPrincipal(t *testing.T) {
	conn := &mockLDAPConn{
		searchRes: &ldap.SearchResult{
			Entries: []*ldap.Entry{ldapEntry("uid=jdoe,ou=people,dc=test", "jdoe")},
		},
	}
	connector := &mockLDAPConnector{conn: conn}
	v := NewLDAPValidator(baseLDAPConfig(), connector)

	// OIDC subject resolves to uid=jdoe, requested principal is jdoe → match
	result, err := v.Validate("tenant1", "client1", "jdoe", "http://cilogon.org/serverE/users/100001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("expected Allowed=true")
	}
	if result.ValidatedPrincipal != "jdoe" {
		t.Errorf("expected principal jdoe, got %s", result.ValidatedPrincipal)
	}
}

func TestLDAPValidator_PrincipalMismatch(t *testing.T) {
	conn := &mockLDAPConn{
		searchRes: &ldap.SearchResult{
			Entries: []*ldap.Entry{ldapEntry("uid=jdoe,ou=people,dc=test", "jdoe")},
		},
	}
	connector := &mockLDAPConnector{conn: conn}
	v := NewLDAPValidator(baseLDAPConfig(), connector)

	// OIDC subject resolves to jdoe, but caller requested "admin" → mismatch
	_, err := v.Validate("tenant1", "client1", "admin", "http://cilogon.org/serverE/users/100001")
	if err == nil {
		t.Fatal("expected error for principal mismatch")
	}
	valErr := err.(*ValidationError)
	if valErr.ReasonCode != "LDAP_PRINCIPAL_MISMATCH" {
		t.Errorf("expected LDAP_PRINCIPAL_MISMATCH, got %s", valErr.ReasonCode)
	}
}

func TestLDAPValidator_IdentityNotFound(t *testing.T) {
	conn := &mockLDAPConn{
		searchRes: &ldap.SearchResult{Entries: []*ldap.Entry{}},
	}
	connector := &mockLDAPConnector{conn: conn}
	v := NewLDAPValidator(baseLDAPConfig(), connector)

	_, err := v.Validate("tenant1", "client1", "jdoe", "http://cilogon.org/serverE/users/999999")
	if err == nil {
		t.Fatal("expected error")
	}
	valErr := err.(*ValidationError)
	if valErr.ReasonCode != "LDAP_IDENTITY_NOT_FOUND" {
		t.Errorf("expected LDAP_IDENTITY_NOT_FOUND, got %s", valErr.ReasonCode)
	}
}

func TestLDAPValidator_MissingUsernameAttribute(t *testing.T) {
	// Entry exists but has no uid attribute
	entry := ldap.NewEntry("cn=someone,ou=people,dc=test", map[string][]string{
		"cn": {"someone"},
	})
	conn := &mockLDAPConn{
		searchRes: &ldap.SearchResult{Entries: []*ldap.Entry{entry}},
	}
	connector := &mockLDAPConnector{conn: conn}
	v := NewLDAPValidator(baseLDAPConfig(), connector)

	_, err := v.Validate("tenant1", "client1", "someone", "sub123")
	if err == nil {
		t.Fatal("expected error for missing uid attribute")
	}
	valErr := err.(*ValidationError)
	if valErr.ReasonCode != "LDAP_USERNAME_MISSING" {
		t.Errorf("expected LDAP_USERNAME_MISSING, got %s", valErr.ReasonCode)
	}
}

func TestLDAPValidator_ConnectionFailure(t *testing.T) {
	connector := &mockLDAPConnector{connErr: errors.New("connection refused")}
	v := NewLDAPValidator(baseLDAPConfig(), connector)

	_, err := v.Validate("t1", "c1", "jdoe", "sub123")
	if err == nil {
		t.Fatal("expected error")
	}
	valErr := err.(*ValidationError)
	if valErr.ReasonCode != "LDAP_UNAVAILABLE" {
		t.Errorf("expected LDAP_UNAVAILABLE, got %s", valErr.ReasonCode)
	}
}

func TestLDAPValidator_BindFailure(t *testing.T) {
	conn := &mockLDAPConn{bindErr: errors.New("invalid credentials")}
	connector := &mockLDAPConnector{conn: conn}
	v := NewLDAPValidator(baseLDAPConfig(), connector)

	_, err := v.Validate("t1", "c1", "jdoe", "sub123")
	if err == nil {
		t.Fatal("expected error")
	}
	valErr := err.(*ValidationError)
	if valErr.ReasonCode != "LDAP_UNAVAILABLE" {
		t.Errorf("expected LDAP_UNAVAILABLE, got %s", valErr.ReasonCode)
	}
}

func TestLDAPValidator_SearchError(t *testing.T) {
	conn := &mockLDAPConn{searchErr: errors.New("search timeout")}
	connector := &mockLDAPConnector{conn: conn}
	v := NewLDAPValidator(baseLDAPConfig(), connector)

	_, err := v.Validate("t1", "c1", "jdoe", "sub123")
	if err == nil {
		t.Fatal("expected error")
	}
	valErr := err.(*ValidationError)
	if valErr.ReasonCode != "LDAP_UNAVAILABLE" {
		t.Errorf("expected LDAP_UNAVAILABLE, got %s", valErr.ReasonCode)
	}
}

func TestLDAPValidator_FilterEscaping(t *testing.T) {
	var capturedFilter string
	conn := &mockLDAPConn{
		searchRes: &ldap.SearchResult{Entries: []*ldap.Entry{}},
	}
	wrappedConn := &filterCapturingConn{
		LDAPConnection: conn,
		capturedFilter: &capturedFilter,
	}
	connector := &mockLDAPConnector{conn: wrappedConn}
	v := NewLDAPValidator(baseLDAPConfig(), connector)

	// OIDC subject with injection characters — should be escaped in filter
	v.Validate("t1", "c1", "jdoe", "http://evil.org/)(objectClass=*)")

	expected := `(&(objectClass=posixAccount)(voPersonExternalID=http://evil.org/\29\28objectClass=\2a\29))`
	if capturedFilter != expected {
		t.Errorf("expected escaped filter %q, got %q", expected, capturedFilter)
	}
}

func TestLDAPValidator_DefaultUsernameAttribute(t *testing.T) {
	cfg := baseLDAPConfig()
	cfg.UsernameAttribute = "" // should default to "uid"
	conn := &mockLDAPConn{
		searchRes: &ldap.SearchResult{
			Entries: []*ldap.Entry{ldapEntry("uid=jdoe,ou=people,dc=test", "jdoe")},
		},
	}
	connector := &mockLDAPConnector{conn: conn}
	v := NewLDAPValidator(cfg, connector)

	result, err := v.Validate("t1", "c1", "jdoe", "sub123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ValidatedPrincipal != "jdoe" {
		t.Errorf("expected jdoe, got %s", result.ValidatedPrincipal)
	}
}

// filterCapturingConn wraps LDAPConnection to capture the search filter.
type filterCapturingConn struct {
	LDAPConnection
	capturedFilter *string
}

func (c *filterCapturingConn) Search(req *ldap.SearchRequest) (*ldap.SearchResult, error) {
	*c.capturedFilter = req.Filter
	return c.LDAPConnection.Search(req)
}
