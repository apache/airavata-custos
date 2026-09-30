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

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/apache/airavata-custos/signer/internal/auth"
	"github.com/apache/airavata-custos/signer/internal/config"
)

type Handlers struct {
	Sign              http.HandlerFunc
	Revoke            http.HandlerFunc
	JWKS              http.HandlerFunc
	CAPublicKey       http.HandlerFunc
	Health            http.HandlerFunc
	Admin             http.HandlerFunc
	Certificates      http.HandlerFunc
	CertificateDetail http.HandlerFunc
	UserInfo          http.HandlerFunc
}

func NewRouter(
	cfg *config.Config,
	authenticator *auth.ClientAuthenticator,
	oidcValidator *auth.OIDCValidator,
	handlers Handlers,
) *chi.Mux {
	r := chi.NewRouter()

	if cfg.CORS.Enabled {
		r.Use(CORSMiddleware(cfg.CORS))
	}
	r.Use(SecurityHeadersMiddleware)
	r.Use(SourceIPMiddleware)
	r.Use(BodyLimitMiddleware(1 << 20)) // 1 MB

	// Health endpoint (no auth)
	r.Get("/api/v1/health", handlers.Health)

	// Metrics endpoint (no auth)
	if cfg.Metrics.Enabled {
		r.Handle(cfg.Metrics.Path, promhttp.Handler())
	}

	// Client-authenticated routes (machine-to-machine)
	r.Group(func(r chi.Router) {
		r.Use(ClientAuthMiddleware(authenticator))

		r.Post("/api/v1/sign", handlers.Sign)
		r.Post("/api/v1/revoke", handlers.Revoke)
		r.Get("/api/v1/jwks", handlers.JWKS)
		r.Get("/api/v1/ca-public-key", handlers.CAPublicKey)
		r.Post("/api/v1/admin/rotate-ca", handlers.Admin)
	})

	// User-authenticated routes (OIDC Bearer token)
	r.Group(func(r chi.Router) {
		r.Use(BearerAuthMiddleware(oidcValidator))

		r.Get("/api/v1/certificates", handlers.Certificates)
		r.Get("/api/v1/certificates/{serial}", handlers.CertificateDetail)
		r.Get("/api/v1/userinfo", handlers.UserInfo)
	})

	return r
}
