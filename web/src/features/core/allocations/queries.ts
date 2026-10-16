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
import { projectKeys } from "@/features/core/projects/queries";
import { resourceKeys } from "@/features/core/resources/queries";
import { userKeys } from "@/features/core/users/queries";
import {
  deleteComputeAllocationChangeRequestEventsById,
  deleteComputeAllocationChangeRequestsById,
  deleteComputeAllocationDiffsById,
  deleteComputeAllocationMembershipResourceOverridesById,
  deleteComputeAllocationMembershipsById,
  deleteComputeAllocationUsagesById,
  deleteComputeAllocationsByIdResourcesByResourceId,
  getComputeAllocationChangeRequestEventsById,
  getComputeAllocationChangeRequestsById,
  getComputeAllocationChangeRequestsByIdEvents,
  getComputeAllocationChangeRequestsByIdEventsLatest,
  getComputeAllocationDiffsById,
  getComputeAllocationMembershipsById,
  getComputeAllocationMembershipsByIdResourceOverrides,
  getComputeAllocationUsagesById,
  getComputeAllocations,
  getComputeAllocationsById,
  getComputeAllocationsByIdChangeRequests,
  getComputeAllocationsByIdDiffs,
  getComputeAllocationsByIdMemberships,
  getComputeAllocationsByIdResources,
  getComputeAllocationsByIdUsages,
  getComputeAllocationsByIdUsagesTotal,
  getComputeAllocationsByIdUsersByUserIdUsagesTotal,
  postComputeAllocationChangeRequestEvents,
  postComputeAllocationChangeRequests,
  postComputeAllocationDiffs,
  postComputeAllocationMembershipResourceOverrides,
  postComputeAllocationMemberships,
  postComputeAllocationUsages,
  postComputeAllocations,
  postComputeAllocationsByIdResources,
  putComputeAllocationChangeRequestsById,
  putComputeAllocationMembershipResourceOverridesById,
  putComputeAllocationMembershipsById,
  putComputeAllocationMembershipsByIdStatus,
} from "@/generated/core/sdk.gen";
import type {
  ComputeAllocationChangeRequest,
  GetComputeAllocationsData,
} from "@/generated/core/types.gen";
import { ApiError, type KnownError, apiFetch, errorCopy } from "@/shared/api/client";
import { listOf } from "@/shared/api/listOf";
import { paged } from "@/shared/api/paged";
import { useInvalidatingWith } from "@/shared/api/useInvalidating";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { enumLabel } from "@/shared/format";
import {
  type MutationMeta,
  type QueryKey,
  skipToken,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import * as React from "react";

export const allocationKeys = {
  all: ["allocations"] as const,
  list: (query: GetComputeAllocationsData["query"]) =>
    [...allocationKeys.all, "list", query] as const,
  detail: (id: string) => [...allocationKeys.all, "detail", id] as const,
  resources: (id: string) => [...allocationKeys.detail(id), "resources"] as const,
  usage: (id: string) => [...allocationKeys.detail(id), "usage"] as const,
  diffs: (id: string) => [...allocationKeys.detail(id), "diffs"] as const,
  usageTotal: (id: string) => [...allocationKeys.usage(id), "total"] as const,
  userUsageTotal: (id: string, userId: string) =>
    [...allocationKeys.usageTotal(id), userId] as const,
  usageRecord: (id: string) => [...allocationKeys.all, "usage-record", id] as const,
  diff: (id: string) => [...allocationKeys.all, "diff", id] as const,
  members: (id: string) => [...allocationKeys.detail(id), "members"] as const,
  membership: (id: string) => [...allocationKeys.all, "membership", id] as const,
  overrides: (membershipId: string) =>
    [...allocationKeys.membership(membershipId), "overrides"] as const,
  changeRequests: (status: string | undefined) =>
    [...allocationKeys.all, "change-requests", "list", status] as const,
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

// Without the read privilege the server lists the caller's allocations whole and ignores the
// query, so the filter and page apply here.
function useScopedList() {
  return !useAbility().can("read", "Allocation");
}

export function useAllocations(
  query: NonNullable<GetComputeAllocationsData["query"]>,
  enabled = true,
) {
  const scoped = useScopedList();
  return useQuery({
    queryKey: allocationKeys.list(scoped ? {} : query),
    queryFn: () => getComputeAllocations({ query: scoped ? undefined : query }),
    enabled,
    ...paged,
    select: ({ items, total }) => {
      if (!scoped) return { items, total };
      const q = query.q?.toLowerCase();
      const rows = listOf(items).filter(
        (a) =>
          (!query.status || a.status === query.status) && (!q || a.name.toLowerCase().includes(q)),
      );
      const offset = query.offset ?? 0;
      return {
        items: rows.slice(offset, offset + (query.limit ?? rows.length)),
        total: rows.length,
      };
    },
  });
}

// Sanctioned cross-feature hook (ADR-0004) for projects to read its allocations.
export function useAllocationsByProject(projectId: string | undefined) {
  const scoped = useScopedList();
  return useQuery({
    queryKey: allocationKeys.list(scoped ? {} : { project_id: projectId }),
    // 200 is the backend's page-size ceiling.
    queryFn: projectId
      ? () =>
          getComputeAllocations({
            query: scoped ? undefined : { project_id: projectId, limit: 200 },
          })
      : skipToken,
    select: (list) => (list.items ?? []).filter((a) => a.project_id === projectId),
  });
}

export function useAllocation(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.detail(id ?? ""),
    queryFn: id ? () => getComputeAllocationsById({ path: { id } }) : skipToken,
  });
}

