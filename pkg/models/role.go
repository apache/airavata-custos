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

// SystemRoleSuperAdmin is the bootstrap role: carries privileges:grant
// and roles:manage. Created on first server start.
const SystemRoleSuperAdmin = "super_admin"

// Role is a named bundle of privileges. Mutating the bundle propagates to
// every holder.
type Role struct {
	ID          string    `json:"id"          db:"id"`
	Name        string    `json:"name"        db:"name"`
	Description *string   `json:"description" db:"description"`
	IsSystem    bool      `json:"is_system"   db:"is_system"`
	CreatedAt   time.Time `json:"created_at"  db:"created_at"`
}

// UserRole is one role assignment. Revoke is DELETE; history lives in
// audit_events.
type UserRole struct {
	UserID    string    `json:"user_id"    db:"user_id"`
	RoleID    string    `json:"role_id"    db:"role_id"`
	GrantedBy *string   `json:"granted_by" db:"granted_by"`
	GrantedAt time.Time `json:"granted_at" db:"granted_at"`
	Reason    *string   `json:"reason"     db:"reason"`
}
