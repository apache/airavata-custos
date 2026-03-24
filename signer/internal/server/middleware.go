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

package server

import (
	"net/http"
	"strings"

	"github.com/apache/airavata-custos/signer/internal/auth"
	"github.com/apache/airavata-custos/signer/internal/httputil"
	"github.com/apache/airavata-custos/signer/internal/metrics"
)

// BearerAuthMiddleware authenticates requests using an OIDC Bearer token in the
// Authorization header. Used for user-facing endpoints (certificates, userinfo).
func BearerAuthMiddleware(oidcValidator *auth.OIDCValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				metrics.AuthFailuresTotal.WithLabelValues("missing_bearer").Inc()
				httputil.WriteJSONError(w, http.StatusUnauthorized, "unauthorized", "Missing Authorization header")
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				metrics.AuthFailuresTotal.WithLabelValues("invalid_bearer").Inc()
				httputil.WriteJSONError(w, http.StatusUnauthorized, "unauthorized", "Authorization header must use Bearer scheme")
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				metrics.AuthFailuresTotal.WithLabelValues("empty_bearer").Inc()
				httputil.WriteJSONError(w, http.StatusUnauthorized, "unauthorized", "Empty Bearer token")
				return
			}

			identity, err := oidcValidator.ValidateAccessToken(r.Context(), token)
			if err != nil {
				metrics.AuthFailuresTotal.WithLabelValues("invalid_token").Inc()
				if authErr, ok := err.(*auth.AuthError); ok {
					httputil.WriteJSONError(w, authErr.Status, authErr.Code, authErr.Message)
					return
				}
				httputil.WriteJSONError(w, http.StatusUnauthorized, "unauthorized", "Token validation failed")
				return
			}

			ctx := httputil.WithUserIdentity(r.Context(), &httputil.UserIdentityContext{
				Issuer:    identity.Issuer,
				Subject:   identity.Subject,
				Email:     identity.Email,
				Principal: identity.Principal,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClientAuthMiddleware authenticates requests using X-Client-Id and X-Client-Secret headers.
func ClientAuthMiddleware(authenticator *auth.ClientAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIDHeader := r.Header.Get("X-Client-Id")
			secretHeader := r.Header.Get("X-Client-Secret")

			result, err := authenticator.Authenticate(r.Context(), clientIDHeader, secretHeader)
			if err != nil {
				if authErr, ok := err.(*auth.AuthError); ok {
					metrics.AuthFailuresTotal.WithLabelValues(authErr.Code).Inc()
					httputil.WriteJSONError(w, authErr.Status, authErr.Code, authErr.Message)
					return
				}
				httputil.WriteJSONError(w, http.StatusInternalServerError, "internal_error", "Authentication error")
				return
			}

			ctx := httputil.WithClientConfig(r.Context(), result.Config)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SourceIPMiddleware extracts the client source IP from X-Forwarded-For or remote address.
func SourceIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractSourceIP(r)
		ctx := httputil.WithSourceIP(r.Context(), ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func BodyLimitMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				httputil.WriteJSONError(w, http.StatusRequestEntityTooLarge, "invalid_request", "Request body too large")
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

func extractSourceIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}

	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		if addr[0] == '[' {
			if bracketIdx := strings.Index(addr, "]"); bracketIdx != -1 {
				return addr[1:bracketIdx]
			}
		}
		return addr[:idx]
	}
	return addr
}
