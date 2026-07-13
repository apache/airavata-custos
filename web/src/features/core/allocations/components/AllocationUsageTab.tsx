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

import { useResourceSummaries } from "@/features/core/resources/queries";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { UsageBar } from "@/shared/ui/UsageBar";
import { useAllocationUsage } from "../queries";
import type { ComputeAllocation } from "../schemas";

export type AllocationUsageTabProps = {
  allocation: ComputeAllocation;
};

function formatNumber(n: number): string {
  return new Intl.NumberFormat().format(n);
}

export function AllocationUsageTab({ allocation }: AllocationUsageTabProps) {
  const query = useAllocationUsage(allocation.id);
  const summariesQuery = useResourceSummaries();
  if (query.isLoading) return <TableSkeleton rows={3} columns={3} />;
  if (query.error) {
    return <ErrorState message={(query.error as Error).message} onRetry={() => query.refetch()} />;
  }
  const rows = query.data ?? [];
  const usedTotal = rows.reduce((acc, r) => acc + r.used_su_amount, 0);

  const nameById = new Map((summariesQuery.data ?? []).map((s) => [s.id, s.name]));
  // Per-resource SU used, summed from usage rows. No per-resource SU allocation
  // exists, so bars show each resource's share of total usage, not a quota.
  const byResource = new Map<string, number>();
  for (const row of rows) {
    const key = row.compute_allocation_resource_id;
    byResource.set(key, (byResource.get(key) ?? 0) + row.used_su_amount);
  }

  return (
    <div className="space-y-4">
      <UsageBar
        value={usedTotal}
        max={allocation.initial_su_amount}
        label={`${formatNumber(usedTotal)} / ${formatNumber(allocation.initial_su_amount)} SUs`}
        ariaLabel="Allocation SU usage"
      />
      {rows.length === 0 ? (
        <EmptyState
          heading="No usage recorded"
          description="Jobs that consume this allocation will appear here."
        />
      ) : (
        <ul className="space-y-3">
          {[...byResource.entries()].map(([resourceId, used]) => {
            const name = nameById.get(resourceId) ?? resourceId;
            return (
              <li key={resourceId} className="rounded-md border bg-card px-3 py-2">
                <div className="mb-1 flex items-baseline justify-between text-sm">
                  <span className="font-medium">{name}</span>
                  <span className="tabular-nums">{formatNumber(used)} SU</span>
                </div>
                <UsageBar
                  value={used}
                  max={usedTotal}
                  ariaLabel={`${name} share of usage`}
                  size="sm"
                />
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
