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
	"bytes"
	"encoding/pem"
	"fmt"
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

// ParseCAPrivateKey reads the CA private key. Only ed25519 keys in OpenSSH format are accepted.
func ParseCAPrivateKey(data []byte) (ssh.Signer, error) {
	if block, _ := pem.Decode(data); block == nil || block.Type != "OPENSSH PRIVATE KEY" {
		return nil, fmt.Errorf("CA private key must be in OpenSSH format")
	}
	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("parsing CA private key: %w", err)
	}
	if signer.PublicKey().Type() != ssh.KeyAlgoED25519 {
		return nil, fmt.Errorf("CA key must be ed25519, got %s", signer.PublicKey().Type())
	}
	return signer, nil
}

// ParseCAPublicKey reads the CA public key in OpenSSH format.
func ParseCAPublicKey(data []byte) (ssh.PublicKey, error) {
	pub, err := ParseSSHPublicKey(string(data))
	if err != nil {
		return nil, fmt.Errorf("parsing CA public key: %w", err)
	}
	if pub.Type() != ssh.KeyAlgoED25519 {
		return nil, fmt.Errorf("CA key must be ed25519, got %s", pub.Type())
	}
	return pub, nil
}

// CheckCAKeysMatch returns an error unless publicKey belongs to privateKey.
func CheckCAKeysMatch(privateKey, publicKey []byte) error {
	signer, err := ParseCAPrivateKey(privateKey)
	if err != nil {
		return err
	}
	pub, err := ParseCAPublicKey(publicKey)
	if err != nil {
		return err
	}
	if !bytes.Equal(signer.PublicKey().Marshal(), pub.Marshal()) {
		return fmt.Errorf("CA public key does not match the private key")
	}
	return nil
}

func CAFingerprint(publicKey []byte) (string, error) {
	pub, err := ParseCAPublicKey(publicKey)
	if err != nil {
		return "", err
	}
	return ssh.FingerprintSHA256(pub), nil
}
