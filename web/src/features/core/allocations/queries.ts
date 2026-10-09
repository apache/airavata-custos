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

import { analyticsKeys } from "@/features/core/analytics/queries";
import { projectKeys, useProject } from "@/features/core/projects/queries";
import { resourceKeys } from "@/features/core/resources/queries";
import { userKeys } from "@/features/core/users/queries";
import {
  deleteComputeAllocationChangeRequestEventsById,
  deleteComputeAllocationChangeRequestsById,
  deleteComputeAllocationDiffsById,
  deleteComputeAllocationMembershipResourceOverridesById,
  deleteComputeAllocationMembershipsById,
  deleteComputeAllocationsByIdResourcesByResourceId,
  deleteComputeAllocationUsagesById,
  getComputeAllocationChangeRequestEventsById,
  getComputeAllocationChangeRequestsByIdEventsLatest,
  getComputeAllocationDiffsById,
  getComputeAllocationMembershipsById,
  getComputeAllocationMembershipsByIdResourceOverrides,
  getComputeAllocationsByIdUsagesTotal,
  getComputeAllocationsByIdUsersByUserIdUsagesTotal,
  getComputeAllocationUsagesById,
  postComputeAllocationChangeRequestEvents,
  postComputeAllocationDiffs,
  postComputeAllocationMembershipResourceOverrides,
  postComputeAllocations,
  postComputeAllocationsByIdResources,
  postComputeAllocationUsages,
  putComputeAllocationMembershipResourceOverridesById,
  putComputeAllocationMembershipsById,
  putComputeAllocationMembershipsByIdStatus,
  putComputeAllocationsByIdResourcesByResourceId,
  getComputeAllocationChangeRequestsById,
  getComputeAllocationChangeRequestsByIdEvents,
  getComputeAllocations,
  getComputeAllocationsById,
  getComputeAllocationChangeRequests,
  getComputeAllocationsByIdChangeRequests,
  getComputeAllocationsByIdDiffs,
  getComputeAllocationsByIdDiffsLatest,
  getComputeAllocationsByIdMemberships,
  getComputeAllocationsByIdResources,
  getComputeAllocationsByIdResourcesByResourceId,
  getComputeAllocationsByIdUsages,
  postComputeAllocationChangeRequests,
  postComputeAllocationMemberships,
  putComputeAllocationChangeRequestsById,
} from "@/generated/core/sdk.gen";
import type {
  GetComputeAllocationChangeRequestsData,
  ComputeAllocation,
  GetComputeAllocationsData,
} from "@/generated/core/types.gen";
import { ApiError } from "@/shared/api/client";
import { useInvalidating } from "@/shared/api/useInvalidating";
import { skipToken, useQuery } from "@tanstack/react-query";

export const allocationKeys = {
  all: ["allocations"] as const,
  list: (query: GetComputeAllocationsData["query"]) =>
    [...allocationKeys.all, "list", query] as const,
  detail: (id: string) => [...allocationKeys.all, "detail", id] as const,
  resources: (id: string) => [...allocationKeys.detail(id), "resources"] as const,
  // Outside resources(id), so a detach's list refresh does not refetch the detached mapping.
  mappings: (id: string) => [...allocationKeys.detail(id), "resource-mappings"] as const,
  usage: (id: string) => [...allocationKeys.detail(id), "usage"] as const,
  diffs: (id: string) => [...allocationKeys.detail(id), "diffs"] as const,
  latestDiff: (id: string) => [...allocationKeys.diffs(id), "latest"] as const,
  usageTotal: (id: string) => [...allocationKeys.usage(id), "total"] as const,
  userUsageTotal: (id: string, userId: string) =>
    [...allocationKeys.usageTotal(id), userId] as const,
  usageRecord: (id: string) => [...allocationKeys.all, "usage-record", id] as const,
  diff: (id: string) => [...allocationKeys.all, "diff", id] as const,
  members: (id: string) => [...allocationKeys.detail(id), "members"] as const,
  membership: (id: string) => [...allocationKeys.all, "membership", id] as const,
  overrides: (membershipId: string) =>
    [...allocationKeys.membership(membershipId), "overrides"] as const,
  changeRequests: (query: GetComputeAllocationChangeRequestsData["query"]) =>
    [...allocationKeys.all, "change-requests", "list", query] as const,
  allocationChangeRequests: (id: string) =>
    [...allocationKeys.all, "change-requests", "allocation", id] as const,
  changeRequestDetail: (id: string) =>
    [...allocationKeys.all, "change-requests", "detail", id] as const,
  changeRequestEvents: (id: string) =>
    [...allocationKeys.all, "change-requests", "events", id] as const,
  latestChangeRequestEvent: (id: string) =>
    [...allocationKeys.changeRequestEvents(id), "latest"] as const,
  changeRequestEvent: (id: string) =>
    [...allocationKeys.all, "change-requests", "event", id] as const,
};

