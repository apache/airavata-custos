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

package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/airavata-custos/signer/internal/httputil"
)

func TestUserInfoHandler_MissingIdentity(t *testing.T) {
	h := NewUserInfoHandler()

	req := httptest.NewRequest("GET", "/api/v1/userinfo", nil)
	rec := httptest.NewRecorder()
	h.Handle(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestUserInfoHandler_Success(t *testing.T) {
	h := NewUserInfoHandler()

	req := httptest.NewRequest("GET", "/api/v1/userinfo", nil)
	ctx := httputil.WithUserIdentity(req.Context(), &httputil.UserIdentityContext{
		Issuer:    "https://cilogon.org",
		Subject:   "http://cilogon.org/serverA/users/12345",
		Principal: "jdoe",
	})
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.Handle(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp UserInfoResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Subject != "http://cilogon.org/serverA/users/12345" {
		t.Errorf("subject: got %q", resp.Subject)
	}
	if resp.Issuer != "https://cilogon.org" {
		t.Errorf("issuer: got %q", resp.Issuer)
	}
	if resp.Principal != "jdoe" {
		t.Errorf("principal: got %q", resp.Principal)
	}
}

func TestUserInfoResponse_JSONFormat(t *testing.T) {
	resp := UserInfoResponse{
		Subject:   "user123",
		Issuer:    "https://example.com",
		Principal: "jdoe",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)

	for _, key := range []string{"subject", "issuer", "principal"} {
		if _, ok := parsed[key]; !ok {
			t.Errorf("missing key %q in JSON output", key)
		}
	}
}
