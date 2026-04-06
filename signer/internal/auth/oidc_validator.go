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

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/sync/singleflight"

	"github.com/apache/airavata-custos/signer/internal/config"
)

type UserIdentity struct {
	Issuer    string
	Subject   string
	Email     string
	Principal string // extracted label for logging
}

type OIDCValidator struct {
	enabled        bool
	defaultEmail   string
	defaultSubject string
	allowedIssuers []string
	cacheTTL       time.Duration
	maxProviders   int
	httpTimeout    time.Duration
	httpClient     *http.Client

	mu      sync.RWMutex
	cache   map[string]*jwksEntry
	sfGroup singleflight.Group
	// JWKSFetcher can be overridden for testing
	JWKSFetcher func(ctx context.Context, issuer string) (jwk.Set, error)
}

type jwksEntry struct {
	set       jwk.Set
	fetchedAt time.Time
}

func NewOIDCValidator(cfg config.AuthConfig, devMode config.DevModeConfig) *OIDCValidator {
	enabled := !devMode.Enabled
	defaultEmail := devMode.DefaultEmail
	if defaultEmail == "" {
		defaultEmail = "dev@localhost"
	}
	defaultSubject := devMode.DefaultSubject
	if defaultSubject == "" {
		defaultSubject = "dev-user"
	}

	v := &OIDCValidator{
		enabled:        enabled,
		defaultEmail:   defaultEmail,
		defaultSubject: defaultSubject,
		allowedIssuers: cfg.AllowedIssuers,
		cacheTTL:       time.Duration(cfg.OIDC.JWKSCacheTTLSeconds) * time.Second,
		maxProviders:   cfg.OIDC.JWKSMaxProviders,
		httpTimeout:    time.Duration(cfg.OIDC.TimeoutSeconds) * time.Second,
		cache:          make(map[string]*jwksEntry),
	}
	if v.cacheTTL == 0 {
		v.cacheTTL = 5 * time.Minute
	}
	if v.maxProviders == 0 {
		v.maxProviders = 10
	}
	if v.httpTimeout == 0 {
		v.httpTimeout = 10 * time.Second
	}
	v.httpClient = &http.Client{Timeout: v.httpTimeout}
	return v
}

