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
	"errors"
	"testing"
)

func policyErrorCode(t *testing.T, err error) string {
	t.Helper()
	var policyErr *PolicyError
	if !errors.As(err, &policyErr) {
		t.Fatalf("expected PolicyError, got %v", err)
	}
	return policyErr.Code
}

// Make sure the signer caps certificate length even for its trusted caller.
func TestEnforce_RefusesLifetimeOverTheLimit(t *testing.T) {
	e := NewEnforcer(3600, []string{"ed25519"})
	if code := policyErrorCode(t, e.Enforce(3601, "ssh-ed25519")); code != "invalid_ttl" {
		t.Errorf("code = %s", code)
	}
	if err := e.Enforce(3600, "ssh-ed25519"); err != nil {
		t.Errorf("lifetime at the limit was refused: %v", err)
	}
}

// Key type that is not allowed will have its own error code
func TestEnforce_RefusesKeyTypeNotAllowed(t *testing.T) {
	e := NewEnforcer(3600, []string{"ed25519"})
	if code := policyErrorCode(t, e.Enforce(60, "ssh-rsa")); code != "key_type_not_allowed" {
		t.Errorf("code = %s", code)
	}
}
