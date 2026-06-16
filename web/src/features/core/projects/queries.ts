"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  addProjectMember,
  createProject,
  getProject,
  listProjectMembers,
  listProjects,
  removeProjectMember,
  updateProjectMember,
  updateProjectStatus,
} from "./api";
import type {
  AddProjectMemberPayload,
  CreateProjectPayload,
  UpdateProjectMemberPayload,
  UpdateProjectStatusPayload,
} from "./schemas";
import type { ProjectListParams } from "./types";

export const projectKeys = {
  all: ["projects"] as const,
  list: (params: ProjectListParams = {}) => [...projectKeys.all, "list", params] as const,
  detail: (id: string) => [...projectKeys.all, "detail", id] as const,
  members: (projectId: string) => [...projectKeys.all, "members", projectId] as const,
};

const DEFAULTS = {
  staleTime: 30_000,
  gcTime: 300_000,
  refetchOnWindowFocus: false,
} as const;

export function useProjects(params: ProjectListParams = {}) {
  return useQuery({
    queryKey: projectKeys.list(params),
    queryFn: () => listProjects(params),
    ...DEFAULTS,
  });
}

export function useProject(id: string | undefined) {
  return useQuery({
    queryKey: id ? projectKeys.detail(id) : [...projectKeys.all, "detail", "none"],
    queryFn: () => getProject(id as string),
    enabled: Boolean(id),
    ...DEFAULTS,
  });
}

export function useProjectMembers(projectId: string | undefined) {
  return useQuery({
    queryKey: projectId
      ? projectKeys.members(projectId)
      : [...projectKeys.all, "members", "none"],
    queryFn: () => listProjectMembers(projectId as string),
    enabled: Boolean(projectId),
    ...DEFAULTS,
  });
}

export function useCreateProject() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateProjectPayload) => createProject(payload),
    onSuccess: () => client.invalidateQueries({ queryKey: projectKeys.all }),
  });
}

export function useUpdateProjectStatus() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateProjectStatusPayload }) =>
      updateProjectStatus(id, payload),
    onSuccess: (_data, variables) => {
      client.invalidateQueries({ queryKey: projectKeys.all });
      client.invalidateQueries({ queryKey: projectKeys.detail(variables.id) });
    },
  });
}

export function useAddProjectMember(projectId: string) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (payload: AddProjectMemberPayload) => addProjectMember(projectId, payload),
    onSuccess: () => client.invalidateQueries({ queryKey: projectKeys.members(projectId) }),
  });
}

export function useUpdateProjectMember(projectId: string) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({
      memberId,
      payload,
    }: {
      memberId: string;
      payload: UpdateProjectMemberPayload;
    }) => updateProjectMember(projectId, memberId, payload),
    onSuccess: () => client.invalidateQueries({ queryKey: projectKeys.members(projectId) }),
  });
}

export function useRemoveProjectMember(projectId: string) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (memberId: string) => removeProjectMember(projectId, memberId),
    onSuccess: () => client.invalidateQueries({ queryKey: projectKeys.members(projectId) }),
  });
}
