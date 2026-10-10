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

import type { ReplyResponse } from "@/generated/amie/types.gen";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import { pluralize } from "../utils";

export type ReplyTrackerProps = {
  rows: ReplyResponse[];
  // Undefined until the list loads.
  total: number | undefined;
  isLoading: boolean;
  error: Error | null;
  onRefresh: () => void;
  // Omitted when the caller lacks amie:replies:write.
  onRetryReply?: (id: string) => void;
};

export function ReplyTracker({
  rows,
  total,
  isLoading,
  error,
  onRefresh,
  onRetryReply,
}: ReplyTrackerProps) {
  const columns: DataTableColumn<ReplyResponse>[] = [
    {
      key: "id",
      header: "Reply ID",
      cell: (r) => <span className="font-mono text-sm">{r.id}</span>,
    },
    {
      key: "actions",
      header: "",
      align: "right",
      interactive: true,
      cell: (r) =>
        onRetryReply ? (
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => onRetryReply(r.id)}
          >
            Retry
          </Button>
        ) : null,
    },
  ];
  return (
    <div className="space-y-4">
      <div className="flex rounded-md border bg-card p-4">
        <p className="ml-auto text-xs text-muted-foreground">
          {total === undefined ? null : `${total} ${pluralize("reply", total, "replies")} in scope`}
        </p>
      </div>

      {error ? (
        <ErrorState message={error.message} onRetry={onRefresh} />
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
