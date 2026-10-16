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

import { AllocationStatusLabel } from "@/features/core/allocations/components/AllocationStatusLabel";
import { allocationKeys } from "@/features/core/allocations/queries";
import { useClusters } from "@/features/core/clusters/queries";
import {
  getComputeAllocationMembershipsById,
  getComputeAllocationsById,
  getComputeAllocationsByIdMemberships,
} from "@/generated/core/sdk.gen";
import type {
  ComputeAllocation,
  ComputeAllocationMembershipResourceOverride,
} from "@/generated/core/types.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { formatNumber } from "@/shared/format";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { Mono, RecordDrawer } from "@/shared/ui/RecordDrawer";
import { useQueries } from "@tanstack/react-query";
import Link from "next/link";
import { useResource, useResourceAllocations, useResourceOverrides } from "../queries";

const membersQuery = (id: string) => ({
  queryKey: allocationKeys.members(id),
  queryFn: () => getComputeAllocationsByIdMemberships({ path: { id } }),
});

// An override names only its membership. The members of the allocations using this resource name
// most; a membership in an allocation the resource was detached from is read by id. settled: every
// read succeeded, so a missing name is unknown rather than still loading or failed.
function useOverrideNames(
  listed: ComputeAllocation[] | undefined,
  overrides: ComputeAllocationMembershipResourceOverride[] = [],
) {
  const members = useQueries({
    queries: (listed ?? []).map((a) => membersQuery(a.id)),
    combine: (results) => ({
      rows: results.flatMap((r) => r.data ?? []),
      done: results.every((r) => !r.isPending),
      ok: results.every((r) => r.isSuccess),
    }),
  });
  const unlisted =
    listed && members.done
      ? overrides
          .map((o) => o.compute_allocation_membership_id)
          .filter((id) => !members.rows.some((m) => m.id === id))
      : [];
  const memberships = useQueries({
    queries: unlisted.map((id) => ({
      queryKey: allocationKeys.membership(id),
      queryFn: () => getComputeAllocationMembershipsById({ path: { id } }),
    })),
    combine: (results) => ({
      ids: [...new Set(results.flatMap((r) => (r.data ? [r.data.compute_allocation_id] : [])))],
      ok: results.every((r) => r.isSuccess),
    }),
  });
  const detachedMembers = useQueries({
    queries: memberships.ids.map(membersQuery),
    combine: (results) => ({
      rows: results.flatMap((r) => r.data ?? []),
      ok: results.every((r) => r.isSuccess),
    }),
  });
  const detached = useQueries({
    queries: memberships.ids.map((id) => ({
      queryKey: allocationKeys.detail(id),
      queryFn: () => getComputeAllocationsById({ path: { id } }),
    })),
    combine: (results) => ({
      rows: results.flatMap((r) => (r.data ? [r.data] : [])),
      ok: results.every((r) => r.isSuccess),
    }),
  });
  const allocationName = new Map([...(listed ?? []), ...detached.rows].map((a) => [a.id, a.name]));
  const names = new Map(
    [...members.rows, ...detachedMembers.rows].map((m) => [
      m.id,
      { member: m.display_name, allocation: allocationName.get(m.compute_allocation_id) },
    ]),
  );
  const settled =
    Boolean(listed) && members.ok && memberships.ok && detachedMembers.ok && detached.ok;
  return { names, settled };
}

export type ResourceDetailDrawerProps = {
  resourceId: string | null;
  onOpenChange: (open: boolean) => void;
};

export function ResourceDetailDrawer({ resourceId, onOpenChange }: ResourceDetailDrawerProps) {
  const id = resourceId ?? undefined;
  const resource = useResource(id);
  const allocations = useResourceAllocations(id);
  const overrides = useResourceOverrides(id);
  const { names, settled } = useOverrideNames(allocations.data, overrides.data);
  const canReadClusters = useAbility().can("read", "Cluster");
  const clusters = useClusters(canReadClusters);

  return (
    <RecordDrawer
      title={resource.data?.name ?? "Resource"}
      id={id}
      onClose={() => onOpenChange(false)}
      query={resource}
      width="lg"
      fields={(r) => ({
        "Resource ID": <Mono>{r.id}</Mono>,
        Type: r.resource_type,
        Amount: <span className="tabular-nums">{formatNumber(r.resource_amount)}</span>,
        ...(canReadClusters && {
          Cluster:
            clusters.data?.find((c) => c.id === r.compute_cluster_id)?.name ??
            (clusters.isSuccess ? "Unknown cluster" : "—"),
        }),
      })}
    >
      {() => (
        <>
          <section className="space-y-2">
            <h3 className="text-sm font-semibold text-foreground">Allocations</h3>
            {allocations.isLoading ? (
              <CardSkeleton />
            ) : allocations.error ? (
              <ErrorState error={allocations.error} onRetry={() => allocations.refetch()} />
            ) : !allocations.data?.length ? (
              <p className="text-sm text-muted-foreground">No allocations use this resource.</p>
            ) : (
              <ul className="space-y-1">
                {allocations.data.map((a) => (
                  <li
                    key={a.id}
                    className="flex items-center justify-between gap-3 rounded-md border bg-card px-3 py-2 text-sm"
                  >
                    <Link
                      href={`/allocations/${a.id}`}
                      className="min-w-0 text-link underline-offset-4 [overflow-wrap:anywhere] hover:underline"
                    >
                      {a.name}
                    </Link>
                    <AllocationStatusLabel
                      status={a.status}
                      startTime={a.start_time}
                      endTime={a.end_time}
                    />
                  </li>
                ))}
              </ul>
            )}
          </section>

          <section className="space-y-2">
            <h3 className="text-sm font-semibold text-foreground">Membership overrides</h3>
            {overrides.isLoading ? (
              <CardSkeleton />
            ) : overrides.error ? (
              <ErrorState error={overrides.error} onRetry={() => overrides.refetch()} />
            ) : !overrides.data?.length ? (
              <p className="text-sm text-muted-foreground">
                No membership overrides on this resource.
              </p>
            ) : (
              <table className="w-full text-left text-sm">
                <caption className="sr-only">Membership overrides</caption>
                <thead className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                  <tr>
                    <th className="py-2 pr-4 font-medium">Member</th>
                    <th className="py-2 pr-4 font-medium">Allocation</th>
                    <th className="py-2 pr-4 text-right font-medium">Amount</th>
                    <th className="py-2 text-right font-medium">Time (min)</th>
                  </tr>
                </thead>
                <tbody>
                  {overrides.data.map((o) => (
                    <tr key={o.id} className="border-t border-border/60">
                      <td className="py-2 pr-4 [overflow-wrap:anywhere]">
                        {names.get(o.compute_allocation_membership_id)?.member ??
                          (settled ? "Unknown user" : "—")}
                      </td>
                      <td className="py-2 pr-4 [overflow-wrap:anywhere]">
                        {names.get(o.compute_allocation_membership_id)?.allocation ?? "—"}
                      </td>
                      <td className="py-2 pr-4 text-right tabular-nums">
                        {formatNumber(o.override_resource_amount)}
                      </td>
                      <td className="py-2 text-right tabular-nums">
                        {formatNumber(o.override_resource_time)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </section>
        </>
      )}
    </RecordDrawer>
  );
}
