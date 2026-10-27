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

import type { ProjectResponse, ProjectStatus } from "@/generated/core/zod.gen";
import { requires } from "@/shared/casl/permissions";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { ErrorState } from "@/shared/ui/ErrorState";
import { GuardedButton } from "@/shared/ui/GuardedButton";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import Link from "next/link";
import { useRouter } from "next/navigation";
import * as React from "react";
import { ProjectStatusIcon } from "./ProjectStatusIcon";

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
}: ProjectsListProps) {
  const router = useRouter();
  const filtered = React.useMemo(() => {
    const needle = search.trim().toLowerCase();
    return rows.filter((row) => {
      if (statusFilter !== "all" && row.status !== statusFilter) return false;
      if (needle) {
        const hay = `${row.title} ${row.ref}`.toLowerCase();
        if (!hay.includes(needle)) return false;
      }
      return true;
    });
  }, [rows, search, statusFilter]);

  const columns: Array<DataTableColumn<ProjectResponse>> = [
    {
      key: "title",
      header: "Project",
      sortable: true,
      sortValue: (row) => row.title,
      cell: (row) => (
        <div className="flex items-center gap-2">
          <ProjectStatusIcon status={row.status} />
          <div className="flex flex-col items-start gap-0.5">
            <Link href={`/projects/${row.id}`} className="font-medium text-foreground">
              {row.title}
            </Link>
            <span className="text-xs text-muted-foreground">{row.ref}</span>
          </div>
        </div>
      ),
    },
    {
      key: "pi",
      header: "PI",
      sortable: true,
      sortValue: (row) => row.project_pi_display_name ?? row.project_pi_id,
      cell: (row) => (
        <div className="flex flex-col gap-0.5">
          <span className="text-sm text-muted-foreground">
            {row.project_pi_display_name ?? row.project_pi_id}
          </span>
          {row.project_pi_email ? (
            <span className="text-xs text-muted-foreground">{row.project_pi_email}</span>
          ) : null}
        </div>
      ),
    },
    {
      key: "origination",
      header: "Origination",
      width: "8rem",
      sortable: true,
      sortValue: (row) => row.origination,
      cell: (row) => <span className="text-sm">{row.origination}</span>,
    },
  ];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="font-display text-[28px] font-bold leading-tight">Projects</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Projects you own, manage, or belong to.
          </p>
        </div>
        {headerCta ? <div>{headerCta}</div> : null}
      </div>

      <div className="flex flex-col gap-3 rounded-md border bg-card p-4 sm:flex-row sm:items-center">
        <Input
          type="search"
          placeholder="Search by title or ID"
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          aria-label="Search projects"
          className="sm:w-64"
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
        <TableSkeleton rows={6} columns={4} />
      ) : error ? (
        <ErrorState message={error.message} onRetry={onRetry} />
      ) : (
        <DataTable
          columns={columns}
          rows={filtered}
          rowKey={(row) => row.id}
          onRowClick={(row) => router.push(`/projects/${row.id}`)}
          empty="No projects match the current filters."
        />
      )}
    </div>
  );
}

export function NewProjectCta({ canCreate, onClick }: { canCreate: boolean; onClick: () => void }) {
  return (
    <GuardedButton allowed={canCreate} reason={requires("projectsWrite")} onClick={onClick}>
      + New project
    </GuardedButton>
  );
}
