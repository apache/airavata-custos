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

package service

import (
	"context"
	"sync"
	"time"

	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

// identityCacheDefaultTTL is the fallback TTL when SetIdentityCacheTTL is never
// invoked or receives a non-positive value. The config loader applies the
// same default and enforces a 60s ceiling before the value reaches the cache.
const identityCacheDefaultTTL = 30 * time.Second

type identityCacheEntry struct {
	caller  *identity.Caller
	privs   []models.PrivilegeKey
	expires time.Time
}

// identityCache is the sub-keyed in-process cache the resolver consults before
// hitting the database. Entries expire on read after TTL.
type identityCache struct {
	mu      sync.Mutex
	entries map[string]identityCacheEntry
	ttl     time.Duration
}

func newIdentityCache() *identityCache {
	return &identityCache{entries: map[string]identityCacheEntry{}, ttl: identityCacheDefaultTTL}
}

func (c *identityCache) get(sub string, now time.Time) (identityCacheEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[sub]
	if !ok {
		return identityCacheEntry{}, false
	}
	if !now.Before(e.expires) {
		delete(c.entries, sub)
		return identityCacheEntry{}, false
	}
	return e, true
}

func (c *identityCache) set(sub string, caller *identity.Caller, privs []models.PrivilegeKey, now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[sub] = identityCacheEntry{caller: caller, privs: privs, expires: now.Add(c.ttl)}
}

func (c *identityCache) invalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = map[string]identityCacheEntry{}
}

func (c *identityCache) setTTL(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if d <= 0 {
		d = identityCacheDefaultTTL
	}
	c.ttl = d
}

// SetIdentityCacheTTL configures the resolver cache's TTL. The config loader
// caps the value at the 60s ceiling before it reaches the cache.
func (s *Service) SetIdentityCacheTTL(d time.Duration) {
	s.identityCache.setTTL(d)
}

// InvalidateAllIdentities empties the cache.
func (s *Service) InvalidateAllIdentities() {
	s.identityCache.invalidateAll()
}

// ResolveCaller maps a verified OIDC sub to a Caller + effective privilege set.
// Returns ErrNotLinked when no user_identities row references sub.
func (s *Service) ResolveCaller(ctx context.Context, oidcSub string) (*identity.Caller, []models.PrivilegeKey, error) {
	now := time.Now()
	if hit, ok := s.identityCache.get(oidcSub, now); ok {
		return hit.caller, hit.privs, nil
	}
	user, err := s.users.GetUserByOIDCSub(ctx, oidcSub)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, identity.ErrNotLinked
	}
	privs, err := s.EffectivePrivileges(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}
	caller := &identity.Caller{UserID: user.ID}
	s.identityCache.set(oidcSub, caller, privs, now)
	return caller, privs, nil
}
