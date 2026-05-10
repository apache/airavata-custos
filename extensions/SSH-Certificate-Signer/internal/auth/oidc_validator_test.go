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

package auth

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/apache/airavata-custos/signer/internal/config"
)

func createTestJWT(t *testing.T, key crypto.Signer, alg jwa.SignatureAlgorithm, claims map[string]interface{}) string {
	t.Helper()

	tok := jwt.New()
	for k, v := range claims {
		tok.Set(k, v)
	}

	jwkKey, err := jwk.FromRaw(key)
	if err != nil {
		t.Fatalf("creating JWK from key: %v", err)
	}
	jwkKey.Set(jwk.KeyIDKey, "test-key-1")
	jwkKey.Set(jwk.AlgorithmKey, alg)

	signed, err := jwt.Sign(tok, jwt.WithKey(alg, jwkKey))
	if err != nil {
		t.Fatalf("signing JWT: %v", err)
	}
	return string(signed)
}

func createJWKSServer(t *testing.T, pubKey crypto.PublicKey, alg jwa.SignatureAlgorithm, requestCount *atomic.Int32) *httptest.Server {
	t.Helper()

	jwkKey, err := jwk.FromRaw(pubKey)
	if err != nil {
		t.Fatalf("creating JWK: %v", err)
	}
	jwkKey.Set(jwk.KeyIDKey, "test-key-1")
	jwkKey.Set(jwk.AlgorithmKey, alg)

	set := jwk.NewSet()
	set.AddKey(jwkKey)

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		if requestCount != nil {
			requestCount.Add(1)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(set)
	})

	return httptest.NewServer(mux)
}

func newTestValidator(allowedIssuers []string) *OIDCValidator {
	return NewOIDCValidator(config.AuthConfig{
		AllowedIssuers: allowedIssuers,
		OIDC: config.OIDCConfig{
			JWKSCacheTTLSeconds: 300,
			JWKSMaxProviders:    10,
			TimeoutSeconds:      10,
		},
	}, config.DevModeConfig{})
}

func TestOIDCValidator_DevMode(t *testing.T) {
	v := NewOIDCValidator(config.AuthConfig{}, config.DevModeConfig{
		Enabled:      true,
		DefaultEmail: "test@dev.local",
	})

	identity, err := v.ValidateAccessToken(context.Background(), "garbage-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.Issuer != "dev-mode" {
		t.Errorf("expected issuer 'dev-mode', got %s", identity.Issuer)
	}
	if identity.Subject != "dev-user" {
		t.Errorf("expected subject 'dev-user', got %s", identity.Subject)
	}
	if identity.Email != "test@dev.local" {
		t.Errorf("expected email 'test@dev.local', got %s", identity.Email)
	}
	if identity.Principal != "dev-user" {
		t.Errorf("expected principal 'dev-user', got %s", identity.Principal)
	}
}

func TestOIDCValidator_DevModeDefaultEmail(t *testing.T) {
	v := NewOIDCValidator(config.AuthConfig{}, config.DevModeConfig{
		Enabled: true,
	})

	identity, err := v.ValidateAccessToken(context.Background(), "any-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.Email != "dev@localhost" {
		t.Errorf("expected email 'dev@localhost', got %s", identity.Email)
	}
}

