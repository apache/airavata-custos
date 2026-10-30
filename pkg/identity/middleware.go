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

package identity

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/apache/airavata-custos/pkg/models"
)

// Middleware verifies the bearer token once, resolves the caller, and
// attaches Caller + privilege set to the request context. Public paths
// bypass verification entirely. Callers who need onboarding or consent only
// reach routes withoutUserConsent matches; nil matches none.
func Middleware(verifier TokenVerifier, resolver UserResolver, publicPaths []string, withoutUserConsent func(*http.Request) bool, next http.Handler) http.Handler {
	publicSet := make(map[string]struct{}, len(publicPaths))
	for _, path := range publicPaths {
		publicSet[path] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, ok := publicSet[req.URL.Path]; ok {
			next.ServeHTTP(w, req)
			return
		}
		rawToken, ok := bearer(req.Header.Get("Authorization"))
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "missing_bearer", "Missing bearer token")
			return
		}
		claims, err := verifier.Verify(req.Context(), rawToken)
		if err != nil {
			writeJSONError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired bearer token")
			return
		}
		caller, privileges, err := resolver.ResolveCaller(req.Context(), claims)
		if errors.Is(err, ErrNotLinked) {
			writeJSONError(w, http.StatusUnauthorized, "identity_not_linked", "OIDC identity is not linked to a portal user")
			return
		}
		if errors.Is(err, ErrUserDisabled) {
			writeJSONError(w, http.StatusForbidden, "user_disabled", "User account is disabled")
			return
		}
		if err != nil {
			slog.ErrorContext(req.Context(), "identity resolution failed", "sub", claims.Sub, "email", claims.Email, "error", err.Error())
			writeJSONError(w, http.StatusServiceUnavailable, "auth_lookup_failed", "Identity resolution failed")
			return
		}

		// Determine that this endpoint requires no user consent
		allowedWithoutUserConsent := withoutUserConsent != nil && withoutUserConsent(req)
		if caller.NeedsOnboarding && !allowedWithoutUserConsent {
			writeJSONError(w, http.StatusForbidden, "onboarding_required", "Accept the user agreement to continue")
			return
		}
		if caller.ConsentPending && !allowedWithoutUserConsent {
			writeJSONError(w, http.StatusForbidden, "consent_required", "Accept the updated user agreement to continue")
			return
		}
		ctx := WithClaims(req.Context(), claims)
		ctx = WithCaller(ctx, caller)
		ctx = withPrivileges(ctx, privileges)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// Router is the only path that lets HTTP routes hit the mux. The mux
// field is unexported on purpose: callers must go through RequirePrivilege,
// RequireAuth, RequireAuthWithoutUserConsent, RequireScoped, or Public.
type Router struct {
	mux                   *http.ServeMux
	withoutUserConsentMux *http.ServeMux
	publicPaths           []string
}

// NewRouter wraps mux.
func NewRouter(mux *http.ServeMux) *Router {
	return &Router{mux: mux, withoutUserConsentMux: http.NewServeMux()}
}

// RequirePrivilege registers an authenticated handler. The wrapper checks privilege
// against the caller's privilege set in ctx; absent → 403.
func (r *Router) RequirePrivilege(pattern string, privilege models.PrivilegeKey, handler http.HandlerFunc) {
	r.mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		if !HasPrivilege(req.Context(), privilege) {
			writeJSONError(w, http.StatusForbidden, "insufficient_privilege", "Caller lacks required privilege")
			return
		}
		handler(w, req)
	})
}

// Public registers an unauthenticated handler. The middleware uses the path
// component to bypass JWT verification.
func (r *Router) Public(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc(pattern, handler)
	r.publicPaths = append(r.publicPaths, methodPath(pattern))
}

// RequireAuth registers a handler that requires a verified caller but no
// specific privilege. (e.g., caller looking up their own data).
func (r *Router) RequireAuth(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc(pattern, handler)
}

// RequireScoped registers a route whose access depends on the specific
// resource, not a site privilege. authorize runs first: it writes the response
// and returns false to deny, or returns true to run handler. The policy is a
// required argument, so a scoped route cannot be registered without one.
func (r *Router) RequireScoped(pattern string, authorize func(http.ResponseWriter, *http.Request) bool, handler http.HandlerFunc) {
	r.mux.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		if !authorize(w, req) {
			return
		}
		handler(w, req)
	})
}

// RequireAuthWithoutUserConsent registers a handler that signed-in callers
// can reach before accepting the user agreement. Handlers read ClaimsFromContext.
func (r *Router) RequireAuthWithoutUserConsent(pattern string, handler http.HandlerFunc) {
	r.mux.HandleFunc(pattern, handler)
	r.withoutUserConsentMux.HandleFunc(pattern, func(http.ResponseWriter, *http.Request) {})
}

// WithoutUserConsentRoutes reports whether a request targets a
// RequireAuthWithoutUserConsent route, with the same pattern rules as the real mux.
func (r *Router) WithoutUserConsentRoutes() func(*http.Request) bool {
	return func(req *http.Request) bool {
		_, pattern := r.withoutUserConsentMux.Handler(req)
		return pattern != ""
	}
}

// PublicPaths returns the path components of every Public() registration,
// suitable for handing to Middleware.
func (r *Router) PublicPaths() []string { return r.publicPaths }

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

func bearer(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	return strings.TrimSpace(header[len(prefix):]), true
}

// methodPath strips the optional "METHOD " prefix from a mux pattern, since
// the middleware bypass check compares against req.URL.Path (path only).
func methodPath(pattern string) string {
	if i := strings.IndexByte(pattern, ' '); i >= 0 {
		return strings.TrimSpace(pattern[i+1:])
	}
	return pattern
}

// TODO: use pkg/common.WriteJSON once its import cycle is broken.
//
//	Cycle: pkg/identity → pkg/common → pkg/service → pkg/identity.
func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "message": message})
}
