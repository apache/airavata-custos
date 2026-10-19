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

"use client";

import { skipToken, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  addProjectMember,
  createProject,
  deleteProject,
  getProject,
  listProjectMembers,
  listProjects,
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
    queryFn: id ? () => getProject(id) : skipToken,
    ...DEFAULTS,
  });
}

export function useProjectMembers(projectId: string | undefined) {
  return useQuery({
    queryKey: projectId
      ? projectKeys.members(projectId)
      : [...projectKeys.all, "members", "none"],
    queryFn: projectId ? () => listProjectMembers(projectId) : skipToken,
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

export function useDeleteProject() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: deleteProject,
    // Marks lists stale without refetching the deleted project's own queries.
    onSuccess: () => client.invalidateQueries({ queryKey: projectKeys.all, refetchType: "none" }),
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
    mutationFn: ({ userId, payload }: { userId: string; payload: UpdateProjectMemberPayload }) =>
      updateProjectMember(projectId, userId, payload),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: projectKeys.members(projectId) });
      // Allocation member rows show the project role; a key prefix avoids importing allocations.
      client.invalidateQueries({ queryKey: ["allocations"] });
    },
  });
}
