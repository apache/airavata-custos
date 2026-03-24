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

// Package cert handles SSH certificate construction and signing.
package cert

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/ssh"
)

// ParseSSHPublicKey parses an SSH public key in authorized_keys format.
func ParseSSHPublicKey(keyStr string) (ssh.PublicKey, error) {
	keyStr = strings.TrimSpace(keyStr)
	if keyStr == "" {
		return nil, fmt.Errorf("empty public key")
	}
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(keyStr))
	if err != nil {
		return nil, fmt.Errorf("parsing SSH public key: %w", err)
	}
	return pubKey, nil
}

func NormalizeKeyType(sshKeyType string) string {
	switch sshKeyType {
	case "ssh-ed25519":
		return "ed25519"
	case "ssh-rsa":
		return "rsa"
	case "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521":
		return "ecdsa"
	default:
		return sshKeyType
	}
}

func SSHPublicKeyFingerprint(key ssh.PublicKey) string {
	return ssh.FingerprintSHA256(key)
}

func ParseCAPrivateKey(pemData []byte) (crypto.Signer, ssh.Signer, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to decode PEM block")
	}

	var signer crypto.Signer
	switch block.Type {
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing PKCS8 private key: %w", err)
		}
		var ok bool
		signer, ok = key.(crypto.Signer)
		if !ok {
			return nil, nil, fmt.Errorf("private key does not implement crypto.Signer")
		}
	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing EC private key: %w", err)
		}
		signer = key
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing RSA private key: %w", err)
		}
		signer = key
	default:
		return nil, nil, fmt.Errorf("unsupported PEM block type: %s", block.Type)
	}

	var sshSigner ssh.Signer
	var err error

	// For RSA keys, use SHA-256 signer
	if _, ok := signer.(*rsa.PrivateKey); ok {
		sshSigner, err = ssh.NewSignerFromKey(signer)
		if err != nil {
			return nil, nil, fmt.Errorf("creating SSH signer: %w", err)
		}
		// Wrap to use rsa-sha2-256 algorithm
		algorithmSigner, ok := sshSigner.(ssh.AlgorithmSigner)
		if ok {
			sshSigner = &rsaSHA256Signer{inner: algorithmSigner}
		}
	} else {
		sshSigner, err = ssh.NewSignerFromKey(signer)
		if err != nil {
			return nil, nil, fmt.Errorf("creating SSH signer: %w", err)
		}
	}

	return signer, sshSigner, nil
}

// rsaSHA256Signer wraps an SSH AlgorithmSigner to always use rsa-sha2-256.
type rsaSHA256Signer struct {
	inner ssh.AlgorithmSigner
}

func (s *rsaSHA256Signer) PublicKey() ssh.PublicKey {
	return s.inner.PublicKey()
}

func (s *rsaSHA256Signer) Sign(rand io.Reader, data []byte) (*ssh.Signature, error) {
	return s.inner.SignWithAlgorithm(rand, data, ssh.KeyAlgoRSASHA256)
}

func ParseCAPublicKey(pemData []byte) (ssh.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing public key: %w", err)
	}

	sshPub, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("converting to SSH public key: %w", err)
	}
	return sshPub, nil
}

// PEMToOpenSSH converts a PEM public key to OpenSSH format, suitable for trusted-user-ca-keys files.
func PEMToOpenSSH(publicKeyPEM []byte) (string, error) {
	sshPub, err := ParseCAPublicKey(publicKeyPEM)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub))), nil
}

func CAFingerprint(publicKeyPEM []byte) (string, error) {
	sshPub, err := ParseCAPublicKey(publicKeyPEM)
	if err != nil {
		return "", err
	}
	return ssh.FingerprintSHA256(sshPub), nil
}

func PublicKeyToJWK(publicKeyPEM []byte) (map[string]interface{}, error) {
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing public key: %w", err)
	}

	sshPub, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("converting to SSH public key: %w", err)
	}
	fingerprint := ssh.FingerprintSHA256(sshPub)

	jwkMap := map[string]interface{}{
		"kid": fingerprint,
	}

	switch key := pubKey.(type) {
	case ed25519.PublicKey:
		jwkMap["kty"] = "OKP"
		jwkMap["crv"] = "Ed25519"
		jwkMap["x"] = base64.RawURLEncoding.EncodeToString([]byte(key))
	case *ecdsa.PublicKey:
		jwkMap["kty"] = "EC"
		switch key.Curve.Params().BitSize {
		case 256:
			jwkMap["crv"] = "P-256"
		case 384:
			jwkMap["crv"] = "P-384"
		case 521:
			jwkMap["crv"] = "P-521"
		}
		byteLen := (key.Curve.Params().BitSize + 7) / 8
		xBytes := key.X.Bytes()
		yBytes := key.Y.Bytes()
		// Pad to full length
		xPadded := make([]byte, byteLen)
		yPadded := make([]byte, byteLen)
		copy(xPadded[byteLen-len(xBytes):], xBytes)
		copy(yPadded[byteLen-len(yBytes):], yBytes)
		jwkMap["x"] = base64.RawURLEncoding.EncodeToString(xPadded)
		jwkMap["y"] = base64.RawURLEncoding.EncodeToString(yPadded)
	case *rsa.PublicKey:
		jwkMap["kty"] = "RSA"
		jwkMap["n"] = base64.RawURLEncoding.EncodeToString(key.N.Bytes())
		eBytes := make([]byte, 0)
		e := key.E
		for e > 0 {
			eBytes = append([]byte{byte(e & 0xff)}, eBytes...)
			e >>= 8
		}
		jwkMap["e"] = base64.RawURLEncoding.EncodeToString(eBytes)
	default:
		return nil, fmt.Errorf("unsupported public key type")
	}

	return jwkMap, nil
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(h[:])
}
