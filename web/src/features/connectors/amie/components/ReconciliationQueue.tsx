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
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import { Label } from "@/shared/ui/label";
import { formatDate } from "../utils";

export type ReconciliationQueueProps = {
  rows: PacketResponse[];
  total: number;
  isLoading: boolean;
  error: Error | null;
  onRefresh: () => void;
  // Omitted when the caller lacks amie:unmapped:write.
  onLink?: (id: string) => void;
};

export function ReconciliationQueue({
  rows,
  total,
  isLoading,
  error,
  onRefresh,
  onLink,
}: ReconciliationQueueProps) {
  const columns: DataTableColumn<PacketResponse>[] = [
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
      key: "received",
      header: "Received",
      cell: (r) => (
        <span className="text-xs tabular-nums text-muted-foreground">
          {formatDate(r.received_at)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "",
      align: "right",
      interactive: true,
      cell: (r) =>
        onLink ? (
          <Button type="button" variant="outline" size="sm" onClick={() => onLink(r.id)}>
            Link
          </Button>
        ) : null,
    },
  ];
  return (
    <div className="space-y-4">
      <div className="rounded-md border bg-card p-4">
        <div className="flex items-end justify-between gap-3">
          <p className="text-xs text-muted-foreground">
            Decoded packets that couldn't be auto-linked to a domain entity.
          </p>
          <Label className="text-xs text-muted-foreground">{total} unmapped</Label>
        </div>
      </div>

      {error ? (
        <ErrorState message={error.message} onRetry={onRefresh} />
      ) : isLoading ? (
        <TableSkeleton />
      ) : rows.length === 0 ? (
        <EmptyState
          heading="Inbox clean"
          description="Every decoded packet is mapped to a domain entity."
        />
      ) : (
        <DataTable columns={columns} rows={rows} rowKey={(r) => r.id} />
      )}
    </div>
  );
}
