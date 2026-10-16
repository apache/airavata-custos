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

import { Timestamp } from "@/features/core/audit/components/primitives/Timestamp";
import type { PacketResponse } from "@/generated/amie/types.gen";
import { cn } from "@/lib/utils";
import { pluralize } from "@/shared/format";
import { DataTable, type DataTableColumn, PastLastPage } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { TriangleAlertIcon } from "lucide-react";
import { ageHoursOf } from "../utils";
import { PacketStatusBadge } from "./PacketStatusBadge";

export type FailedQueueProps = {
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
}: FailedQueueProps) {
  const columns: DataTableColumn<PacketResponse>[] = [
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
              loud ? "font-semibold text-destructive-text" : "text-muted-foreground",
            )}
          >
            {age < 1 ? `${Math.round(age * 60)}m` : `${Math.round(age)}h`}
          </span>
        );
      },
    },
    {
      key: "retries",
      header: "Failed attempts",
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
        <Timestamp
          iso={row.received_at}
          className="whitespace-nowrap text-xs tabular-nums text-muted-foreground"
        />
      ),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => (
        <PacketStatusBadge status={row.status} ageHours={ageHoursOf(row.received_at)} />
      ),
    },
  ];

  return (
    <div className="space-y-4">
      {failedOver24h > 0 ? (
        <div
          role="alert"
          className="sticky top-16 z-10 flex items-center gap-3 rounded-md border border-[color:var(--banner-error-border)] bg-[color:var(--banner-error-bg)] px-4 py-2 text-sm text-[color:var(--banner-error-fg)]"
        >
          <TriangleAlertIcon className="size-4" aria-hidden />
          <span>
            <strong>{failedOver24h}</strong> failed {pluralize("packet", failedOver24h)} older than
            24 hours {pluralize("needs", failedOver24h, "need")} attention.
          </span>
        </div>
      ) : null}

      {error ? (
        <ErrorState error={error} onRetry={onRefresh} />
      ) : isLoading ? (
        <TableSkeleton />
      ) : rows.length === 0 && page > 1 ? (
        <PastLastPage things="failed packets" page={page} onPageChange={onPageChange} />
      ) : rows.length === 0 ? (
        <EmptyState
          icon={<TriangleAlertIcon className="size-6" aria-hidden />}
          heading="No packet is in the Failed state"
        />
      ) : (
        <DataTable
          columns={columns}
          rows={rows}
          rowKey={(row) => row.id}
          onRowClick={onRowClick}
          caption={`${total} failed ${pluralize("packet", total)}`}
          pagination={{ page, pageSize, total, onPageChange }}
        />
      )}
    </div>
  );
}