func TestOIDCValidator_ValidRS256Token(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	var reqCount atomic.Int32
	server := createJWKSServer(t, &rsaKey.PublicKey, jwa.RS256, &reqCount)
	defer server.Close()

	v := newTestValidator(nil)
	v.JWKSFetcher = func(ctx context.Context, issuer string) (jwk.Set, error) {
		return jwk.Fetch(ctx, server.URL+"/.well-known/jwks.json")
	}

	token := createTestJWT(t, rsaKey, jwa.RS256, map[string]interface{}{
		"iss": server.URL,
		"sub": "user123",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	identity, err := v.ValidateAccessToken(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.Subject != "user123" {
		t.Errorf("expected subject user123, got %s", identity.Subject)
	}
}

func TestOIDCValidator_ValidES256Token(t *testing.T) {
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	server := createJWKSServer(t, &ecKey.PublicKey, jwa.ES256, nil)
	defer server.Close()

	v := newTestValidator(nil)
	v.JWKSFetcher = func(ctx context.Context, issuer string) (jwk.Set, error) {
		return jwk.Fetch(ctx, server.URL+"/.well-known/jwks.json")
	}

	token := createTestJWT(t, ecKey, jwa.ES256, map[string]interface{}{
		"iss": server.URL,
		"sub": "ecuser",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	identity, err := v.ValidateAccessToken(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.Subject != "ecuser" {
		t.Errorf("expected subject ecuser, got %s", identity.Subject)
	}
}

func TestOIDCValidator_ValidEdDSAToken(t *testing.T) {
	_, edKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	server := createJWKSServer(t, edKey.Public(), jwa.EdDSA, nil)
	defer server.Close()

	v := newTestValidator(nil)
	v.JWKSFetcher = func(ctx context.Context, issuer string) (jwk.Set, error) {
		return jwk.Fetch(ctx, server.URL+"/.well-known/jwks.json")
	}

	token := createTestJWT(t, edKey, jwa.EdDSA, map[string]interface{}{
		"iss": server.URL,
		"sub": "eduser",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	identity, err := v.ValidateAccessToken(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.Subject != "eduser" {
		t.Errorf("expected subject eduser, got %s", identity.Subject)
	}
}

func TestOIDCValidator_ExpiredToken(t *testing.T) {
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	server := createJWKSServer(t, &rsaKey.PublicKey, jwa.RS256, nil)
	defer server.Close()

	v := newTestValidator(nil)
	v.JWKSFetcher = func(ctx context.Context, issuer string) (jwk.Set, error) {
		return jwk.Fetch(ctx, server.URL+"/.well-known/jwks.json")
	}

	token := createTestJWT(t, rsaKey, jwa.RS256, map[string]interface{}{
		"iss": server.URL,
		"sub": "user",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})

	_, err := v.ValidateAccessToken(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	authErr, ok := err.(*AuthError)
	if !ok {
		t.Fatalf("expected AuthError, got %T", err)
	}
	if authErr.Code != "invalid_token" {
		t.Errorf("expected code invalid_token, got %s", authErr.Code)
	}
}

func TestOIDCValidator_FutureNBF(t *testing.T) {
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	server := createJWKSServer(t, &rsaKey.PublicKey, jwa.RS256, nil)
	defer server.Close()

	v := newTestValidator(nil)
	v.JWKSFetcher = func(ctx context.Context, issuer string) (jwk.Set, error) {
		return jwk.Fetch(ctx, server.URL+"/.well-known/jwks.json")
	}

	token := createTestJWT(t, rsaKey, jwa.RS256, map[string]interface{}{
		"iss": server.URL,
		"sub": "user",
		"exp": time.Now().Add(2 * time.Hour).Unix(),
		"nbf": time.Now().Add(time.Hour).Unix(),
	})

	_, err := v.ValidateAccessToken(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for future nbf token")
	}
}

func TestOIDCValidator_InvalidSignature(t *testing.T) {
	rsaKey1, _ := rsa.GenerateKey(rand.Reader, 2048)
	rsaKey2, _ := rsa.GenerateKey(rand.Reader, 2048)

	// JWKS serves key2's public key, but token signed with key1
	server := createJWKSServer(t, &rsaKey2.PublicKey, jwa.RS256, nil)
	defer server.Close()

	v := newTestValidator(nil)
	v.JWKSFetcher = func(ctx context.Context, issuer string) (jwk.Set, error) {
		return jwk.Fetch(ctx, server.URL+"/.well-known/jwks.json")
	}

	token := createTestJWT(t, rsaKey1, jwa.RS256, map[string]interface{}{
		"iss": server.URL,
		"sub": "user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	_, err := v.ValidateAccessToken(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestOIDCValidator_PrincipalFromSub(t *testing.T) {
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	server := createJWKSServer(t, &rsaKey.PublicKey, jwa.RS256, nil)
	defer server.Close()

	v := newTestValidator(nil)
	v.JWKSFetcher = func(ctx context.Context, issuer string) (jwk.Set, error) {
		return jwk.Fetch(ctx, server.URL+"/.well-known/jwks.json")
	}

	token := createTestJWT(t, rsaKey, jwa.RS256, map[string]interface{}{
		"iss":   server.URL,
		"sub":   "user123",
		"email": "user@example.com",
		"exp":   time.Now().Add(time.Hour).Unix(),
	})

	identity, err := v.ValidateAccessToken(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.Principal != "user123" {
		t.Errorf("expected principal user123, got %s", identity.Principal)
	}
}

func TestOIDCValidator_EmptySubject(t *testing.T) {
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	server := createJWKSServer(t, &rsaKey.PublicKey, jwa.RS256, nil)
	defer server.Close()

	v := newTestValidator(nil)
	v.JWKSFetcher = func(ctx context.Context, issuer string) (jwk.Set, error) {
		return jwk.Fetch(ctx, server.URL+"/.well-known/jwks.json")
	}

	token := createTestJWT(t, rsaKey, jwa.RS256, map[string]interface{}{
		"iss": server.URL,
		"sub": "",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	_, err := v.ValidateAccessToken(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for empty subject")
	}
	authErr, ok := err.(*AuthError)
	if !ok {
		t.Fatalf("expected AuthError, got %T", err)
	}
	if authErr.Message != "No valid principal found in token claims" {
		t.Errorf("expected 'No valid principal found' message, got %s", authErr.Message)
	}
}

func TestOIDCValidator_MissingIssuer(t *testing.T) {
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	v := newTestValidator(nil)

	token := createTestJWT(t, rsaKey, jwa.RS256, map[string]interface{}{
		"sub": "user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	_, err := v.ValidateAccessToken(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for missing issuer")
	}
}

func TestOIDCValidator_JWKSCaching(t *testing.T) {
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	var fetchCount atomic.Int32
	server := createJWKSServer(t, &rsaKey.PublicKey, jwa.RS256, nil)
	defer server.Close()

	v := newTestValidator(nil)
	v.JWKSFetcher = func(ctx context.Context, issuer string) (jwk.Set, error) {
		fetchCount.Add(1)
		return jwk.Fetch(ctx, server.URL+"/.well-known/jwks.json")
	}

	for i := 0; i < 5; i++ {
		token := createTestJWT(t, rsaKey, jwa.RS256, map[string]interface{}{
			"iss": server.URL,
			"sub": fmt.Sprintf("user%d", i),
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		_, err := v.ValidateAccessToken(context.Background(), token)
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
	}

	if fetchCount.Load() != 1 {
		t.Errorf("expected 1 JWKS fetch, got %d", fetchCount.Load())
	}
}

func TestOIDCValidator_DisallowedIssuer(t *testing.T) {
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	server := createJWKSServer(t, &rsaKey.PublicKey, jwa.RS256, nil)
	defer server.Close()

	v := newTestValidator([]string{"https://allowed.example.com"})
	v.JWKSFetcher = func(ctx context.Context, issuer string) (jwk.Set, error) {
		return jwk.Fetch(ctx, server.URL+"/.well-known/jwks.json")
	}

	token := createTestJWT(t, rsaKey, jwa.RS256, map[string]interface{}{
		"iss": "https://other.example.com",
		"sub": "user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	_, err := v.ValidateAccessToken(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for disallowed issuer")
	}
	authErr := err.(*AuthError)
	if authErr.Message != "Issuer not allowed" {
		t.Errorf("expected 'Issuer not allowed', got %s", authErr.Message)
	}
}

func TestOIDCValidator_EmptyAllowedIssuersAcceptsAll(t *testing.T) {
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	server := createJWKSServer(t, &rsaKey.PublicKey, jwa.RS256, nil)
	defer server.Close()

	v := newTestValidator([]string{})
	v.JWKSFetcher = func(ctx context.Context, issuer string) (jwk.Set, error) {
		return jwk.Fetch(ctx, server.URL+"/.well-known/jwks.json")
	}

	token := createTestJWT(t, rsaKey, jwa.RS256, map[string]interface{}{
		"iss": "https://any-issuer.example.com",
		"sub": "user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	_, err := v.ValidateAccessToken(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOIDCValidator_EmptyToken(t *testing.T) {
	v := newTestValidator(nil)
	_, err := v.ValidateAccessToken(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty token")
	}
}
