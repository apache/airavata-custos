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
	"errors"
	"regexp"
	"testing"
)

func TestPrincipalRegex(t *testing.T) {
	re := regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)

	tests := []struct {
		name  string
		input string
		valid bool
	}{
		// Valid cases
		{"simple lowercase", "jdoe", true},
		{"underscore start", "_user", true},
		{"with hyphen", "j-doe", true},
		{"with underscore", "j_doe", true},
		{"with numbers", "user123", true},
		{"single char", "a", true},
		{"32 chars", "abcdefghijklmnopqrstuvwxyz012345", true},

		// Invalid cases
		{"uppercase", "JDoe", false},
		{"dot", "j.doe", false},
		{"at sign", "jdoe@example", false},
		{"digit start", "1user", false},
		{"33 chars", "abcdefghijklmnopqrstuvwxyz0123456", false},
		{"empty", "", false},
		{"space", "j doe", false},
		{"special chars", "user$", false},
		{"John.Doe", "John.Doe", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := re.MatchString(tt.input)
			if got != tt.valid {
				t.Errorf("principalRegex.MatchString(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestIsDuplicateKeyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"generic error", errors.New("connection refused"), false},
		{"mysql duplicate entry", errors.New("Error 1062 (23000): Duplicate entry '42' for key 'serial_number'"), true},
		{"wrapped duplicate entry", errors.New("failed to insert: Duplicate entry '99' for key 'PRIMARY'"), true},
		{"partial match", errors.New("Duplicate entry"), true},
		{"case sensitive no match", errors.New("duplicate entry"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDuplicateKeyError(tt.err)
			if got != tt.want {
				t.Errorf("isDuplicateKeyError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