// ValidateAccessToken returns a default identity when dev mode is active;
// otherwise performs full OIDC token validation via JWKS.
func (v *OIDCValidator) ValidateAccessToken(ctx context.Context, tokenString string) (*UserIdentity, error) {
	if !v.enabled {
		return &UserIdentity{
			Issuer:    "dev-mode",
			Subject:   v.defaultSubject,
			Email:     v.defaultEmail,
			Principal: v.defaultSubject,
		}, nil
	}

	if tokenString == "" {
		return nil, &AuthError{Code: "invalid_token", Message: "Missing access token", Status: 401}
	}

	// Parse without verification first to get issuer
	insecureToken, err := jwt.Parse([]byte(tokenString), jwt.WithVerify(false), jwt.WithValidate(false))
	if err != nil {
		return nil, &AuthError{Code: "invalid_token", Message: fmt.Sprintf("Failed to parse token: %v", err), Status: 401}
	}

	issuer := insecureToken.Issuer()
	if issuer == "" {
		return nil, &AuthError{Code: "invalid_token", Message: "Token missing issuer claim", Status: 401}
	}

	if len(v.allowedIssuers) > 0 {
		allowed := false
		for _, ai := range v.allowedIssuers {
			if ai == issuer {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, &AuthError{Code: "invalid_token", Message: "Issuer not allowed", Status: 401}
		}
	}

	keySet, err := v.getJWKS(ctx, issuer)
	if err != nil {
		return nil, &AuthError{Code: "invalid_token", Message: fmt.Sprintf("Failed to fetch JWKS: %v", err), Status: 401}
	}

	// Parse, verify signature, and validate claims in a single pass
	token, err := jwt.Parse([]byte(tokenString),
		jwt.WithKeySet(keySet),
		jwt.WithValidate(true),
	)
	if err != nil {
		errMsg := err.Error()
		if containsAny(errMsg, "exp", "expired") {
			return nil, &AuthError{Code: "invalid_token", Message: "Token has expired", Status: 401}
		}
		if containsAny(errMsg, "nbf", "not yet valid") {
			return nil, &AuthError{Code: "invalid_token", Message: "Token is not yet valid", Status: 401}
		}
		return nil, &AuthError{Code: "invalid_token", Message: fmt.Sprintf("Token validation failed: %v", err), Status: 401}
	}

	principal := extractPrincipal(token)
	if principal == "" {
		return nil, &AuthError{Code: "invalid_token", Message: "No valid principal found in token claims", Status: 401}
	}

	var email string
	if v, ok := token.Get("email"); ok {
		if s, ok := v.(string); ok {
			email = s
		}
	}

	return &UserIdentity{
		Issuer:    issuer,
		Subject:   token.Subject(),
		Email:     email,
		Principal: principal,
	}, nil
}

func extractPrincipal(token jwt.Token) string {
	// Priority: sub, preferred_username, email, username
	if s := token.Subject(); s != "" {
		return s
	}
	if v, ok := token.Get("preferred_username"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if v, ok := token.Get("email"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if v, ok := token.Get("username"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

func (v *OIDCValidator) getJWKS(ctx context.Context, issuer string) (jwk.Set, error) {
	v.mu.RLock()
	entry, ok := v.cache[issuer]
	v.mu.RUnlock()

	if ok && time.Since(entry.fetchedAt) < v.cacheTTL {
		return entry.set, nil
	}

	// Use singleflight to coalesce concurrent fetches for the same issuer
	result, err, _ := v.sfGroup.Do(issuer, func() (interface{}, error) {
		// Double-check cache after acquiring singleflight slot
		v.mu.RLock()
		entry2, ok2 := v.cache[issuer]
		v.mu.RUnlock()
		if ok2 && time.Since(entry2.fetchedAt) < v.cacheTTL {
			return entry2.set, nil
		}

		var keySet jwk.Set
		var fetchErr error
		if v.JWKSFetcher != nil {
			keySet, fetchErr = v.JWKSFetcher(ctx, issuer)
		} else {
			keySet, fetchErr = v.fetchJWKSFromIssuer(ctx, issuer)
		}
		if fetchErr != nil {
			return nil, fetchErr
		}

		v.mu.Lock()
		defer v.mu.Unlock()

		// Evict oldest if at capacity
		if len(v.cache) >= v.maxProviders {
			var oldestKey string
			var oldestTime time.Time
			for k, e := range v.cache {
				if oldestKey == "" || e.fetchedAt.Before(oldestTime) {
					oldestKey = k
					oldestTime = e.fetchedAt
				}
			}
			delete(v.cache, oldestKey)
		}

		v.cache[issuer] = &jwksEntry{set: keySet, fetchedAt: time.Now()}
		return keySet, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(jwk.Set), nil
}

func (v *OIDCValidator) fetchJWKSFromIssuer(ctx context.Context, issuer string) (jwk.Set, error) {
	ctx2, cancel := context.WithTimeout(ctx, v.httpTimeout)
	defer cancel()

	// Try standard OIDC discovery to get jwks_uri
	discoveryURL := issuer + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx2, http.MethodGet, discoveryURL, nil)
	if err == nil {
		resp, err := v.httpClient.Do(req)
		if err == nil && resp.StatusCode == 200 {
			var discovery struct {
				JWKSURI string `json:"jwks_uri"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&discovery); err == nil && discovery.JWKSURI != "" {
				resp.Body.Close()
				keySet, err := jwk.Fetch(ctx2, discovery.JWKSURI)
				if err == nil {
					return keySet, nil
				}
			} else {
				resp.Body.Close()
			}
		} else if resp != nil {
			resp.Body.Close()
		}
	}

	// Fallback: try standard JWKS URI
	keySet, err := jwk.Fetch(ctx2, issuer+"/.well-known/jwks.json")
	if err != nil {
		// Try certs endpoint (Google-style)
		keySet, err = jwk.Fetch(ctx2, issuer+"/certs")
		if err != nil {
			return nil, fmt.Errorf("fetching JWKS from %s: %w", issuer, err)
		}
	}
	return keySet, nil
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
