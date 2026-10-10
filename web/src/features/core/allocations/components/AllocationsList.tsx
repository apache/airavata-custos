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

import { formatDate, formatNumber } from "@/shared/format";
import Link from "next/link";
import * as React from "react";
import { DataTable, type DataTableColumn, type DataTablePagination } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { StatusBadge, statusBadgeVariantFromAllocationStatus } from "@/shared/ui/StatusBadge";
import { STATUSES, statusLabel } from "@/shared/ui/FormDialog";
import type { AllocationStatus, ComputeAllocation } from "@/generated/core/types.gen";
import { useClusterName } from "@/features/core/clusters/queries";
import { useProject } from "@/features/core/projects/queries";

export function ProjectLink({ id, className }: { id: string; className?: string }) {
  const title = useProject(id).data?.title;
  return (
    <Link
      href={`/projects/${id}`}
      className={className ?? "break-all text-sm text-muted-foreground hover:underline"}
    >
      {title ?? id}
    </Link>
  );
}

export type AllocationsListProps = {
  rows: ComputeAllocation[];
  isLoading: boolean;
  error: Error | null;
  onRetry?: () => void;
  search: string;
  onSearchChange: (next: string) => void;
  statusFilter: AllocationStatus | "all";
  onStatusFilterChange: (next: AllocationStatus | "all") => void;
  pagination?: DataTablePagination;
};

export function AllocationsList({
  rows,
  isLoading,
  error,
  onRetry,
  search,
  onSearchChange,
  statusFilter,
  onStatusFilterChange,
  pagination,
}: AllocationsListProps) {
  const clusterName = useClusterName();
  // The backend skips filters for callers without allocations read.
  const filtered = React.useMemo(() => {
    const needle = search.trim().toLowerCase();
    return rows.filter(
      (row) =>
        (statusFilter === "all" || row.status === statusFilter) &&
        (row.name ?? "").toLowerCase().includes(needle),
    );
  }, [rows, search, statusFilter]);

  const columns: Array<DataTableColumn<ComputeAllocation>> = [
    {
      key: "name",
      header: "Allocation",
      sortable: true,
      sortValue: (row) => row.name,
      width: "14rem",
      cell: (row) => (
        <div className="flex flex-col gap-0.5">
          <Link
            href={`/allocations/${row.id}`}
            className="font-medium text-foreground hover:underline"
          >
            {row.name}
          </Link>
          <span className="hidden font-mono text-xs text-muted-foreground">{row.id}</span>
        </div>
      ),
    },
    {
      key: "project",
      header: "Project",
      width: "12rem",
      cell: (row) => <ProjectLink id={row.project_id ?? ""} />,
    },
    {
      key: "cluster",
      header: "Cluster",
      sortable: true,
      sortValue: (row) => clusterName(row.compute_cluster_id),
      width: "12rem",
      cell: (row) => (
        <span className="break-all text-sm">
          {clusterName(row.compute_cluster_id)}
        </span>
      ),
    },
    {
      key: "initial",
      header: "Initial SUs",
      sortable: true,
      sortValue: (row) => row.initial_su_amount,
      cell: (row) => <span className="tabular-nums">{formatNumber(row.initial_su_amount)}</span>,
    },
    {
      key: "endDate",
      header: "End date",
      sortable: true,
      sortValue: (row) => new Date(row.end_time ?? ""),
      cell: (row) => (
        <span className="text-sm text-muted-foreground">{formatDate(row.end_time)}</span>
      ),
    },
    {
      key: "status",
      header: "Status",
      sortable: true,
      sortValue: (row) => row.status,
      cell: (row) => (
        <StatusBadge
          variant={statusBadgeVariantFromAllocationStatus(row.status)}
          label={row.status}
        />
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-display text-[28px] font-bold leading-tight">Allocations</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Compute resources granted to you across projects.
        </p>
      </div>

      <div className="flex flex-col gap-3 rounded-md border bg-card p-4 sm:flex-row sm:items-center">
        <Input
          type="search"
          placeholder="Search by allocation name"
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          aria-label="Search allocations"
          className="sm:w-72"
        />
        <Select
          value={statusFilter}
          onValueChange={(value) => value && onStatusFilterChange(value)}
        >
          <SelectTrigger aria-label="Filter by status" className="h-9 w-36 px-3">
            <SelectValue>{statusLabel}</SelectValue>
          </SelectTrigger>
          <SelectContent>
            {["all", ...STATUSES].map((s) => (
              <SelectItem key={s} value={s}>
                {statusLabel(s)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {isLoading ? (
        <TableSkeleton rows={6} columns={6} />
      ) : error ? (
        <ErrorState message={error.message} onRetry={onRetry} />
      ) : filtered.length === 0 ? (
        <EmptyState
          heading="No allocations yet"
          description="No allocations match the current filters."
        />
      ) : (
        <DataTable
          columns={columns}
          rows={filtered}
          rowKey={(row) => row.id ?? ""}
          pagination={pagination}
        />
      )}
    </div>
  );
}
