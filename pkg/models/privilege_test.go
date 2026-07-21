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

// Area 6: the signer certificate privileges are registered, returned by the
// catalog exactly once, and their string values match the shared contract.
func TestKnownPrivileges_IncludesSignerCertificates(t *testing.T) {
	if SignerCertificatesRead != "signer:certificates:read" {
		t.Errorf("SignerCertificatesRead = %q, want %q", SignerCertificatesRead, "signer:certificates:read")
	}
	if SignerCertificatesWrite != "signer:certificates:write" {
		t.Errorf("SignerCertificatesWrite = %q, want %q", SignerCertificatesWrite, "signer:certificates:write")
	}

	known := KnownPrivileges()
	c := counts(known)

	for _, want := range []PrivilegeKey{SignerCertificatesRead, SignerCertificatesWrite} {
		switch c[want] {
		case 0:
			t.Errorf("privilege %q is not registered in the catalog", want)
		case 1:
			// exactly once — good
		default:
			t.Errorf("privilege %q registered %d times, want exactly 1", want, c[want])
		}
		if !IsKnownPrivilege(want) {
			t.Errorf("IsKnownPrivilege(%q) = false, want true", want)
		}
	}

	// The catalog as a whole must contain no duplicates.
	if len(c) != len(known) {
		t.Errorf("catalog contains duplicates: %d entries but %d unique keys", len(known), len(c))
	}
}

// Area 6: adding the signer privileges did not disturb the existing catalog.
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

// Area 6: Register is idempotent — re-registering the signer keys must not
// create duplicate catalog entries.
func TestRegister_IdempotentForSignerCertificates(t *testing.T) {
	Register(SignerCertificatesRead, SignerCertificatesWrite)

	c := counts(KnownPrivileges())
	if c[SignerCertificatesRead] != 1 {
		t.Errorf("after re-register, %q count = %d, want 1", SignerCertificatesRead, c[SignerCertificatesRead])
	}
	if c[SignerCertificatesWrite] != 1 {
		t.Errorf("after re-register, %q count = %d, want 1", SignerCertificatesWrite, c[SignerCertificatesWrite])
	}
}
