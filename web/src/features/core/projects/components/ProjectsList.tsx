"use client";

import Link from "next/link";
import * as React from "react";
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import { TableSkeleton } from "@/shared/ui/Loading";
import {
  StatusBadge,
  statusBadgeVariantFromAllocationStatus,
} from "@/shared/ui/StatusBadge";
import type { Project, ProjectStatus } from "../schemas";

export type ProjectsListProps = {
  rows: Project[];
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
  const filtered = React.useMemo(() => {
    const needle = search.trim().toLowerCase();
    return rows.filter((row) => {
      if (statusFilter !== "all" && row.status !== statusFilter) return false;
      if (needle) {
        const hay = `${row.title} ${row.originated_id}`.toLowerCase();
        if (!hay.includes(needle)) return false;
      }
      return true;
    });
  }, [rows, search, statusFilter]);

  const columns: Array<DataTableColumn<Project>> = [
    {
      key: "title",
      header: "Project",
      sortable: true,
      sortValue: (row) => row.title,
      cell: (row) => (
        <div className="flex flex-col gap-0.5">
          <Link
            href={`/projects/${row.id}`}
            className="font-medium text-foreground hover:underline"
          >
            {row.title}
          </Link>
          <span className="font-mono text-xs text-muted-foreground">{row.originated_id}</span>
        </div>
      ),
    },
    {
      key: "pi",
      header: "PI",
      sortable: true,
      sortValue: (row) => row.project_pi_display_name ?? row.project_pi_id,
      cell: (row) => (
        <span className="text-sm text-muted-foreground">
          {row.project_pi_display_name ?? row.project_pi_id}
        </span>
      ),
    },
    {
      key: "origination",
      header: "Origination",
      sortable: true,
      sortValue: (row) => row.origination,
      cell: (row) => <span className="text-sm">{row.origination}</span>,
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
        <select
          value={statusFilter}
          onChange={(e) => onStatusFilterChange(e.target.value as ProjectStatus | "all")}
          aria-label="Filter by status"
          className="h-9 rounded-md border bg-background px-3 text-sm"
        >
          <option value="all">All statuses</option>
          <option value="ACTIVE">Active</option>
          <option value="INACTIVE">Inactive</option>
          <option value="DELETED">Deleted</option>
        </select>
      </div>

      {isLoading ? (
        <TableSkeleton rows={6} columns={4} />
      ) : error ? (
        <ErrorState message={error.message} onRetry={onRetry} />
      ) : filtered.length === 0 ? (
        <EmptyState
          heading="No projects yet"
          description="No projects match the current filters."
        />
      ) : (
        <DataTable columns={columns} rows={filtered} rowKey={(row) => row.id} />
      )}
    </div>
  );
}

export function NewProjectCta({ canCreate = true }: { canCreate?: boolean }) {
  if (!canCreate) return null;
  return (
    <Button render={<Link href="/projects?new=1" aria-label="New project" />}>
      + New project
    </Button>
  );
}
