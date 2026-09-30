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

import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge, type StatusBadgeVariant } from "@/shared/ui/StatusBadge";
import type { ClusterAccount } from "../schemas";

export function clusterAccountStatus(row: ClusterAccount): {
  variant: StatusBadgeVariant;
  label: string;
} {
  if (row.approval_status === "DENIED") return { variant: "rejected", label: "Denied" };
  if (row.approval_status === "APPROVED") {
    return row.provisioned_at
      ? { variant: "active", label: "Active" }
      : { variant: "pending", label: "Approved · provisioning" };
  }
  return { variant: "warning", label: "Pending approval" };
}

function formatDate(iso?: string | null): string {
  if (!iso) return "";
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

export type ClusterAccountsTableProps = {
  rows: ClusterAccount[];
  isLoading: boolean;
  error: Error | null;
  onRetry?: () => void;
  canReview: boolean;
  onApprove: (row: ClusterAccount) => void;
  onDeny: (row: ClusterAccount) => void;
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
  const columns: Array<DataTableColumn<ClusterAccount>> = [
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
      key: "status",
      header: "Status",
      cell: (row) => {
        const status = clusterAccountStatus(row);
        return <StatusBadge variant={status.variant} label={status.label} />;
      },
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
  ];

  if (canReview) {
    columns.push({
      key: "actions",
      header: "",
      align: "right",
      interactive: true,
      cell: (row) => {
        if (row.approval_status === "APPROVED") return null;
        return (
          <div className="flex justify-end gap-2">
            {row.approval_status === "PENDING" ? (
              <Button size="sm" variant="destructive" onClick={() => onDeny(row)}>
                Deny
              </Button>
            ) : null}
            <Button size="sm" variant="success" onClick={() => onApprove(row)}>
              Approve
            </Button>
          </div>
        );
      },
    });
  }

  if (isLoading) return <TableSkeleton rows={6} columns={columns.length} />;
  if (error) return <ErrorState message={error.message} onRetry={onRetry} />;
  if (rows.length === 0) return <EmptyState heading={emptyHeading} />;

  return (
    <DataTable
      columns={columns}
      rows={rows}
      rowKey={(row) => row.id}
      pagination={{ page, pageSize, total, onPageChange }}
    />
  );
}
