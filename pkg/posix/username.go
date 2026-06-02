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

// Package posix builds and validates POSIX-conformant usernames for HPC
// account provisioning.
package posix

import (
	"os"
	"strings"
	"unicode"

	"github.com/apache/airavata-custos/pkg/models"
)

const MaxCollisionSuffix = 999

// BuildBase returns the unsuffixed username and a flag set when the name
// portion was truncated to fit the 32-char Unix login limit.
func BuildBase(u *models.User, prefix string) (string, bool) {
	first := Normalize(u.FirstName)
	last := Normalize(u.LastName)

	var local string
	switch {
	case first != "" && last != "":
		local = string(first[0]) + last
	case last != "":
		local = last
	case first != "":
		local = first
	default:
		local = "user"
	}

	// Reserve 3 chars for a numeric collision suffix (up to "999").
	maxLocal := 32 - len(prefix) - 1 - 3
	truncated := false
	if len(local) > maxLocal {
		local = local[:maxLocal]
		truncated = true
	}
	return prefix + "-" + local, truncated
}

func Normalize(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if r > unicode.MaxASCII {
			continue
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func Prefix() string {
	if v := os.Getenv("POSIX_USERNAME_PREFIX"); v != "" {
		return v
	}
	return "custos"
}
