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

import "time"

// PrivilegeKey names a fine-grained admin capability. The set is closed and
// declared in code; service-layer validation rejects grants of any key not
// returned by KnownPrivileges.
type PrivilegeKey string

const (
	PrivilegeAMIERead    PrivilegeKey = "amie:read"
	PrivilegeAMIEWrite   PrivilegeKey = "amie:write"
	PrivilegeHPCRead     PrivilegeKey = "hpc:read"
	PrivilegeHPCWrite    PrivilegeKey = "hpc:write"
	PrivilegeSignerRead  PrivilegeKey = "signer:read"
	PrivilegeSignerWrite PrivilegeKey = "signer:write"
	PrivilegeGrant       PrivilegeKey = "privileges:grant"
)

// KnownPrivileges returns the static catalog of declared privilege keys.
func KnownPrivileges() []PrivilegeKey {
	return []PrivilegeKey{
		PrivilegeAMIERead,
		PrivilegeAMIEWrite,
		PrivilegeHPCRead,
		PrivilegeHPCWrite,
		PrivilegeSignerRead,
		PrivilegeSignerWrite,
		PrivilegeGrant,
	}
}

// IsKnownPrivilege reports whether p is in the declared catalog.
func IsKnownPrivilege(p PrivilegeKey) bool {
	for _, k := range KnownPrivileges() {
		if k == p {
			return true
		}
	}
	return false
}

// UserPrivilege is one active grant in user_privileges. Revoked grants are
// deleted from the table; their history lives in audit_events.
type UserPrivilege struct {
	ID        string       `json:"id"          db:"id"`
	UserID    string       `json:"user_id"     db:"user_id"`
	Privilege PrivilegeKey `json:"privilege"   db:"privilege"`
	GrantedBy *string      `json:"granted_by"  db:"granted_by"`
	GrantedAt time.Time    `json:"granted_at"  db:"granted_at"`
	Reason    *string      `json:"reason"      db:"reason"`
}
