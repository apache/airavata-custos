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

// TODO(openapi): replace with generated from projects.openapi.yaml
import { apiFetch } from "@/shared/api/client";
import {
  type AddProjectMemberPayload,
  type CreateProjectPayload,
  type Project,
  type ProjectListEnvelope,
  type ProjectMember,
  type UpdateProjectMemberPayload,
  type UpdateProjectStatusPayload,
  addProjectMemberPayloadSchema,
  createProjectPayloadSchema,
  projectListEnvelopeSchema,
  projectMemberSchema,
  projectMembersResponseSchema,
  projectSchema,
  updateProjectMemberPayloadSchema,
  updateProjectStatusPayloadSchema,
} from "./schemas";
import type { ProjectListParams } from "./types";

export async function listProjects(params: ProjectListParams = {}): Promise<ProjectListEnvelope> {
  const search = new URLSearchParams();
  if (typeof params.limit === "number") search.set("limit", String(params.limit));
  if (typeof params.offset === "number") search.set("offset", String(params.offset));
  if (params.pi_id) search.set("pi_id", params.pi_id);
  if (params.status) search.set("status", params.status);
  if (params.q) search.set("q", params.q);
  const qs = search.toString();
  const raw = await apiFetch(`/projects${qs ? `?${qs}` : ""}`);
  return projectListEnvelopeSchema.parse(raw);
}

export async function getProject(id: string): Promise<Project> {
  const raw = await apiFetch(`/projects/${id}`);
  return projectSchema.parse(raw);
}

export async function createProject(payload: CreateProjectPayload): Promise<Project> {
  const validated = createProjectPayloadSchema.parse(payload);
  const raw = await apiFetch("/projects", { method: "POST", body: validated });
  return projectSchema.parse(raw);
}

export async function updateProjectStatus(
  id: string,
  payload: UpdateProjectStatusPayload,
): Promise<Project> {
  const validated = updateProjectStatusPayloadSchema.parse(payload);
  const raw = await apiFetch(`/projects/${id}/status`, { method: "PUT", body: validated });
  return projectSchema.parse(raw);
}

export async function listProjectMembers(projectId: string): Promise<ProjectMember[]> {
  const raw = await apiFetch(`/projects/${projectId}/members`);
  return projectMembersResponseSchema.parse(raw ?? []);
}

export async function addProjectMember(
  projectId: string,
  payload: AddProjectMemberPayload,
): Promise<ProjectMember> {
  const validated = addProjectMemberPayloadSchema.parse(payload);
  const raw = await apiFetch(`/projects/${projectId}/members`, {
    method: "POST",
    body: validated,
  });
  return projectMemberSchema.parse(raw);
}

export async function updateProjectMember(
  projectId: string,
  memberId: string,
  payload: UpdateProjectMemberPayload,
): Promise<ProjectMember> {
  const validated = updateProjectMemberPayloadSchema.parse(payload);
  const raw = await apiFetch(`/projects/${projectId}/members/${memberId}`, {
    method: "PUT",
    body: validated,
  });
  return projectMemberSchema.parse(raw);
}

export async function removeProjectMember(projectId: string, memberId: string): Promise<void> {
  await apiFetch(`/projects/${projectId}/members/${memberId}`, { method: "DELETE" });
}
