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

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  addMember,
  approveChangeRequest,
  getAllocation,
  getChangeRequest,
  listAllocationMembers,
  listAllocationResources,
  listAllocationUsage,
  listAllocations,
  listChangeRequestEvents,
  listChangeRequests,
  rejectChangeRequest,
  removeMember,
  submitChangeRequest,
  updateMember,
} from "./api";
import type {
  CreateChangeRequestPayload,
  CreateMembershipPayload,
  UpdateMembershipPayload,
} from "./schemas";
import type { AllocationListParams, ChangeRequestListParams } from "./types";

export const allocationKeys = {
  all: ["allocations"] as const,
  list: (params: AllocationListParams = {}) =>
    [...allocationKeys.all, "list", params] as const,
  detail: (id: string) => [...allocationKeys.all, "detail", id] as const,
  resources: (id: string) => [...allocationKeys.detail(id), "resources"] as const,
  usage: (id: string) => [...allocationKeys.detail(id), "usage"] as const,
  members: (id: string) => [...allocationKeys.detail(id), "members"] as const,
  changeRequests: (params: ChangeRequestListParams = {}) =>
    [...allocationKeys.all, "change-requests", "list", params] as const,
  changeRequestDetail: (id: string) =>
    [...allocationKeys.all, "change-requests", "detail", id] as const,
  changeRequestEvents: (id: string) =>
    [...allocationKeys.all, "change-requests", "events", id] as const,
};

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
    queryFn: () => listAllocations({ project_id: projectId as string }),
    enabled: Boolean(projectId),
    ...DEFAULTS,
  });
}

export function useAllocation(id: string | undefined) {
  return useQuery({
    queryKey: id ? allocationKeys.detail(id) : [...allocationKeys.all, "detail", "none"],
    queryFn: () => getAllocation(id as string),
    enabled: Boolean(id),
    ...DEFAULTS,
  });
}

export function useAllocationResources(id: string | undefined) {
  return useQuery({
    queryKey: id ? allocationKeys.resources(id) : [...allocationKeys.all, "resources", "none"],
    queryFn: () => listAllocationResources(id as string),
    enabled: Boolean(id),
    ...DEFAULTS,
  });
}

export function useAllocationUsage(id: string | undefined) {
  return useQuery({
    queryKey: id ? allocationKeys.usage(id) : [...allocationKeys.all, "usage", "none"],
    queryFn: () => listAllocationUsage(id as string),
    enabled: Boolean(id),
    ...DEFAULTS,
  });
}

export function useAllocationMembers(id: string | undefined) {
  return useQuery({
    queryKey: id ? allocationKeys.members(id) : [...allocationKeys.all, "members", "none"],
    queryFn: () => listAllocationMembers(id as string),
    enabled: Boolean(id),
    ...DEFAULTS,
  });
}

export function useAddMember(allocationId: string) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateMembershipPayload) => addMember(payload),
    onSuccess: () => client.invalidateQueries({ queryKey: allocationKeys.members(allocationId) }),
  });
}

export function useUpdateMember(allocationId: string) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: UpdateMembershipPayload }) =>
      updateMember(id, patch),
    onSuccess: () => client.invalidateQueries({ queryKey: allocationKeys.members(allocationId) }),
  });
}

export function useRemoveMember(allocationId: string) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => removeMember(id),
    onSuccess: () => client.invalidateQueries({ queryKey: allocationKeys.members(allocationId) }),
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
    queryFn: () => getChangeRequest(id as string),
    enabled: Boolean(id),
    ...DEFAULTS,
  });
}

export function useChangeRequestEvents(id: string | undefined) {
  return useQuery({
    queryKey: id
      ? allocationKeys.changeRequestEvents(id)
      : [...allocationKeys.all, "change-requests", "events", "none"],
    queryFn: () => listChangeRequestEvents(id as string),
    enabled: Boolean(id),
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
      client.invalidateQueries({ queryKey: [...allocationKeys.all, "change-requests"] });
    },
  });
}

export function useRejectChangeRequest() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, approverId, comment }: { id: string; approverId: string; comment?: string }) =>
      rejectChangeRequest(id, approverId, comment),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: [...allocationKeys.all, "change-requests"] });
    },
  });
}
