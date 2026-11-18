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

package krl

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// Make sure OpenSSH reads the correct serials it was given.
func TestBuild_RevokesExactlyTheListedSerials(t *testing.T) {
	keygen, err := exec.LookPath("ssh-keygen")
	if err != nil {
		t.Skip("ssh-keygen not installed")
	}
	dir := t.TempDir()
	_, caKey, _ := ed25519.GenerateKey(rand.Reader)
	ca, _ := ssh.NewSignerFromKey(caKey)
	certFile := func(serial uint64) string {
		pub, _, _ := ed25519.GenerateKey(rand.Reader)
		key, _ := ssh.NewPublicKey(pub)
		cert := &ssh.Certificate{Key: key, Serial: serial, CertType: ssh.UserCert, ValidPrincipals: []string{"jdoe"}, ValidBefore: ssh.CertTimeInfinity}
		if err := cert.SignCert(rand.Reader, ca); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, fmt.Sprintf("cert-%d.pub", serial))
		if err := os.WriteFile(path, ssh.MarshalAuthorizedKey(cert), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	verdict := func(serials []int64, certPath string) string {
		krlPath := filepath.Join(dir, "revoked.krl")
		if err := os.WriteFile(krlPath, Build(1, time.Now(), serials), 0o600); err != nil {
			t.Fatal(err)
		}
		out, _ := exec.Command(keygen, "-Q", "-f", krlPath, certPath).CombinedOutput()
		return string(out)
	}

	revoked, other := certFile(77), certFile(78)
	if out := verdict([]int64{77}, revoked); !strings.Contains(out, "REVOKED") {
		t.Fatalf("serial 77 should be revoked: %s", out)
	}
	if out := verdict([]int64{77}, other); !strings.Contains(out, "ok") {
		t.Fatalf("serial 78 should pass: %s", out)
	}
	if out := verdict(nil, revoked); !strings.Contains(out, "ok") {
		t.Fatalf("an empty list should revoke nothing: %s", out)
	}
}
