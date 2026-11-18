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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/airavata-custos/signer/internal/krl"
)

type fixedKRL krl.Snapshot

func (f fixedKRL) Snapshot() krl.Snapshot { return krl.Snapshot(f) }

// Make sure a node gets the list with its version, and a 304 with no body once it has that version.
func TestKRL_ServesTheListOnceUntilTheVersionChanges(t *testing.T) {
	h := NewKRLHandler(fixedKRL{Bytes: []byte("list"), Version: 7})

	first := httptest.NewRecorder()
	h.Handle(first, httptest.NewRequest(http.MethodGet, "/api/v1/krl", nil))
	if first.Code != http.StatusOK || first.Body.String() != "list" || first.Header().Get("ETag") != `"7"` {
		t.Fatalf("got %d %q etag %s", first.Code, first.Body, first.Header().Get("ETag"))
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/krl", nil)
	req.Header.Set("If-None-Match", `"7"`)
	again := httptest.NewRecorder()
	h.Handle(again, req)
	if again.Code != http.StatusNotModified || again.Body.Len() != 0 {
		t.Fatalf("got %d with %d bytes, want 304 and no body", again.Code, again.Body.Len())
	}
}
