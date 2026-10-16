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
import type { ClusterAccountResponse, ComputeClusterUser } from "@/generated/core/types.gen";
import { formatDate } from "@/shared/format";
import { DataTable, type DataTableColumn, NameCell, PastLastPage } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge, type StatusBadgeVariant } from "@/shared/ui/StatusBadge";
import { emailBeside } from "@/shared/users-admin/displayName";

export function clusterAccountStatus(row: ComputeClusterUser): {
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

export type ClusterAccountsTableProps = {
  rows: ClusterAccountResponse[];
  isLoading: boolean;
  error: Error | null;
  onRetry?: () => void;
  onRowClick?: (row: ClusterAccountResponse) => void;
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
  onRowClick,
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
        <div className="min-w-0">
          <NameCell text={row.display_name || row.email} className="font-medium text-foreground" />
          <NameCell
            text={emailBeside(row.display_name || row.email, row.email)}
            className="text-xs text-muted-foreground"
          />
        </div>
      ),
    },
    {
      key: "cluster",
      header: "Cluster",
      cell: (row) => <NameCell text={row.cluster_name} className="font-mono text-xs" />,
    },
    {
      key: "username",
      header: "Username",
      cell: (row) => <NameCell text={row.local_username} className="font-mono text-xs" />,
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
            <div>
              <Timestamp iso={row.reviewed_at}>{formatDate(row.reviewed_at)}</Timestamp>
            </div>
            {row.review_note ? (
              <div className="max-w-[28ch] [overflow-wrap:anywhere]">{row.review_note}</div>
            ) : null}
          </div>
        ) : null,
    },
  ];

  if (isLoading) return <TableSkeleton rows={6} columns={columns.length} />;
  if (error) return <ErrorState error={error} onRetry={onRetry} />;
  if (rows.length === 0 && page > 1) {
    return <PastLastPage things="cluster accounts" page={page} onPageChange={onPageChange} />;
  }
  if (rows.length === 0) return <EmptyState heading={emptyHeading} />;

  return (
    <DataTable
      caption="Cluster accounts"
      columns={columns}
      rows={rows}
      rowKey={(row) => row.id}
      onRowClick={onRowClick}
      pagination={{ page, pageSize, total, onPageChange }}
    />
  );
}
