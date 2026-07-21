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

package posix

import (
	"errors"
	"strings"
	"testing"

	"github.com/apache/airavata-custos/pkg/models"
)

func TestBuildBase(t *testing.T) {
	tests := []struct {
		name      string
		first     string
		middle    string
		last      string
		prefix    string
		wantBase  string
		wantTrunc bool
	}{
		{
			name:  "standard",
			first: "Alice", last: "Smith", prefix: "custos",
			wantBase: "custos-asmith", wantTrunc: false,
		},
		{
			name:  "middle ignored",
			first: "Alice", middle: "Marie", last: "Smith", prefix: "custos",
			wantBase: "custos-asmith", wantTrunc: false,
		},
		{
			name:  "single name as last",
			first: "", last: "Madonna", prefix: "custos",
			wantBase: "custos-madonna", wantTrunc: false,
		},
		{
			name:  "first only",
			first: "Alice", last: "", prefix: "custos",
			wantBase: "custos-alice", wantTrunc: false,
		},
		{
			name:  "non-ASCII stripped",
			first: "Aña", last: "Şəkili", prefix: "custos",
			wantBase: "custos-akili", wantTrunc: false,
		},
		{
			name:  "prefix swap",
			first: "Alice", last: "Smith", prefix: "alt",
			wantBase: "alt-asmith", wantTrunc: false,
		},
		{
			name:  "truncation at 32-char limit",
			first: "L", last: strings.Repeat("a", 50), prefix: "custos",
			wantTrunc: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &models.User{FirstName: tc.first, MiddleName: tc.middle, LastName: tc.last}
			got, trunc, err := BuildBase(u, tc.prefix)
			if err != nil {
				t.Fatalf("BuildBase: %v", err)
			}

			if trunc != tc.wantTrunc {
				t.Errorf("truncated = %v, want %v", trunc, tc.wantTrunc)
			}
			if tc.wantBase != "" && got != tc.wantBase {
				t.Errorf("base = %q, want %q", got, tc.wantBase)
			}
			if len(got) > 32 {
				t.Errorf("base len %d > 32: %q", len(got), got)
			}
			if !strings.HasPrefix(got, tc.prefix+"-") {
				t.Errorf("base %q does not start with prefix %q", got, tc.prefix+"-")
			}
		})
	}
}

func TestBuildBase_UnbuildableReturnsError(t *testing.T) {
	cases := []struct {
		name  string
		first string
		last  string
	}{
		{"both empty", "", ""},
		{"both non-ASCII only", "李", "王"},
		{"both punctuation only", "...", "---"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := &models.User{ID: "u-1", FirstName: tc.first, LastName: tc.last}
			got, _, err := BuildBase(u, "custos")
			if !errors.Is(err, ErrUnbuildableUsername) {
				t.Fatalf("err = %v, want ErrUnbuildableUsername", err)
			}
			if got != "" {
				t.Errorf("expected empty username on error, got %q", got)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Alice", "alice"},
		{"ALLCAPS", "allcaps"},
		{"abc123", "abc123"},
		{"Aña", "aa"},
		{"Şəkili", "kili"},
		{"hello-world", "helloworld"},
		{"O'Brien", "obrien"},
		{"", ""},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got := Normalize(tc.in)
			if got != tc.want {
				t.Errorf("Normalize(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestPrefix(t *testing.T) {
	t.Setenv("POSIX_USERNAME_PREFIX", "")
	if got := Prefix(); got != "custos" {
		t.Errorf("default = %q, want %q", got, "custos")
	}
	t.Setenv("POSIX_USERNAME_PREFIX", "alt")
	if got := Prefix(); got != "alt" {
		t.Errorf("override = %q, want %q", got, "alt")
	}
}

func TestValidChosen(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"jdoe", true},
		{"a", true},
		{"user-1", true},
		{"user_1", true},
		{"j2", true},
		{"", false},
		{"1abc", false},   // must start with a letter
		{"-abc", false},   // must start with a letter
		{"JDoe", false},   // no uppercase
		{"jo hn", false},  // no space
		{"jdoe!", false},  // no punctuation
		{"nexus-x", true}, // a prefixed-looking name is still a valid literal
		{strings.Repeat("a", 33), false},
		{strings.Repeat("a", 32), true},
	}
	for _, c := range cases {
		if got := ValidChosen(c.in); got != c.want {
			t.Errorf("ValidChosen(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
