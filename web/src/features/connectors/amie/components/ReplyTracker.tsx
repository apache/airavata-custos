"use client";

import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import { Label } from "@/shared/ui/label";
import type { Reply, ReplyStatus } from "../types";
import { formatDate } from "../utils";
import { ReplyStatusBadge } from "./PacketStatusBadge";

export type ReplyTrackerProps = {
  rows: Reply[];
  total: number;
  isLoading: boolean;
  error: Error | null;
  statusFilter: ReplyStatus | "all";
  onStatusChange: (next: ReplyStatus | "all") => void;
  onRetry: (id: string) => void;
  onRefresh: () => void;
};

export function ReplyTracker({
  rows,
  total,
  isLoading,
  error,
  statusFilter,
  onStatusChange,
  onRetry,
  onRefresh,
}: ReplyTrackerProps) {
  const columns: DataTableColumn<Reply>[] = [
    {
      key: "amie_id",
      header: "AMIE ID",
      cell: (r) => <span className="font-mono text-sm">{r.amie_id}</span>,
    },
    {
      key: "type",
      header: "Type",
      cell: (r) => <span className="text-sm">{r.type}</span>,
    },
    {
      key: "status",
      header: "Status",
      cell: (r) => <ReplyStatusBadge status={r.status} />,
    },
    {
      key: "in_reply_to",
      header: "Replies to",
      cell: (r) => (
        <span className="font-mono text-xs text-muted-foreground">
          {r.in_reply_to_packet_id ?? "—"}
        </span>
      ),
    },
    {
      key: "created",
      header: "Created",
      cell: (r) => <span className="text-xs tabular-nums">{formatDate(r.created_at)}</span>,
    },
    {
      key: "sent",
      header: "Sent",
      cell: (r) => <span className="text-xs tabular-nums">{formatDate(r.sent_at)}</span>,
    },
    {
      key: "acked",
      header: "Acked",
      cell: (r) => <span className="text-xs tabular-nums">{formatDate(r.acked_at)}</span>,
    },
    {
      key: "retries",
      header: "Retries",
      cell: (r) => <span className="text-xs tabular-nums">{r.retries}</span>,
    },
    {
      key: "actions",
      header: "Actions",
      align: "right",
      cell: (r) =>
        r.status === "FAILED" || r.status === "PENDING" ? (
          <Button type="button" variant="outline" size="sm" onClick={() => onRetry(r.id)}>
            Retry
          </Button>
        ) : null,
    },
  ];

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-end gap-3 rounded-md border bg-card p-4">
        <div className="flex flex-col gap-1">
          <Label htmlFor="reply-status">Status</Label>
          <select
            id="reply-status"
            value={statusFilter}
            onChange={(e) => onStatusChange(e.currentTarget.value as ReplyStatus | "all")}
            className="rounded-md border bg-background px-3 py-1.5 text-sm"
          >
            <option value="all">All</option>
            <option value="PENDING">PENDING</option>
            <option value="SENT">SENT</option>
            <option value="ACKED">ACKED</option>
            <option value="FAILED">FAILED</option>
          </select>
        </div>
        <p className="ml-auto text-xs text-muted-foreground">
          {total} reply{total === 1 ? "" : "s"} in scope
        </p>
      </div>

      {error ? (
        <ErrorState message={error.message ?? "Failed to load replies"} onRetry={onRefresh} />
      ) : isLoading ? (
        <TableSkeleton />
      ) : rows.length === 0 ? (
        <EmptyState
          heading="No outgoing replies"
          description="inform_* packets the connector emits back to ACCESS will appear here."
        />
      ) : (
        <DataTable columns={columns} rows={rows} rowKey={(r) => r.id} />
      )}
    </div>
  );
}
