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
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/go-ldap/ldap/v3"
)

// LDAPConnection abstracts LDAP operations for testability.
type LDAPConnection interface {
	Bind(username, password string) error
	Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error)
	Close() error
}

// LDAPConnector creates LDAP connections.
type LDAPConnector interface {
	Connect(url string, verifySSL bool) (LDAPConnection, error)
}

// DefaultLDAPConnector uses go-ldap/ldap to establish real LDAP connections.
type DefaultLDAPConnector struct{}

func NewDefaultLDAPConnector() *DefaultLDAPConnector {
	return &DefaultLDAPConnector{}
}

func (c *DefaultLDAPConnector) Connect(url string, verifySSL bool) (LDAPConnection, error) {
	conn, err := ldap.DialURL(url, ldap.DialWithTLSConfig(&tls.Config{
		InsecureSkipVerify: !verifySSL,
	}))
	if err != nil {
		return nil, fmt.Errorf("connecting to LDAP at %s: %w", url, err)
	}
	return conn, nil
}

// LDAPConfig holds the configuration needed to validate principals via LDAP.
//
// The lookup flow:
//  1. Search LDAP using the OIDC subject (identitySubject) against SearchFilter
//     e.g. (&(objectClass=posixAccount)(voPersonExternalID=%s))
//  2. Extract the POSIX username from UsernameAttribute (e.g. "uid")
//  3. Compare the extracted username with the requested principal
type LDAPConfig struct {
	URL               string
	BindDN            string
	BindPassword      string
	BaseDN            string
	SearchFilter      string // must contain %s — substituted with the OIDC subject
	UsernameAttribute string // LDAP attribute holding the POSIX username (e.g. "uid")
	VerifySSL         bool
}

// LDAPValidator resolves an OIDC subject to a POSIX username via LDAP directory lookup.
// It maintains a persistent LDAP connection that is reused across requests and
// automatically reconnects on failure.
type LDAPValidator struct {
	config    LDAPConfig
	connector LDAPConnector
	mu        sync.Mutex
	conn      LDAPConnection
}

func NewLDAPValidator(config LDAPConfig, connector LDAPConnector) *LDAPValidator {
	usernameAttr := config.UsernameAttribute
	if usernameAttr == "" {
		usernameAttr = "uid"
	}
	config.UsernameAttribute = usernameAttr
	return &LDAPValidator{
		config:    config,
		connector: connector,
	}
}

func (v *LDAPValidator) Validate(tenantID, clientID, principal, identitySubject string) (*ValidationResult, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	result, err := v.searchLDAP(identitySubject)
	if err != nil {
		// Connection may be stale — close, reconnect, and retry once
		v.closeConn()
		result, err = v.searchLDAP(identitySubject)
		if err != nil {
			v.closeConn()
			return nil, &ValidationError{
				Message:    "Principal validation unavailable: " + err.Error(),
				ReasonCode: "LDAP_UNAVAILABLE",
			}
		}
	}

	if len(result.Entries) == 0 {
		return nil, &ValidationError{
			Message:    "No POSIX account found for identity subject in directory",
			ReasonCode: "LDAP_IDENTITY_NOT_FOUND",
		}
	}

	resolvedUsername := result.Entries[0].GetAttributeValue(v.config.UsernameAttribute)
	if resolvedUsername == "" {
		return nil, &ValidationError{
			Message:    fmt.Sprintf("Directory entry missing %s attribute", v.config.UsernameAttribute),
			ReasonCode: "LDAP_USERNAME_MISSING",
		}
	}

	if resolvedUsername != principal {
		return nil, &ValidationError{
			Message:    fmt.Sprintf("Requested principal %q does not match directory account %q", principal, resolvedUsername),
			ReasonCode: "LDAP_PRINCIPAL_MISMATCH",
		}
	}

	return &ValidationResult{
		Allowed:            true,
		ValidatedPrincipal: resolvedUsername,
	}, nil
}

// ensureConn returns the existing connection or creates a new one with bind.
func (v *LDAPValidator) ensureConn() (LDAPConnection, error) {
	if v.conn != nil {
		return v.conn, nil
	}

	conn, err := v.connector.Connect(v.config.URL, v.config.VerifySSL)
	if err != nil {
		return nil, fmt.Errorf("LDAP connection failed")
	}

	if err := conn.Bind(v.config.BindDN, v.config.BindPassword); err != nil {
		conn.Close()
		return nil, fmt.Errorf("LDAP bind failed")
	}

	v.conn = conn
	return conn, nil
}

// searchLDAP executes the principal lookup on the current or newly established connection.
func (v *LDAPValidator) searchLDAP(identitySubject string) (*ldap.SearchResult, error) {
	conn, err := v.ensureConn()
	if err != nil {
		return nil, err
	}

	filter := fmt.Sprintf(v.config.SearchFilter, ldap.EscapeFilter(identitySubject))

	return conn.Search(ldap.NewSearchRequest(
		v.config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1,  // SizeLimit
		10, // TimeLimit (seconds)
		false,
		filter,
		[]string{v.config.UsernameAttribute},
		nil,
	))
}

func (v *LDAPValidator) closeConn() {
	if v.conn != nil {
		v.conn.Close()
		v.conn = nil
	}
}

// Close releases the persistent LDAP connection.
func (v *LDAPValidator) Close() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.closeConn()
}