export function useAllocations(query: GetComputeAllocationsData["query"]) {
  return useQuery({
    queryKey: allocationKeys.list(query),
    queryFn: () => getComputeAllocations({ query }),
  });
}

// Sanctioned cross-feature hook (ADR-0004) for projects to read its allocations.
export function useAllocationsByProject(projectId: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.list({ project_id: projectId }),
    // 200 is the backend's page-size ceiling.
    queryFn: projectId
      ? () => getComputeAllocations({ query: { project_id: projectId, limit: 200 } })
      : skipToken,
    // The backend ignores project_id for callers without allocations read.
    select: (list) => list.items.filter((a) => a.project_id === projectId),
  });
}

export function useAllocation(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.detail(id ?? ""),
    queryFn: id ? () => getComputeAllocationsById({ path: { id } }) : skipToken,
  });
}

// The backend rejects writes for allocations of projects owned by an external origination.
export function useCustosManaged(allocationId: string | undefined): boolean {
  const projectId = useAllocation(allocationId).data?.project_id;
  return useProject(projectId).data?.origination === "internal";
}

export function useAllocationResources(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.resources(id ?? ""),
    queryFn: id ? () => getComputeAllocationsByIdResources({ path: { id } }) : skipToken,
  });
}

export function useAllocationResourceMapping(id: string, resourceId: string | undefined) {
  return useQuery({
    queryKey: [...allocationKeys.mappings(id), resourceId ?? ""],
    queryFn: resourceId
      ? () => getComputeAllocationsByIdResourcesByResourceId({ path: { id, resourceId } })
      : skipToken,
  });
}

export function useAllocationUsage(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.usage(id ?? ""),
    queryFn: id ? () => getComputeAllocationsByIdUsages({ path: { id } }) : skipToken,
  });
}

export function useAllocationDiffs(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.diffs(id ?? ""),
    queryFn: id ? () => getComputeAllocationsByIdDiffs({ path: { id } }) : skipToken,
  });
}

// The latest diff carries the allocation's current SUs; before any diff (404) the initial amount stands.
export function useCurrentSuAmount(allocation: ComputeAllocation, enabled = true) {
  const id = enabled ? allocation.id : undefined;
  return useQuery({
    queryKey: allocationKeys.latestDiff(id ?? ""),
    queryFn: id ? () => orNull(getComputeAllocationsByIdDiffsLatest({ path: { id } })) : skipToken,
    select: (diff) => diff?.new_su_amount ?? allocation.initial_su_amount ?? 0,
  });
}

export function useAllocationMembers(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.members(id ?? ""),
    queryFn: id ? () => getComputeAllocationsByIdMemberships({ path: { id } }) : skipToken,
  });
}

// Membership rows also list under project members and the user drawer. A function because
// projects/queries imports this module.
export const membershipLists = () => [
  [...projectKeys.all, "members"],
  [...userKeys.all, "memberships"],
];

