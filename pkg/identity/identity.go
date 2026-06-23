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

	"github.com/apache/airavata-custos/pkg/models"
)

// Caller is the verified identity behind an HTTP request.
type Caller struct {
	UserID string
}

// ErrNotLinked is returned by a UserResolver when the OIDC sub has no row in
// the Custos user table. The middleware translates this into 401
// identity_not_linked.
var ErrNotLinked = errors.New("identity not linked")

// UserResolver maps a verified OIDC sub to a Caller and the effective
// privilege set.
type UserResolver interface {
	ResolveCaller(ctx context.Context, oidcSub string) (*Caller, []models.PrivilegeKey, error)
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

// WithPrivilegesForTest is the test-only entry point to seed privileges on ctx.
func WithPrivilegesForTest(ctx context.Context, keys []models.PrivilegeKey) context.Context {
	return withPrivileges(ctx, keys)
}
