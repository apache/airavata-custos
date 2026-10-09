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
	"github.com/apache/airavata-custos/internal/store"
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

// RoleDetailResponse is a role with its privilege bundle.
type RoleDetailResponse struct {
	Role       models.Role           `json:"role" binding:"required"`
	Privileges []models.PrivilegeKey `json:"privileges" extensions:"x-nullable"`
}

// ProjectListResponse is the paginated list envelope for projects.
type ProjectListResponse struct {
	Items []ProjectResponse `json:"items" binding:"required"`
	Total int               `json:"total" binding:"required"`
}

// OrganizationListResponse is the paginated list envelope for organizations.
type OrganizationListResponse struct {
	Items []models.Organization `json:"items" binding:"required"`
	Total int                   `json:"total" binding:"required"`
}

// UserListResponse is the paginated list envelope for users.
type UserListResponse struct {
	Items []models.User `json:"items" binding:"required"`
	Total int           `json:"total" binding:"required"`
}

// ProjectMemberResponse is one row in the project members tab.
type ProjectMemberResponse = store.ProjectMember

// AllocationMembershipResponse embeds the persisted membership and surfaces
// the joined user display fields plus the project-level role (defaulted to
// MEMBER when no project_roles row exists).
type AllocationMembershipResponse struct {
	models.ComputeAllocationMembership
	Role        string `json:"role" binding:"required"`
	DisplayName string `json:"display_name,omitempty"`
	Email       string `json:"email,omitempty"`
}

// ClusterAccountResponse is a cluster user with the user's name and email and the cluster name, for the review list.
type ClusterAccountResponse struct {
	models.ComputeClusterUser
	DisplayName string `json:"display_name" binding:"required"`
	Email       string `json:"email" binding:"required"`
	ClusterName string `json:"cluster_name" binding:"required"`
}

// ClusterAccountListResponse is the paginated list envelope for cluster accounts.
type ClusterAccountListResponse struct {
	Items []ClusterAccountResponse `json:"items" binding:"required"`
	Total int                      `json:"total" binding:"required"`
}

// ComputeAllocationListResponse is the paginated list envelope for compute
// allocations.
type ComputeAllocationListResponse struct {
	Items []models.ComputeAllocation `json:"items" binding:"required"`
	Total int                        `json:"total" binding:"required"`
}

// AllocationSUTotalResponse is the response for total SUs consumed on an
// allocation.
type AllocationSUTotalResponse struct {
	ComputeAllocationID string `json:"compute_allocation_id" binding:"required"`
	TotalSUAmount       int64  `json:"total_su_amount" binding:"required"`
}

// UserAllocationSUTotalResponse is the response for SUs consumed by one user
// on one allocation.
type UserAllocationSUTotalResponse struct {
	ComputeAllocationID string `json:"compute_allocation_id" binding:"required"`
	UserID              string `json:"user_id" binding:"required"`
	TotalSUAmount       int64  `json:"total_su_amount" binding:"required"`
}
