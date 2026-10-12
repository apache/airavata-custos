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

import {
  membershipStatusText,
  membershipStatusVariant,
} from "@/features/core/allocations/components/AllocationStatusLabel";
import type {
  AllocationStatus,
  ProjectMemberAllocationRef,
  ProjectMemberResponse,
} from "@/generated/core/types.gen";
import { enumLabel, pluralize } from "@/shared/format";
import { DataTable, type DataTableColumn, NameCell } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { QueryErrorState } from "@/shared/ui/ErrorState";
import { Field, FormDialog, SelectField, idText, text } from "@/shared/ui/FormDialog";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import { Button } from "@/shared/ui/button";
import { Link } from "@/shared/ui/link";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { emailBeside, initials } from "@/shared/users-admin/displayName";
import { idInput } from "@/shared/users-admin/idInput";
import * as React from "react";
import { toast } from "sonner";
import { useProjectMembers, useSetProjectRole } from "../queries";

// The spec types roles as plain strings; these maps cover the known values.
const CHIP_ROLE_SUFFIX: Record<string, string> = {
  PI: "P",
  CO_PI: "C",
  ALLOCATION_MANAGER: "A",
  MEMBER: "M",
};

const CHIP_ROLE_TINT: Record<string, string> = {
  PI: "bg-[color:var(--tone-accent-bg)] text-[color:var(--tone-accent-fg)]",
  CO_PI: "bg-[color:var(--tone-info-bg)] text-[color:var(--tone-info-fg)]",
  ALLOCATION_MANAGER: "bg-muted text-muted-foreground",
  MEMBER: "bg-muted text-muted-foreground",
};

function AllocationsCell({
  allocations,
  linked,
}: { allocations: ProjectMemberAllocationRef[]; linked: boolean }) {
  if (allocations.length === 0) return <span className="text-xs text-muted-foreground">—</span>;
  return (
    <div className="flex flex-wrap items-center gap-1">
      {allocations.map(({ id, name, role }) => {
        const chip = {
          className: `inline-flex h-6 max-w-[14ch] items-center gap-1 truncate rounded px-1.5 text-xs ${CHIP_ROLE_TINT[role] ?? ""}`,
          title: `${name} — ${roleLabel(role)}`,
          children: (
            <>
              <span className="truncate">{name}</span>
              <span className="font-mono opacity-70">{CHIP_ROLE_SUFFIX[role]}</span>
            </>
          ),
        };
        return linked ? (
          <Link key={id} href={`/allocations/${id}`} {...chip} />
        ) : (
          <span key={id} {...chip} />
        );
      })}
    </div>
  );
}

export type ProjectMembersTabProps = {
  projectId: string;
  canManage: boolean;
  // Chips link only for callers who can open any allocation, not just their own.
  canReadAllocations: boolean;
};

export const ROLE_LABELS: Record<string, string> = {
  PI: "PI",
  CO_PI: "Co-PI",
  ALLOCATION_MANAGER: "Allocation manager",
  MEMBER: "Member",
};

export const roleLabel = (role: string) => ROLE_LABELS[role] ?? enumLabel(role);

const ASSIGNABLE_ROLES = ["CO_PI", "ALLOCATION_MANAGER"];

