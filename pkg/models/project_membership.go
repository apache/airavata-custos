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

// ProjectRole names the project-level governance roles persisted in
// project_memberships. MEMBER is not stored here; it is derived from a user
// having at least one compute allocation membership on the project.
type ProjectRole string

const (
	ProjectRolePI                ProjectRole = "PI"
	ProjectRoleCoPI              ProjectRole = "CO_PI"
	ProjectRoleAllocationManager ProjectRole = "ALLOCATION_MANAGER"
)

// ProjectMembership ties a user to a project-level governance role.
type ProjectMembership struct {
	ProjectID string      `json:"project_id" db:"project_id"`
	UserID    string      `json:"user_id"    db:"user_id"`
	Role      ProjectRole `json:"role"       db:"role"`
	AddedTime time.Time   `json:"added_time" db:"added_time"`
}
