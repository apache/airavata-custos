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

import { formatNumber } from "@/shared/format";
import { useResourceSummaries } from "@/features/core/resources/queries";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { UsageBar } from "@/shared/ui/UsageBar";
import {
  useAllocationMembers,
  useAllocationUsage,
  useCurrentSuAmount,
  useUsageTotal,
  useUserUsageTotal,
} from "../queries";
import type {
  AllocationMembershipResponse,
  ComputeAllocation,
} from "@/generated/core/types.gen";
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
      <span className="font-medium">{member.display_name ?? member.user_id}</span>
      <span className="tabular-nums">{total === undefined ? "—" : `${formatNumber(total)} SU`}</span>
    </li>
  );
}

const dash = (n?: number) => (n === undefined ? "—" : formatNumber(n));

export function AllocationUsageTab({ allocation, canRead, canManage }: AllocationUsageTabProps) {
  const query = useAllocationUsage(allocation.id);
  const latestSu = useCurrentSuAmount(allocation, canRead).data;
  const currentSu = canRead ? latestSu : allocation.initial_su_amount;
  const usedTotal = useUsageTotal(allocation.id).data;
  const members = useAllocationMembers(canRead ? allocation.id : undefined).data ?? [];
  const summariesQuery = useResourceSummaries();
  if (query.isLoading) return <TableSkeleton rows={3} columns={3} />;
  if (query.error) {
    return <ErrorState message={query.error.message} onRetry={() => query.refetch()} />;
  }
  const rows = query.data ?? [];

  const nameById = new Map((summariesQuery.data ?? []).map((s) => [s.id, s.name]));
  const resourceName = (id = "") => nameById.get(id) ?? id;
  const userName = (id = "") => members.find((m) => m.user_id === id)?.display_name ?? id;
  // Per-resource SU used, summed from usage rows. No per-resource SU allocation
  // exists, so bars show each resource's share of total usage, not a quota.
  const byResource = new Map<string, number>();
  for (const row of rows) {
    const key = row.compute_allocation_resource_id ?? "";
    byResource.set(key, (byResource.get(key) ?? 0) + (row.used_su_amount ?? 0));
  }

  return (
    <div className="space-y-4">
      <UsageBar
        value={usedTotal ?? 0}
        max={currentSu ?? 0}
        label={`${dash(usedTotal)} / ${dash(currentSu)} SUs`}
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
                    <span className="font-medium">{name}</span>
                    <span className="tabular-nums">{formatNumber(used)} SU</span>
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
          {members.length > 0 ? (
            <section className="space-y-2">
              <h2 className="text-sm font-semibold text-foreground">By member</h2>
              <ul className="space-y-1">
                {members.map((m) => (
                  <MemberUsage key={m.id} allocationId={allocation.id ?? ""} member={m} />
                ))}
              </ul>
            </section>
          ) : null}
        </>
      )}
      {rows.length > 0 || canManage ? (
        <UsageRecords
          allocationId={allocation.id ?? ""}
          rows={rows}
          resourceName={resourceName}
          userName={userName}
          canRead={canRead}
          canManage={canManage}
        />
      ) : null}
    </div>
  );
}