export function useAllocationResources(id: string) {
  return useQuery({
    queryKey: allocationKeys.resources(id),
    queryFn: () => getComputeAllocationsByIdResources({ path: { id } }),
  });
}

export function useAllocationUsage(id: string) {
  return useQuery({
    queryKey: allocationKeys.usage(id),
    queryFn: () => getComputeAllocationsByIdUsages({ path: { id } }),
  });
}

export function useAllocationDiffs(id: string) {
  return useQuery({
    queryKey: allocationKeys.diffs(id),
    queryFn: () => getComputeAllocationsByIdDiffs({ path: { id } }),
  });
}

export function useAllocationMembers(id: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.members(id ?? ""),
    queryFn: id ? () => getComputeAllocationsByIdMemberships({ path: { id } }) : skipToken,
  });
}

// The causes master returns for allocation writes, matched on its error text.
const allocationGone: KnownError = {
  status: 400,
  when: /compute allocation "[^"]*" not found/,
  copy: "This allocation no longer exists.",
};
const resourceGone: KnownError = {
  status: 400,
  when: /compute allocation resource "[^"]*" not found/,
  copy: "That resource no longer exists.",
};
const changeRequestGone: KnownError[] = [
  { status: 404, copy: "This change request no longer exists." },
  {
    status: 400,
    when: /change request "[^"]*" not found/,
    copy: "This change request no longer exists.",
  },
];
const membershipGone: KnownError[] = [
  { status: 404, copy: "This membership no longer exists." },
  { status: 400, when: /membership "[^"]*" not found/, copy: "This membership no longer exists." },
];
const membershipRead = {
  on: membershipGone,
  key: ({ path }: { path: { id: string } }) => allocationKeys.membership(path.id),
};
const errors = (...known: Array<KnownError | KnownError[]>) => ({ errors: known.flat() });
const notFound = (copy: string): KnownError => ({ status: 404, copy });

const isRefusal = (error: unknown, known: KnownError[]) =>
  known.some((k) => errorCopy(error, [k]) === k.copy);

