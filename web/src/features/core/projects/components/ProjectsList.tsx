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

import type { ProjectResponse, ProjectStatus } from "@/generated/core/types.gen";
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
import { StatusBadge, statusBadgeVariantFromAllocationStatus } from "@/shared/ui/StatusBadge";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import type * as React from "react";
import { originationLabel } from "../format";

export type ProjectsListProps = {
  rows: ProjectResponse[];
  isLoading: boolean;
  error: Error | null;
  onRetry?: () => void;
  search: string;
  onSearchChange: (next: string) => void;
  statusFilter: ProjectStatus | "all";
  onStatusFilterChange: (next: ProjectStatus | "all") => void;
  headerCta?: React.ReactNode;
  // Project readers list every project, not only their own.
  canReadAll?: boolean;
  pagination?: DataTablePagination;
};

export function ProjectsList({
  rows,
  isLoading,
  error,
  onRetry,
  search,
  onSearchChange,
  statusFilter,
  onStatusFilterChange,
  headerCta,
  canReadAll,
  pagination,
}: ProjectsListProps) {
  const scope = canReadAll ? "All projects" : "Projects you own, manage, or belong to";
  const columns: Array<DataTableColumn<ProjectResponse>> = [
    {
      key: "title",
      header: "Project",
      sortable: true,
      sortValue: (row) => row.title,
      cell: (row) => (
        <div className="flex flex-col gap-0.5">
          <NameCell
            href={`/projects/${row.id}`}
            text={row.title}
            className="font-medium text-foreground hover:underline"
          />
          <NameCell text={row.originated_id} className="font-mono text-xs text-muted-foreground" />
        </div>
      ),
    },
    {
      key: "pi",
      header: "PI",
      sortable: true,
      sortValue: (row) => row.project_pi_display_name ?? row.project_pi_id,
      cell: (row) => (
        <NameCell
          text={row.project_pi_display_name ?? row.project_pi_id}
          className="text-sm text-muted-foreground"
        />
      ),
    },
    {
      key: "origination",
      header: "Origination",
      sortable: true,
      sortValue: (row) => row.origination,
      cell: (row) => <span className="text-sm">{originationLabel(row.origination)}</span>,
    },
    {
      key: "status",
      header: "Status",
      sortable: true,
      sortValue: (row) => row.status,
      cell: (row) => (
        <StatusBadge
          variant={statusBadgeVariantFromAllocationStatus(row.status)}
          label={statusLabel(row.status)}
        />
      ),
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="font-display text-[28px] font-bold leading-tight">
            {ROUTE_LABELS["/projects"]}
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">{scope}.</p>
        </div>
        {headerCta ? <div>{headerCta}</div> : null}
      </div>

      <div className="flex flex-col gap-3 rounded-md border bg-card p-4 sm:flex-row sm:items-center">
        <Input
          type="search"
          placeholder="Search by title or originated ID"
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          aria-label="Search projects"
          className="sm:w-64"
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
        <ErrorState error={error} onRetry={onRetry} />
      ) : rows.length === 0 && pagination && pagination.page > 1 ? (
        <PastLastPage
          things="projects"
          page={pagination.page}
          onPageChange={pagination.onPageChange}
        />
      ) : rows.length === 0 ? (
        search.trim() || statusFilter !== "all" ? (
          <EmptyState
            heading="No matching projects"
            description="No projects match the current filters."
          />
        ) : (
          <EmptyState heading="No projects yet" description={`${scope} appear here.`} />
        )
      ) : (
        <DataTable
          caption="Projects"
          columns={columns}
          rows={rows}
          rowKey={(row) => row.id}
          pagination={pagination}
        />
      )}
    </div>
  );
}
