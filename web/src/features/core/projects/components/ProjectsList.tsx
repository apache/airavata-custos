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
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { GuardedButton } from "@/shared/ui/GuardedButton";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import * as React from "react";
import { ProjectCard } from "./ProjectCard";

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
  origination: string | null;
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
  origination,
  onStatusFilterChange,
  headerCta,
}: ProjectsListProps) {
  const filtered = React.useMemo(() => {
    const needle = search.trim().toLowerCase();
    return rows.filter((row) => {
      if (statusFilter !== "all" && row.status !== statusFilter) return false;
      if (origination && row.origination.toLowerCase() !== origination.toLowerCase()) return false;
      if (needle) {
        const hay = `${row.title} ${row.ref}`.toLowerCase();
        if (!hay.includes(needle)) return false;
      }
      return true;
    });
  }, [rows, search, statusFilter, origination]);

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="font-display text-[28px] font-bold leading-tight">
            {origination
              ? `${origination[0]?.toUpperCase()}${origination.slice(1)} projects`
              : "Projects"}
          </h1>
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
      ) : filtered.length === 0 ? (
        <EmptyState heading="No projects match the current filters." />
      ) : (
        <div className="space-y-4">
          {filtered.map((project) => (
            <ProjectCard key={project.id} project={project} />
          ))}
        </div>
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
