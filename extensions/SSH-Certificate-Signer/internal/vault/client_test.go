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

package vault

import (
	"errors"
	"fmt"
	"testing"

	vaultapi "github.com/openbao/openbao/api/v2"
)

func TestIsCASMismatch(t *testing.T) {
	casErr := &vaultapi.ResponseError{
		StatusCode: 400,
		Errors:     []string{"check-and-set parameter did not match the current version"},
	}
	if !isCASMismatch(casErr) {
		t.Error("CAS mismatch response not detected")
	}
	if !isCASMismatch(fmt.Errorf("wrapped: %w", casErr)) {
		t.Error("wrapped CAS mismatch response not detected")
	}

	for _, msg := range []string{
		"permission denied",
		"check-and-set parameter required for this call",
		"error parsing check-and-set parameter",
	} {
		other := &vaultapi.ResponseError{StatusCode: 400, Errors: []string{msg}}
		if isCASMismatch(other) {
			t.Errorf("misclassified as conflict: %q", msg)
		}
	}
	if isCASMismatch(errors.New("connection refused")) {
		t.Error("plain error misclassified as conflict")
	}
}

func TestSerialConflictWrapping(t *testing.T) {
	// same shape writeMetadataCAS produces
	err := fmt.Errorf("%w (CAS=7): %v", ErrSerialConflict, errors.New("underlying"))
	if !errors.Is(err, ErrSerialConflict) {
		t.Error("wrapped conflict not matched by errors.Is")
	}
}
