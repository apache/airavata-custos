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

package middleware_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/apache/airavata-custos/internal/config"
	"github.com/apache/airavata-custos/internal/server/middleware"
	"github.com/apache/airavata-custos/pkg/identity"
)

const (
	testIssuer   = "https://issuer.test"
	testAudience = "test-audience"
)

// signer holds a fresh RSA key pair and the JWKS server backing it. Tests
// instantiate one per case so cache state never leaks between cases.
type signer struct {
	privateKey jwk.Key
	publicSet  jwk.Set
	jwksServer *httptest.Server
}

func newSigner(t *testing.T) *signer {
	t.Helper()
	raw, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa: %v", err)
	}
	priv, err := jwk.FromRaw(raw)
	if err != nil {
		t.Fatalf("priv from raw: %v", err)
	}
	if err := priv.Set(jwk.KeyIDKey, "test-key"); err != nil {
		t.Fatalf("set kid: %v", err)
	}
	if err := priv.Set(jwk.AlgorithmKey, jwa.RS256); err != nil {
		t.Fatalf("set alg: %v", err)
	}

	pub, err := priv.PublicKey()
	if err != nil {
		t.Fatalf("public from priv: %v", err)
	}
	set := jwk.NewSet()
	if err := set.AddKey(pub); err != nil {
		t.Fatalf("add public key: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(set)
	}))
	t.Cleanup(srv.Close)

	return &signer{privateKey: priv, publicSet: set, jwksServer: srv}
}

type tokenOpts struct {
	issuer    string
	audience  string
	subject   string
	email     string
	expiresAt time.Time
}

func (s *signer) mint(t *testing.T, o tokenOpts) string {
	t.Helper()
	builder := jwt.NewBuilder().
		Issuer(o.issuer).
		Audience([]string{o.audience}).
		Subject(o.subject).
		Expiration(o.expiresAt).
		IssuedAt(time.Now())
	if o.email != "" {
		builder = builder.Claim("email", o.email)
	}
	tok, err := builder.Build()
	if err != nil {
		t.Fatalf("build token: %v", err)
	}
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, s.privateKey))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return string(signed)
}

func resolverAlways(userID string) middleware.UserResolver {
	return func(_ context.Context, _ string) (string, error) { return userID, nil }
}

func resolverErr(err error) middleware.UserResolver {
	return func(_ context.Context, _ string) (string, error) { return "", err }
}

func newAuth(t *testing.T, s *signer, resolver middleware.UserResolver, opts ...middleware.Option) *middleware.Auth {
	t.Helper()
	opts = append(opts, middleware.WithJWKSURL(s.jwksServer.URL))
	a, err := middleware.NewAuth(config.AuthConfig{Issuer: testIssuer, Audience: testAudience}, resolver, opts...)
	if err != nil {
		t.Fatalf("new auth: %v", err)
	}
	return a
}

func newRequest(token string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/projects", nil)
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	return r
}

func successHandler(captured **identity.Caller) http.HandlerFunc {
	return func(_ http.ResponseWriter, r *http.Request) {
		*captured = identity.CallerFromContext(r.Context())
	}
}

