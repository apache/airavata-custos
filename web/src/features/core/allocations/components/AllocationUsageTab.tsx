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

import { useUsageSummary } from "@/features/core/analytics/queries";
import { useResources } from "@/features/core/resources/queries";
import type { AllocationMembershipResponse, ComputeAllocation } from "@/generated/core/types.gen";
import { formatSU } from "@/shared/format";
import { NameCell } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { UsageBar } from "@/shared/ui/UsageBar";
import {
  useAllocationMembers,
  useAllocationResources,
  useAllocationUsage,
  useUsageTotal,
  useUserUsageTotal,
} from "../queries";
import { MemberName, memberName } from "./MemberName";
import { UsageRecords } from "./UsageRecords";

export type AllocationUsageTabProps = {
  allocation: ComputeAllocation;
  // Diffs, per-member totals and record detail require the allocations read privilege.
  canRead: boolean;
  canManage: boolean;
};

function MemberUsage({
  allocationId,
  member,
}: {
  allocationId: string;
  member: AllocationMembershipResponse;
}) {
  const total = useUserUsageTotal(allocationId, member.user_id).data;
  return (
    <li className="flex items-baseline justify-between rounded-md border bg-card px-3 py-2 text-sm">
      <NameCell text={memberName(member)} className="font-medium" />
      <span className="tabular-nums">{total === undefined ? "—" : `${formatSU(total)} SU`}</span>
    </li>
  );
}

export function AllocationUsageTab({ allocation, canRead, canManage }: AllocationUsageTabProps) {
  const query = useAllocationUsage(allocation.id);
  const initialSu = allocation.initial_su_amount;
  const usedTotal = useUsageTotal(allocation.id).data;
  const members = useAllocationMembers(allocation.id).data;
  const attachedQuery = useAllocationResources(allocation.id);
  const attached = attachedQuery.data;
  // A detached resource's past usage is named from the catalog, which readers alone may list, else
  // from the analytics summary, which names every resource the allocation has used.
  const catalog = useResources({ enabled: canRead });
  const detached =
    !canRead &&
    attached !== undefined &&
    (query.data ?? []).some(
      (r) => !attached.some((a) => a.id === r.compute_allocation_resource_id),
    );
  const summary = useUsageSummary(detached ? allocation.id : undefined);
  if (query.isLoading) return <TableSkeleton rows={3} columns={3} />;
  if (query.error) {
    return <ErrorState onRetry={() => query.refetch()} />;
  }
  const rows = query.data ?? [];

  const looked =
    attachedQuery.isSuccess && (canRead ? catalog.isSuccess : !detached || summary.isSuccess);
  const resourceName = (id: string) =>
    [...(attached ?? []), ...(catalog.data ?? [])].find((r) => r.id === id)?.name ??
    summary.data?.by_resource?.find((r) => r.resource_id === id)?.name ??
    (looked ? "Unknown resource" : "—");
  const userName = (id: string) => {
    const member = members?.find((m) => m.user_id === id);
    return member ? memberName(member) : <MemberName id={id} allocationId={allocation.id} />;
  };
  // Per-resource SU used, summed from usage rows. No per-resource SU allocation
  // exists, so bars show each resource's share of total usage, not a quota.
  const byResource = new Map<string, number>();
  for (const row of rows) {
    const key = row.compute_allocation_resource_id;
    byResource.set(key, (byResource.get(key) ?? 0) + row.used_su_amount);
  }
  const memberIds = new Set(members?.map((m) => m.user_id));
  const otherUsed = rows
    .filter((r) => !memberIds.has(r.user_id))
    .reduce((sum, r) => sum + r.used_su_amount, 0);

  return (
    <div className="space-y-4">
      <UsageBar
        value={usedTotal ?? 0}
        max={initialSu}
        label={`${usedTotal === undefined ? "—" : formatSU(usedTotal)} / ${formatSU(initialSu)} initial SUs`}
        ariaLabel="Allocation SU usage"
      />
      {rows.length === 0 ? (
        <EmptyState
          heading="No usage recorded"
          description="Jobs that consume this allocation will appear here."
        />
      ) : (
        <>
          <ul className="space-y-3">
            {[...byResource.entries()].map(([resourceId, used]) => {
              const name = resourceName(resourceId);
              return (
                <li key={resourceId} className="rounded-md border bg-card px-3 py-2">
                  <div className="mb-1 flex items-baseline justify-between text-sm">
                    <NameCell text={name} className="font-medium" />
                    <span className="tabular-nums">{formatSU(used)} SU</span>
                  </div>
                  <UsageBar
                    value={used}
                    max={usedTotal ?? 0}
                    ariaLabel={`${name} share of usage`}
                    size="sm"
                  />
                </li>
              );
            })}
          </ul>
          {canRead && members && (members.length > 0 || otherUsed > 0) ? (
            <section className="space-y-2">
              <h2 className="text-sm font-semibold text-foreground">By member</h2>
              <ul className="space-y-1">
                {members.map((m) => (
                  <MemberUsage key={m.id} allocationId={allocation.id} member={m} />
                ))}
                {otherUsed > 0 ? (
                  <li className="flex items-baseline justify-between rounded-md border bg-card px-3 py-2 text-sm">
                    <span className="font-medium">Other (not a member)</span>
                    <span className="tabular-nums">{formatSU(otherUsed)} SU</span>
                  </li>
                ) : null}
              </ul>
            </section>
          ) : null}
        </>
      )}
      <UsageRecords
        allocationId={allocation.id}
        rows={rows}
        userName={userName}
        resourceName={resourceName}
        canRead={canRead}
        canManage={canManage}
      />
    </div>
  );
}
