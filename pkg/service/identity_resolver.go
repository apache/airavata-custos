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
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

const (
	identityAuditLinked = "IDENTITY_LINKED"
	identitySourceOIDC  = "oidc"
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

// ResolveCaller maps verified OIDC claims to a Caller + effective privilege set
func (s *Service) ResolveCaller(ctx context.Context, claims *identity.Claims) (*identity.Caller, []models.PrivilegeKey, error) {
	if claims == nil || claims.Sub == "" {
		return nil, nil, identity.ErrNotLinked
	}
	now := time.Now()
	if hit, ok := s.identityCache.get(claims.Sub, now); ok {
		return hit.caller, hit.privs, nil
	}
	user, err := s.users.GetUserByOIDCSub(ctx, claims.Sub)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		user, err = s.linkBySub(ctx, claims)
		if err != nil {
			return nil, nil, err
		}
	}
	privs, err := s.EffectivePrivileges(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}
	caller := &identity.Caller{UserID: user.ID}
	s.identityCache.set(claims.Sub, caller, privs, now)
	return caller, privs, nil
}

// linkBySub creates the first OIDC binding for a PENDING user matched by verified email.
func (s *Service) linkBySub(ctx context.Context, claims *identity.Claims) (*models.User, error) {
	slog.Debug("linkBySub: evaluating claims",
		"sub", claims.Sub,
		"email", claims.Email,
		"email_verified", claims.EmailVerified,
	)
	if !claims.EmailVerified || claims.Email == "" {
		return nil, identity.ErrEmailNotVerified
	}
	user, err := s.users.FindByEmail(ctx, claims.Email)
	if err != nil {
		return nil, fmt.Errorf("email-fallback lookup: %w", err)
	}
	if user == nil {
		return nil, identity.ErrNoUserMatchesEmail
	}
	existingOIDC, err := s.userIdentities.FindByUserAndSource(ctx, user.ID, identitySourceOIDC)
	if err != nil {
		return nil, fmt.Errorf("email-fallback identity scan: %w", err)
	}
	if existingOIDC != nil {
		return nil, identity.ErrUserAlreadyOIDCLinked
	}
	if user.Status != models.UserPending {
		// A non-MERGED user without an OIDC binding is an inconsistent state;
		// merged users legitimately lose theirs to the surviving user.
		if user.Status != models.UserMerged {
			slog.Error("inconsistent user state: reached non-PENDING lifecycle without an OIDC binding",
				"user_id", user.ID,
				"email", claims.Email,
				"sub", claims.Sub,
				"status", user.Status,
			)
		}
		return nil, fmt.Errorf("%w: status=%s", identity.ErrUserNotPending, user.Status)
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		binding := &models.UserIdentity{
			ID:         newID(),
			UserID:     user.ID,
			Source:     identitySourceOIDC,
			ExternalID: claims.Sub,
			Email:      claims.Email,
			OIDCSub:    claims.Sub,
		}
		if err := s.userIdentities.Create(ctx, tx, binding); err != nil {
			return fmt.Errorf("create oidc identity: %w", err)
		}
		if err := s.users.UpdateStatus(ctx, tx, user.ID, models.UserActive); err != nil {
			return fmt.Errorf("activate user: %w", err)
		}
		return s.writeIdentityAuditTx(ctx, tx, user.ID, map[string]any{
			"oidc_sub": claims.Sub,
			"email":    claims.Email,
			"source":   "oidc_email_fallback",
		})
	}); err != nil {
		return nil, err
	}
	slog.Info("identity linked via email fallback", "user_id", user.ID, "email", claims.Email)
	user.Status = models.UserActive
	return user, nil
}

func (s *Service) writeIdentityAuditTx(ctx context.Context, tx *sql.Tx, userID string, details map[string]any) error {
	payload, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("marshal audit details: %w", err)
	}
	return s.auditEvents.Create(ctx, tx, &models.AuditEvent{
		ID:         newID(),
		EventType:  identityAuditLinked,
		EventTime:  nowUTC(),
		EntityID:   userID,
		EntityType: "user",
		Details:    string(payload),
	})
}
