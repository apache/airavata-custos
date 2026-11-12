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
	"crypto/x509"
	"encoding/pem"
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

// openSSHPrivateKey encodes key the way ssh-keygen writes it.
func openSSHPrivateKey(t *testing.T, key any) []byte {
	t.Helper()
	block, err := ssh.MarshalPrivateKey(key, "")
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(block)
}

// Make sure an ed25519 CA key in OpenSSH format, as ssh-keygen writes it, is accepted.
func TestParseCAPrivateKey_AcceptsOpenSSHEd25519(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	if _, err := ParseCAPrivateKey(openSSHPrivateKey(t, priv)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Make sure a CA key in PEM format or of a type other than ed25519 is refused.
func TestParseCAPrivateKey_RefusesOtherFormatsAndTypes(t *testing.T) {
	_, ed25519Key, _ := ed25519.GenerateKey(rand.Reader)
	pkcs8, _ := x509.MarshalPKCS8PrivateKey(ed25519Key)
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	for name, key := range map[string][]byte{
		"ed25519 in PEM":   pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8}),
		"rsa in OpenSSH":   openSSHPrivateKey(t, rsaKey),
		"not a key at all": []byte("not a key"),
	} {
		if _, err := ParseCAPrivateKey(key); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// Make sure the CA public key is read as an ed25519 OpenSSH line and a key of another type is refused.
func TestParseCAPublicKey_NeedsEd25519Line(t *testing.T) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	sshPub, _ := ssh.NewPublicKey(pub)
	if _, err := ParseCAPublicKey(ssh.MarshalAuthorizedKey(sshPub)); err != nil {
		t.Errorf("ed25519 line: %v", err)
	}

	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	rsaPub, _ := ssh.NewPublicKey(&rsaKey.PublicKey)
	if _, err := ParseCAPublicKey(ssh.MarshalAuthorizedKey(rsaPub)); err == nil {
		t.Error("rsa line: accepted")
	}
}
