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

package service

import (
	"errors"
	"testing"
	"time"

	"github.com/apache/airavata-custos/pkg/identity"
)

func TestResolveLoginLocalDateBucketsByUserCalendar(t *testing.T) {
	// 01:30 UTC belongs to the previous day in New York and the same day in
	// Shanghai.
	instant := time.Date(2026, 9, 20, 1, 30, 0, 0, time.UTC)
	cases := []struct {
		timezone string
		wantZone string
		wantDate string
	}{
		{"", "UTC", "2026-09-20"},
		{"UTC", "UTC", "2026-09-20"},
		{"America/New_York", "America/New_York", "2026-09-19"},
		{"Asia/Shanghai", "Asia/Shanghai", "2026-09-20"},
		{"Not/AZone", "UTC", "2026-09-20"},
	}
	for _, tc := range cases {
		zone, date := resolveLoginLocalDate(instant, tc.timezone)
		if zone != tc.wantZone || date != tc.wantDate {
			t.Errorf("timezone %q: got %s/%s, want %s/%s",
				tc.timezone, zone, date, tc.wantZone, tc.wantDate)
		}
	}
}

func TestResolveLoginLocalDateAcrossDaylightSaving(t *testing.T) {
	// The US spring-forward transition: 06:30 UTC is still March 8 locally.
	instant := time.Date(2026, 3, 8, 6, 30, 0, 0, time.UTC)
	if _, date := resolveLoginLocalDate(instant, "America/New_York"); date != "2026-03-08" {
		t.Errorf("spring forward: got %s, want 2026-03-08", date)
	}
	// After the transition the same wall clock maps to a different offset,
	// which must not shift the calendar day.
	instant = time.Date(2026, 3, 9, 3, 30, 0, 0, time.UTC)
	if _, date := resolveLoginLocalDate(instant, "America/New_York"); date != "2026-03-08" {
		t.Errorf("after spring forward: got %s, want 2026-03-08", date)
	}
}

func TestRecordLoginEventValidatesTrustedInput(t *testing.T) {
	// A nil store proves validation happens before any persistence.
	svc := &Service{}
	cases := map[string]struct {
		userID string
		input  RecordLoginEventInput
	}{
		"missing user": {"", RecordLoginEventInput{EventKey: "k", OccurredAt: nowUTC()}},
		"missing key":  {"u-1", RecordLoginEventInput{OccurredAt: nowUTC()}},
		"missing time": {"u-1", RecordLoginEventInput{EventKey: "k"}},
		"future time": {"u-1", RecordLoginEventInput{
			EventKey:   "k",
			OccurredAt: nowUTC().Add(loginEventFutureSkew + time.Minute),
		}},
	}
	for name, tc := range cases {
		if _, err := svc.RecordLoginEvent(t.Context(), tc.userID, tc.input); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%s: got %v, want ErrInvalidInput", name, err)
		}
	}
}

func TestCallerWithSessionClaimsNeverServesStaleTokenEvidence(t *testing.T) {
	// The resolver caches one caller per subject, so per-token claims must be
	// layered on for each request instead of being cached with it.
	cached := &identity.Caller{UserID: "u-1"}
	first := callerWithSessionClaims(cached, &identity.Claims{
		Issuer: "https://issuer.example.org", Sub: "sub-1",
		SessionID: "sid-1", AuthTime: 100, TokenID: "jti-1", IssuedAt: 110,
	})
	second := callerWithSessionClaims(cached, &identity.Claims{
		Issuer: "https://issuer.example.org", Sub: "sub-1",
		SessionID: "sid-2", AuthTime: 200, TokenID: "jti-2", IssuedAt: 210,
	})
	if first.TokenID == second.TokenID || first.IssuedAt == second.IssuedAt {
		t.Error("each request must carry its own token evidence")
	}
	if first.UserID != "u-1" || second.UserID != "u-1" {
		t.Error("the resolved user must survive the merge")
	}
	if cached.TokenID != "" || cached.SessionID != "" {
		t.Error("the cached caller must not be mutated")
	}
}
