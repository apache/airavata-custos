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
	"log/slog"
	"sync"
	"time"

	"github.com/apache/airavata-custos/signer/internal/vault"
)

// CredentialFetcher abstracts Vault credential reads for validating.
type CredentialFetcher interface {
	GetValidationCredentials(ctx context.Context) (*vault.ValidationCredentials, error)
}

type cachedCreds struct {
	creds     *vault.ValidationCredentials
	err       error
	fetchedAt time.Time
}

// ValidatorDispatcher runs the deployment's principal validator (noop, ldap, comanage).
type ValidatorDispatcher struct {
	credFetcher   CredentialFetcher
	ldapConnector LDAPConnector
	source        string
	cacheTTL      time.Duration
	negativeTTL   time.Duration
	cached        *cachedCreds
	ldap          *LDAPValidator
	mu            sync.RWMutex
	logger        *slog.Logger
}

func NewValidatorDispatcher(
	credFetcher CredentialFetcher,
	ldapConnector LDAPConnector,
	source string,
	cacheTTL time.Duration,
	logger *slog.Logger,
) *ValidatorDispatcher {
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Minute
	}
	return &ValidatorDispatcher{
		credFetcher:   credFetcher,
		ldapConnector: ldapConnector,
		source:        source,
		cacheTTL:      cacheTTL,
		negativeTTL:   30 * time.Second,
		logger:        logger,
	}
}

// ValidateForClient checks the principal with the deployment's validator.
// Called directly by the sign handler.
func (d *ValidatorDispatcher) ValidateForClient(ctx context.Context, clientID, principal, identitySubject string) (*ValidationResult, error) {
	switch d.source {
	case "noop", "":
		return &ValidationResult{
			Allowed:            true,
			ValidatedPrincipal: principal,
		}, nil

	case "ldap":
		return d.validateLDAP(ctx, clientID, principal, identitySubject)

	case "comanage":
		return nil, &ValidationError{
			Message:    "COmanage principal validation is not yet implemented",
			ReasonCode: "COMANAGE_NOT_IMPLEMENTED",
		}

	default:
		return nil, &ValidationError{
			Message:    "Unknown principal source: " + d.source,
			ReasonCode: "UNKNOWN_PRINCIPAL_SOURCE",
		}
	}
}

func (d *ValidatorDispatcher) validateLDAP(ctx context.Context, clientID, principal, identitySubject string) (*ValidationResult, error) {
	creds, err := d.fetchCredentials(ctx)
	if err != nil {
		d.logger.Error("failed to fetch validation credentials", "error", err)
		return nil, &ValidationError{
			Message:    "Principal validation unavailable: credential fetch failed",
			ReasonCode: "LDAP_UNAVAILABLE",
		}
	}
	if creds == nil {
		return nil, &ValidationError{
			Message:    "LDAP validation not configured",
			ReasonCode: "VALIDATION_NOT_CONFIGURED",
		}
	}

	validator := d.getOrCreateValidator(creds)
	return validator.Validate(clientID, principal, identitySubject)
}

func (d *ValidatorDispatcher) getOrCreateValidator(creds *vault.ValidationCredentials) *LDAPValidator {
	d.mu.RLock()
	if d.ldap != nil {
		d.mu.RUnlock()
		return d.ldap
	}
	d.mu.RUnlock()

	verifySSL := true
	if creds.VerifySSL != nil {
		verifySSL = *creds.VerifySSL
	}

	validator := NewLDAPValidator(LDAPConfig{
		URL:               creds.LDAPUrl,
		BindDN:            creds.BindDN,
		BindPassword:      creds.BindPassword,
		BaseDN:            creds.BaseDN,
		SearchFilter:      creds.SearchFilter,
		UsernameAttribute: creds.UsernameAttribute,
		VerifySSL:         verifySSL,
	}, d.ldapConnector)

	d.mu.Lock()
	defer d.mu.Unlock()
	// Double-check after acquiring write lock
	if d.ldap == nil {
		d.ldap = validator
	}
	return d.ldap
}

func (d *ValidatorDispatcher) fetchCredentials(ctx context.Context) (*vault.ValidationCredentials, error) {
	d.mu.RLock()
	if entry := d.cached; entry != nil {
		ttl := d.cacheTTL
		if entry.err != nil {
			ttl = d.negativeTTL
		}
		if time.Since(entry.fetchedAt) < ttl {
			d.mu.RUnlock()
			return entry.creds, entry.err
		}
	}
	d.mu.RUnlock()

	creds, err := d.credFetcher.GetValidationCredentials(ctx)

	d.mu.Lock()
	d.cached = &cachedCreds{
		creds:     creds,
		err:       err,
		fetchedAt: time.Now(),
	}
	d.mu.Unlock()

	return creds, err
}
