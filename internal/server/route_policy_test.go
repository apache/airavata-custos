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

package server

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// A RequireAuth route with a path id would ship without any per-resource
// check: RequireAuth verifies the caller only. Id-scoped reads must register
// through RequireScoped (or RequirePrivilege when admin-only).
func TestNoRequireAuthOnIdScopedRoutes(t *testing.T) {
	src, err := os.ReadFile("server.go")
	if err != nil {
		t.Fatalf("read server.go: %v", err)
	}
	re := regexp.MustCompile(`RequireAuth\("[A-Z]+ [^"]*\{[^"]*"`)
	for _, line := range strings.Split(string(src), "\n") {
		if re.MatchString(line) {
			t.Errorf("id-scoped route registered with RequireAuth (needs RequireScoped or RequirePrivilege): %s", strings.TrimSpace(line))
		}
	}
}
