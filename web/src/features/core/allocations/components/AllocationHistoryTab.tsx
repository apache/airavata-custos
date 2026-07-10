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
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge, statusBadgeVariantFromAllocationStatus } from "@/shared/ui/StatusBadge";
import { useAllocationDiffs } from "../queries";
import type { AllocationDiff, ComputeAllocation } from "../schemas";

export type AllocationHistoryTabProps = {
  allocation: ComputeAllocation;
};

function formatDateTime(iso: string): string {
  try {
    return new Date(iso).toLocaleString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return iso;
  }
}

export function AllocationHistoryTab({ allocation }: AllocationHistoryTabProps) {
  const query = useAllocationDiffs(allocation.id);

  if (query.isLoading) return <TableSkeleton rows={3} columns={5} />;
  if (query.error) {
    return <ErrorState message={(query.error as Error).message} onRetry={() => query.refetch()} />;
  }
  const rows = [...(query.data ?? [])].sort(
    (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime(),
  );

  if (rows.length === 0) {
    return <EmptyState heading="No history recorded for this allocation." />;
  }

  const columns: Array<DataTableColumn<AllocationDiff>> = [
    {
      key: "when",
      header: "When",
      cell: (row) => (
        <span className="text-sm text-muted-foreground">{formatDateTime(row.timestamp)}</span>
      ),
    },
    {
      key: "type",
      header: "Type",
      cell: (row) => <span className="text-sm">{row.diff_type}</span>,
    },
    {
      key: "su",
      header: "New SU amount",
      align: "right",
      cell: (row) => (
        <span className="tabular-nums">
          {typeof row.new_su_amount === "number"
            ? new Intl.NumberFormat().format(row.new_su_amount)
            : "-"}
        </span>
      ),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => (
        <StatusBadge
          variant={statusBadgeVariantFromAllocationStatus(row.status)}
          label={row.status}
        />
      ),
    },
    {
      key: "description",
      header: "Description",
      cell: (row) =>
        row.description ? (
          <span className="text-sm text-muted-foreground">{row.description}</span>
        ) : null,
    },
  ];

  return <DataTable columns={columns} rows={rows} rowKey={(row) => row.id} />;
}
