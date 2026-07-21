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
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/apache/airavata-custos/pkg/models"
)

const MaxCollisionSuffix = 999

var ErrUnbuildableUsername = errors.New("posix: cannot build username from empty first and last name")

// BuildBase returns the unsuffixed username. 'truncated' is set when the name
// portion was shortened to fit the 32-char POSIX login cap.
func BuildBase(u *models.User, prefix string) (string, bool, error) {
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
		return "", false, fmt.Errorf("%w: user %q (first=%q last=%q)", ErrUnbuildableUsername, u.ID, u.FirstName, u.LastName)
	}

	// 32 = POSIX login cap; -1 separator, -3 reserved for collision suffix (up to "999").
	maxLocal := 32 - len(prefix) - 1 - 3
	truncated := false
	if len(local) > maxLocal {
		local = local[:maxLocal]
		truncated = true
	}
	return prefix + "-" + local, truncated, nil
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

// ValidChosen reports whether s is acceptable as a user-picked login: 1 to 32
// characters, starting with a lowercase letter, then lowercase letters, digits,
// hyphen, or underscore. Unlike a generated name it carries no prefix.
func ValidChosen(s string) bool {
	if s == "" || len(s) > 32 {
		return false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case i > 0 && (r >= '0' && r <= '9' || r == '-' || r == '_'):
		default:
			return false
		}
	}
	return true
}
