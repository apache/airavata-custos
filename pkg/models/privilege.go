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

import (
	"slices"
	"sync"
	"time"
)

// PrivilegeKey names a fine-grained admin capability.
type PrivilegeKey string

const (
	ClustersRead          PrivilegeKey = "core:clusters:read"
	ClustersWrite         PrivilegeKey = "core:clusters:write"
	AllocationsRead       PrivilegeKey = "core:allocations:read"
	AllocationsWrite      PrivilegeKey = "core:allocations:write"
	ProjectsRead          PrivilegeKey = "core:projects:read"
	ProjectsWrite         PrivilegeKey = "core:projects:write"
	UsersRead             PrivilegeKey = "core:users:read"
	UsersWrite            PrivilegeKey = "core:users:write"
	OrganizationsRead     PrivilegeKey = "core:organizations:read"
	OrganizationsWrite    PrivilegeKey = "core:organizations:write"
	TracesRead            PrivilegeKey = "core:traces:read"
	PrivilegesGrant       PrivilegeKey = "core:privileges:grant"
	RolesManage           PrivilegeKey = "core:roles:manage"
	EventsManage          PrivilegeKey = "core:events:manage"
	AgreementsRead        PrivilegeKey = "core:agreements:read"
	AgreementsManage      PrivilegeKey = "core:agreements:manage"
	ServiceAccountsManage PrivilegeKey = "core:service-accounts:manage"
)

// privilegeRules holds what we know about a privilege.
type privilegeRules struct {
	allowedForServiceAccounts bool
}

var (
	registryMu sync.RWMutex
	registry   = map[PrivilegeKey]privilegeRules{}
)

func init() {
	Register(
		ClustersRead,
		ClustersWrite,
		AllocationsRead,
		AllocationsWrite,
		ProjectsRead,
		ProjectsWrite,
		UsersRead,
		UsersWrite,
		OrganizationsRead,
		OrganizationsWrite,
		TracesRead,
		PrivilegesGrant,
		RolesManage,
		EventsManage,
		AgreementsRead,
		AgreementsManage,
		ServiceAccountsManage,
	)
}

// Register adds privilege keys to the catalog. Idempotent.
func Register(keys ...PrivilegeKey) {
	registryMu.Lock()
	defer registryMu.Unlock()
	for _, k := range keys {
		if _, ok := registry[k]; !ok {
			registry[k] = privilegeRules{}
		}
	}
}

// AllowForServiceAccounts registers privileges that can be given to a service account.
// Never use it for privileges that manage user access, roles, privileges, or service accounts.
func AllowForServiceAccounts(keys ...PrivilegeKey) {
	registryMu.Lock()
	defer registryMu.Unlock()
	for _, k := range keys {
		registry[k] = privilegeRules{allowedForServiceAccounts: true}
	}
}

// KnownPrivileges returns the registered keys, sorted.
func KnownPrivileges() []PrivilegeKey {
	return registeredKeys(func(privilegeRules) bool { return true })
}

// ServiceAccountPrivileges returns the privileges a service account can hold, sorted.
func ServiceAccountPrivileges() []PrivilegeKey {
	return registeredKeys(func(rules privilegeRules) bool { return rules.allowedForServiceAccounts })
}

func registeredKeys(include func(privilegeRules) bool) []PrivilegeKey {
	registryMu.RLock()
	defer registryMu.RUnlock()
	keys := []PrivilegeKey{} // not nil, so the API returns [] instead of null
	for k, rules := range registry {
		if include(rules) {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	return keys
}

// IsServiceAccountPrivilege reports whether a service account can hold p.
func IsServiceAccountPrivilege(p PrivilegeKey) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[p].allowedForServiceAccounts
}

// IsKnownPrivilege reports whether p is registered.
func IsKnownPrivilege(p PrivilegeKey) bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	_, ok := registry[p]
	return ok
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
