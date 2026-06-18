"use client";

import { TriangleAlertIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import type { Packet } from "../types";
import { ageHoursOf, formatDate } from "../utils";
import { PacketStatusBadge } from "./PacketStatusBadge";

export type FailedQueueProps = {
  rows: Packet[];
  total: number;
  isLoading: boolean;
  error: Error | null;
  failedOver24h: number;
  selected: Set<string>;
  onSelectChange: (next: Set<string>) => void;
  onRowClick: (packet: Packet) => void;
  onRetryRow: (id: string) => void;
  onResolveRow: (packet: Packet) => void;
  onBulkRetry: () => void;
  onRefresh: () => void;
};

export function FailedQueue({
  rows,
  total,
  isLoading,
  error,
  failedOver24h,
  selected,
  onSelectChange,
  onRowClick,
  onRetryRow,
  onResolveRow,
  onBulkRetry,
  onRefresh,
}: FailedQueueProps) {
  const allSelected = rows.length > 0 && rows.every((r) => selected.has(r.id));

  function toggleAll() {
    if (allSelected) onSelectChange(new Set());
    else onSelectChange(new Set(rows.map((r) => r.id)));
  }

  function toggleOne(id: string) {
    const next = new Set(selected);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    onSelectChange(next);
  }

  const columns: DataTableColumn<Packet>[] = [
    {
      key: "select",
      header: (
        <input type="checkbox" aria-label="Select all" checked={allSelected} onChange={toggleAll} />
      ),
      width: "32px",
      interactive: true,
      cell: (row) => (
        <input
          type="checkbox"
          aria-label={`Select ${row.amie_id}`}
          checked={selected.has(row.id)}
          onChange={() => toggleOne(row.id)}
          onClick={(e) => e.stopPropagation()}
        />
      ),
    },
    {
      key: "amie_id",
      header: "AMIE ID",
      cell: (row) => <span className="font-mono text-sm">{row.amie_id}</span>,
    },
    {
      key: "type",
      header: "Type",
      cell: (row) => <span className="text-sm">{row.type}</span>,
    },
    {
      key: "age",
      header: "Age",
      cell: (row) => {
        const age = ageHoursOf(row.received_at);
        const loud = age > 24;
        return (
          <span
            className={cn(
              "text-xs tabular-nums",
              loud ? "font-semibold text-destructive" : "text-muted-foreground",
            )}
          >
            {age < 1 ? `${Math.round(age * 60)}m` : `${Math.round(age)}h`}
          </span>
        );
      },
    },
    {
      key: "retries",
      header: "Retries",
      cell: (row) => <span className="text-xs tabular-nums">{row.retries}</span>,
    },
    {
      key: "error",
      header: "Last error",
      cell: (row) => (
        <span className="line-clamp-1 text-xs text-muted-foreground">{row.last_error ?? "—"}</span>
      ),
    },
    {
      key: "received",
      header: "Received",
      cell: (row) => (
        <span className="text-xs tabular-nums text-muted-foreground">
          {formatDate(row.received_at)}
        </span>
      ),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => (
        <PacketStatusBadge status={row.status} ageHours={ageHoursOf(row.received_at)} />
      ),
    },
    {
      key: "actions",
      header: "Actions",
      align: "right",
      interactive: true,
      cell: (row) => (
        <span className="flex justify-end gap-2">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={(e) => {
              e.stopPropagation();
              onRowClick(row);
            }}
          >
            View
          </Button>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={(e) => {
              e.stopPropagation();
              onRetryRow(row.id);
            }}
          >
            Retry
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={(e) => {
              e.stopPropagation();
              onResolveRow(row);
            }}
          >
            Resolve
          </Button>
        </span>
      ),
    },
  ];

  return (
    <div className="space-y-4">
      {failedOver24h > 0 ? (
        <div
          role="alert"
          className="sticky top-0 z-10 flex items-center gap-3 rounded-md border border-[color:var(--custos-red-200)] bg-[color:var(--custos-red-50)] px-4 py-2 text-sm text-[color:var(--custos-red-700)]"
        >
          <TriangleAlertIcon className="size-4" aria-hidden />
          <span>
            <strong>{failedOver24h}</strong> failed packet{failedOver24h === 1 ? "" : "s"} older
            than 24 hours need attention.
          </span>
        </div>
      ) : null}

      <div className="flex items-center justify-between rounded-md border bg-card p-4">
        <p className="text-xs text-muted-foreground">
          Pre-filtered packets with status = FAILED. Use bulk retry to requeue all selected.
        </p>
        <div className="flex items-center gap-2">
          <span className="text-xs text-muted-foreground">{selected.size} selected</span>
          <Button variant="outline" disabled={selected.size === 0} onClick={onBulkRetry}>
            Bulk retry
          </Button>
        </div>
      </div>

      {error ? (
        <ErrorState message={error.message ?? "Failed to load queue"} onRetry={onRefresh} />
      ) : isLoading ? (
        <TableSkeleton />
      ) : rows.length === 0 ? (
        <EmptyState
          icon={<TriangleAlertIcon className="size-6" aria-hidden />}
          heading="Nothing failed"
          description="All packets have either processed cleanly or are still in-flight."
        />
      ) : (
        <DataTable
          columns={columns}
          rows={rows}
          rowKey={(row) => row.id}
          caption={`${total} failed packet${total === 1 ? "" : "s"}`}
        />
      )}
    </div>
  );
}
