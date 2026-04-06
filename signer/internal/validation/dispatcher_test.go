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

func (m *mockCredentialFetcher) GetValidationCredentials(ctx context.Context, tenantID, clientID string) (*vault.ValidationCredentials, error) {
	m.callCount++
	return m.creds, m.err
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestDispatcher_NoopSource(t *testing.T) {
	d := NewValidatorDispatcher(nil, nil, "noop", time.Minute, testLogger())

	result, err := d.ValidateForClient(context.Background(), "t1", "c1", "jdoe", "sub1", "noop")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("expected Allowed=true for noop")
	}
	if result.ValidatedPrincipal != "jdoe" {
		t.Errorf("expected principal jdoe, got %s", result.ValidatedPrincipal)
	}
}

func TestDispatcher_EmptySourceUsesFallback(t *testing.T) {
	d := NewValidatorDispatcher(nil, nil, "noop", time.Minute, testLogger())

	result, err := d.ValidateForClient(context.Background(), "t1", "c1", "jdoe", "sub1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("expected Allowed=true for noop fallback")
	}
}

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
	d := NewValidatorDispatcher(fetcher, connector, "noop", time.Minute, testLogger())

	result, err := d.ValidateForClient(context.Background(), "t1", "c1", "jdoe", "sub1", "ldap")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("expected Allowed=true")
	}
}

func TestDispatcher_LDAPSource_NoCredentials(t *testing.T) {
	fetcher := &mockCredentialFetcher{creds: nil}
	d := NewValidatorDispatcher(fetcher, nil, "noop", time.Minute, testLogger())

	_, err := d.ValidateForClient(context.Background(), "t1", "c1", "jdoe", "sub1", "ldap")
	if err == nil {
		t.Fatal("expected error")
	}
	valErr := err.(*ValidationError)
	if valErr.ReasonCode != "VALIDATION_NOT_CONFIGURED" {
		t.Errorf("expected VALIDATION_NOT_CONFIGURED, got %s", valErr.ReasonCode)
	}
}

func TestDispatcher_LDAPSource_VaultError(t *testing.T) {
	fetcher := &mockCredentialFetcher{err: errors.New("vault sealed")}
	d := NewValidatorDispatcher(fetcher, nil, "noop", time.Minute, testLogger())

	_, err := d.ValidateForClient(context.Background(), "t1", "c1", "jdoe", "sub1", "ldap")
	if err == nil {
		t.Fatal("expected error")
	}
	valErr := err.(*ValidationError)
	if valErr.ReasonCode != "LDAP_UNAVAILABLE" {
		t.Errorf("expected LDAP_UNAVAILABLE, got %s", valErr.ReasonCode)
	}
}

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
	d := NewValidatorDispatcher(fetcher, connector, "noop", time.Minute, testLogger())

	// First call — fetches from Vault
	d.ValidateForClient(context.Background(), "t1", "c1", "jdoe", "sub1", "ldap")
	// Second call — should use cache
	d.ValidateForClient(context.Background(), "t1", "c1", "jdoe", "sub1", "ldap")

	if fetcher.callCount != 1 {
		t.Errorf("expected 1 Vault fetch (cached), got %d", fetcher.callCount)
	}
}

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
	d := NewValidatorDispatcher(fetcher, connector, "noop", 1*time.Millisecond, testLogger())

	// First call
	d.ValidateForClient(context.Background(), "t1", "c1", "jdoe", "sub1", "ldap")
	// Wait for cache to expire
	time.Sleep(5 * time.Millisecond)
	// Second call — should re-fetch
	d.ValidateForClient(context.Background(), "t1", "c1", "jdoe", "sub1", "ldap")

	if fetcher.callCount != 2 {
		t.Errorf("expected 2 Vault fetches (cache expired), got %d", fetcher.callCount)
	}
}

func TestDispatcher_UnknownSource(t *testing.T) {
	d := NewValidatorDispatcher(nil, nil, "noop", time.Minute, testLogger())

	_, err := d.ValidateForClient(context.Background(), "t1", "c1", "jdoe", "sub1", "unknown")
	if err == nil {
		t.Fatal("expected error")
	}
	valErr := err.(*ValidationError)
	if valErr.ReasonCode != "UNKNOWN_PRINCIPAL_SOURCE" {
		t.Errorf("expected UNKNOWN_PRINCIPAL_SOURCE, got %s", valErr.ReasonCode)
	}
}

func TestDispatcher_ComanageStub(t *testing.T) {
	d := NewValidatorDispatcher(nil, nil, "noop", time.Minute, testLogger())

	_, err := d.ValidateForClient(context.Background(), "t1", "c1", "jdoe", "sub1", "comanage")
	if err == nil {
		t.Fatal("expected error")
	}
	valErr := err.(*ValidationError)
	if valErr.ReasonCode != "COMANAGE_NOT_IMPLEMENTED" {
		t.Errorf("expected COMANAGE_NOT_IMPLEMENTED, got %s", valErr.ReasonCode)
	}
}
