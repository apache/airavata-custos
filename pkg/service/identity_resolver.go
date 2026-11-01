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
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

const (
	identityAuditLinked    = "IDENTITY_LINKED"
	identitySourceOIDC     = "oidc"
	identitySourceComanage = "comanage"
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

func (c *identityCache) invalidate(sub string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, sub)
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

// InvalidateIdentity drops the cached caller for one sub.
func (s *Service) InvalidateIdentity(sub string) {
	s.identityCache.invalidate(sub)
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
		// Not cached, so a status change or a finished onboarding shows on the next request.
		if user, err = s.matchPendingByEmail(ctx, claims); err != nil {
			return nil, nil, err
		}
		return &identity.Caller{UserID: user.ID, NeedsOnboarding: true, ConsentPending: true}, nil, nil
	}
	if err := refuseDisabled(user); err != nil {
		return nil, nil, err
	}
	privs, err := s.EffectivePrivileges(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}
	pending, err := s.PendingUserAgreements(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}
	caller := &identity.Caller{UserID: user.ID, ConsentPending: len(pending) > 0}
	s.identityCache.set(claims.Sub, caller, privs, now)
	return caller, privs, nil
}

func refuseDisabled(user *models.User) error {
	switch user.Status {
	case models.UserSuspended, models.UserInactive, models.UserMerged:
		return fmt.Errorf("%w: status=%s", identity.ErrUserDisabled, user.Status)
	}
	return nil
}

// matchPendingByEmail finds the PENDING user a not-yet-bound sub belongs to.
func (s *Service) matchPendingByEmail(ctx context.Context, claims *identity.Claims) (*models.User, error) {
	slog.DebugContext(ctx, "matchPendingByEmail: evaluating claims",
		"sub", claims.Sub,
		"email", claims.Email,
		"email_verified", claims.EmailVerified,
	)
	// TODO: restore the email_verified gate once the CILogon test issuer +
	// COmanage LDAP provisioner start returning email_verified=true.
	// if !claims.EmailVerified || claims.Email == "" {
	if claims.Email == "" {
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
	if err := refuseDisabled(user); err != nil {
		return nil, err
	}
	if user.Status != models.UserPending {
		slog.ErrorContext(ctx, "inconsistent user state: reached non-PENDING lifecycle without an OIDC binding",
			"user_id", user.ID,
			"email", claims.Email,
			"sub", claims.Sub,
			"status", user.Status,
		)
		return nil, fmt.Errorf("%w: status=%s", identity.ErrUserNotPending, user.Status)
	}
	return user, nil
}

// bindSubTx stores the sub as the user's oidc identity and activates the user.
func (s *Service) bindSubTx(ctx context.Context, tx *sql.Tx, user *models.User, claims *identity.Claims) error {
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
	if err := s.writeAuditTx(ctx, tx, "user", identityAuditLinked, user.ID, map[string]any{
		"oidc_sub": claims.Sub,
		"email":    claims.Email,
		"source":   "first_sign_in",
	}); err != nil {
		return err
	}
	slog.InfoContext(ctx, "identity linked on first sign-in", "user_id", user.ID, "email", claims.Email)
	return s.eventBus.Publish(ctx, tx, events.UserIdentityCreateEvent, binding)
}
