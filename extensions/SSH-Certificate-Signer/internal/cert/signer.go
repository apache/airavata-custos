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
	"crypto/rand"
	"fmt"
	"maps"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// These go in to the extensions of a certificate
const (
	ExtensionUserSub     = "user-sub@custos.apache.org"
	ExtensionUserEmail   = "user-email@custos.apache.org"
	ExtensionRequestedBy = "requested-by@custos.apache.org"
)

type SignRequest struct {
	PublicKey       ssh.PublicKey
	CAPrivateKeyPEM []byte
	Serial          uint64
	Principal       string
	UserEmail       string
	UserSub         string
	RequestedBy     string
	TTLSeconds      uint64
	Extensions      map[string]string
	CriticalOptions map[string]string
}

type SignResult struct {
	CertBytes     []byte // raw certificate bytes
	AuthorizedKey string // the certificate as one OpenSSH line
	Serial        uint64
	ValidAfter    uint64
	ValidBefore   uint64
	CAFingerprint string
	KeyID         string
}

func SignCertificate(req *SignRequest) (*SignResult, error) {
	_, sshSigner, err := ParseCAPrivateKey(req.CAPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("parsing CA key: %w", err)
	}

	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	now := time.Now().UTC()
	validAfter := uint64(now.Unix())
	validBefore := uint64(now.Unix()) + req.TTLSeconds

	// sshd writes the key ID to the login log
	keyID := req.UserEmail + ":" + req.RequestedBy

	extensions := make(map[string]string, len(req.Extensions)+3)
	maps.Copy(extensions, req.Extensions)
	extensions[ExtensionUserSub] = req.UserSub
	extensions[ExtensionUserEmail] = req.UserEmail
	extensions[ExtensionRequestedBy] = req.RequestedBy

	cert := &ssh.Certificate{
		Nonce:           nonce,
		Key:             req.PublicKey,
		Serial:          req.Serial,
		CertType:        ssh.UserCert,
		KeyId:           keyID,
		ValidPrincipals: []string{req.Principal},
		ValidAfter:      validAfter,
		ValidBefore:     validBefore,
		Permissions: ssh.Permissions{
			CriticalOptions: req.CriticalOptions,
			Extensions:      extensions,
		},
	}

	if err := cert.SignCert(rand.Reader, sshSigner); err != nil {
		return nil, fmt.Errorf("signing certificate: %w", err)
	}

	caFingerprint := ssh.FingerprintSHA256(sshSigner.PublicKey())

	return &SignResult{
		CertBytes:     cert.Marshal(),
		AuthorizedKey: strings.TrimSpace(string(ssh.MarshalAuthorizedKey(cert))),
		Serial:        req.Serial,
		ValidAfter:    validAfter,
		ValidBefore:   validBefore,
		CAFingerprint: caFingerprint,
		KeyID:         keyID,
	}, nil
}