func TestNewAuth_Validation(t *testing.T) {
	cases := []struct {
		name     string
		cfg      config.AuthConfig
		resolver middleware.UserResolver
	}{
		{"missing issuer", config.AuthConfig{Audience: "a"}, resolverAlways("u")},
		{"missing audience", config.AuthConfig{Issuer: "i"}, resolverAlways("u")},
		{"missing resolver", config.AuthConfig{Issuer: "i", Audience: "a"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := middleware.NewAuth(tc.cfg, tc.resolver); err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestAuth_AcceptsValidToken(t *testing.T) {
	s := newSigner(t)
	a := newAuth(t, s, resolverAlways("user-1"))

	token := s.mint(t, tokenOpts{
		issuer:    testIssuer,
		audience:  testAudience,
		subject:   "sub-1",
		email:     "user@example.com",
		expiresAt: time.Now().Add(time.Minute),
	})

	var captured *identity.Caller
	a.Wrap(successHandler(&captured)).ServeHTTP(httptest.NewRecorder(), newRequest(token))

	if captured == nil {
		t.Fatalf("caller not set on context")
	}
	if captured.UserID != "user-1" || captured.OIDCSub != "sub-1" || captured.Email != "user@example.com" {
		t.Fatalf("unexpected caller %+v", captured)
	}
}

func TestAuth_RejectsRequests(t *testing.T) {
	s := newSigner(t)
	good := tokenOpts{issuer: testIssuer, audience: testAudience, subject: "sub-1", expiresAt: time.Now().Add(time.Minute)}

	cases := []struct {
		name     string
		token    func() string
		resolver middleware.UserResolver
		header   string // overrides the Authorization header when set
	}{
		{
			name:     "no bearer",
			token:    func() string { return "" },
			resolver: resolverAlways("user-1"),
		},
		{
			name:     "non-bearer scheme",
			token:    func() string { return "" },
			resolver: resolverAlways("user-1"),
			header:   "Basic abc",
		},
		{
			name:     "malformed jwt",
			token:    func() string { return "not-a-jwt" },
			resolver: resolverAlways("user-1"),
		},
		{
			name: "expired",
			token: func() string {
				bad := good
				bad.expiresAt = time.Now().Add(-time.Minute)
				return s.mint(t, bad)
			},
			resolver: resolverAlways("user-1"),
		},
		{
			name: "wrong issuer",
			token: func() string {
				bad := good
				bad.issuer = "https://other.test"
				return s.mint(t, bad)
			},
			resolver: resolverAlways("user-1"),
		},
		{
			name: "wrong audience",
			token: func() string {
				bad := good
				bad.audience = "other-audience"
				return s.mint(t, bad)
			},
			resolver: resolverAlways("user-1"),
		},
		{
			name:     "unknown sub",
			token:    func() string { return s.mint(t, good) },
			resolver: resolverAlways(""),
		},
		{
			name:     "resolver error",
			token:    func() string { return s.mint(t, good) },
			resolver: resolverErr(errors.New("db down")),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := newAuth(t, s, tc.resolver)
			req := newRequest(tc.token())
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()
			handlerCalled := false
			a.Wrap(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				handlerCalled = true
			})).ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
			if handlerCalled {
				t.Fatalf("handler should not run for rejected requests")
			}
		})
	}
}

func TestAuth_RejectsTokenSignedByForeignKey(t *testing.T) {
	good := newSigner(t)
	foreign := newSigner(t)

	a := newAuth(t, good, resolverAlways("user-1"))
	token := foreign.mint(t, tokenOpts{
		issuer:    testIssuer,
		audience:  testAudience,
		subject:   "sub-1",
		expiresAt: time.Now().Add(time.Minute),
	})

	rec := httptest.NewRecorder()
	a.Wrap(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("handler should not run")
	})).ServeHTTP(rec, newRequest(token))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestAuth_SkipPrefixBypassesVerification(t *testing.T) {
	s := newSigner(t)
	a := newAuth(t, s, resolverAlways("user-1"), middleware.WithSkipPrefixes("/healthz"))

	called := false
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	a.Wrap(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		called = true
		if c := identity.CallerFromContext(r.Context()); c != nil {
			t.Fatalf("skipped route should not carry a caller, got %+v", c)
		}
	})).ServeHTTP(rec, req)

	if !called {
		t.Fatalf("handler should run for skipped paths")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestAuth_DiscoversJWKSURLFromIssuer(t *testing.T) {
	s := newSigner(t)

	var issuerSrv *httptest.Server
	issuerSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":   issuerSrv.URL,
			"jwks_uri": s.jwksServer.URL,
		})
	}))
	defer issuerSrv.Close()

	a, err := middleware.NewAuth(
		config.AuthConfig{Issuer: issuerSrv.URL, Audience: testAudience},
		resolverAlways("user-1"),
	)
	if err != nil {
		t.Fatalf("new auth: %v", err)
	}

	token := s.mint(t, tokenOpts{
		issuer:    issuerSrv.URL,
		audience:  testAudience,
		subject:   "sub-1",
		expiresAt: time.Now().Add(time.Minute),
	})

	rec := httptest.NewRecorder()
	called := false
	a.Wrap(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { called = true })).ServeHTTP(rec, newRequest(token))

	if !called || rec.Code != http.StatusOK {
		t.Fatalf("discovery path failed: code=%d called=%v", rec.Code, called)
	}
}
