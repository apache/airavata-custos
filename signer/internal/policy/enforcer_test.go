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

package policy

import (
	"testing"

	"github.com/apache/airavata-custos/signer/internal/store"
)

func clientCfg(maxTTL int, keyTypes []string, sourceRestriction *string, critOpts map[string]string) *store.ClientConfig {
	return &store.ClientConfig{
		MaxTTLSeconds:            maxTTL,
		AllowedKeyTypes:          keyTypes,
		SourceAddressRestriction: sourceRestriction,
		CriticalOptions:          critOpts,
	}
}

func strPtr(s string) *string {
	return &s
}

func TestEnforce_TTLWithinLimit(t *testing.T) {
	e := NewEnforcer(86400, []string{"ed25519", "rsa", "ecdsa"})
	cfg := clientCfg(86400, []string{"ed25519"}, nil, nil)
	err := e.Enforce(3600, "ssh-ed25519", "10.0.0.1", cfg)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestEnforce_TTLExceedsLimit(t *testing.T) {
	e := NewEnforcer(86400, []string{"ed25519"})
	cfg := clientCfg(3600, []string{"ed25519"}, nil, nil)
	err := e.Enforce(7200, "ssh-ed25519", "10.0.0.1", cfg)
	if err == nil {
		t.Fatal("expected error for TTL exceeding limit")
	}
	pErr, ok := err.(*PolicyError)
	if !ok {
		t.Fatalf("expected PolicyError, got %T", err)
	}
	if pErr.Message != "Requested TTL 7200 seconds exceeds maximum allowed TTL 3600 seconds" {
		t.Errorf("unexpected message: %s", pErr.Message)
	}
}

func TestEnforce_TTLAtExactLimit(t *testing.T) {
	e := NewEnforcer(86400, []string{"ed25519"})
	cfg := clientCfg(3600, []string{"ed25519"}, nil, nil)
	err := e.Enforce(3600, "ssh-ed25519", "10.0.0.1", cfg)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestEnforce_NullMaxTTLUsesDefault(t *testing.T) {
	e := NewEnforcer(86400, []string{"ed25519"})
	cfg := clientCfg(0, []string{"ed25519"}, nil, nil) // 0 = null equivalent
	err := e.Enforce(86400, "ssh-ed25519", "10.0.0.1", cfg)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestEnforce_TTLZero(t *testing.T) {
	e := NewEnforcer(86400, []string{"ed25519"})
	cfg := clientCfg(86400, []string{"ed25519"}, nil, nil)
	err := e.Enforce(0, "ssh-ed25519", "10.0.0.1", cfg)
	if err == nil {
		t.Fatal("expected error for TTL 0")
	}
}

func TestEnforce_NegativeTTL(t *testing.T) {
	e := NewEnforcer(86400, []string{"ed25519"})
	cfg := clientCfg(86400, []string{"ed25519"}, nil, nil)
	err := e.Enforce(-100, "ssh-ed25519", "10.0.0.1", cfg)
	if err == nil {
		t.Fatal("expected error for negative TTL")
	}
}

func TestEnforce_DisallowedKeyType(t *testing.T) {
	e := NewEnforcer(86400, []string{"ed25519"})
	cfg := clientCfg(86400, []string{"ed25519"}, nil, nil)
	err := e.Enforce(3600, "ssh-rsa", "10.0.0.1", cfg)
	if err == nil {
		t.Fatal("expected error for disallowed key type")
	}
	pErr := err.(*PolicyError)
	if pErr == nil {
		t.Fatal("expected PolicyError")
	}
}

func TestEnforce_AllowedKeyType(t *testing.T) {
	e := NewEnforcer(86400, []string{"ed25519", "rsa"})
	cfg := clientCfg(86400, []string{"ed25519", "rsa"}, nil, nil)
	err := e.Enforce(3600, "ssh-rsa", "10.0.0.1", cfg)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestEnforce_SourceAddressRejected(t *testing.T) {
	e := NewEnforcer(86400, []string{"ed25519"})
	cidr := "10.0.0.0/8"
	cfg := clientCfg(86400, []string{"ed25519"}, &cidr, nil)
	err := e.Enforce(3600, "ssh-ed25519", "192.168.1.1", cfg)
	if err == nil {
		t.Fatal("expected error for source address outside CIDR")
	}
}

func TestEnforce_SourceAddressAccepted(t *testing.T) {
	e := NewEnforcer(86400, []string{"ed25519"})
	cidr := "10.0.0.0/8"
	cfg := clientCfg(86400, []string{"ed25519"}, &cidr, nil)
	err := e.Enforce(3600, "ssh-ed25519", "10.1.2.3", cfg)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestEnforce_NoSourceRestriction(t *testing.T) {
	e := NewEnforcer(86400, []string{"ed25519"})
	cfg := clientCfg(86400, []string{"ed25519"}, nil, nil)
	err := e.Enforce(3600, "ssh-ed25519", "192.168.1.1", cfg)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestGetCriticalOptions(t *testing.T) {
	opts := map[string]string{"source-address": "10.0.0.0/8"}
	cfg := clientCfg(86400, nil, nil, opts)
	result := GetCriticalOptions(cfg)
	if result == nil {
		t.Fatal("expected non-nil critical options")
	}
	if result["source-address"] != "10.0.0.0/8" {
		t.Errorf("expected source-address, got %v", result)
	}
}

func TestGetCriticalOptions_Nil(t *testing.T) {
	cfg := clientCfg(86400, nil, nil, nil)
	result := GetCriticalOptions(cfg)
	if result != nil {
		t.Errorf("expected nil critical options, got %v", result)
	}
}
