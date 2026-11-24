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
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/apache/airavata-custos/signer/internal/vault"
	"github.com/go-ldap/ldap/v3"
)

// mockCredentialFetcher implements CredentialFetcher for testing.
type mockCredentialFetcher struct {
	creds     *vault.ValidationCredentials
	err       error
	callCount int
}

func (m *mockCredentialFetcher) GetValidationCredentials(ctx context.Context) (*vault.ValidationCredentials, error) {
	m.callCount++
	return m.creds, m.err
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// Make sure no validator accepts the requested principal as is.
func TestDispatcher_NoneSource(t *testing.T) {
	d := NewValidatorDispatcher(nil, nil, "none", time.Minute, testLogger())

	result, err := d.Validate(context.Background(), "jdoe", "sub1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("expected Allowed=true for none")
	}
	if result.ValidatedPrincipal != "jdoe" {
		t.Errorf("expected principal jdoe, got %s", result.ValidatedPrincipal)
	}
}

// Make sure the ldap validator allows a principal that matches the directory entry.
func TestDispatcher_LDAPSource_PrincipalFound(t *testing.T) {
	boolTrue := true
	fetcher := &mockCredentialFetcher{
		creds: &vault.ValidationCredentials{
			Type:         "ldap",
			LDAPUrl:      "ldaps://ldap.test:636",
			BindDN:       "cn=reader,dc=test",
			BindPassword: "secret",
			BaseDN:       "ou=people,dc=test",
			SearchFilter: "(&(uid=%s)(objectClass=posixAccount))",
			VerifySSL:    &boolTrue,
		},
	}
	conn := &mockLDAPConn{
		searchRes: &ldap.SearchResult{
			Entries: []*ldap.Entry{ldap.NewEntry("uid=jdoe,ou=people,dc=test", map[string][]string{"uid": {"jdoe"}})},
		},
	}
	connector := &mockLDAPConnector{conn: conn}
	d := NewValidatorDispatcher(fetcher, connector, "ldap", time.Minute, testLogger())

	result, err := d.Validate(context.Background(), "jdoe", "sub1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("expected Allowed=true")
	}
}

// Make sure the principal is refused when the ldap validator has no settings stored.
func TestDispatcher_LDAPSource_NoCredentials(t *testing.T) {
	fetcher := &mockCredentialFetcher{creds: nil}
	d := NewValidatorDispatcher(fetcher, nil, "ldap", time.Minute, testLogger())

	_, err := d.Validate(context.Background(), "jdoe", "sub1")
	if err == nil {
		t.Fatal("expected error")
	}
	valErr := err.(*ValidationError)
	if valErr.ReasonCode != "VALIDATION_NOT_CONFIGURED" {
		t.Errorf("expected VALIDATION_NOT_CONFIGURED, got %s", valErr.ReasonCode)
	}
}

// Make sure a failed read of the validation settings is reported as the
// directory being unavailable, not as a denied principal.
func TestDispatcher_LDAPSource_VaultError(t *testing.T) {
	fetcher := &mockCredentialFetcher{err: errors.New("vault sealed")}
	d := NewValidatorDispatcher(fetcher, nil, "ldap", time.Minute, testLogger())

	_, err := d.Validate(context.Background(), "jdoe", "sub1")
	if err == nil {
		t.Fatal("expected error")
	}
	valErr := err.(*ValidationError)
	if valErr.ReasonCode != "LDAP_UNAVAILABLE" {
		t.Errorf("expected LDAP_UNAVAILABLE, got %s", valErr.ReasonCode)
	}
}

// Make sure the validation settings are read once and reused for later requests.
func TestDispatcher_CacheHit(t *testing.T) {
	boolTrue := true
	fetcher := &mockCredentialFetcher{
		creds: &vault.ValidationCredentials{
			Type:         "ldap",
			LDAPUrl:      "ldaps://ldap.test:636",
			BindDN:       "cn=reader,dc=test",
			BindPassword: "secret",
			BaseDN:       "ou=people,dc=test",
			SearchFilter: "(&(uid=%s)(objectClass=posixAccount))",
			VerifySSL:    &boolTrue,
		},
	}
	conn := &mockLDAPConn{
		searchRes: &ldap.SearchResult{
			Entries: []*ldap.Entry{ldap.NewEntry("uid=jdoe,ou=people,dc=test", map[string][]string{"uid": {"jdoe"}})},
		},
	}
	connector := &mockLDAPConnector{conn: conn}
	d := NewValidatorDispatcher(fetcher, connector, "ldap", time.Minute, testLogger())

	// First call — fetches from Vault
	d.Validate(context.Background(), "jdoe", "sub1")
	// Second call — should use cache
	d.Validate(context.Background(), "jdoe", "sub1")

	if fetcher.callCount != 1 {
		t.Errorf("expected 1 Vault fetch (cached), got %d", fetcher.callCount)
	}
}

// Make sure the validation settings are read again once the cache TTL has passed.
func TestDispatcher_CacheExpiry(t *testing.T) {
	boolTrue := true
	fetcher := &mockCredentialFetcher{
		creds: &vault.ValidationCredentials{
			Type:         "ldap",
			LDAPUrl:      "ldaps://ldap.test:636",
			BindDN:       "cn=reader,dc=test",
			BindPassword: "secret",
			BaseDN:       "ou=people,dc=test",
			SearchFilter: "(&(uid=%s)(objectClass=posixAccount))",
			VerifySSL:    &boolTrue,
		},
	}
	conn := &mockLDAPConn{
		searchRes: &ldap.SearchResult{
			Entries: []*ldap.Entry{ldap.NewEntry("uid=jdoe,ou=people,dc=test", map[string][]string{"uid": {"jdoe"}})},
		},
	}
	connector := &mockLDAPConnector{conn: conn}
	d := NewValidatorDispatcher(fetcher, connector, "ldap", 1*time.Millisecond, testLogger())

	// First call
	d.Validate(context.Background(), "jdoe", "sub1")
	// Wait for cache to expire
	time.Sleep(5 * time.Millisecond)
	// Second call — should re-fetch
	d.Validate(context.Background(), "jdoe", "sub1")

	if fetcher.callCount != 2 {
		t.Errorf("expected 2 Vault fetches (cache expired), got %d", fetcher.callCount)
	}
}

// Make sure an unknown validator name refuses every request.
func TestDispatcher_UnknownSource(t *testing.T) {
	d := NewValidatorDispatcher(nil, nil, "unknown", time.Minute, testLogger())

	_, err := d.Validate(context.Background(), "jdoe", "sub1")
	if err == nil {
		t.Fatal("expected error")
	}
	valErr := err.(*ValidationError)
	if valErr.ReasonCode != "UNKNOWN_PRINCIPAL_SOURCE" {
		t.Errorf("expected UNKNOWN_PRINCIPAL_SOURCE, got %s", valErr.ReasonCode)
	}
}
