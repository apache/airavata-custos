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

// Package identity carries the verified caller across the request lifecycle.
// HTTP handlers reach for [CallerFromContext] and [HasPrivilege]; the auth
// middleware installs both after verifying the bearer token.
package identity

import (
	"context"
	"errors"
	"fmt"

	"github.com/apache/airavata-custos/pkg/models"
)

// Caller is the verified identity behind an HTTP request.
type Caller struct {
	UserID string
}

// Wrap the ErrNotLinked to keep the specific cause in the log while the middleware
// returns a single 401 identity_not_linked to clients.
var (
	ErrNotLinked             = errors.New("identity not linked")
	ErrEmailNotVerified      = fmt.Errorf("%w: verified email claim missing", ErrNotLinked)
	ErrNoUserMatchesEmail    = fmt.Errorf("%w: no user matches email", ErrNotLinked)
	ErrUserNotPending        = fmt.Errorf("%w: user is not pending", ErrNotLinked)
	ErrUserAlreadyOIDCLinked = fmt.Errorf("%w: user already has an OIDC binding", ErrNotLinked)
)

// UserResolver maps verified OIDC claims to a Caller and the effective
// privilege set.
type UserResolver interface {
	ResolveCaller(ctx context.Context, claims *Claims) (*Caller, []models.PrivilegeKey, error)
}

type ctxKey string

const (
	callerKey     ctxKey = "caller"
	privilegesKey ctxKey = "privileges"
)

type privilegeSet map[models.PrivilegeKey]struct{}

// WithCaller attaches caller to ctx.
func WithCaller(ctx context.Context, caller *Caller) context.Context {
	return context.WithValue(ctx, callerKey, caller)
}

// CallerFromContext returns the verified caller, or nil if none is set.
func CallerFromContext(ctx context.Context) *Caller {
	if caller, ok := ctx.Value(callerKey).(*Caller); ok {
		return caller
	}
	return nil
}

// withPrivileges attaches the caller's effective privilege set to ctx;
// RequirePrivilege wrappers only read it.
func withPrivileges(ctx context.Context, keys []models.PrivilegeKey) context.Context {
	set := make(privilegeSet, len(keys))
	for _, key := range keys {
		set[key] = struct{}{}
	}
	return context.WithValue(ctx, privilegesKey, set)
}

// HasPrivilege reports whether privilege is present in the set attached to
// ctx. A context that never went through the middleware returns false.
func HasPrivilege(ctx context.Context, privilege models.PrivilegeKey) bool {
	set, _ := ctx.Value(privilegesKey).(privilegeSet)
	_, ok := set[privilege]
	return ok
}

// PrivilegesFromContext returns the caller's effective privilege set. Returns
// nil for both the absent and the empty-set cases.
func PrivilegesFromContext(ctx context.Context) []models.PrivilegeKey {
	set, _ := ctx.Value(privilegesKey).(privilegeSet)
	if len(set) == 0 {
		return nil
	}
	keys := make([]models.PrivilegeKey, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	return keys
}

// WithPrivilegesForTest is the test-only entry point to seed privileges on ctx.
func WithPrivilegesForTest(ctx context.Context, keys []models.PrivilegeKey) context.Context {
	return withPrivileges(ctx, keys)
}
