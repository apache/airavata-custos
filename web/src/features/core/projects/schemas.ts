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

import { z } from "zod";
import { zProjectStatus, zUserStatus, zUserType } from "@/generated/core/zod.gen";

export const projectStatusSchema = zProjectStatus;
export type ProjectStatus = z.infer<typeof projectStatusSchema>;

export const userStatusSchema = zUserStatus;
export type UserStatus = z.infer<typeof userStatusSchema>;

// Mirrors pkg/models/project.go#Project. The backend only exposes
// POST /projects, GET /projects/{id}, PUT /projects/{id}/status today; list and
// member endpoints below are MSW-served until the backend catches up.
export const projectSchema = z.object({
  id: z.string(),
  originated_id: z.string(),
  title: z.string(),
  origination: z.string(),
  project_pi_id: z.string(),
  project_pi_display_name: z.string().optional(),
  project_pi_email: z.string().optional(),
  status: projectStatusSchema,
  created_time: z.string(),
});
export type Project = z.infer<typeof projectSchema>;

export const userSchema = z.object({
  id: z.string(),
  organization_id: z.string(),
  first_name: z.string(),
  last_name: z.string(),
  middle_name: z.string().optional(),
  email: z.string(),
  status: userStatusSchema,
});
export type User = z.infer<typeof userSchema>;

// Portal-shaped paginated envelope. Backend has no list endpoint yet — MSW
// returns this shape, the real route will mirror it.
export const projectListEnvelopeSchema = z.object({
  items: z.array(projectSchema),
  total: z.number().int().nonnegative(),
});
export type ProjectListEnvelope = z.infer<typeof projectListEnvelopeSchema>;

export const projectMemberRoleSchema = z.enum(["PI", "CO_PI", "ALLOCATION_MANAGER", "MEMBER"]);
export type ProjectMemberRole = z.infer<typeof projectMemberRoleSchema>;

// Portal-side aggregation. No backend ProjectMember model yet; in practice
// project membership is derived from per-allocation memberships (see Phase 5).
// MSW returns this composite shape so the UI is ready when the contract lands.
export const projectMemberAllocationSchema = z.object({
  id: z.string(),
  name: z.string(),
  role: z.enum(["PI", "CO_PI", "ALLOCATION_MANAGER", "MEMBER"]),
});
export type ProjectMemberAllocation = z.infer<typeof projectMemberAllocationSchema>;

export const userTypeSchema = zUserType;
export type UserType = z.infer<typeof userTypeSchema>;

export const projectMemberSchema = z.object({
  id: z.string(),
  project_id: z.string(),
  user_id: z.string(),
  email: z.string(),
  display_name: z.string(),
  role: projectMemberRoleSchema,
  status: projectStatusSchema,
  type: userTypeSchema.optional(),
  added_time: z.string(),
  allocations: z.array(projectMemberAllocationSchema).default([]),
});
export type ProjectMember = z.infer<typeof projectMemberSchema>;

export const projectMembersResponseSchema = z.array(projectMemberSchema);
export type ProjectMembersResponse = z.infer<typeof projectMembersResponseSchema>;

export const createProjectPayloadSchema = z.object({
  title: z.string().min(1).max(200),
  origination: z.string().min(1).max(64),
  project_pi_id: z.string().min(1),
  originated_id: z.string().optional(),
});
export type CreateProjectPayload = z.infer<typeof createProjectPayloadSchema>;

export const updateProjectStatusPayloadSchema = z.object({
  status: projectStatusSchema,
});
export type UpdateProjectStatusPayload = z.infer<typeof updateProjectStatusPayloadSchema>;

export const addProjectMemberPayloadSchema = z.object({
  user_id: z.string().min(1),
  role: projectMemberRoleSchema.default("MEMBER"),
});
export type AddProjectMemberPayload = z.input<typeof addProjectMemberPayloadSchema>;

export const updateProjectMemberPayloadSchema = z
  .object({
    role: projectMemberRoleSchema.optional(),
    status: projectStatusSchema.optional(),
  })
  .refine((v) => Object.keys(v).length > 0, { message: "patch is empty" });
export type UpdateProjectMemberPayload = z.infer<typeof updateProjectMemberPayloadSchema>;
