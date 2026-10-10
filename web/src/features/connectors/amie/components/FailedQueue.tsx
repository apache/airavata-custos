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

import type { PacketResponse } from "@/generated/amie/types.gen";
import { TriangleAlertIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import { ageHoursOf, formatDate, pluralize } from "../utils";
import { PacketStatusBadge } from "./PacketStatusBadge";
import { PacketActionButtons, type PacketActions, selectColumn } from "./packetActions";

export type FailedQueueProps = PacketActions & {
  rows: PacketResponse[];
  total: number;
  isLoading: boolean;
  error: Error | null;
  failedOver24h: number;
  page: number;
  pageSize: number;
  onRowClick: (packet: PacketResponse) => void;
  onPageChange: (page: number) => void;
  onRefresh: () => void;
  selected: Set<string>;
  onSelectChange: (selected: Set<string>) => void;
};

export function FailedQueue({
  rows,
  total,
  isLoading,
  error,
  failedOver24h,
  page,
  pageSize,
  onRowClick,
  onPageChange,
  onRefresh,
  selected,
  onSelectChange,
  ...actions
}: FailedQueueProps) {
  const columns: DataTableColumn<PacketResponse>[] = [
    ...(actions.onRetryPackets ? [selectColumn(rows, selected, onSelectChange)] : []),
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
      header: "",
      align: "right",
      interactive: true,
      cell: (row) => (
        <span className="flex justify-end gap-2">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={() => onRowClick(row)}
          >
            View
          </Button>
          <PacketActionButtons ids={[row.id]} {...actions} />
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
            <strong>{failedOver24h}</strong> failed {pluralize("packet", failedOver24h)} older
            than 24 hours need attention.
          </span>
        </div>
      ) : null}

      <div className="flex items-center gap-2 rounded-md border bg-card p-4">
        <p className="text-xs text-muted-foreground">Pre-filtered packets with status = FAILED.</p>
        {actions.onRetryPackets ? (
          <div className="ml-auto flex items-center gap-2" role="toolbar" aria-label="Bulk actions">
            <span className="text-xs text-muted-foreground">{selected.size} selected</span>
            <PacketActionButtons ids={[...selected]} bulk {...actions} />
          </div>
        ) : null}
      </div>

      {error ? (
        <ErrorState message={error.message} onRetry={onRefresh} />
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
          caption={`${total} failed ${pluralize("packet", total)}`}
          pagination={{ page, pageSize, total, onPageChange }}
        />
      )}
    </div>
  );
}