// A write refused because its record is gone resets that record's read, so its gone state shows.
function useWrite<O, R>(
  meta: MutationMeta,
  mutationFn: (options: O) => Promise<R>,
  gone: { on: KnownError | KnownError[]; key: (options: O) => QueryKey },
  ...keys: QueryKey[]
) {
  const client = useQueryClient();
  const on = [gone.on].flat();
  return useMutation({
    mutationFn,
    meta,
    onError: (error, options) => {
      if (isRefusal(error, on)) void client.resetQueries({ queryKey: gone.key(options) });
    },
    onSettled: () => Promise.all(keys.map((queryKey) => client.invalidateQueries({ queryKey }))),
  });
}
const allocationRead = (id: string) => ({
  on: allocationGone,
  key: () => allocationKeys.detail(id),
});

// Membership rows also list under project members and the user drawer. A function because
// projects/queries imports this module.
export const membershipLists = () => [
  [...projectKeys.all, "members"],
  [...userKeys.all, "memberships"],
];

export function useAddMember(allocationId: string) {
  return useWrite(
    errors(
      allocationGone,
      { status: 400, when: /user "[^"]*" not found/, copy: "No user has this ID." },
      { status: 409, copy: "This user is already a member of this allocation." },
    ),
    postComputeAllocationMemberships<true>,
    allocationRead(allocationId),
    allocationKeys.members(allocationId),
    ...membershipLists(),
  );
}

export function useRemoveMember(allocationId: string) {
  return useWrite(
    errors(membershipGone),
    deleteComputeAllocationMembershipsById<true>,
    membershipRead,
    allocationKeys.members(allocationId),
    ...membershipLists(),
  );
}

// The newest first, at most 200 (the server's ceiling). The endpoint is served but absent from the
// spec, so it has no generated client.
export const CHANGE_REQUEST_LIMIT = 200;

export function useChangeRequests(status: string | undefined) {
  return useQuery({
    queryKey: allocationKeys.changeRequests(status),
    queryFn: () =>
      apiFetch<ComputeAllocationChangeRequest[]>(
        `/compute-allocation-change-requests?${new URLSearchParams({
          ...(status ? { status } : {}),
          limit: String(CHANGE_REQUEST_LIMIT),
        })}`,
      ),
    ...paged,
  });
}

function useRefresh(...queryKeys: QueryKey[]) {
  const client = useQueryClient();
  return () => {
    for (const queryKey of queryKeys) void client.invalidateQueries({ queryKey });
  };
}

// Refetch the list behind a record that turned out gone, so its stale row leaves.
export const useRefreshAllocationLists = () => useRefresh([...allocationKeys.all, "list"]);
export const useRefreshAllocationResources = (id: string) =>
  useRefresh(allocationKeys.resources(id));
export const useRefreshAllocationMembers = (id: string) => useRefresh(allocationKeys.members(id));
// Never the detail prefix: a gone detail that refreshed itself would refetch its 404 forever.
export const useRefreshChangeRequests = () =>
  useRefresh(
    [...allocationKeys.all, "change-requests", "list"],
    [...allocationKeys.all, "change-requests", "allocation"],
  );
// A request found decided under an open edit: its timeline and every list showing it.
export const useRefreshDecidedChangeRequest = (id: string) =>
  useRefresh(
    allocationKeys.changeRequestEvents(id),
    [...allocationKeys.all, "change-requests", "list"],
    [...allocationKeys.all, "change-requests", "allocation"],
    [...userKeys.all, "change-requests"],
  );

