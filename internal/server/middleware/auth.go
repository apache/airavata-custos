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

// Package middleware carries the HTTP-layer cross-cutting concerns: bearer
// token verification, CORS, and any future per-request seam that runs before
// the application handlers in internal/server. Each middleware here is
// constructed once at boot and wraps the server's mux.
package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/sync/singleflight"

	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/pkg/identity"
)

// UserResolver looks up a Custos user by verified OIDC subject. The auth
// middleware calls this after signature + claim validation; a nil result
// means the sub is valid but unknown to Custos, which the middleware turns
// into a 401.
type UserResolver func(ctx context.Context, oidcSub string) (userID string, err error)

// Auth verifies OIDC bearer tokens against the configured issuer's JWKS,
// resolves the verified `sub` to a Custos user, and attaches an
// [identity.Caller] to the request context.
type Auth struct {
	issuer       string
	audience     string
	jwksOverride string
	resolveUser  UserResolver
	skipPrefixes []string

	httpClient   *http.Client
	jwksCacheTTL time.Duration

	mu      sync.RWMutex
	jwksURL string
	cache   *jwksEntry
	sfGroup singleflight.Group
}

type jwksEntry struct {
	set       jwk.Set
	fetchedAt time.Time
}

// Option tunes an Auth middleware at construction time.
type Option func(*Auth)

// WithJWKSURL pins the JWKS URL, bypassing discovery via the issuer's
// .well-known endpoint. Used by integration tests to point at an
// in-process JWKS server.
func WithJWKSURL(u string) Option {
	return func(a *Auth) { a.jwksOverride = u; a.jwksURL = u }
}

// WithSkipPrefixes lets routes such as /healthz bypass authentication.
// Matching is by exact prefix on the request path.
func WithSkipPrefixes(prefixes ...string) Option {
	return func(a *Auth) { a.skipPrefixes = append(a.skipPrefixes, prefixes...) }
}

// NewAuth constructs the middleware. issuer and audience must be set; the
// config block's JWKSURL is treated as an override (equivalent to passing
// [WithJWKSURL]). resolveUser is mandatory.
func NewAuth(cfg config.AuthConfig, resolveUser UserResolver, opts ...Option) (*Auth, error) {
	if strings.TrimSpace(cfg.Issuer) == "" {
		return nil, errors.New("auth: issuer is required")
	}
	if strings.TrimSpace(cfg.Audience) == "" {
		return nil, errors.New("auth: audience is required")
	}
	if resolveUser == nil {
		return nil, errors.New("auth: user resolver is required")
	}
	a := &Auth{
		issuer:       cfg.Issuer,
		audience:     cfg.Audience,
		jwksOverride: cfg.JWKSURL,
		jwksURL:      cfg.JWKSURL,
		resolveUser:  resolveUser,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		jwksCacheTTL: 10 * time.Minute,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a, nil
}

// Wrap returns a handler that rejects unauthenticated requests and forwards
// authenticated ones with an [identity.Caller] in the context.
func (a *Auth) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.shouldSkip(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		token := bearerFromHeader(r.Header.Get("Authorization"))
		if token == "" {
			writeAuthError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		caller, err := a.verify(r.Context(), token)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, err.Error())
			return
		}

		ctx := identity.WithCaller(r.Context(), caller)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Auth) shouldSkip(path string) bool {
	for _, p := range a.skipPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// verify validates the token signature, issuer, audience, and expiry, then
// resolves the OIDC sub to a Custos user. Any failure collapses to a single
// 401 message for the client; the specific reason is intentionally generic
// so callers cannot distinguish "bad signature" from "unknown sub".
func (a *Auth) verify(ctx context.Context, tokenString string) (*identity.Caller, error) {
	keys, err := a.getJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("token verification unavailable")
	}

	tok, err := jwt.Parse([]byte(tokenString),
		jwt.WithKeySet(keys),
		jwt.WithValidate(true),
		jwt.WithIssuer(a.issuer),
		jwt.WithAudience(a.audience),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token")
	}

	sub := tok.Subject()
	if sub == "" {
		return nil, fmt.Errorf("invalid token")
	}

	userID, err := a.resolveUser(ctx, sub)
	if err != nil {
		return nil, fmt.Errorf("identity lookup failed")
	}
	if userID == "" {
		return nil, fmt.Errorf("unknown caller")
	}

	email, _ := tok.Get("email")
	emailStr, _ := email.(string)

	return &identity.Caller{
		UserID:  userID,
		OIDCSub: sub,
		Email:   emailStr,
	}, nil
}

func (a *Auth) getJWKS(ctx context.Context) (jwk.Set, error) {
	a.mu.RLock()
	entry := a.cache
	a.mu.RUnlock()
	if entry != nil && time.Since(entry.fetchedAt) < a.jwksCacheTTL {
		return entry.set, nil
	}

	result, err, _ := a.sfGroup.Do("jwks", func() (interface{}, error) {
		a.mu.RLock()
		entry2 := a.cache
		a.mu.RUnlock()
		if entry2 != nil && time.Since(entry2.fetchedAt) < a.jwksCacheTTL {
			return entry2.set, nil
		}
		set, err := a.fetchJWKS(ctx)
		if err != nil {
			return nil, err
		}
		a.mu.Lock()
		a.cache = &jwksEntry{set: set, fetchedAt: time.Now()}
		a.mu.Unlock()
		return set, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(jwk.Set), nil
}

// fetchJWKS resolves the JWKS URL (via discovery if not overridden) and
// fetches the key set. Discovery results are cached on the Auth instance so
// subsequent fetches skip the well-known round trip.
func (a *Auth) fetchJWKS(ctx context.Context) (jwk.Set, error) {
	ctx, cancel := context.WithTimeout(ctx, a.httpClient.Timeout)
	defer cancel()

	if a.jwksURL == "" {
		url, err := a.discoverJWKSURL(ctx)
		if err != nil {
			return nil, err
		}
		a.mu.Lock()
		a.jwksURL = url
		a.mu.Unlock()
	}
	return jwk.Fetch(ctx, a.jwksURL, jwk.WithHTTPClient(a.httpClient))
}

func (a *Auth) discoverJWKSURL(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(a.issuer, "/")+"/.well-known/openid-configuration", nil)
	if err != nil {
		return "", fmt.Errorf("auth: build discovery request: %w", err)
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("auth: discovery request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth: discovery returned %d", resp.StatusCode)
	}
	var doc struct {
		JWKSURI string `json:"jwks_uri"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return "", fmt.Errorf("auth: decode discovery: %w", err)
	}
	if doc.JWKSURI == "" {
		return "", fmt.Errorf("auth: discovery missing jwks_uri")
	}
	return doc.JWKSURI, nil
}

func bearerFromHeader(h string) string {
	if h == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(h[len(prefix):])
}

func writeAuthError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
