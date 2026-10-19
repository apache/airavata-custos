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

package server

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/apache/airavata-custos/signer/internal/config"
)

func TestCertificateRevokeRoutes(t *testing.T) {
	router := NewRouter(&config.Config{}, nil, nil, nil, Handlers{})
	routes := map[string]bool{}
	if err := chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		routes[method+" "+route] = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !routes[http.MethodPost+" /api/v1/certificates/{serial}/revoke"] {
		t.Fatal("shared certificate revoke route is missing")
	}
	if routes[http.MethodPost+" /api/v1/admin/certificates/{serial}/revoke"] {
		t.Fatal("old administrator revoke route is still registered")
	}
}
