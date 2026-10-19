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

import { Server } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import * as React from "react";
import { ClusterName } from "@/features/core/clusters/components/ClusterName";
import { ProjectGroupHeader } from "@/features/core/projects/components/ProjectGroupHeader";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { StatusIcon } from "@/shared/ui/StatusIcon";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import type { AllocationStatus, ComputeAllocation } from "@/generated/core/zod.gen";
import { useAllocationUsedSU } from "../queries";

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  } catch {
    return iso;
  }
}

function statusLabelFor(value: string): string {
  switch (value) {
    case "ACTIVE":
      return "Active";
    case "INACTIVE":
      return "Inactive";
    case "DELETED":
      return "Deleted";
    default:
      return "All statuses";
  }
}

function formatSU(n: number): string {
  return new Intl.NumberFormat().format(n);
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
  showCluster: boolean;
  headerCta?: React.ReactNode;
};

function RemainingSU({ allocation }: { allocation: ComputeAllocation }) {
  const used = useAllocationUsedSU(allocation.id).data;
  return (
    <span className="tabular-nums" title={`of ${formatSU(allocation.initial_su_amount)}`}>
      {used === undefined ? "…" : formatSU(allocation.initial_su_amount - used)}
    </span>
  );
}

export function AllocationsList({
  rows,
  isLoading,
  error,
  onRetry,
  search,
  onSearchChange,
  statusFilter,
  onStatusFilterChange,
  showCluster,
  headerCta,
}: AllocationsListProps) {
  const router = useRouter();
  const filtered = React.useMemo(() => {
    const needle = search.trim().toLowerCase();
    return rows.filter((row) => {
      if (statusFilter !== "all" && row.status !== statusFilter) return false;
      if (needle) {
        const hay = `${row.name} ${row.project_id} ${row.id}`.toLowerCase();
        if (!hay.includes(needle)) return false;
      }
      return true;
    });
  }, [rows, search, statusFilter]);

  const columns: Array<DataTableColumn<ComputeAllocation>> = [
    {
      key: "name",
      header: "Allocation",
      sortable: true,
      sortValue: (row) => row.name,
      cell: (row) => (
        <span className="-ml-7 inline-flex items-center gap-3 whitespace-nowrap">
          <StatusIcon
            icon={Server}
            active={row.status === "ACTIVE"}
            label={statusLabelFor(row.status)}
          />
          <Link href={`/allocations/${row.id}`} className="font-medium text-foreground">
            {row.name}
          </Link>
        </span>
      ),
    },
    {
      key: "cluster",
      header: "Cluster",
      width: "12rem",
      cell: (row) => <ClusterName id={row.compute_cluster_id} />,
    },
    {
      key: "remaining",
      header: "Remaining SUs",
      width: "9rem",
      cell: (row) => <RemainingSU allocation={row} />,
    },
    {
      key: "endDate",
      header: "End date",
      width: "9rem",
      sortable: true,
      sortValue: (row) => new Date(row.end_time),
      cell: (row) => (
        <span className="text-sm text-muted-foreground">{formatDate(row.end_time)}</span>
      ),
    },
  ];
  const visibleColumns = columns.filter((col) => showCluster || col.key !== "cluster");

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="font-display text-[28px] font-bold leading-tight">Allocations</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Compute resources granted to you across projects.
          </p>
        </div>
        {headerCta ? <div>{headerCta}</div> : null}
      </div>

      <div className="flex flex-col gap-3 rounded-md border bg-card p-4 sm:flex-row sm:items-center">
        <Input
          type="search"
          placeholder="Search by allocation name, project, or ID"
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          aria-label="Search allocations"
          className="sm:w-72"
        />
        <Select
          value={statusFilter}
          onValueChange={(value) => value !== null && onStatusFilterChange(value)}
        >
          <SelectTrigger aria-label="Filter by status" className="h-9 w-36 px-3">
            <SelectValue>{(value: string) => statusLabelFor(value)}</SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All statuses</SelectItem>
            <SelectItem value="ACTIVE">Active</SelectItem>
            <SelectItem value="INACTIVE">Inactive</SelectItem>
            <SelectItem value="DELETED">Deleted</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {isLoading ? (
        <TableSkeleton rows={6} columns={5} />
      ) : error ? (
        <ErrorState message={error.message} onRetry={onRetry} />
      ) : filtered.length === 0 ? (
        <DataTable
          columns={visibleColumns}
          rows={[]}
          rowKey={(row) => row.id}
          empty="No allocations match the current filters."
        />
      ) : (
        <div className="space-y-3">
          {[...Map.groupBy(filtered, (row) => row.project_id)].map(([projectId, rows]) => (
            <section key={projectId} className="overflow-hidden rounded-lg border border-foreground/20 bg-card">
              <div className="border-b border-border bg-muted px-4 py-2.5">
                <ProjectGroupHeader id={projectId} />
              </div>
              <DataTable
                className="rounded-none border-0 [&_th]:py-1.5 [&_thead]:bg-muted/40 [&_td:first-child]:pl-11 [&_th:first-child]:pl-11 [&_td]:align-middle [&_td]:whitespace-nowrap"
                columns={visibleColumns}
                rows={rows}
                rowKey={(row) => row.id}
                onRowClick={(row) => router.push(`/allocations/${row.id}`)}
              />
            </section>
          ))}
        </div>
      )}
    </div>
  );
}
