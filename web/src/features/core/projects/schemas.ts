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

import {
  zGetProjectsResponse,
  zPostProjectsBody,
  zProjectMemberAllocationRef,
  zProjectMemberResponse,
  zProjectStatus,
  zPutProjectsByIdMembersByUserIdBody,
  zPutProjectsByIdStatusBody,
} from "@/generated/core/zod.gen";
import { z } from "zod";

export const projectListEnvelopeSchema = zGetProjectsResponse.transform((value) => ({
  ...value,
  items: value.items ?? [],
}));
export type ProjectListEnvelope = z.infer<typeof projectListEnvelopeSchema>;

export const projectMemberRoleSchema = z.enum(["PI", "CO_PI", "ALLOCATION_MANAGER", "MEMBER"]);
export type ProjectMemberRole = z.infer<typeof projectMemberRoleSchema>;

// The backend types role and status as plain strings; the portal narrows them to the values it
// renders.
export const projectMemberAllocationSchema = zProjectMemberAllocationRef.extend({
  role: projectMemberRoleSchema,
});
export type ProjectMemberAllocation = z.infer<typeof projectMemberAllocationSchema>;

export const projectMemberSchema = zProjectMemberResponse.extend({
  role: projectMemberRoleSchema,
  status: zProjectStatus,
  allocations: z
    .array(projectMemberAllocationSchema)
    .nullable()
    .transform((value) => value ?? []),
});
export type ProjectMember = z.infer<typeof projectMemberSchema>;

export const projectMembersResponseSchema = z.array(projectMemberSchema);

export const createProjectPayloadSchema = zPostProjectsBody.pick({ originated_id: true }).extend({
  title: z.string().min(1).max(200),
  origination: z.string().min(1).max(64),
  project_pi_id: z.string().min(1),
});
export type CreateProjectPayload = z.infer<typeof createProjectPayloadSchema>;

export const updateProjectStatusPayloadSchema = zPutProjectsByIdStatusBody.required();
export type UpdateProjectStatusPayload = z.infer<typeof updateProjectStatusPayloadSchema>;

// The backend has no add-member route; MSW serves it.
export const addProjectMemberPayloadSchema = z.object({
  user_id: z.string().min(1),
  role: projectMemberRoleSchema.default("MEMBER"),
});
export type AddProjectMemberPayload = z.input<typeof addProjectMemberPayloadSchema>;

export const updateProjectMemberPayloadSchema = zPutProjectsByIdMembersByUserIdBody
  .extend({ role: projectMemberRoleSchema.optional() })
  .refine((v) => Object.keys(v).length > 0, { message: "patch is empty" });
export type UpdateProjectMemberPayload = z.infer<typeof updateProjectMemberPayloadSchema>;
