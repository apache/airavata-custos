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

package server

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

// callerOrUnauthorized returns the verified caller's user ID, writing a 401
// response if the request is unauthenticated. Returns "" on rejection;
// the caller must return immediately when it sees an empty string.
func callerOrUnauthorized(w http.ResponseWriter, r *http.Request) string {
	caller := identity.CallerFromContext(r.Context())
	if caller == nil {
		writeError(w, http.StatusUnauthorized, errors.New("unauthenticated"))
		return ""
	}
	return caller.UserID
}

// authProfileTTL bounds how long the middleware will trust a cached
// privilege set before re-reading the DB.
//
// TODO: make configurable via env (eg. AUTH_CACHE_TTL_SECONDS), cap at 60s.
// TODO support caching for multi-instance, the cache is per-process.
const authProfileTTL = 5 * time.Second

// authProfile is the cached snapshot of a user's effective privileges.
type authProfile struct {
	privileges map[models.PrivilegeKey]struct{}
}

func (p *authProfile) has(privilege models.PrivilegeKey) bool {
	if p == nil {
		return false
	}
	_, ok := p.privileges[privilege]
	return ok
}

// authProfileCache is a tiny in-process TTL cache for "userID -> privilege set".
// An empty profile (zero privileges) is still cached so users with no grants do not read the DB everytime.
type authProfileCache struct {
	mu      sync.Mutex
	entries map[string]authProfileCacheEntry
	ttl     time.Duration
}

type authProfileCacheEntry struct {
	profile *authProfile
	expires time.Time
}

func newAuthProfileCache(ttl time.Duration) *authProfileCache {
	return &authProfileCache{
		entries: make(map[string]authProfileCacheEntry),
		ttl:     ttl,
	}
}

func (c *authProfileCache) get(userID string) (*authProfile, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[userID]
	if !ok || time.Now().After(e.expires) {
		return nil, false
	}
	return e.profile, true
}

func (c *authProfileCache) set(userID string, profile *authProfile) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[userID] = authProfileCacheEntry{
		profile: profile,
		expires: time.Now().Add(c.ttl),
	}
}

// invalidate drops the cache entry for userID. Called after any grant or
// revoke so subsequent requests reflect the new state without waiting for TTL.
func (c *authProfileCache) invalidate(userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, userID)
}

// invalidateAll empties the cache. Used when one mutation affects many
// users at once: adding or removing a privilege from a role (every holder
// gains/loses it) or deleting a role.
func (c *authProfileCache) invalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]authProfileCacheEntry)
}

// requirePrivilege returns a middleware that admits the request only if the
// verified caller holds the named active privilege.
//
// Responses:
//   - 401 Unauthorized - caller not on context (middleware should have set it)
//   - 403 Forbidden - caller is identified but does not hold the privilege
//   - 503 Service Unavailable - auth-profile lookup failed
//
// Fail-closed: a DB failure NEVER reads as 403
func (s *Server) requirePrivilege(p models.PrivilegeKey, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		callerID := callerOrUnauthorized(w, r)
		if callerID == "" {
			return
		}
		profile, err := s.lookupAuthProfile(r.Context(), callerID)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, errors.New("auth lookup failed"))
			return
		}
		if !profile.has(p) {
			writeError(w, http.StatusForbidden, errors.New("insufficient privilege"))
			return
		}
		next(w, r)
	}
}

// lookupAuthProfile returns the caller's effective privilege snapshot
// (direct grants + role-derived privileges), hitting the cache first
// and falling back to the DB. Errors propagate so middleware can fail closed.
func (s *Server) lookupAuthProfile(ctx context.Context, userID string) (*authProfile, error) {
	if cached, ok := s.authCache.get(userID); ok {
		return cached, nil
	}
	keys, err := s.svc.EffectivePrivileges(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile := &authProfile{privileges: make(map[models.PrivilegeKey]struct{}, len(keys))}
	for _, k := range keys {
		profile.privileges[k] = struct{}{}
	}
	s.authCache.set(userID, profile)
	return profile, nil
}
