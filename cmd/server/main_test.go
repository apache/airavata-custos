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

package main

import (
	"testing"

	"github.com/apache/airavata-custos/pkg/models"
)

func TestRegisterExtensionPrivileges(t *testing.T) {
	key := "test-main:extension:read"
	registerExtensionPrivileges([]string{key, key})
	if !models.IsKnownPrivilege(models.PrivilegeKey(key)) {
		t.Fatalf("extension privilege %q was not registered", key)
	}
	count := 0
	for _, known := range models.KnownPrivileges() {
		if known == models.PrivilegeKey(key) {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("extension privilege registration count = %d, want 1", count)
	}
}
