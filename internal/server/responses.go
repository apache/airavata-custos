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
	"time"

	"github.com/apache/airavata-custos/pkg/models"
)

// ProjectResponse is the API response shape for a project. It embeds the core
// Project entity and adds display fields the portal renders so the entity
// itself stays free of presentation concerns.
type ProjectResponse struct {
	models.Project
	ProjectPIDisplayName string `json:"project_pi_display_name,omitempty"`
	ProjectPIEmail       string `json:"project_pi_email,omitempty"`
}

// ProjectListResponse is the paginated list envelope for projects.
type ProjectListResponse struct {
	Items []ProjectResponse `json:"items"`
	Total int               `json:"total"`
}

// ProjectMemberAllocationRef carries the (allocation, role) the user holds on
// one of the project's allocations.
type ProjectMemberAllocationRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// ProjectMemberResponse is one row in the project members tab. Role is the
// strongest role the user holds across the project's allocations; the raw
// per-allocation roles live in Allocations.
type ProjectMemberResponse struct {
	ID          string                       `json:"id"`
	ProjectID   string                       `json:"project_id"`
	UserID      string                       `json:"user_id"`
	Email       string                       `json:"email"`
	DisplayName string                       `json:"display_name"`
	Role        string                       `json:"role"`
	Status      string                       `json:"status"`
	AddedTime   time.Time                    `json:"added_time"`
	Allocations []ProjectMemberAllocationRef `json:"allocations"`
}

// AllocationMembershipResponse embeds the persisted membership and surfaces
// the joined user display fields plus the project-level role (defaulted to
// MEMBER when no project_memberships row exists).
type AllocationMembershipResponse struct {
	models.ComputeAllocationMembership
	Role        string `json:"role"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email,omitempty"`
}

// ComputeAllocationListResponse is the paginated list envelope for compute
// allocations.
type ComputeAllocationListResponse struct {
	Items []models.ComputeAllocation `json:"items"`
	Total int                        `json:"total"`
}

// AllocationSUTotalResponse is the response for total SUs consumed on an
// allocation.
type AllocationSUTotalResponse struct {
	ComputeAllocationID string `json:"compute_allocation_id"`
	TotalSUAmount       int64  `json:"total_su_amount"`
}

// UserAllocationSUTotalResponse is the response for SUs consumed by one user
// on one allocation.
type UserAllocationSUTotalResponse struct {
	ComputeAllocationID string `json:"compute_allocation_id"`
	UserID              string `json:"user_id"`
	TotalSUAmount       int64  `json:"total_su_amount"`
}
