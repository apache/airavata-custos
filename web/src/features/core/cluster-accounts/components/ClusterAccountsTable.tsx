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

import { useProject } from "@/features/core/projects/queries";
import type { ClusterAccountResponse } from "@/generated/core/zod.gen";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import * as React from "react";
import { AllocationRef } from "./AllocationRef";
import { ApprovalGroupHeader } from "./ApprovalGroupHeader";
import { clusterAccountStatus } from "./ClusterAccountBadge";

function formatDate(iso: string | undefined, withYear = true): string {
  if (!iso) return "";
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleDateString(undefined, {
    ...(withYear ? { year: "numeric" } : {}),
    month: "short",
    day: "numeric",
  });
}

type ColumnOptions = {
  grouped: boolean;
  piId?: string | undefined;
  canReview: boolean;
  onApprove: (row: ClusterAccountResponse) => void;
  onDeny: (row: ClusterAccountResponse) => void;
};

function buildColumns({ grouped, piId, canReview, onApprove, onDeny }: ColumnOptions) {
  const columns: Array<DataTableColumn<ClusterAccountResponse> | false> = [
    {
      key: "user",
      header: "User",
      cell: (row) => (
        <div>
          <div className="flex items-center gap-2 font-medium text-foreground">
            {row.display_name || row.email}
            {piId && row.user_id === piId ? (
              <Badge
                variant="outline"
                className="h-4 px-1.5 py-0 text-[10px] text-muted-foreground"
              >
                PI
              </Badge>
            ) : null}
          </div>
          <div className="text-xs text-muted-foreground">{row.email}</div>
        </div>
      ),
    },
    !grouped && {
      key: "allocation",
      header: "Allocation",
      cell: (row) => <AllocationRef account={row} withTitle />,
    },
    !grouped && {
      key: "cluster",
      header: "Cluster",
      cell: (row) => (
        <span className="whitespace-nowrap font-mono text-xs">{row.cluster_name}</span>
      ),
    },
    {
      key: "username",
      header: "Username",
      ...(grouped ? { width: "14rem" } : {}),
      cell: (row) => (
        <span className="whitespace-nowrap font-mono text-xs">{row.local_username}</span>
      ),
    },
    grouped
      ? {
          key: "requested",
          header: "Waiting since",
          width: "9rem",
          cell: (row) => (
            <span className="text-sm text-muted-foreground">
              {formatDate(row.requested_at, false)}
            </span>
          ),
        }
      : {
          key: "reviewed",
          header: "Reviewed",
          cell: (row) =>
            row.reviewed_at ? (
              <div className="text-xs text-muted-foreground">
                <div className="whitespace-nowrap">{formatDate(row.reviewed_at)}</div>
                {row.reviewer_name ? <div>by {row.reviewer_name}</div> : null}
                {row.review_note ? <div className="max-w-[28ch]">{row.review_note}</div> : null}
              </div>
            ) : null,
        },
    {
      key: "approval",
      header: "Approval",
      align: "right",
      ...(grouped ? { width: "10rem" } : {}),
      interactive: true,
      cell: (row) => {
        if (canReview && row.approval_status === "PENDING") {
          return (
            <div className="flex justify-end gap-1">
              <Button size="xs" variant="success" onClick={() => onApprove(row)}>
                Approve
              </Button>
              <Button size="xs" variant="destructive" onClick={() => onDeny(row)}>
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
  return columns.filter((c) => c !== false);
}

type GroupProps = Omit<ColumnOptions, "grouped" | "piId"> & {
  rows: ClusterAccountResponse[];
  onApproveAll: (rows: ClusterAccountResponse[]) => void;
};

function ApprovalGroup({ rows, onApproveAll, ...options }: GroupProps) {
  const labelId = React.useId();
  const head = rows[0];
  const { data: project } = useProject(head?.project_id);
  const piId = project?.project_pi_id;
  const isPi = (row: ClusterAccountResponse) => Number(row.user_id === piId);
  const sorted = [...rows].sort((a, b) => isPi(b) - isPi(a));
  const canApproveAll = options.canReview && Boolean(head?.allocation_id) && rows.length >= 2;
  return (
    <section
      aria-labelledby={labelId}
      className="overflow-hidden rounded-lg border border-foreground/20 bg-card"
    >
      <div className="border-b border-border bg-muted px-4 py-2.5">
        <ApprovalGroupHeader
          labelId={labelId}
          allocationId={head?.allocation_id}
          allocationName={head?.allocation_name}
          project={project}
          cluster={head?.cluster_name ?? ""}
          count={rows.length}
          onApproveAll={canApproveAll ? () => onApproveAll(rows) : undefined}
        />
      </div>
      <DataTable
        className="rounded-none border-0 [&_th]:py-1.5 [&_thead]:bg-muted/40 [&_td:first-child]:pl-11 [&_th:first-child]:pl-11 [&_td]:align-middle"
        columns={buildColumns({ ...options, grouped: true, piId })}
        rows={sorted}
        rowKey={(row) => row.id}
      />
    </section>
  );
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
  grouped?: boolean;
  onApproveAll: (rows: ClusterAccountResponse[]) => void;
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
  grouped = false,
  onApproveAll,
}: ClusterAccountsTableProps) {
  const options = { canReview, onApprove, onDeny };
  const columns = buildColumns({ ...options, grouped });

  if (isLoading) return <TableSkeleton rows={6} columns={columns.length} />;
  if (error) return <ErrorState message={error.message} onRetry={onRetry} />;

  if (!grouped || rows.length === 0) {
    return (
      <DataTable
        columns={columns}
        rows={rows}
        rowKey={(row) => row.id}
        empty={emptyHeading}
        {...(grouped ? {} : { pagination: { page, pageSize, total, onPageChange } })}
      />
    );
  }

  const groups = [...Map.groupBy(rows, (row) => row.allocation_id)];
  return (
    <div className="space-y-3">
      {groups.map(([allocationId, groupRows]) => (
        <ApprovalGroup
          key={allocationId ?? "none"}
          rows={groupRows}
          onApproveAll={onApproveAll}
          {...options}
        />
      ))}
    </div>
  );
}