export function useAllocationChangeRequests(id: string) {
  return useQuery({
    queryKey: allocationKeys.allocationChangeRequests(id),
    queryFn: () => getComputeAllocationsByIdChangeRequests({ path: { id } }),
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
  return useWrite(
    errors(allocationGone),
    postComputeAllocationChangeRequests<true>,
    {
      on: allocationGone,
      key: ({ body }) => allocationKeys.detail(body.compute_allocation_id ?? ""),
    },
    [...allocationKeys.all, "change-requests"],
    userChangeRequests,
  );
}

export function useUpdateChangeRequest() {
  return useWrite(
    errors(changeRequestGone, allocationGone),
    putComputeAllocationChangeRequestsById<true>,
    { on: changeRequestGone, key: ({ path }) => allocationKeys.changeRequestDetail(path.id) },
    [...allocationKeys.all, "change-requests"],
    userChangeRequests,
  );
}

// The refusal of a write against a request someone else has decided since it was shown.
export const alreadyDecided = (status: string) =>
  new Error(`This request was already ${enumLabel(status).toLowerCase()}.`);

// A decision sends the request as the server now holds it, with the new status and the approver,
// and is refused when the request is no longer pending. Approval writes a diff, so every allocation
// read refreshes.
export function useDecideChangeRequest() {
  return useWrite(
    errors(changeRequestGone, allocationGone),

    async ({
      id,
      ...decision
    }: Required<Pick<ComputeAllocationChangeRequest, "id" | "change_status" | "approver_id">>) => {
      const current = await getComputeAllocationChangeRequestsById({ path: { id } });
      if (current.change_status !== "PENDING") throw alreadyDecided(current.change_status);
      return putComputeAllocationChangeRequestsById({
        path: { id },
        body: { ...current, ...decision },
      });
    },
    { on: changeRequestGone, key: ({ id }) => allocationKeys.changeRequestDetail(id) },
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

// A typed ID that matches no record reads as such; a picked one was deleted since the list loaded.
export function useCreateAllocation(typed: { project: boolean; cluster: boolean }) {
  return useInvalidatingWith(
    errors(
      {
        status: 400,
        when: /project "[^"]*" not found/,
        copy: typed.project ? "No project has that ID." : "That project no longer exists.",
      },
      {
        status: 400,
        when: /compute cluster "[^"]*" not found/,
        copy: typed.cluster ? "No cluster has that ID." : "That cluster no longer exists.",
      },
    ),
    postComputeAllocations<true>,
    [...allocationKeys.all, "list"],
  );
}

// Attachments feed the resource summaries' allocation counts and totals.
export function useAttachResource(allocationId: string) {
  return useWrite(
    errors(allocationGone, resourceGone, {
      status: 409,
      copy: "That resource is already attached to this allocation.",
    }),
    postComputeAllocationsByIdResources<true>,
    allocationRead(allocationId),
    allocationKeys.resources(allocationId),
    resourceKeys.all,
  );
}

export function useDetachResource(allocationId: string) {
  return useInvalidatingWith(
    errors({ status: 404, copy: "That resource is no longer attached to this allocation." }),

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
  return useWrite(
    errors(membershipGone),
    putComputeAllocationMembershipsById<true>,
    membershipRead,
    allocationKeys.members(allocationId),
    [...allocationKeys.all, "membership"],
    ...membershipLists(),
  );
}

export function useSetMembershipStatus(allocationId: string) {
  return useWrite(
    errors(membershipGone),
    putComputeAllocationMembershipsByIdStatus<true>,
    membershipRead,
    allocationKeys.members(allocationId),
    [...allocationKeys.all, "membership"],
    ...membershipLists(),
  );
}

export function useMembershipOverrides(membershipId: string) {
  return useQuery({
    queryKey: allocationKeys.overrides(membershipId),
    queryFn: () =>
      getComputeAllocationMembershipsByIdResourceOverrides({ path: { id: membershipId } }),
  });
}

// Overrides also list under their resource's detail; the allocation's resources refresh so a
// resource another tab detached leaves the picker.
export function useCreateOverride(allocationId: string, membershipId: string) {
  return useWrite(
    errors(membershipGone, resourceGone, {
      status: 409,
      copy: "This member already has an override for that resource.",
    }),
    postComputeAllocationMembershipResourceOverrides<true>,
    { on: membershipGone, key: () => allocationKeys.membership(membershipId) },
    allocationKeys.overrides(membershipId),
    allocationKeys.resources(allocationId),
    resourceKeys.all,
  );
}

export function useUpdateOverride(allocationId: string, membershipId: string) {
  return useInvalidatingWith(
    errors({ status: 404, copy: "This override no longer exists." }),

    putComputeAllocationMembershipResourceOverridesById<true>,
    allocationKeys.overrides(membershipId),
    allocationKeys.resources(allocationId),
    resourceKeys.all,
  );
}

export function useDeleteOverride(membershipId: string) {
  return useInvalidatingWith(
    errors({ status: 404, copy: "This override no longer exists." }),

    deleteComputeAllocationMembershipResourceOverridesById<true>,
    allocationKeys.overrides(membershipId),
    resourceKeys.all,
  );
}

export function useUsageTotal(id: string) {
  return useQuery({
    queryKey: allocationKeys.usageTotal(id),
    queryFn: () => getComputeAllocationsByIdUsagesTotal({ path: { id } }),
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
  return useWrite(
    errors(allocationGone, resourceGone),
    postComputeAllocationUsages<true>,
    allocationRead(allocationId),
    allocationKeys.usage(allocationId),
    ...usageLists,
  );
}

const usageGone = notFound("This usage record no longer exists.");
export function useDeleteUsage(allocationId: string) {
  return useWrite(
    errors(usageGone),
    (id: string) => deleteComputeAllocationUsagesById({ path: { id } }),
    { on: usageGone, key: allocationKeys.usageRecord },
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
  return useWrite(
    errors(allocationGone),
    postComputeAllocationDiffs<true>,
    {
      on: allocationGone,
      key: ({ body }) => allocationKeys.detail(body.compute_allocation_id ?? ""),
    },
    allocationKeys.all,
  );
}

const diffGone = notFound("This history entry no longer exists.");
export function useDeleteDiff(allocationId: string) {
  return useWrite(
    errors(diffGone),
    (id: string) => deleteComputeAllocationDiffsById({ path: { id } }),
    { on: diffGone, key: allocationKeys.diff },
    allocationKeys.detail(allocationId),
  );
}

export function useDeleteChangeRequest() {
  const client = useQueryClient();
  const deleted = React.useRef<string>(undefined);
  // Back to the deleted request must refetch (404), not render the cached PENDING copy. The cache
  // drops it once its page unmounts: while the page still observes it, removal refetches a 404.
  React.useEffect(
    () => () => {
      const id = deleted.current;
      if (!id) return;
      client.removeQueries({ queryKey: allocationKeys.changeRequestDetail(id) });
      client.removeQueries({ queryKey: allocationKeys.changeRequestEvents(id) });
    },
    [client],
  );
  return useMutation({
    meta: errors(changeRequestGone),
    mutationFn: (id: string) => deleteComputeAllocationChangeRequestsById({ path: { id } }),
    onError: (error, id) => {
      if (isRefusal(error, changeRequestGone))
        void client.resetQueries({ queryKey: allocationKeys.changeRequestDetail(id) });
    },
    onSettled: () =>
      Promise.all(
        [
          [...allocationKeys.all, "change-requests", "list"],
          [...allocationKeys.all, "change-requests", "allocation"],
          userChangeRequests,
        ].map((queryKey) => client.invalidateQueries({ queryKey })),
      ),
    onSuccess: (_data, id) => {
      deleted.current = id;
    },
  });
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

// The request rereads too, so one deleted elsewhere turns its page gone instead of keeping controls.
export function useCreateChangeRequestEvent(changeRequestId: string) {
  return useWrite(
    errors(changeRequestGone),
    postComputeAllocationChangeRequestEvents<true>,
    { on: changeRequestGone, key: () => allocationKeys.changeRequestDetail(changeRequestId) },
    allocationKeys.changeRequestEvents(changeRequestId),
    allocationKeys.changeRequestDetail(changeRequestId),
  );
}

const eventGone = notFound("This event no longer exists.");
export function useDeleteChangeRequestEvent(changeRequestId: string) {
  return useWrite(
    errors(eventGone),
    deleteComputeAllocationChangeRequestEventsById<true>,
    { on: eventGone, key: ({ path }) => allocationKeys.changeRequestEvent(path.id) },
    allocationKeys.changeRequestEvents(changeRequestId),
  );
}
