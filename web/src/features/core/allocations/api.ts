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
  zGetComputeAllocationChangeRequestsByIdEventsResponse,
  zGetComputeAllocationsByIdDiffsResponse,
  zGetComputeAllocationsByIdResourcesResponse,
  zGetComputeAllocationsByIdResponse,
  zGetComputeAllocationsByIdUsagesResponse,
  zGetComputeAllocationsResponse,
  zPostComputeAllocationMembershipsResponse,
  zPutComputeAllocationMembershipsByIdResponse,
} from "@/generated/core/zod.gen";
import { apiFetch } from "@/shared/api/client";
import {
  type AllocationDiff,
  type AllocationMembership,
  type AllocationUsage,
  type AttachedResource,
  type ChangeRequest,
  type ChangeRequestEvent,
  type ComputeAllocation,
  type ComputeAllocationList,
  type ComputeAllocationMembership,
  type CreateChangeRequestPayload,
  type CreateMembershipPayload,
  type UpdateChangeRequestPayload,
  type UpdateMembershipPayload,
  allocationMembershipSchema,
  changeRequestSchema,
  createChangeRequestPayloadSchema,
  createMembershipPayloadSchema,
  updateChangeRequestPayloadSchema,
  updateMembershipPayloadSchema,
} from "./schemas";
import type { AllocationListParams, ChangeRequestListParams } from "./types";

export async function listAllocations(
  params: AllocationListParams = {},
): Promise<ComputeAllocationList> {
  const search = new URLSearchParams();
  if (typeof params.limit === "number") search.set("limit", String(params.limit));
  if (typeof params.offset === "number") search.set("offset", String(params.offset));
  if (params.project_id) search.set("project_id", params.project_id);
  if (params.status) search.set("status", params.status);
  if (params.q) search.set("q", params.q);
  const qs = search.toString();
  const raw = await apiFetch(`/compute-allocations${qs ? `?${qs}` : ""}`);
  const list = zGetComputeAllocationsResponse.parse(raw);
  return { ...list, items: list.items ?? [] };
}

export async function getAllocation(id: string): Promise<ComputeAllocation> {
  const raw = await apiFetch(`/compute-allocations/${id}`);
  return zGetComputeAllocationsByIdResponse.parse(raw);
}

export async function listAllocationResources(id: string): Promise<AttachedResource[]> {
  const raw = await apiFetch(`/compute-allocations/${id}/resources`);
  return zGetComputeAllocationsByIdResourcesResponse.parse(raw ?? []);
}

export async function listAllocationDiffs(id: string): Promise<AllocationDiff[]> {
  const raw = await apiFetch(`/compute-allocations/${id}/diffs`);
  return zGetComputeAllocationsByIdDiffsResponse.parse(raw ?? []);
}

export async function listAllocationUsage(id: string): Promise<AllocationUsage[]> {
  const raw = await apiFetch(`/compute-allocations/${id}/usages`);
  return zGetComputeAllocationsByIdUsagesResponse.parse(raw ?? []);
}

export async function listAllocationMembers(id: string): Promise<AllocationMembership[]> {
  const raw = await apiFetch(`/compute-allocations/${id}/memberships`);
  return allocationMembershipSchema.array().parse(raw ?? []);
}

export async function addMember(
  payload: CreateMembershipPayload,
): Promise<ComputeAllocationMembership> {
  const validated = createMembershipPayloadSchema.parse(payload);
  const raw = await apiFetch("/compute-allocation-memberships", {
    method: "POST",
    body: validated,
  });
  return zPostComputeAllocationMembershipsResponse.parse(raw);
}

export async function updateMember(
  id: string,
  patch: UpdateMembershipPayload,
): Promise<ComputeAllocationMembership> {
  const validated = updateMembershipPayloadSchema.parse(patch);
  const raw = await apiFetch(`/compute-allocation-memberships/${id}`, {
    method: "PUT",
    body: validated,
  });
  return zPutComputeAllocationMembershipsByIdResponse.parse(raw);
}

export async function removeMember(id: string): Promise<void> {
  await apiFetch(`/compute-allocation-memberships/${id}`, { method: "DELETE" });
}

export async function listChangeRequests(
  params: ChangeRequestListParams = {},
): Promise<ChangeRequest[]> {
  if (params.allocation_id) {
    const raw = await apiFetch(`/compute-allocations/${params.allocation_id}/change-requests`);
    return changeRequestSchema.array().parse(raw ?? []);
  }
  if (params.requester_id) {
    const raw = await apiFetch(`/users/${params.requester_id}/change-requests`);
    return changeRequestSchema.array().parse(raw ?? []);
  }
  const search = new URLSearchParams();
  if (params.status) search.set("status", params.status);
  const qs = search.toString();
  const raw = await apiFetch(`/compute-allocation-change-requests${qs ? `?${qs}` : ""}`);
  return changeRequestSchema.array().parse(raw ?? []);
}

export async function getChangeRequest(id: string): Promise<ChangeRequest> {
  const raw = await apiFetch(`/compute-allocation-change-requests/${id}`);
  return changeRequestSchema.parse(raw);
}

export async function submitChangeRequest(
  payload: CreateChangeRequestPayload,
): Promise<ChangeRequest> {
  const validated = createChangeRequestPayloadSchema.parse(payload);
  const raw = await apiFetch("/compute-allocation-change-requests", {
    method: "POST",
    body: validated,
  });
  return changeRequestSchema.parse(raw);
}

async function updateChangeRequest(
  id: string,
  patch: UpdateChangeRequestPayload,
): Promise<ChangeRequest> {
  const validated = updateChangeRequestPayloadSchema.parse(patch);
  const raw = await apiFetch(`/compute-allocation-change-requests/${id}`, {
    method: "PUT",
    body: validated,
  });
  return changeRequestSchema.parse(raw);
}

export async function approveChangeRequest(
  id: string,
  approverId: string,
  comment?: string,
): Promise<ChangeRequest> {
  return updateChangeRequest(id, {
    change_status: "APPROVED",
    approver_id: approverId,
    ...(comment ? { reason: comment } : {}),
  });
}

export async function rejectChangeRequest(
  id: string,
  approverId: string,
  comment?: string,
): Promise<ChangeRequest> {
  return updateChangeRequest(id, {
    change_status: "REJECTED",
    approver_id: approverId,
    ...(comment ? { reason: comment } : {}),
  });
}

export async function listChangeRequestEvents(id: string): Promise<ChangeRequestEvent[]> {
  const raw = await apiFetch(`/compute-allocation-change-requests/${id}/events`);
  return zGetComputeAllocationChangeRequestsByIdEventsResponse.parse(raw ?? []);
}

export { updateChangeRequest };
