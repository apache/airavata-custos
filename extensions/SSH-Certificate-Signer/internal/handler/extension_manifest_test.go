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

package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/apache/airavata-custos/signer/internal/config"
)

func TestExtensionManifest(t *testing.T) {
	h := NewExtensionManifestHandler(config.WebConfig{BaseURL: "https://signer.example.org"})
	recorder := httptest.NewRecorder()
	h.Handle(recorder, httptest.NewRequest("GET", "/.well-known/custos-extension.json", nil))

	if recorder.Code != 200 {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q", got)
	}
	var manifest extensionManifest
	if err := json.NewDecoder(recorder.Body).Decode(&manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.SchemaVersion != 1 || manifest.BasePath != "/signer" {
		t.Fatalf("unexpected manifest: %+v", manifest)
	}
	if manifest.WebURL != "https://signer.example.org" {
		t.Fatalf("web_url = %q", manifest.WebURL)
	}
	if len(manifest.Navigation) != 1 || manifest.Navigation[0].RequiredPrivilege != "signer:certificates:read" {
		t.Fatalf("unexpected navigation: %+v", manifest.Navigation)
	}
}
