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

import {
  type QueryClient,
  skipToken,
  useMutation,
  useQueries,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { toast } from "sonner";
import {
  addMember,
  approveChangeRequest,
  createAllocation,
  deleteAllocation,
  getAllocation,
  getAllocationUsedSU,
  getChangeRequest,
  listAllocationDiffs,
  listAllocationMembers,
  listAllocationResources,
  listAllocationUsage,
  listUserMemberships,
  listAllocations,
  listChangeRequestEvents,
  listChangeRequests,
  rejectChangeRequest,
  removeMember,
  setAllocationResources,
  submitChangeRequest,
} from "./api";
import type {
  AllocationMembership,
  CreateAllocationPayload,
  CreateChangeRequestPayload,
  CreateMembershipPayload,
} from "./schemas";
import type { ComputeAllocation } from "@/generated/core/zod.gen";
import type { AllocationListParams, ChangeRequestListParams } from "./types";

export const allocationKeys = {
  all: ["allocations"] as const,
  list: (params: AllocationListParams = {}) =>
    [...allocationKeys.all, "list", params] as const,
  detail: (id: string) => [...allocationKeys.all, "detail", id] as const,
  resources: (id: string) => [...allocationKeys.detail(id), "resources"] as const,
  usage: (id: string) => [...allocationKeys.detail(id), "usage"] as const,
  usedSU: (id: string) => [...allocationKeys.detail(id), "used-su"] as const,
  diffs: (id: string) => [...allocationKeys.detail(id), "diffs"] as const,
  members: (id: string) => [...allocationKeys.detail(id), "members"] as const,
  userMemberships: (userId: string) => [...allocationKeys.all, "user-memberships", userId] as const,
  changeRequests: (params: ChangeRequestListParams = {}) =>
    [...allocationKeys.all, "change-requests", "list", params] as const,
  changeRequestDetail: (id: string) =>
    [...allocationKeys.all, "change-requests", "detail", id] as const,
  changeRequestEvents: (id: string) =>
    [...allocationKeys.all, "change-requests", "events", id] as const,
};

// Project members are derived from allocation memberships; a key prefix avoids importing projects.
function invalidateMembers(client: QueryClient, allocationId: string) {
  client.invalidateQueries({ queryKey: allocationKeys.members(allocationId) });
  client.invalidateQueries({ queryKey: ["projects", "members"] });
}

const DEFAULTS = {
  staleTime: 30_000,
  gcTime: 300_000,
  refetchOnWindowFocus: false,
} as const;

export function useAllocations(params: AllocationListParams = {}) {
  return useQuery({
    queryKey: allocationKeys.list(params),
    queryFn: () => listAllocations(params),
    ...DEFAULTS,
  });
}

// Sanctioned cross-feature hook (ADR-0004) for projects to read its allocations.
export function useAllocationsByProject(projectId: string | undefined) {
  return useQuery({
    queryKey: projectId
      ? allocationKeys.list({ project_id: projectId })
      : [...allocationKeys.all, "list", "none"],
    queryFn: projectId ? () => listAllocations({ project_id: projectId }) : skipToken,
    ...DEFAULTS,
  });
}

export function useCreateAllocation() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: async ({
      resourceIds,
      ...payload
    }: CreateAllocationPayload & { resourceIds: string[] }) => {
      const created = await createAllocation(payload);
      // The allocation exists either way, so a failed grant is reported rather than retried as a create.
      if (resourceIds.length) {
        await setAllocationResources(created.id, resourceIds).catch((err: Error) =>
          toast.error(`Allocation created without its resources: ${err.message}`),
        );
      }
      return created;
    },
    onSuccess: () => client.invalidateQueries({ queryKey: allocationKeys.all }),
  });
}

export function useDeleteAllocation() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: deleteAllocation,
    // The allocation stays until an admin approves, so only the pending deletions change.
    onSuccess: () => client.invalidateQueries({ queryKey: ["deletion-requests"] }),
  });
}

export function useSetAllocationResources(allocationId: string) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (resourceIds: string[]) => setAllocationResources(allocationId, resourceIds),
    onSuccess: () => client.invalidateQueries({ queryKey: allocationKeys.all }),
  });
}

export function useAllocation(id: string | undefined) {
  return useQuery({
    queryKey: id ? allocationKeys.detail(id) : [...allocationKeys.all, "detail", "none"],
    queryFn: id ? () => getAllocation(id) : skipToken,
    ...DEFAULTS,
  });
}

export function useAllocationResources(id: string | undefined) {
  return useQuery({
    queryKey: id ? allocationKeys.resources(id) : [...allocationKeys.all, "resources", "none"],
    queryFn: id ? () => listAllocationResources(id) : skipToken,
    ...DEFAULTS,
  });
}

