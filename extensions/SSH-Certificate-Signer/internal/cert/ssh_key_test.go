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

package cert

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestParseSSHPublicKey_Invalid(t *testing.T) {
	_, err := ParseSSHPublicKey("not-a-valid-ssh-key")
	if err == nil {
		t.Error("expected error for invalid SSH key")
	}
}

func TestParseSSHPublicKey_Empty(t *testing.T) {
	_, err := ParseSSHPublicKey("")
	if err == nil {
		t.Error("expected error for empty SSH key")
	}
}

func TestParseSSHPublicKey_ValidEd25519(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	sshPub, _ := ssh.NewPublicKey(pub)
	authorizedKey := ssh.MarshalAuthorizedKey(sshPub)

	parsed, err := ParseSSHPublicKey(string(authorizedKey))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Type() != "ssh-ed25519" {
		t.Errorf("expected ssh-ed25519, got %s", parsed.Type())
	}
}

func TestParseSSHPublicKey_ValidRSA(t *testing.T) {
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	sshPub, _ := ssh.NewPublicKey(&rsaKey.PublicKey)
	authorizedKey := ssh.MarshalAuthorizedKey(sshPub)

	parsed, err := ParseSSHPublicKey(string(authorizedKey))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Type() != "ssh-rsa" {
		t.Errorf("expected ssh-rsa, got %s", parsed.Type())
	}
}

func TestParseSSHPublicKey_ValidECDSA(t *testing.T) {
	ecKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	sshPub, _ := ssh.NewPublicKey(&ecKey.PublicKey)
	authorizedKey := ssh.MarshalAuthorizedKey(sshPub)

	parsed, err := ParseSSHPublicKey(string(authorizedKey))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Type() != "ecdsa-sha2-nistp256" {
		t.Errorf("expected ecdsa-sha2-nistp256, got %s", parsed.Type())
	}
}

func TestNormalizeKeyType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ssh-ed25519", "ed25519"},
		{"ssh-rsa", "rsa"},
		{"ecdsa-sha2-nistp256", "ecdsa"},
		{"ecdsa-sha2-nistp384", "ecdsa"},
		{"unknown", "unknown"},
	}
	for _, tt := range tests {
		got := NormalizeKeyType(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeKeyType(%s): expected %s, got %s", tt.input, tt.expected, got)
		}
	}
}

func TestHashToken(t *testing.T) {
	hash1 := HashToken("eyJhbGciOiJSUzI1NiIs")
	hash2 := HashToken("eyJhbGciOiJSUzI1NiIs")

	// Consistent
	if hash1 != hash2 {
		t.Error("hash should be consistent for the same input")
	}

	// Verify it is SHA-256 base64
	decoded, err := base64.StdEncoding.DecodeString(hash1)
	if err != nil {
		t.Fatalf("not valid base64: %v", err)
	}
	if len(decoded) != sha256.Size {
		t.Errorf("expected %d bytes, got %d", sha256.Size, len(decoded))
	}

	// Different tokens produce different hashes
	hash3 := HashToken("different-token")
	if hash1 == hash3 {
		t.Error("different tokens should produce different hashes")
	}
}

func TestSSHPublicKeyFingerprint(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	sshPub, _ := ssh.NewPublicKey(pub)
	fp := SSHPublicKeyFingerprint(sshPub)
	if fp == "" {
		t.Error("expected non-empty fingerprint")
	}
	if len(fp) < 7 || fp[:7] != "SHA256:" {
		t.Errorf("expected SHA256: prefix, got %s", fp)
	}
}

func TestParseCAPrivateKey_Ed25519(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	privBytes, _ := x509.MarshalPKCS8PrivateKey(priv)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	signer, sshSigner, err := ParseCAPrivateKey(privPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if signer == nil || sshSigner == nil {
		t.Fatal("expected non-nil signers")
	}
}

func TestParseCAPrivateKey_RSA(t *testing.T) {
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	privBytes, _ := x509.MarshalPKCS8PrivateKey(rsaKey)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	_, sshSigner, err := ParseCAPrivateKey(privPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sshSigner == nil {
		t.Fatal("expected non-nil SSH signer")
	}
}

func TestParseCAPrivateKey_ECDSA(t *testing.T) {
	ecKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	privBytes, _ := x509.MarshalPKCS8PrivateKey(ecKey)
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	_, sshSigner, err := ParseCAPrivateKey(privPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sshSigner == nil {
		t.Fatal("expected non-nil SSH signer")
	}
}

func TestParseCAPrivateKey_InvalidPEM(t *testing.T) {
	_, _, err := ParseCAPrivateKey([]byte("not a PEM"))
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

func TestPublicKeyToJWK_Ed25519(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	pubBytes, _ := x509.MarshalPKIXPublicKey(pub)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	jwkMap, err := PublicKeyToJWK(pubPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jwkMap["kty"] != "OKP" {
		t.Errorf("expected kty OKP, got %v", jwkMap["kty"])
	}
	if jwkMap["crv"] != "Ed25519" {
		t.Errorf("expected crv Ed25519, got %v", jwkMap["crv"])
	}
	if jwkMap["x"] == nil || jwkMap["x"] == "" {
		t.Error("expected non-empty x")
	}
	if jwkMap["kid"] == nil || jwkMap["kid"] == "" {
		t.Error("expected non-empty kid")
	}
}

func TestCAFingerprint(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	pubBytes, _ := x509.MarshalPKIXPublicKey(pub)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	fp, err := CAFingerprint(pubPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fp == "" || fp[:7] != "SHA256:" {
		t.Errorf("expected SHA256: fingerprint, got %s", fp)
	}
}

func TestPEMToOpenSSH_Ed25519(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	pubBytes, _ := x509.MarshalPKIXPublicKey(pub)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	openssh, err := PEMToOpenSSH(pubPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(openssh, "ssh-ed25519 ") {
		t.Errorf("expected ssh-ed25519 prefix, got %q", openssh)
	}

	// Should be parseable back
	parsed, err := ParseSSHPublicKey(openssh)
	if err != nil {
		t.Fatalf("failed to re-parse OpenSSH key: %v", err)
	}
	if parsed.Type() != "ssh-ed25519" {
		t.Errorf("expected ssh-ed25519, got %s", parsed.Type())
	}
}

func TestPEMToOpenSSH_RSA(t *testing.T) {
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubBytes, _ := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	openssh, err := PEMToOpenSSH(pubPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(openssh, "ssh-rsa ") {
		t.Errorf("expected ssh-rsa prefix, got %q", openssh)
	}
}

func TestPEMToOpenSSH_ECDSA(t *testing.T) {
	ecKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pubBytes, _ := x509.MarshalPKIXPublicKey(&ecKey.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	openssh, err := PEMToOpenSSH(pubPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(openssh, "ecdsa-sha2-nistp256 ") {
		t.Errorf("expected ecdsa-sha2-nistp256 prefix, got %q", openssh)
	}
}

func TestPEMToOpenSSH_InvalidPEM(t *testing.T) {
	_, err := PEMToOpenSSH([]byte("not valid PEM"))
	if err == nil {
		t.Error("expected error for invalid PEM")
	}
}

func TestPEMToOpenSSH_NoTrailingNewline(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	pubBytes, _ := x509.MarshalPKIXPublicKey(pub)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	openssh, err := PEMToOpenSSH(pubPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.HasSuffix(openssh, "\n") {
		t.Error("PEMToOpenSSH should not include trailing newline")
	}
}
