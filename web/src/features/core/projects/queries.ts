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
import type { GetProjectsData, ProjectListResponse } from "@/generated/core/types.gen";
import { ApiError, type KnownError } from "@/shared/api/client";
import { listOf } from "@/shared/api/listOf";
import { paged } from "@/shared/api/paged";
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
  record: (id: string) => [...projectKeys.all, "record", id] as const,
  members: (projectId: string) => [...projectKeys.all, "members", projectId] as const,
};

// Allocation member rows carry the project role.
function invalidateAllocationMembers(client: QueryClient) {
  client.invalidateQueries({
    queryKey: allocationKeys.all,
    predicate: (q) => q.queryKey.at(-1) === "members",
  });
}

// Master answers 400 when a user ID names no user; malformed IDs are refused before sending.
const UNKNOWN_USER: KnownError = { status: 400, copy: "No user has this ID." };
const PROJECT_GONE: KnownError = { status: 404, copy: "This project no longer exists." };

// A 404 means the project is gone; resetting its detail lets the page's gone state take over.
function forgetIfGone(client: QueryClient, err: unknown, id: string) {
  if (err instanceof ApiError && err.status === 404)
    void client.resetQueries({ queryKey: projectKeys.detail(id) });
}

// Master ignores the query for callers without projects:read and returns every project they take part in.
export function participantPage(
  { items }: ProjectListResponse,
  { q, status, limit = 50, offset = 0 }: NonNullable<GetProjectsData["query"]>,
): ProjectListResponse {
  const needle = q?.toLowerCase();
  const matches = (items ?? []).filter(
    (p) =>
      (!status || p.status === status) &&
      (!needle ||
        p.title.toLowerCase().includes(needle) ||
        p.originated_id.toLowerCase().includes(needle)),
  );
  return { items: matches.slice(offset, offset + limit), total: matches.length };
}

export function useProjects(query: NonNullable<GetProjectsData["query"]>, readAll: boolean) {
  return useQuery({
    queryKey: readAll ? projectKeys.list(query) : projectKeys.list({}),
    queryFn: () => getProjects(readAll ? { query } : {}),
    select: readAll ? undefined : (data: ProjectListResponse) => participantPage(data, query),
    ...paged,
  });
}

export function useProject(id: string | undefined) {
  return useQuery({
    queryKey: projectKeys.detail(id ?? ""),
    queryFn: id ? () => getProjectsById({ path: { id } }) : skipToken,
  });
}

// The edit dialog's own read, so a 404 shows in the dialog instead of flipping the page.
export function useProjectRecord(id: string | undefined) {
  return useQuery({
    queryKey: projectKeys.record(id ?? ""),
    queryFn: id ? () => getProjectsById({ path: { id } }) : skipToken,
  });
}

// Drops a project another tab deleted from every cached list.
export function useRefreshProjectLists() {
  const client = useQueryClient();
  return () => client.invalidateQueries({ queryKey: [...projectKeys.all, "list"] });
}

// The edit dialog found the project gone: the page refetches its 404 instead of offering stale controls.
export function useForgetProject(id: string | undefined) {
  const client = useQueryClient();
  const refreshLists = useRefreshProjectLists();
  return () => {
    refreshLists();
    if (id) void client.resetQueries({ queryKey: projectKeys.detail(id) });
  };
}

export function useProjectMembers(projectId: string | undefined) {
  const client = useQueryClient();
  return useQuery({
    queryKey: projectKeys.members(projectId ?? ""),
    queryFn: projectId
      ? () =>
          getProjectsByIdMembers({ path: { id: projectId } })
            .then(listOf)
            .catch((err) => {
              // The project is gone; the page must stop offering Edit and Delete on its cached copy.
              if (err instanceof ApiError && err.status === 404) {
                void client.resetQueries({ queryKey: projectKeys.detail(projectId) });
              }
              throw err;
            })
      : skipToken,
  });
}

export function useCreateProject() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: postProjects<true>,
    meta: { errors: [UNKNOWN_USER] },
    onSettled: () => client.invalidateQueries({ queryKey: projectKeys.all }),
  });
}

export function useUpdateProject() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: putProjectsById<true>,
    meta: { errors: [UNKNOWN_USER, PROJECT_GONE] },
    onError: (err, { path }) => forgetIfGone(client, err, path.id),
    onSettled: () => {
      client.invalidateQueries({ queryKey: projectKeys.all });
      // A PI change retags project roles shown on allocation member rows.
      invalidateAllocationMembers(client);
    },
  });
}

export function useUpdateProjectStatus() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: putProjectsByIdStatus<true>,
    meta: { errors: [PROJECT_GONE] },
    onError: (err, { path }) => forgetIfGone(client, err, path.id),
    onSettled: () => client.invalidateQueries({ queryKey: projectKeys.all }),
  });
}

export function useDeleteProject() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: deleteProjectsById<true>,
    // Master refuses a project delete with 409 only while allocations remain.
    meta: {
      errors: [{ status: 409, copy: "Delete this project's allocations first." }, PROJECT_GONE],
    },
    onError: (err, { path }) => {
      // The allocation added elsewhere must show, and this project's Delete must go.
      if (err instanceof ApiError && err.status === 409)
        client.invalidateQueries({ queryKey: [...allocationKeys.all, "list"] });
      forgetIfGone(client, err, path.id);
    },
    onSettled: () => client.invalidateQueries({ queryKey: [...projectKeys.all, "list"] }),
    // Back to the deleted project must refetch (404), not render the cached ACTIVE copy.
    onSuccess: (_data, { path }) => client.removeQueries({ queryKey: projectKeys.detail(path.id) }),
  });
}

export function useSetProjectRole() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (options: Parameters<typeof putProjectsByIdMembersByUserId<true>>[0]) =>
      putProjectsByIdMembersByUserId<true>(options).catch(async (err) => {
        // Master answers 400 for a deleted project as for an unknown user; a re-read tells them apart.
        if (err instanceof ApiError && err.status === 400) {
          const gone = await getProjectsById({ path: { id: options.path.id } }).then(
            () => false,
            (read) => read instanceof ApiError && read.status === 404,
          );
          if (gone) {
            void client.resetQueries({ queryKey: projectKeys.detail(options.path.id) });
            throw new Error(PROJECT_GONE.copy);
          }
        }
        throw err;
      }),
    meta: { errors: [UNKNOWN_USER] },
    onSettled: (_data, _error, { path }) => {
      client.invalidateQueries({ queryKey: projectKeys.members(path.id) });
      invalidateAllocationMembers(client);
    },
  });
}
