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

import type { ClusterAccountResponse } from "@/generated/core/zod.gen";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { Button } from "@/shared/ui/button";
import { clusterAccountStatus } from "./ClusterAccountBadge";

function formatDate(iso?: string | null): string {
  if (!iso) return "";
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

export type ClusterAccountsTableProps = {
  rows: ClusterAccountResponse[];
  isLoading: boolean;
  error: Error | null;
  onRetry?: () => void;
  canReview: boolean;
  onApprove: (row: ClusterAccountResponse) => void;
  onDeny: (row: ClusterAccountResponse) => void;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  emptyHeading: string;
};

export function ClusterAccountsTable({
  rows,
  isLoading,
  error,
  onRetry,
  canReview,
  onApprove,
  onDeny,
  page,
  pageSize,
  total,
  onPageChange,
  emptyHeading,
}: ClusterAccountsTableProps) {
  const columns: Array<DataTableColumn<ClusterAccountResponse>> = [
    {
      key: "user",
      header: "User",
      cell: (row) => (
        <div>
          <div className="font-medium text-foreground">{row.display_name || row.email}</div>
          <div className="text-xs text-muted-foreground">{row.email}</div>
        </div>
      ),
    },
    {
      key: "cluster",
      header: "Cluster",
      cell: (row) => <span className="font-mono text-xs">{row.cluster_name}</span>,
    },
    {
      key: "username",
      header: "Username",
      cell: (row) => <span className="font-mono text-xs">{row.local_username}</span>,
    },
    {
      key: "reviewed",
      header: "Reviewed",
      cell: (row) =>
        row.reviewed_at ? (
          <div className="text-xs text-muted-foreground">
            <div>{formatDate(row.reviewed_at)}</div>
            {row.review_note ? <div className="max-w-[28ch]">{row.review_note}</div> : null}
          </div>
        ) : null,
    },
    {
      key: "approval",
      header: "Approval",
      align: "right",
      interactive: true,
      cell: (row) => {
        if (canReview && row.approval_status === "PENDING") {
          return (
            <div className="flex justify-end gap-1">
              <Button size="xs" variant="success" onClick={() => onApprove(row)}>
                Approve
              </Button>
              <Button size="xs" variant="outline" onClick={() => onDeny(row)}>
                Deny
              </Button>
            </div>
          );
        }
        const status = clusterAccountStatus(row);
        return <StatusBadge variant={status.variant} label={status.label} />;
      },
    },
  ];

  if (isLoading) return <TableSkeleton rows={6} columns={columns.length} />;
  if (error) return <ErrorState message={error.message} onRetry={onRetry} />;

  return (
    <DataTable
      columns={columns}
      rows={rows}
      rowKey={(row) => row.id}
      empty={emptyHeading}
      pagination={{ page, pageSize, total, onPageChange }}
    />
  );
}
