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

package auth

import (
	"testing"
)

func TestParseClientID_NoColon(t *testing.T) {
	_, _, err := ParseClientID("nocolonhere")
	if err == nil {
		t.Error("expected error for client ID without colon")
	}
}

func TestParseClientID_MultipleColons(t *testing.T) {
	tenant, client, err := ParseClientID("t1:c1:extra")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenant != "t1" {
		t.Errorf("expected tenant t1, got %s", tenant)
	}
	if client != "c1:extra" {
		t.Errorf("expected client c1:extra, got %s", client)
	}
}

func TestParseClientID_Empty(t *testing.T) {
	_, _, err := ParseClientID("")
	if err == nil {
		t.Error("expected error for empty client ID")
	}
}

func TestParseClientID_Valid(t *testing.T) {
	tenant, client, err := ParseClientID("tenant1:webapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenant != "tenant1" {
		t.Errorf("expected tenant1, got %s", tenant)
	}
	if client != "webapp" {
		t.Errorf("expected webapp, got %s", client)
	}
}

func TestParseClientID_EmptyTenant(t *testing.T) {
	_, _, err := ParseClientID(":client")
	if err == nil {
		t.Error("expected error for empty tenant")
	}
}

func TestParseClientID_EmptyClient(t *testing.T) {
	_, _, err := ParseClientID("tenant:")
	if err == nil {
		t.Error("expected error for empty client")
	}
}

func TestAuthError_Interface(t *testing.T) {
	err := &AuthError{Code: "unauthorized", Message: "test", Status: 401}
	if err.Error() != "test" {
		t.Errorf("expected 'test', got %s", err.Error())
	}
}
