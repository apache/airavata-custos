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

package validation

import (
	"testing"
)

func TestNoOpValidator_AllowsAll(t *testing.T) {
	v := NewNoOpValidator()
	result, err := v.Validate("t1", "c1", "anyuser", "sub123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Error("expected allowed")
	}
	if result.ValidatedPrincipal != "anyuser" {
		t.Errorf("expected anyuser, got %s", result.ValidatedPrincipal)
	}
}

func TestCOmanageValidator_AlwaysDenies(t *testing.T) {
	v := NewCOmanageValidator()
	result, err := v.Validate("t1", "c1", "user", "sub123")
	if result != nil {
		t.Error("expected nil result")
	}
	if err == nil {
		t.Fatal("expected error")
	}
	valErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if valErr.ReasonCode != "COMANAGE_NOT_IMPLEMENTED" {
		t.Errorf("expected COMANAGE_NOT_IMPLEMENTED, got %s", valErr.ReasonCode)
	}
	if valErr.Message != "Requested principal is not allowed" {
		t.Errorf("expected 'Requested principal is not allowed', got %s", valErr.Message)
	}
}

func TestValidationError_Interface(t *testing.T) {
	err := &ValidationError{
		Message:    "Requested principal is not allowed",
		ReasonCode: "NOT_ALLOWED",
	}
	if err.Error() != "Requested principal is not allowed" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}
