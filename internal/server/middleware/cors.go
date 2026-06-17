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

package middleware

import (
	"net/http"
	"strings"

	"github.com/apache/airavata-custos/internal/config"
)

// CORS enforces an explicit origin allowlist. Preflight (OPTIONS) requests
// short-circuit here so they never reach the auth middleware that follows.
type CORS struct {
	allowed map[string]struct{}
	methods string
	headers string
}

// NewCORS reads the allowlist from cfg. An empty allowlist makes the
// middleware a no-op pass-through, which is the right behaviour when CORS is
// not configured (server-to-server deployment, integration tests).
func NewCORS(cfg config.CORSConfig) *CORS {
	c := &CORS{
		allowed: make(map[string]struct{}, len(cfg.AllowedOrigins)),
		methods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		headers: "Authorization, Content-Type, X-Trace-Id",
	}
	for _, o := range cfg.AllowedOrigins {
		o = strings.TrimSpace(o)
		if o != "" {
			c.allowed[o] = struct{}{}
		}
	}
	return c
}

// Wrap returns a handler that adds CORS response headers for allowlisted
// origins and answers preflight requests directly. Non-allowlisted origins
// pass through without CORS headers; the browser will reject the response.
func (c *CORS) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			// Vary on every cross-origin response so caches don't serve the
			// allowlisted response to a different origin.
			w.Header().Add("Vary", "Origin")
			if _, ok := c.allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", c.methods)
				w.Header().Set("Access-Control-Allow-Headers", c.headers)
			}
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