export function useAddMember(allocationId: string) {
  return useInvalidating(
    postComputeAllocationMemberships<true>,
    allocationKeys.members(allocationId),
    ...membershipLists(),
  );
}

export function useRemoveMember(allocationId: string) {
  return useInvalidating(
    deleteComputeAllocationMembershipsById<true>,
    allocationKeys.members(allocationId),
    ...membershipLists(),
  );
}

export function useChangeRequests(query: GetComputeAllocationChangeRequestsData["query"]) {
  return useQuery({
    queryKey: allocationKeys.changeRequests(query),
    queryFn: () => getComputeAllocationChangeRequests({ query }),
  });
}

export function useAllocationChangeRequests(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.allocationChangeRequests(id ?? ""),
    queryFn: id ? () => getComputeAllocationsByIdChangeRequests({ path: { id } }) : skipToken,
  });
}

export function useChangeRequest(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.changeRequestDetail(id ?? ""),
    queryFn: id ? () => getComputeAllocationChangeRequestsById({ path: { id } }) : skipToken,
  });
}

export function useChangeRequestEvents(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.changeRequestEvents(id ?? ""),
    queryFn: id ? () => getComputeAllocationChangeRequestsByIdEvents({ path: { id } }) : skipToken,
  });
}

// The user drawer lists each submitter's change requests and usage; usage also feeds the
// resource summaries' used SUs and the analytics views.
const userChangeRequests = [...userKeys.all, "change-requests"];
const usageLists = [[...userKeys.all, "usages"], resourceKeys.list(), analyticsKeys.all];

export function useSubmitChangeRequest() {
  return useInvalidating(
    postComputeAllocationChangeRequests<true>,
    [...allocationKeys.all, "change-requests"],
    userChangeRequests,
  );
}

// The backend PUT replaces every column, so a decision sends the full request. Approval
// writes a diff that sets the allocation's SU amount and status, so every allocation read refreshes.
export function useDecideChangeRequest() {
  return useInvalidating(
    putComputeAllocationChangeRequestsById<true>,
    allocationKeys.all,
    userChangeRequests,
  );
}

// A 404 on a "latest" read means nothing is recorded yet.
function orNull<T>(promise: Promise<T>): Promise<T | null> {
  return promise.catch((err) => {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  });
}

export function useCreateAllocation() {
  return useInvalidating(postComputeAllocations<true>, [...allocationKeys.all, "list"]);
}

// Attachments feed the resource summaries' allocation counts and totals.
export function useAttachResource(allocationId: string) {
  return useInvalidating(
    postComputeAllocationsByIdResources<true>,
    allocationKeys.resources(allocationId),
    resourceKeys.all,
  );
}

export function useUpdateResourceMapping(allocationId: string) {
  return useInvalidating(
    putComputeAllocationsByIdResourcesByResourceId<true>,
    allocationKeys.mappings(allocationId),
    resourceKeys.all,
  );
}

export function useDetachResource(allocationId: string) {
  return useInvalidating(
    deleteComputeAllocationsByIdResourcesByResourceId<true>,
    allocationKeys.resources(allocationId),
    resourceKeys.all,
  );
}

export function useMembership(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.membership(id ?? ""),
    queryFn: id ? () => getComputeAllocationMembershipsById({ path: { id } }) : skipToken,
  });
}

export function useUpdateMembership(allocationId: string) {
  return useInvalidating(
    putComputeAllocationMembershipsById<true>,
    allocationKeys.members(allocationId),
    [...allocationKeys.all, "membership"],
    ...membershipLists(),
  );
}

export function useSetMembershipStatus(allocationId: string) {
  return useInvalidating(
    putComputeAllocationMembershipsByIdStatus<true>,
    allocationKeys.members(allocationId),
    [...allocationKeys.all, "membership"],
    ...membershipLists(),
  );
}

