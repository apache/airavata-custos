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

import type { AllocationStatus, ComputeAllocation } from "@/generated/core/types.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { formatSU } from "@/shared/format";
import { ROUTE_LABELS } from "@/shared/layout/nav";
import {
  DataTable,
  type DataTableColumn,
  type DataTablePagination,
  NameCell,
  PastLastPage,
} from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { STATUSES, statusLabel } from "@/shared/ui/FormDialog";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { AllocationStatusLabel, knownDay } from "./AllocationStatusLabel";
import { CreateAllocationDialog } from "./CreateAllocationDialog";

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
  // Allocation read lists every allocation; without it the backend scopes the list to the caller.
  const ability = useAbility();
  const readsAll = ability.can("read", "Allocation");
  const columns: Array<DataTableColumn<ComputeAllocation>> = [
    {
      key: "name",
      header: "Allocation",
      sortable: true,
      sortValue: (row) => row.name,
      width: "14rem",
      cell: (row) => (
        <div className="flex flex-col gap-0.5">
          <NameCell
            href={`/allocations/${row.id}`}
            text={row.name}
            className="font-medium text-foreground hover:underline"
          />
          <span className="hidden font-mono text-xs text-muted-foreground">{row.id}</span>
        </div>
      ),
    },
    {
      key: "initial",
      header: "Initial SUs",
      sortable: true,
      sortValue: (row) => row.initial_su_amount,
      cell: (row) => <span className="tabular-nums">{formatSU(row.initial_su_amount)}</span>,
    },
    {
      key: "endDate",
      header: "End date",
      sortable: true,
      sortValue: (row) => new Date(row.end_time),
      cell: (row) => (
        <span className="text-sm text-muted-foreground">{knownDay(row.end_time)}</span>
      ),
    },
    {
      key: "status",
      header: "Status",
      sortable: true,
      sortValue: (row) => row.status,
      cell: (row) => (
        <AllocationStatusLabel
          status={row.status}
          startTime={row.start_time}
          endTime={row.end_time}
        />
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="font-display text-[28px] font-bold leading-tight">
            {ROUTE_LABELS["/allocations"]}
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {readsAll
              ? "All compute allocations."
              : "Compute resources granted to you across projects."}
          </p>
        </div>
        {ability.can("write", "Allocation") ? <CreateAllocationDialog /> : null}
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
        <TableSkeleton rows={6} columns={4} />
      ) : error ? (
        <ErrorState onRetry={onRetry} />
      ) : rows.length === 0 && pagination && pagination.page > 1 ? (
        <PastLastPage
          things="allocations"
          page={pagination.page}
          onPageChange={pagination.onPageChange}
        />
      ) : rows.length === 0 ? (
        search.trim() || statusFilter !== "all" ? (
          <EmptyState
            heading="No matching allocations"
            description="No allocations match the current filters."
          />
        ) : (
          <EmptyState
            heading="No allocations yet"
            description={
              readsAll
                ? "Allocations appear here once they are created."
                : "Allocations you are a member of appear here."
            }
          />
        )
      ) : (
        <DataTable
          caption="Allocations"
          columns={columns}
          rows={rows}
          rowKey={(row) => row.id}
          pagination={pagination}
        />
      )}
    </div>
  );
}
