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

import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { Button } from "@/shared/ui/button";
import type { Delivery } from "../types";
import { MAX_ATTEMPTS, absoluteTime, deliveryState, relativeTime } from "../utils";
import { DeliveryStatusBadge } from "./DeliveryStatusBadge";

export type DeliveryTableProps = {
  rows: Delivery[];
  runningSubscribers?: Set<string>;
  canRetry: boolean;
  empty: React.ReactNode;
  onOpen: (delivery: Delivery) => void;
  onRetry: (delivery: Delivery) => void;
};

function whenCell(row: Delivery) {
  // A delivery that is retrying matters by when it runs next, not when it was made.
  if (row.status === "PENDING" && row.attempts > 0) {
    return (
      <span title={absoluteTime(row.next_run_at)}>next try {relativeTime(row.next_run_at)}</span>
    );
  }
  return <span title={absoluteTime(row.created_at)}>{relativeTime(row.created_at)}</span>;
}

export function DeliveryTable({
  rows,
  runningSubscribers,
  canRetry,
  empty,
  onOpen,
  onRetry,
}: DeliveryTableProps) {
  const columns: DataTableColumn<Delivery>[] = [
    {
      key: "status",
      header: "Status",
      width: "130px",
      cell: (row) => <DeliveryStatusBadge state={deliveryState(row, runningSubscribers)} />,
    },
    {
      key: "event",
      header: "Event",
      cell: (row) => <span className="font-mono text-xs">{row.event.event_type}</span>,
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
      header: "Attempts",
      align: "right",
      width: "90px",
      cell: (row) => (
        <span className="text-xs tabular-nums">
          {row.attempts} / {MAX_ATTEMPTS}
        </span>
      ),
    },
    {
      key: "error",
      header: "Last error",
      cell: (row) =>
        row.last_error ? (
          <span
            title={row.last_error}
            className="line-clamp-1 max-w-80 text-xs text-[color:var(--tone-error-fg)]"
          >
            {row.last_error}
          </span>
        ) : (
          <span className="text-xs text-muted-foreground">-</span>
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
            aria-label={`Retry ${row.event.event_type} for ${row.subscriber}`}
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
