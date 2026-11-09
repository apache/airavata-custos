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
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/apache/airavata-custos/signer/internal/config"
)

const testServiceToken = "0123456789abcdef0123456789abcdef"

func testRouter() *chi.Mux {
	ok := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
	cfg := &config.Config{}
	cfg.Signer.ServiceToken = testServiceToken
	return NewRouter(cfg, Handlers{Sign: ok, CAPublicKey: ok, Health: ok, Certificates: ok, CertificateDetail: ok, CertificateRevoke: ok})
}

// Make sure every route except health needs the service token since nothing else protects the signer.
func TestEveryRouteNeedsTheServiceToken(t *testing.T) {
	router := testRouter()
	err := chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		if route == "/api/v1/health" {
			return nil
		}
		for _, header := range []string{"", "Bearer wrong-token"} {
			r := httptest.NewRequest(method, strings.ReplaceAll(route, "{serial}", "1"), nil)
			r.Header.Set("Authorization", header)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, r)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("%s %s with %q: got %d", method, route, header, rec.Code)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Make sure the right token gets through.
func TestServiceTokenAdmitsTheCaller(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/certificates", nil)
	r.Header.Set("Authorization", "Bearer "+testServiceToken)
	rec := httptest.NewRecorder()
	testRouter().ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d", rec.Code)
	}
}
