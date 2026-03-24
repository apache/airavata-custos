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

package httputil

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/apache/airavata-custos/signer/internal/store"
)

func TestClientConfigContext(t *testing.T) {
	cfg := &store.ClientConfig{
		TenantID: "t1",
		ClientID: "c1",
	}
	ctx := WithClientConfig(context.Background(), cfg)
	got := ClientConfigFromContext(ctx)
	if got == nil {
		t.Fatal("expected non-nil config")
	}
	if got.TenantID != "t1" || got.ClientID != "c1" {
		t.Errorf("unexpected config: %+v", got)
	}
}

func TestClientConfigContext_Nil(t *testing.T) {
	got := ClientConfigFromContext(context.Background())
	if got != nil {
		t.Error("expected nil for empty context")
	}
}

func TestSourceIPContext(t *testing.T) {
	ctx := WithSourceIP(context.Background(), "10.0.0.1")
	got := SourceIPFromContext(ctx)
	if got != "10.0.0.1" {
		t.Errorf("expected 10.0.0.1, got %s", got)
	}
}

func TestSourceIPContext_Empty(t *testing.T) {
	got := SourceIPFromContext(context.Background())
	if got != "" {
		t.Errorf("expected empty, got %s", got)
	}
}

func TestWriteJSONError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSONError(rec, 401, "unauthorized", "Missing header")

	if rec.Code != 401 {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Error("expected application/json content type")
	}
	body := rec.Body.String()
	if body == "" {
		t.Error("expected non-empty body")
	}
}

func TestWriteJSONErrorWithExtra(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSONErrorWithExtra(rec, 403, "principal_denied", "Not allowed", map[string]string{
		"reason_code": "COMANAGE_NOT_IMPLEMENTED",
	})

	if rec.Code != 403 {
		t.Errorf("expected 403, got %d", rec.Code)
	}
	body := rec.Body.String()
	if body == "" {
		t.Error("expected non-empty body")
	}
}
