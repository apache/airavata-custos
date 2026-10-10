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

import { allocationKeys } from "@/features/core/allocations/queries";
import {
  deleteProjectsById,
  getProjects,
  getProjectsById,
  getProjectsByIdMembers,
  postProjects,
  putProjectsById,
  putProjectsByIdMembersByUserId,
  putProjectsByIdStatus,
} from "@/generated/core/sdk.gen";
import type { GetProjectsData } from "@/generated/core/types.gen";
import { useInvalidating } from "@/shared/api/useInvalidating";
import {
  type QueryClient,
  skipToken,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";

export const projectKeys = {
  all: ["projects"] as const,
  list: (query: GetProjectsData["query"]) => [...projectKeys.all, "list", query] as const,
  detail: (id: string) => [...projectKeys.all, "detail", id] as const,
  members: (projectId: string) => [...projectKeys.all, "members", projectId] as const,
};

// Allocation member rows carry the project role.
function invalidateAllocationMembers(client: QueryClient) {
  client.invalidateQueries({
    queryKey: allocationKeys.all,
    predicate: (q) => q.queryKey.at(-1) === "members",
  });
}

export function useProjects(query: GetProjectsData["query"]) {
  return useQuery({
    queryKey: projectKeys.list(query),
    queryFn: () => getProjects({ query }),
  });
}

export function useProject(id: string | undefined) {
  return useQuery({
    queryKey: projectKeys.detail(id ?? ""),
    queryFn: id ? () => getProjectsById({ path: { id } }) : skipToken,
  });
}

export function useProjectMembers(projectId: string | undefined) {
  return useQuery({
    queryKey: projectKeys.members(projectId ?? ""),
    queryFn: projectId ? () => getProjectsByIdMembers({ path: { id: projectId } }) : skipToken,
  });
}

export function useCreateProject() {
  return useInvalidating(postProjects<true>, projectKeys.all);
}

export function useUpdateProject() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: putProjectsById<true>,
    onSuccess: () => {
      client.invalidateQueries({ queryKey: projectKeys.all });
      // A PI change retags project roles shown on allocation member rows.
      invalidateAllocationMembers(client);
    },
  });
}

export function useUpdateProjectStatus() {
  return useInvalidating(putProjectsByIdStatus<true>, projectKeys.all);
}

export function useDeleteProject() {
  return useInvalidating(deleteProjectsById<true>, [...projectKeys.all, "list"]);
}

export function useSetProjectRole() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: putProjectsByIdMembersByUserId<true>,
    onSuccess: (_data, { path }) => {
      client.invalidateQueries({ queryKey: projectKeys.members(path.id) });
      invalidateAllocationMembers(client);
    },
  });
}
