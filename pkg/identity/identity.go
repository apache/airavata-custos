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
//
// HTTP handlers reach for [CallerFromContext]; event-driven flows (AMIE
// subscribers, COmanage provisioner, scheduled jobs) reach for [WithSystem]
// and [IsSystem]. The two paths are mutually exclusive: HTTP middleware
// never sets a system actor, and a system context never carries a Caller.
package identity

import "context"

// Caller is the verified identity behind an HTTP request.
type Caller struct {
	UserID  string
	OIDCSub string
	Email   string
}

type ctxKey int

const (
	callerKey ctxKey = iota
	systemKey
)

// WithCaller attaches c to ctx. Used by the auth middleware after JWT
// verification + OIDC sub -> Custos user resolution.
func WithCaller(ctx context.Context, c *Caller) context.Context {
	return context.WithValue(ctx, callerKey, c)
}

// CallerFromContext returns the verified caller, or nil if none is set.
// Handlers that require a caller treat nil as a programming error: the
// middleware is supposed to reject unauthenticated requests before they
// reach a handler.
func CallerFromContext(ctx context.Context) *Caller {
	if v, ok := ctx.Value(callerKey).(*Caller); ok {
		return v
	}
	return nil
}

// WithSystem marks ctx as a system-actor flow (event subscribers, jobs).
// Service code that branches on caller identity uses [IsSystem] to skip
// per-user privilege checks for these flows.
func WithSystem(ctx context.Context) context.Context {
	return context.WithValue(ctx, systemKey, true)
}

// IsSystem reports whether ctx is a system-actor flow.
func IsSystem(ctx context.Context) bool {
	v, _ := ctx.Value(systemKey).(bool)
	return v
}
