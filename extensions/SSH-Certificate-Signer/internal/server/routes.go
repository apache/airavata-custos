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

	"github.com/apache/airavata-custos/signer/internal/config"
)

type Handlers struct {
	Sign              http.HandlerFunc
	CAPublicKey       http.HandlerFunc
	Health            http.HandlerFunc
	Certificates      http.HandlerFunc
	CertificateDetail http.HandlerFunc
	CertificateRevoke http.HandlerFunc
}

func NewRouter(cfg *config.Config, handlers Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(SecurityHeadersMiddleware)
	r.Use(BodyLimitMiddleware(1 << 20)) // 1 MB

	// Health endpoint (no auth)
	r.Get("/api/v1/health", handlers.Health)

	r.Group(func(r chi.Router) {
		r.Use(ServiceTokenMiddleware(cfg.Signer.ServiceToken))

		r.Post("/api/v1/sign", handlers.Sign)
		r.Get("/api/v1/ca-public-key", handlers.CAPublicKey)
		r.Get("/api/v1/certificates", handlers.Certificates)
		r.Get("/api/v1/certificates/{serial}", handlers.CertificateDetail)
		r.Post("/api/v1/certificates/{serial}/revoke", handlers.CertificateRevoke)
	})

	return r
}
