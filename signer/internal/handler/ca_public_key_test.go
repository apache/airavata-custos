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
)

func TestCAPublicKeyHandler_MissingClientConfig(t *testing.T) {
	h := NewCAPublicKeyHandler(nil, nil)

	req := httptest.NewRequest("GET", "/api/v1/ca-public-key", nil)
	rec := httptest.NewRecorder()
	h.Handle(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] != "internal_error" {
		t.Errorf("expected error 'internal_error', got %q", body["error"])
	}
	if body["message"] != "Missing client config" {
		t.Errorf("expected message 'Missing client config', got %q", body["message"])
	}
}

func TestCAPublicKeyJSONResponse_Format(t *testing.T) {
	resp := CAPublicKeyJSONResponse{
		PublicKey:   "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample",
		Fingerprint: "SHA256:abcdef123456",
		Algorithm:   "ed25519",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, key := range []string{"public_key", "fingerprint", "algorithm"} {
		if _, ok := parsed[key]; !ok {
			t.Errorf("missing key %q in JSON output", key)
		}
	}

	if parsed["public_key"] != "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample" {
		t.Errorf("unexpected public_key: %v", parsed["public_key"])
	}
	if parsed["fingerprint"] != "SHA256:abcdef123456" {
		t.Errorf("unexpected fingerprint: %v", parsed["fingerprint"])
	}
	if parsed["algorithm"] != "ed25519" {
		t.Errorf("unexpected algorithm: %v", parsed["algorithm"])
	}
}
