"use client";

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
  if (query.isLoading) return <TableSkeleton rows={3} columns={3} />;
  if (query.error) {
    return <ErrorState message={(query.error as Error).message} onRetry={() => query.refetch()} />;
  }
  const rows = query.data ?? [];
  const usedTotal = rows.reduce((acc, r) => acc + r.used_su_amount, 0);

  if (rows.length === 0) {
    return (
      <div className="space-y-4">
        <UsageBar
          value={0}
          max={allocation.initial_su_amount}
          label={`0 / ${formatNumber(allocation.initial_su_amount)} SUs`}
          ariaLabel="Allocation SU usage"
        />
        <EmptyState
          heading="No usage recorded"
          description="Jobs that consume this allocation will appear here."
        />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <UsageBar
        value={usedTotal}
        max={allocation.initial_su_amount}
        label={`${formatNumber(usedTotal)} / ${formatNumber(allocation.initial_su_amount)} SUs`}
        ariaLabel="Allocation SU usage"
      />
      <ul className="space-y-1">
        {rows.map((row) => (
          <li
            key={row.id}
            className="flex items-center justify-between rounded-md border bg-card px-3 py-2 text-sm"
          >
            <div>
              <div className="font-medium">{row.user_id}</div>
              <div className="font-mono text-xs text-muted-foreground">{row.job_id}</div>
            </div>
            <span className="tabular-nums">{formatNumber(row.used_su_amount)}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