export function useMembershipOverrides(membershipId: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.overrides(membershipId ?? ""),
    queryFn: membershipId
      ? () => getComputeAllocationMembershipsByIdResourceOverrides({ path: { id: membershipId } })
      : skipToken,
  });
}

// Overrides also list under their resource's detail.
export function useCreateOverride(membershipId: string) {
  return useInvalidating(
    postComputeAllocationMembershipResourceOverrides<true>,
    allocationKeys.overrides(membershipId),
    resourceKeys.all,
  );
}

export function useUpdateOverride(membershipId: string) {
  return useInvalidating(
    putComputeAllocationMembershipResourceOverridesById<true>,
    allocationKeys.overrides(membershipId),
    resourceKeys.all,
  );
}

export function useDeleteOverride(membershipId: string) {
  return useInvalidating(
    deleteComputeAllocationMembershipResourceOverridesById<true>,
    allocationKeys.overrides(membershipId),
    resourceKeys.all,
  );
}

export function useUsageTotal(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.usageTotal(id ?? ""),
    queryFn: id ? () => getComputeAllocationsByIdUsagesTotal({ path: { id } }) : skipToken,
    select: (total) => total.total_su_amount,
  });
}

export function useUserUsageTotal(id: string, userId: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.userUsageTotal(id, userId ?? ""),
    queryFn: userId
      ? () => getComputeAllocationsByIdUsersByUserIdUsagesTotal({ path: { id, userId } })
      : skipToken,
    select: (total) => total.total_su_amount,
  });
}

export function useUsageRecord(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.usageRecord(id ?? ""),
    queryFn: id ? () => getComputeAllocationUsagesById({ path: { id } }) : skipToken,
  });
}

export function useRecordUsage(allocationId: string) {
  return useInvalidating(
    postComputeAllocationUsages<true>,
    allocationKeys.usage(allocationId),
    ...usageLists,
  );
}

export function useDeleteUsage(allocationId: string) {
  return useInvalidating(
    (id: string) => deleteComputeAllocationUsagesById({ path: { id } }),
    allocationKeys.usage(allocationId),
    ...usageLists,
  );
}

export function useDiff(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.diff(id ?? ""),
    queryFn: id ? () => getComputeAllocationDiffsById({ path: { id } }) : skipToken,
  });
}

export function useRecordDiff() {
  return useInvalidating(postComputeAllocationDiffs<true>, allocationKeys.all);
}

export function useDeleteDiff(allocationId: string) {
  return useInvalidating(
    (id: string) => deleteComputeAllocationDiffsById({ path: { id } }),
    allocationKeys.detail(allocationId),
  );
}

// Only the lists refresh: refetching the deleted request's detail would 404 before the page leaves.
export function useDeleteChangeRequest() {
  return useInvalidating(
    (id: string) => deleteComputeAllocationChangeRequestsById({ path: { id } }),
    [...allocationKeys.all, "change-requests", "list"],
    [...allocationKeys.all, "change-requests", "allocation"],
    userChangeRequests,
  );
}

export function useLatestChangeRequestEvent(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.latestChangeRequestEvent(id ?? ""),
    queryFn: id
      ? () => orNull(getComputeAllocationChangeRequestsByIdEventsLatest({ path: { id } }))
      : skipToken,
  });
}

export function useChangeRequestEvent(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.changeRequestEvent(id ?? ""),
    queryFn: id ? () => getComputeAllocationChangeRequestEventsById({ path: { id } }) : skipToken,
  });
}

export function useCreateChangeRequestEvent(changeRequestId: string) {
  return useInvalidating(
    postComputeAllocationChangeRequestEvents<true>,
    allocationKeys.changeRequestEvents(changeRequestId),
  );
}

export function useDeleteChangeRequestEvent(changeRequestId: string) {
  return useInvalidating(
    deleteComputeAllocationChangeRequestEventsById<true>,
    allocationKeys.changeRequestEvents(changeRequestId),
  );
}