export function useAllocationUsedSU(id: string) {
  return useQuery({
    queryKey: allocationKeys.usedSU(id),
    queryFn: () => getAllocationUsedSU(id),
    ...DEFAULTS,
  });
}

export function useAllocationUsage(id: string | undefined) {
  return useQuery({
    queryKey: id ? allocationKeys.usage(id) : [...allocationKeys.all, "usage", "none"],
    queryFn: id ? () => listAllocationUsage(id) : skipToken,
    ...DEFAULTS,
  });
}

export function useAllocationDiffs(id: string | undefined) {
  return useQuery({
    queryKey: id ? allocationKeys.diffs(id) : [...allocationKeys.all, "diffs", "none"],
    queryFn: id ? () => listAllocationDiffs(id) : skipToken,
    ...DEFAULTS,
  });
}

export function useAllocationMembers(id: string | undefined) {
  return useQuery({
    queryKey: id ? allocationKeys.members(id) : [...allocationKeys.all, "members", "none"],
    queryFn: id ? () => listAllocationMembers(id) : skipToken,
    ...DEFAULTS,
  });
}

export function useAddMember(allocationId: string) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateMembershipPayload) => addMember(payload),
    onSuccess: () => invalidateMembers(client, allocationId),
  });
}

export function useRemoveMember(allocationId: string) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => removeMember(id),
    onSuccess: () => invalidateMembers(client, allocationId),
  });
}

// Sanctioned cross-feature hooks (ADR-0004) for the project members grid.
export function useMembershipsByAllocation(allocationIds: string[]) {
  return useQueries({
    queries: allocationIds.map((id) => ({
      queryKey: allocationKeys.members(id),
      queryFn: () => listAllocationMembers(id),
      ...DEFAULTS,
    })),
    combine: (results) => ({
      byAllocation: Object.fromEntries(
        allocationIds.map((id, i) => [id, results[i]?.data ?? []]),
      ),
      isLoading: results.some((r) => r.isLoading),
      error: results.find((r) => r.error)?.error ?? null,
      refetch: () => Promise.all(results.map((r) => r.refetch())),
    }),
  });
}

// Revokes access by deleting the membership, or grants it for the allocation's remaining term.
export function useToggleAllocationAccess() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: async ({
      allocation,
      userId,
      membership,
    }: {
      allocation: ComputeAllocation;
      userId: string;
      membership: AllocationMembership | undefined;
    }) =>
      membership
        ? removeMember(membership.id)
        : addMember({
            compute_allocation_id: allocation.id,
            user_id: userId,
            start_time: new Date().toISOString(),
            end_time: allocation.end_time,
            membership_status: "ACTIVE",
          }),
    onSuccess: (_data, { allocation }) => invalidateMembers(client, allocation.id),
  });
}

export function useChangeRequests(params: ChangeRequestListParams = {}) {
  return useQuery({
    queryKey: allocationKeys.changeRequests(params),
    queryFn: () => listChangeRequests(params),
    ...DEFAULTS,
  });
}

export function useChangeRequest(id: string | undefined) {
  return useQuery({
    queryKey: id
      ? allocationKeys.changeRequestDetail(id)
      : [...allocationKeys.all, "change-requests", "detail", "none"],
    queryFn: id ? () => getChangeRequest(id) : skipToken,
    ...DEFAULTS,
  });
}

export function useChangeRequestEvents(id: string | undefined) {
  return useQuery({
    queryKey: id
      ? allocationKeys.changeRequestEvents(id)
      : [...allocationKeys.all, "change-requests", "events", "none"],
    queryFn: id ? () => listChangeRequestEvents(id) : skipToken,
    ...DEFAULTS,
  });
}

export function useSubmitChangeRequest() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateChangeRequestPayload) => submitChangeRequest(payload),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: [...allocationKeys.all, "change-requests"] });
    },
  });
}

export function useApproveChangeRequest() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, approverId, comment }: { id: string; approverId: string; comment?: string }) =>
      approveChangeRequest(id, approverId, comment),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: allocationKeys.all });
    },
  });
}

export function useRejectChangeRequest() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, approverId, comment }: { id: string; approverId: string; comment?: string }) =>
      rejectChangeRequest(id, approverId, comment),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: allocationKeys.all });
    },
  });
}

export function useUserMemberships(userId: string | undefined, enabled: boolean) {
  return useQuery({
    queryKey: userId ? allocationKeys.userMemberships(userId) : [...allocationKeys.all, "none"],
    queryFn: userId && enabled ? () => listUserMemberships(userId) : skipToken,
    ...DEFAULTS,
  });
}
