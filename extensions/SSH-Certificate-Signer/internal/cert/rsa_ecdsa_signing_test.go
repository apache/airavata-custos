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

func generateRSACA(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
}

func generateECDSACA(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
}

func generateRSAUserKey(t *testing.T) ssh.PublicKey {
	t.Helper()
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	sshPub, _ := ssh.NewPublicKey(&key.PublicKey)
	return sshPub
}

func generateECDSAUserKey(t *testing.T) ssh.PublicKey {
	t.Helper()
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	sshPub, _ := ssh.NewPublicKey(&key.PublicKey)
	return sshPub
}

func TestSignCertificate_RSAKeyWithRSACA(t *testing.T) {
	caPEM := generateRSACA(t)
	userKey := generateRSAUserKey(t)

	result, err := SignCertificate(&SignRequest{
		PublicKey:       userKey,
		CAPrivateKeyPEM: caPEM,
		Serial:          10,
		Principal:       "rsauser",
		ClientID:        "c1",
		TTLSeconds:      3600,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsedKey, err := ssh.ParsePublicKey(result.CertBytes)
	if err != nil {
		t.Fatalf("failed to parse cert: %v", err)
	}
	cert := parsedKey.(*ssh.Certificate)
	if cert.Serial != 10 {
		t.Errorf("expected serial 10, got %d", cert.Serial)
	}
}

func TestSignCertificate_ECDSAKeyWithECDSACA(t *testing.T) {
	caPEM := generateECDSACA(t)
	userKey := generateECDSAUserKey(t)

	result, err := SignCertificate(&SignRequest{
		PublicKey:       userKey,
		CAPrivateKeyPEM: caPEM,
		Serial:          20,
		Principal:       "ecuser",
		ClientID:        "c1",
		TTLSeconds:      3600,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsedKey, err := ssh.ParsePublicKey(result.CertBytes)
	if err != nil {
		t.Fatalf("failed to parse cert: %v", err)
	}
	cert := parsedKey.(*ssh.Certificate)
	if cert.Serial != 20 {
		t.Errorf("expected serial 20, got %d", cert.Serial)
	}
}

// Sign an Ed25519 user key with an RSA CA
func TestSignCertificate_Ed25519KeyWithRSACA(t *testing.T) {
	caPEM := generateRSACA(t)
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	sshPub, _ := ssh.NewPublicKey(pub)

	result, err := SignCertificate(&SignRequest{
		PublicKey:       sshPub,
		CAPrivateKeyPEM: caPEM,
		Serial:          30,
		Principal:       "crossuser",
		ClientID:        "c1",
		TTLSeconds:      3600,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsedKey, err := ssh.ParsePublicKey(result.CertBytes)
	if err != nil {
		t.Fatalf("failed to parse cert: %v", err)
	}
	cert := parsedKey.(*ssh.Certificate)
	if cert.Serial != 30 {
		t.Errorf("expected serial 30, got %d", cert.Serial)
	}
}

// Sign an Ed25519 user key with ECDSA CA
func TestSignCertificate_Ed25519KeyWithECDSACA(t *testing.T) {
	caPEM := generateECDSACA(t)
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	sshPub, _ := ssh.NewPublicKey(pub)

	result, err := SignCertificate(&SignRequest{
		PublicKey:       sshPub,
		CAPrivateKeyPEM: caPEM,
		Serial:          40,
		Principal:       "ecdsacauser",
		ClientID:        "c1",
		TTLSeconds:      3600,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Serial != 40 {
		t.Errorf("expected serial 40, got %d", result.Serial)
	}
}

// Test JWK conversion for RSA and ECDSA keys
func TestPublicKeyToJWK_RSA(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubBytes, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	jwkMap, err := PublicKeyToJWK(pubPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jwkMap["kty"] != "RSA" {
		t.Errorf("expected kty RSA, got %v", jwkMap["kty"])
	}
	if jwkMap["n"] == nil {
		t.Error("expected n field")
	}
	if jwkMap["e"] == nil {
		t.Error("expected e field")
	}
}

func TestPublicKeyToJWK_ECDSA(t *testing.T) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pubBytes, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	jwkMap, err := PublicKeyToJWK(pubPEM)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if jwkMap["kty"] != "EC" {
		t.Errorf("expected kty EC, got %v", jwkMap["kty"])
	}
	if jwkMap["crv"] != "P-256" {
		t.Errorf("expected crv P-256, got %v", jwkMap["crv"])
	}
}
