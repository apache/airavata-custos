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
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
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
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})

	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatal(err)
	}
	return privPEM, sshPub
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
	result, err := SignCertificate(&SignRequest{
		PublicKey:       userKey,
		CAPrivateKeyPEM: caPrivPEM,
		Serial:          42,
		Principal:       "testuser",
		ClientID:        "webapp",
		TTLSeconds:      7200,
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

	for _, ext := range []string{"permit-pty", "permit-port-forwarding", "permit-user-rc"} {
		if _, ok := cert.Extensions[ext]; !ok {
			t.Errorf("expected extension %s", ext)
		}
	}

	if len(cert.Nonce) != 32 {
		t.Errorf("expected 32-byte nonce, got %d", len(cert.Nonce))
	}
}

func TestSignCertificate_UniqueNonces(t *testing.T) {
	caPrivPEM, _ := generateTestEd25519CA(t)
	userKey := generateTestEd25519UserKey(t)

	var nonces [][]byte
	for i := 0; i < 10; i++ {
		result, err := SignCertificate(&SignRequest{
			PublicKey:       userKey,
			CAPrivateKeyPEM: caPrivPEM,
			Serial:          uint64(i + 1),
			Principal:       "user",
			ClientID:        "c1",
			TTLSeconds:      3600,
		})
		if err != nil {
			t.Fatal(err)
		}
		parsedKey, _ := ssh.ParsePublicKey(result.CertBytes)
		cert := parsedKey.(*ssh.Certificate)
		nonces = append(nonces, cert.Nonce)
	}

	// Check all nonces are unique
	for i := 0; i < len(nonces); i++ {
		for j := i + 1; j < len(nonces); j++ {
			if string(nonces[i]) == string(nonces[j]) {
				t.Errorf("nonces %d and %d are identical", i, j)
			}
		}
	}
}

func TestSignCertificate_KeyIDFormat(t *testing.T) {
	caPrivPEM, _ := generateTestEd25519CA(t)
	userKey := generateTestEd25519UserKey(t)

	result, err := SignCertificate(&SignRequest{
		PublicKey:       userKey,
		CAPrivateKeyPEM: caPrivPEM,
		Serial:          1,
		Principal:       "jdoe",
		ClientID:        "webapp",
		TTLSeconds:      3600,
	})
	if err != nil {
		t.Fatal(err)
	}

	parsedKey, _ := ssh.ParsePublicKey(result.CertBytes)
	cert := parsedKey.(*ssh.Certificate)

	// Key ID should match jdoe@webapp-{timestamp}
	if len(cert.KeyId) < len("jdoe@webapp-") {
		t.Errorf("key ID too short: %s", cert.KeyId)
	}
	expected := "jdoe@webapp-"
	if cert.KeyId[:len(expected)] != expected {
		t.Errorf("key ID prefix: expected %s, got %s", expected, cert.KeyId[:len(expected)])
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
		ClientID:        "c1",
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

func TestSignCertificate_Base64Decodable(t *testing.T) {
	caPrivPEM, _ := generateTestEd25519CA(t)
	userKey := generateTestEd25519UserKey(t)

	result, err := SignCertificate(&SignRequest{
		PublicKey:       userKey,
		CAPrivateKeyPEM: caPrivPEM,
		Serial:          1,
		Principal:       "user",
		ClientID:        "c1",
		TTLSeconds:      3600,
	})
	if err != nil {
		t.Fatal(err)
	}

	encoded := base64.StdEncoding.EncodeToString(result.CertBytes)
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("base64 decode failed: %v", err)
	}

	_, err = ssh.ParsePublicKey(decoded)
	if err != nil {
		t.Fatalf("failed to parse decoded certificate: %v", err)
	}
}
