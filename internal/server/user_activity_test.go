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

package server

import (
	"strings"
	"testing"
	"time"

	"github.com/apache/airavata-custos/pkg/identity"
)

func TestActivityWindowParam(t *testing.T) {
	if _, err := activityWindowParam("", 0); err == nil {
		t.Error("the list endpoint must require an explicit window")
	}
	window, err := activityWindowParam("", defaultActivityWindow)
	if err != nil || window != defaultActivityWindow {
		t.Errorf("analytics default: got %d, %v", window, err)
	}
	if window, err = activityWindowParam("7", defaultActivityWindow); err != nil || window != 7 {
		t.Errorf("explicit window: got %d, %v", window, err)
	}
	// A malformed window must not silently render another window's data.
	if _, err := activityWindowParam("thirty", defaultActivityWindow); err == nil {
		t.Error("a non-numeric window must be rejected")
	}
	// Range checking belongs to the service, so the handler passes it through.
	if window, err = activityWindowParam("400", defaultActivityWindow); err != nil || window != 400 {
		t.Errorf("out-of-range windows are the service's to reject: got %d, %v", window, err)
	}
}

func TestTrustedLoginEventPrefersTokenIdentity(t *testing.T) {
	caller := &identity.Caller{
		UserID:    "u-1",
		Issuer:    "https://issuer.example.org",
		Subject:   "sub-1",
		SessionID: "sid-1",
		AuthTime:  1_700_000_000,
		TokenID:   "jti-1",
		IssuedAt:  1_700_000_600,
	}
	input, err := trustedLoginEvent(caller)
	if err != nil {
		t.Fatalf("derive event: %v", err)
	}
	if !input.OccurredAt.Equal(time.Unix(caller.IssuedAt, 0).UTC()) {
		t.Errorf("occurred_at should follow the token issue time, got %s", input.OccurredAt)
	}
	if input.Provider != "oidc" || input.SessionID != "sid-1" {
		t.Errorf("provider/session: got %q/%q", input.Provider, input.SessionID)
	}
	if !strings.HasPrefix(input.EventKey, "v1:") {
		t.Errorf("event key should carry its scheme version, got %q", input.EventKey)
	}
}

func TestTrustedLoginEventFallsBackToSessionClaims(t *testing.T) {
	caller := &identity.Caller{
		UserID:    "u-1",
		Issuer:    "https://issuer.example.org",
		Subject:   "sub-1",
		SessionID: "sid-1",
		AuthTime:  1_700_000_000,
	}
	input, err := trustedLoginEvent(caller)
	if err != nil {
		t.Fatalf("derive event: %v", err)
	}
	if !input.OccurredAt.Equal(time.Unix(caller.AuthTime, 0).UTC()) {
		t.Errorf("occurred_at should follow auth_time, got %s", input.OccurredAt)
	}
}

func TestTrustedLoginEventRejectsMissingEvidence(t *testing.T) {
	cases := map[string]*identity.Caller{
		"no claims at all": {UserID: "u-1"},
		"token id without an issue time": {
			UserID: "u-1", Subject: "sub-1", TokenID: "jti-1",
		},
		"session without an auth time": {
			UserID: "u-1", Subject: "sub-1", SessionID: "sid-1",
		},
	}
	for name, caller := range cases {
		if _, err := trustedLoginEvent(caller); err == nil {
			t.Errorf("%s: expected rejection", name)
		}
	}
}

func TestHashedEventKeySeparatesDistinctSignIns(t *testing.T) {
	const issuer = "https://issuer.example.org"
	base := hashedEventKey("token", issuer, "sub-1", "jti-1", 100)

	// Two sign-ins under one identity-provider session still issue distinct
	// tokens, so they must not collapse into a single event.
	if base == hashedEventKey("token", issuer, "sub-1", "jti-2", 100) {
		t.Error("a different token id must produce a different key")
	}
	if base == hashedEventKey("token", issuer, "sub-1", "jti-1", 101) {
		t.Error("a different issue time must produce a different key")
	}
	if base == hashedEventKey("token", issuer, "sub-2", "jti-1", 100) {
		t.Error("a different subject must produce a different key")
	}
	if base == hashedEventKey("token", "https://other.example.org", "sub-1", "jti-1", 100) {
		t.Error("a different issuer must produce a different key")
	}
	if base == hashedEventKey("session", issuer, "sub-1", "jti-1", 100) {
		t.Error("token and session evidence must not share a key")
	}
	if base != hashedEventKey("token", issuer, "sub-1", "jti-1", 100) {
		t.Error("the same evidence must always produce the same key")
	}
	// NUL separation stops a value that contains the separator from forging
	// the key of a different claim layout.
	if hashedEventKey("token", issuer, "sub-1\x00jti-9", "", 100) ==
		hashedEventKey("token", issuer, "sub-1", "jti-9", 100) {
		t.Error("claim values must not be able to span separators")
	}
}
