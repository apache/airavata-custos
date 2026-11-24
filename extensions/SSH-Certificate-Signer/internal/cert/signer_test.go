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
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func generateTestEd25519CA(t *testing.T) ([]byte, ssh.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	return openSSHPrivateKey(t, priv), sshPub
}

func generateTestEd25519UserKey(t *testing.T) ssh.PublicKey {
	t.Helper()
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	return sshPub
}

func TestSignCertificate_Ed25519(t *testing.T) {
	caPrivPEM, _ := generateTestEd25519CA(t)
	userKey := generateTestEd25519UserKey(t)

	before := time.Now().UTC()
	exts := ExtensionsToMap(AllStandardExtensions())
	result, err := SignCertificate(&SignRequest{
		PublicKey:       userKey,
		CAPrivateKeyPEM: caPrivPEM,
		Serial:          42,
		Principal:       "testuser",
		UserEmail:       "jdoe@example.edu",
		TTLSeconds:      7200,
		Extensions:      exts,
	})
	after := time.Now().UTC()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check serial
	if result.Serial != 42 {
		t.Errorf("expected serial 42, got %d", result.Serial)
	}

	// Check validity window
	if int64(result.ValidAfter) < before.Unix()-1 || int64(result.ValidAfter) > after.Unix()+1 {
		t.Errorf("valid_after %d not within expected range [%d, %d]", result.ValidAfter, before.Unix(), after.Unix())
	}
	expectedBefore := result.ValidAfter + 7200
	if result.ValidBefore != expectedBefore {
		t.Errorf("expected valid_before %d, got %d", expectedBefore, result.ValidBefore)
	}

	// Check CA fingerprint
	if result.CAFingerprint == "" {
		t.Error("expected non-empty CA fingerprint")
	}

	// Parse the cert and validate internals
	parsedKey, err := ssh.ParsePublicKey(result.CertBytes)
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}
	cert, ok := parsedKey.(*ssh.Certificate)
	if !ok {
		t.Fatal("parsed key is not a certificate")
	}

	if cert.Serial != 42 {
		t.Errorf("cert serial: expected 42, got %d", cert.Serial)
	}

	if len(cert.ValidPrincipals) != 1 || cert.ValidPrincipals[0] != "testuser" {
		t.Errorf("expected principals [testuser], got %v", cert.ValidPrincipals)
	}

	// User cert type
	if cert.CertType != ssh.UserCert {
		t.Errorf("expected UserCert type, got %d", cert.CertType)
	}

	for _, ext := range AllStandardExtensions() {
		if _, ok := cert.Extensions[string(ext)]; !ok {
			t.Errorf("expected extension %s", ext)
		}
	}

	if len(cert.Nonce) != 32 {
		t.Errorf("expected 32-byte nonce, got %d", len(cert.Nonce))
	}
}

// Make sure the signed certificate says who it is for and who asked, and that the key ID names both.
func TestSignCertificate_CarriesUserAndRequester(t *testing.T) {
	caPrivPEM, _ := generateTestEd25519CA(t)
	result, err := SignCertificate(&SignRequest{
		PublicKey:       generateTestEd25519UserKey(t),
		CAPrivateKeyPEM: caPrivPEM,
		Serial:          1,
		Principal:       "jdoe",
		UserEmail:       "jdoe@example.edu",
		UserSub:         "http://idp.example.edu/users/9",
		RequestedBy:     "app/123",
		TTLSeconds:      3600,
	})
	if err != nil {
		t.Fatal(err)
	}
	parsedKey, err := ssh.ParsePublicKey(result.CertBytes)
	if err != nil {
		t.Fatal(err)
	}
	cert := parsedKey.(*ssh.Certificate)

	if cert.KeyId != "jdoe@example.edu:app/123" {
		t.Errorf("key ID: %s", cert.KeyId)
	}
	want := map[string]string{
		ExtensionUserSub:     "http://idp.example.edu/users/9",
		ExtensionUserEmail:   "jdoe@example.edu",
		ExtensionRequestedBy: "app/123",
	}
	for name, value := range want {
		if cert.Extensions[name] != value {
			t.Errorf("extension %s = %q, want %q", name, cert.Extensions[name], value)
		}
	}
}

func TestSignCertificate_WithCriticalOptions(t *testing.T) {
	caPrivPEM, _ := generateTestEd25519CA(t)
	userKey := generateTestEd25519UserKey(t)

	result, err := SignCertificate(&SignRequest{
		PublicKey:       userKey,
		CAPrivateKeyPEM: caPrivPEM,
		Serial:          1,
		Principal:       "user",
		UserEmail:       "jdoe@example.edu",
		TTLSeconds:      3600,
		CriticalOptions: map[string]string{"source-address": "10.0.0.0/8"},
	})
	if err != nil {
		t.Fatal(err)
	}

	parsedKey, _ := ssh.ParsePublicKey(result.CertBytes)
	cert := parsedKey.(*ssh.Certificate)

	if v, ok := cert.Permissions.CriticalOptions["source-address"]; !ok || v != "10.0.0.0/8" {
		t.Errorf("expected critical option source-address=10.0.0.0/8, got %v", cert.Permissions.CriticalOptions)
	}
}

// Make sure the ed25519 CA still signs RSA and ECDSA user keys, since users may bring either.
func TestSignCertificate_SignsRSAAndECDSAUserKeys(t *testing.T) {
	caPrivPEM, _ := generateTestEd25519CA(t)
	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	p256, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	p384, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	p521, _ := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)

	for name, key := range map[string]any{
		"rsa":        &rsaKey.PublicKey,
		"ecdsa-p256": &p256.PublicKey,
		"ecdsa-p384": &p384.PublicKey,
		"ecdsa-p521": &p521.PublicKey,
	} {
		userKey, _ := ssh.NewPublicKey(key)
		if _, err := SignCertificate(&SignRequest{
			PublicKey:       userKey,
			CAPrivateKeyPEM: caPrivPEM,
			Serial:          10,
			Principal:       "jdoe",
			UserEmail:       "jdoe@example.edu",
			TTLSeconds:      3600,
		}); err != nil {
			t.Errorf("%s user key: %v", name, err)
		}
	}
}
