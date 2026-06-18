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