export function ProjectMembersTab({
  projectId,
  canManage,
  canReadAllocations,
}: ProjectMembersTabProps) {
  const query = useProjectMembers(projectId);
  const setMutation = useSetProjectRole();
  // null assigns a role to a user by id; undefined keeps the dialog closed.
  const [editing, setEditing] = React.useState<ProjectMemberResponse | null>();
  const piId = query.data?.find((m) => m.role === "PI")?.user_id;
  const setRole = (userId: string, role: string, onSuccess?: () => void) =>
    userId === piId
      ? toast.error("The PI already holds every project role.")
      : toastOnSuccess(
          setMutation.mutateAsync({ path: { id: projectId, userId }, body: { role } }),
          role === "MEMBER" ? "Role removed" : "Role saved",
          onSuccess,
        );

  if (query.isLoading) return <TableSkeleton rows={4} columns={4} />;
  if (query.error) {
    return <QueryErrorState error={query.error} what="project" onRetry={() => query.refetch()} />;
  }
  const members = query.data ?? [];

  const columns: Array<DataTableColumn<ProjectMemberResponse>> = [
    {
      key: "member",
      header: "Member",
      cell: (row) => (
        <div className="flex items-center gap-3">
          <Avatar className="size-8">
            <AvatarFallback>{initials(row.display_name)}</AvatarFallback>
          </Avatar>
          <div className="min-w-0">
            <NameCell text={row.display_name} className="font-medium text-foreground" />
            <NameCell
              text={emailBeside(row.display_name, row.email)}
              className="text-xs text-muted-foreground"
            />
          </div>
        </div>
      ),
    },
    {
      key: "role",
      header: "Role",
      cell: (row) => <span className="text-sm">{roleLabel(row.role)}</span>,
    },
    {
      key: "allocations",
      header: "Allocations",
      sortValue: (row) => row.allocations?.length ?? 0,
      cell: (row) => (
        <AllocationsCell allocations={row.allocations ?? []} linked={canReadAllocations} />
      ),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => {
        if (!row.status) return <span className="text-muted-foreground">—</span>;
        const status = row.status as AllocationStatus;
        return (
          <StatusBadge
            variant={membershipStatusVariant(status)}
            label={membershipStatusText(status)}
          />
        );
      },
    },
    {
      key: "actions",
      header: "",
      align: "right",
      interactive: true,
      cell: (row) =>
        !canManage || row.role === "PI" ? null : row.role === "MEMBER" ? (
          <div className="flex justify-end">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setEditing(row)}
              aria-label={`Assign a role to ${row.display_name}`}
            >
              Assign role
            </Button>
          </div>
        ) : (
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
              onClick={() =>
                confirmToast(`Remove ${row.display_name}'s role on this project?`, "Remove", () =>
                  setRole(row.user_id, "MEMBER"),
                )
              }
              aria-label={`Remove role of ${row.display_name}`}
              pending={setMutation.isPending && setMutation.variables.path.userId === row.user_id}
            >
              Remove role
            </Button>
          </div>
        ),
    },
  ];

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          {members.length} {pluralize("member", members.length)}
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
          description="People appear here once they join one of this project's allocations or are assigned a project role."
        />
      ) : (
        <DataTable
          caption="Members"
          columns={
            canManage && members.some((m) => m.role !== "PI")
              ? columns
              : columns.filter((c) => c.key !== "actions")
          }
          rows={members}
          rowKey={(row) => row.user_id}
        />
      )}
      <RoleDialog
        open={editing !== undefined}
        member={editing ?? null}
        members={query}
        onClose={() => setEditing(undefined)}
        onSubmit={(userId, role) => setRole(userId, role, () => setEditing(undefined))}
        isPending={setMutation.isPending}
      />
    </div>
  );
}

function RoleDialog({
  open,
  member: opened,
  members,
  onClose,
  onSubmit,
  isPending,
}: {
  open: boolean;
  member: ProjectMemberResponse | null;
  members: ReturnType<typeof useProjectMembers>;
  onClose: () => void;
  onSubmit: (userId: string, role: string) => void;
  isPending: boolean;
}) {
  // A member without a role is assigned one; a role holder's role is edited while they hold it.
  const assigning = !opened || opened.role === "MEMBER";
  const member = opened
    ? members.data?.find(
        (m) => m.user_id === opened.user_id && (assigning || ASSIGNABLE_ROLES.includes(m.role)),
      )
    : undefined;
  return (
    <FormDialog
      open={open}
      onOpenChange={(next) => (next ? null : onClose())}
      title={assigning ? "Assign role" : "Edit role"}
      description={opened?.display_name ?? "Give a user a project role."}
      submitLabel="Save"
      isPending={isPending}
      record={opened ? members : undefined}
      gone={
        assigning
          ? `${opened?.display_name} is no longer a member of this project`
          : `${opened?.display_name} no longer holds a role on this project`
      }
      missing={Boolean(opened) && !member}
      onSubmit={(form) => {
        const role = text(form, "role");
        if (!opened) return onSubmit(idText(form, "user"), role);
        if (member && member.role !== role) onSubmit(member.user_id, role);
        else onClose();
      }}
    >
      {opened ? null : <Field label="User ID" name="user" required {...idInput("user")} />}
      <SelectField
        label="Role"
        name="role"
        defaultValue={assigning ? "" : member?.role}
        // The empty choice fails the select's required check, so Save never assigns a default role.
        options={[
          ...(assigning ? [{ value: "", label: "Choose a role" }] : []),
          ...ASSIGNABLE_ROLES.map((r) => ({ value: r, label: roleLabel(r) })),
        ]}
      />
    </FormDialog>
  );
}
