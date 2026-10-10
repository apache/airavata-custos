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

import type { ProjectMemberAllocationRef, ProjectMemberResponse } from "@/generated/core/types.gen";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import { Button } from "@/shared/ui/button";
import { Field, FormDialog, SelectField, text } from "@/shared/ui/FormDialog";
import Link from "next/link";
import * as React from "react";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { useProjectMembers, useSetProjectRole } from "../queries";

// The spec types roles as plain strings; these maps cover the known values.
const CHIP_ROLE_SUFFIX: Record<string, string> = {
  PI: "P",
  CO_PI: "C",
  ALLOCATION_MANAGER: "M",
  MEMBER: "M",
};

const CHIP_ROLE_TINT: Record<string, string> = {
  PI: "bg-indigo-50 text-indigo-700",
  CO_PI: "bg-sky-50 text-sky-700",
  ALLOCATION_MANAGER: "bg-slate-100 text-slate-600",
  MEMBER: "bg-slate-100 text-slate-600",
};

function AllocationsCell({ allocations }: { allocations: ProjectMemberAllocationRef[] }) {
  if (allocations.length === 0) return <span className="text-xs text-muted-foreground">—</span>;
  return (
    <div className="flex flex-wrap items-center gap-1">
      {allocations.map((a) => (
        <Link
          key={a.id}
          href={`/allocations/${a.id}`}
          className={`inline-flex h-6 max-w-[14ch] items-center gap-1 truncate rounded px-1.5 text-xs ${CHIP_ROLE_TINT[a.role] ?? ""}`}
          title={`${a.name} — ${a.role.replace("_", "-")}`}
        >
          <span className="truncate">{a.name}</span>
          <span className="font-mono opacity-70">{CHIP_ROLE_SUFFIX[a.role]}</span>
        </Link>
      ))}
    </div>
  );
}

export type ProjectMembersTabProps = {
  projectId: string;
  canManage: boolean;
};

export function initialsFrom(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length >= 2) return `${parts[0]?.[0] ?? ""}${parts[1]?.[0] ?? ""}`.toUpperCase();
  return name.slice(0, 2).toUpperCase();
}

export const ROLE_LABELS: Record<string, string> = {
  PI: "PI",
  CO_PI: "Co-PI",
  ALLOCATION_MANAGER: "Allocation Manager",
  MEMBER: "Member",
};

const ASSIGNABLE_ROLES = ["CO_PI", "ALLOCATION_MANAGER"];

export function ProjectMembersTab({ projectId, canManage }: ProjectMembersTabProps) {
  const query = useProjectMembers(projectId);
  const setMutation = useSetProjectRole();
  // null assigns a role to a user by id; undefined keeps the dialog closed.
  const [editing, setEditing] = React.useState<ProjectMemberResponse | null>();
  const setRole = (userId: string, role: string, onSuccess?: () => void) =>
    setMutation.mutate(
      { path: { id: projectId, userId }, body: { role } },
      toastOnSuccess(role === "MEMBER" ? "Role removed" : "Role saved", onSuccess),
    );

  if (query.isLoading) return <TableSkeleton rows={4} columns={4} />;
  if (query.error) {
    return <ErrorState message={query.error.message} onRetry={() => query.refetch()} />;
  }
  const members = query.data ?? [];

  const columns: Array<DataTableColumn<ProjectMemberResponse>> = [
    {
      key: "member",
      header: "Member",
      cell: (row) => (
        <div className="flex items-center gap-3">
          <Avatar className="size-8">
            <AvatarFallback>{initialsFrom(row.display_name)}</AvatarFallback>
          </Avatar>
          <div>
            <div className="font-medium text-foreground">{row.display_name}</div>
            <div className="text-xs text-muted-foreground">{row.email}</div>
          </div>
        </div>
      ),
    },
    {
      key: "role",
      header: "Role",
      cell: (row) => <span className="text-sm">{ROLE_LABELS[row.role] ?? row.role}</span>,
    },
    {
      key: "allocations",
      header: "Allocations",
      sortValue: (row) => row.allocations.length,
      cell: (row) => <AllocationsCell allocations={row.allocations} />,
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => (
        <StatusBadge
          variant={
            row.status === "ACTIVE" ? "active" : row.status === "INACTIVE" ? "inactive" : "deleted"
          }
          label={row.status}
        />
      ),
    },
    {
      key: "actions",
      header: "",
      align: "right",
      interactive: true,
      cell: (row) =>
        canManage && ASSIGNABLE_ROLES.includes(row.role) ? (
          <div className="flex justify-end gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setEditing(row)}
              aria-label={`Edit role of ${row.display_name}`}
            >
              Edit role
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => setRole(row.user_id, "MEMBER")}
              aria-label={`Remove role of ${row.display_name}`}
              disabled={setMutation.isPending}
            >
              Remove role
            </Button>
          </div>
        ) : null,
    },
  ];

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          {members.length} member{members.length === 1 ? "" : "s"}
        </p>
        {canManage ? (
          <Button size="sm" onClick={() => setEditing(null)}>
            + Assign role
          </Button>
        ) : null}
      </div>
      {members.length === 0 ? (
        <EmptyState
          heading="No members yet"
          description="Add members to grant access to this project's allocations."
        />
      ) : (
        <DataTable columns={columns} rows={members} rowKey={(row) => row.user_id} />
      )}
      <RoleDialog
        open={editing !== undefined}
        member={editing ?? null}
        onClose={() => setEditing(undefined)}
        onSubmit={(userId, role) => setRole(userId, role, () => setEditing(undefined))}
        isPending={setMutation.isPending}
      />
    </div>
  );
}

function RoleDialog({
  open,
  member,
  onClose,
  onSubmit,
  isPending,
}: {
  open: boolean;
  member: ProjectMemberResponse | null;
  onClose: () => void;
  onSubmit: (userId: string, role: string) => void;
  isPending: boolean;
}) {
  return (
    <FormDialog
      open={open}
      onOpenChange={(next) => (next ? null : onClose())}
      title={member ? "Edit role" : "Assign role"}
      description={member?.display_name ?? "Give a user a project role."}
      submitLabel="Save"
      isPending={isPending}
      onSubmit={(form) => onSubmit(member?.user_id ?? text(form, "user"), text(form, "role"))}
    >
      {member ? null : <Field label="User ID" name="user" required />}
      <SelectField
        label="Role"
        name="role"
        defaultValue={member?.role}
        options={ASSIGNABLE_ROLES.map((r) => ({ value: r, label: ROLE_LABELS[r] ?? r }))}
      />
    </FormDialog>
  );
}
