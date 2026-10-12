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
import type { PendingDelivery } from "@/generated/core/types.gen";
import { formatTimestamp } from "@/shared/format";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { Button } from "@/shared/ui/button";
import { deliveryState, leftThisRun, relativeTime } from "../utils";
import { DeliveryStatusBadge } from "./DeliveryStatusBadge";

export type DeliveryTableProps = {
  rows: PendingDelivery[];
  runningSubscribers?: Set<string>;
  canRetry: boolean;
  empty: React.ReactNode;
  onOpen: (delivery: PendingDelivery) => void;
  onRetry: (delivery: PendingDelivery) => void;
};

function whenCell(row: PendingDelivery) {
  // A delivery that is retrying matters by when it runs next, not when it was made.
  if (row.status === "PENDING" && row.attempts > 0) {
    return <Timestamp iso={row.next_run_at}>next try {relativeTime(row.next_run_at)}</Timestamp>;
  }
  return <Timestamp iso={row.created_at}>{relativeTime(row.created_at)}</Timestamp>;
}

export function DeliveryTable({
  rows,
  runningSubscribers,
  canRetry,
  empty,
  onOpen,
  onRetry,
}: DeliveryTableProps) {
  const columns: DataTableColumn<PendingDelivery>[] = [
    {
      key: "event",
      header: "Event",
      // The opener names its delivery, not only the event type many rows share.
      cell: (row) => (
        <>
          <span className="font-mono text-xs">{row.event.event_type}</span>
          <span className="sr-only">
            {" "}
            for {row.subscriber}, created {formatTimestamp(row.created_at)}, delivery{" "}
            {row.id.slice(0, 8)}
          </span>
        </>
      ),
    },
    {
      key: "status",
      header: "Status",
      width: "130px",
      cell: (row) => <DeliveryStatusBadge state={deliveryState(row, runningSubscribers)} />,
    },
    {
      key: "connector",
      header: "Connector",
      cell: (row) => <span className="font-mono text-xs">{row.subscriber}</span>,
    },
    {
      key: "source",
      header: "Sent by",
      cell: (row) => <span className="text-xs">{row.event.source}</span>,
    },
    {
      key: "attempts",
      header: "Tries",
      align: "right",
      width: "150px",
      cell: (row) => (
        <span className="text-xs tabular-nums">
          {row.status === "PENDING"
            ? leftThisRun(row.attempts)
            : `${row.attempts} ${row.attempts === 1 ? "try" : "tries"} this run`}
        </span>
      ),
    },
    {
      key: "error",
      header: "Last error",
      // The server keeps the last error after a later try succeeds.
      cell: (row) =>
        row.last_error && row.status !== "SUCCEEDED" ? (
          <span
            title={row.last_error}
            className="line-clamp-1 max-w-80 text-xs text-[color:var(--tone-error-fg)]"
          >
            {row.last_error}
          </span>
        ) : (
          <span className="text-xs text-muted-foreground">—</span>
        ),
    },
    {
      key: "when",
      header: "When",
      width: "140px",
      cell: (row) => (
        <span className="text-xs tabular-nums text-muted-foreground">{whenCell(row)}</span>
      ),
    },
  ];

  if (canRetry) {
    columns.push({
      key: "actions",
      header: <span className="sr-only">Actions</span>,
      align: "right",
      width: "90px",
      interactive: true,
      cell: (row) =>
        row.status === "FAILED" ? (
          <Button
            type="button"
            variant="outline"
            size="sm"
            aria-label={`Retry ${row.event.event_type} for ${row.subscriber}, delivery ${row.id.slice(0, 8)}`}
            onClick={(e) => {
              e.stopPropagation();
              onRetry(row);
            }}
          >
            Retry
          </Button>
        ) : null,
    });
  }

  return (
    <DataTable
      columns={columns}
      rows={rows}
      rowKey={(row) => row.id}
      onRowClick={onOpen}
      caption="Event deliveries"
      empty={empty}
    />
  );
}
