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

package models

import "testing"

// counts returns how many times each privilege appears in the catalog.
func counts(keys []PrivilegeKey) map[PrivilegeKey]int {
	m := make(map[PrivilegeKey]int, len(keys))
	for _, k := range keys {
		m[k]++
	}
	return m
}

func TestKnownPrivileges_ExistingRegistrationUnchanged(t *testing.T) {
	existing := []PrivilegeKey{
		ClustersRead, ClustersWrite,
		AllocationsRead, AllocationsWrite,
		ProjectsRead, ProjectsWrite,
		UsersRead, UsersWrite,
		OrganizationsRead, OrganizationsWrite,
		TracesRead, PrivilegesGrant, RolesManage,
	}
	for _, k := range existing {
		if !IsKnownPrivilege(k) {
			t.Errorf("existing privilege %q is no longer registered", k)
		}
	}
}

func TestRegister_IdempotentForExtensionPrivileges(t *testing.T) {
	key := PrivilegeKey("test-extension:certificates:read")
	Register(key, key)

	c := counts(KnownPrivileges())
	if c[key] != 1 {
		t.Errorf("after duplicate registration, %q count = %d, want 1", key, c[key])
	}
	if !IsKnownPrivilege(key) {
		t.Errorf("IsKnownPrivilege(%q) = false, want true", key)
	}
}
