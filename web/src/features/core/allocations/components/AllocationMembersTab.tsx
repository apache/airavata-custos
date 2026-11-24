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

import { useUpdateProjectMember } from "@/features/core/projects/queries";
import { UserDetailsDrawer } from "@/features/core/users/components/UserDetailsDrawer";
import type { ComputeAllocation } from "@/generated/core/zod.gen";
import { requires } from "@/shared/casl/permissions";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { ErrorState } from "@/shared/ui/ErrorState";
import { GuardedButton } from "@/shared/ui/GuardedButton";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import * as React from "react";
import { toast } from "sonner";
import { useAddMember, useAllocationMembers, useRemoveMember } from "../queries";
import type {
  AllocationMembership,
  AllocationMembershipRole,
  CreateMembershipPayload,
} from "../schemas";
import { AddMemberDialog, type Role } from "./AddMemberDialog";
import { MemberEditDialog, type MemberEditPayload } from "./MemberEditDialog";

export type AllocationMembersTabProps = {
  allocation: ComputeAllocation;
  canEditAccess: boolean;
  canEditRoles: boolean;
  canSearchUsers: boolean;
};

const ROLE_LABELS: Record<AllocationMembershipRole, string> = {
  PI: "PI",
  CO_PI: "Co-PI",
  ALLOCATION_MANAGER: "Allocation Manager",
  MEMBER: "Member",
};

function initialsFrom(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length >= 2) return `${parts[0]?.[0] ?? ""}${parts[1]?.[0] ?? ""}`.toUpperCase();
  return name.slice(0, 2).toUpperCase();
}

export function AllocationMembersTab({
  allocation,
  canEditAccess,
  canEditRoles,
  canSearchUsers,
}: AllocationMembersTabProps) {
  const query = useAllocationMembers(allocation.id);
  const addMutation = useAddMember(allocation.id);
  const roleMutation = useUpdateProjectMember(allocation.project_id);
  const removeMutation = useRemoveMember(allocation.id);
  const [editing, setEditing] = React.useState<AllocationMembership | null>(null);
  const [viewing, setViewing] = React.useState<string | null>(null);
  const [adding, setAdding] = React.useState(false);

  function handleAdd(payload: CreateMembershipPayload, role: Role) {
    addMutation.mutate(payload, {
      onSuccess: () => {
        setAdding(false);
        if (role !== "MEMBER") {
          roleMutation.mutate(
            { userId: payload.user_id, payload: { role } },
            { onError: (err) => toast.error(err.message) },
          );
        }
      },
      onError: (err) => toast.error(err.message),
    });
  }

  if (query.isLoading) return <TableSkeleton rows={4} columns={4} />;
  if (query.error) {
    return <ErrorState message={query.error.message} onRetry={() => query.refetch()} />;
  }
  const members = query.data ?? [];

  const headerCta = (
    <GuardedButton
      size="sm"
      allowed={canEditAccess && canSearchUsers}
      reason={requires("allocationsWrite", "usersRead")}
      onClick={() => setAdding(true)}
    >
      + Add member
    </GuardedButton>
  );

  const columns: Array<DataTableColumn<AllocationMembership>> = [
    {
      key: "member",
      header: "Member",
      cell: (row) => {
        const name = row.display_name ?? row.user_id;
        return (
          <div className="flex items-center gap-3">
            <Avatar className="size-8">
              <AvatarFallback>{initialsFrom(name)}</AvatarFallback>
            </Avatar>
            <div>
              <div className="font-medium text-foreground">{name}</div>
              {row.email ? (
                <div className="text-xs text-muted-foreground">{row.email}</div>
              ) : (
                <div className="font-mono text-xs text-muted-foreground">{row.user_id}</div>
              )}
            </div>
          </div>
        );
      },
    },
    {
      key: "role",
      header: "Role",
      width: "11rem",
      cell: (row) => <span className="text-sm">{ROLE_LABELS[row.role]}</span>,
    },
    {
      key: "status",
      header: "Status",
      width: "8rem",
      cell: (row) => (
        <StatusBadge
          variant={
            row.membership_status === "ACTIVE"
              ? "active"
              : row.membership_status === "INACTIVE"
                ? "inactive"
                : "deleted"
          }
          label={row.membership_status}
        />
      ),
    },
    {
      key: "actions",
      header: "",
      width: "11rem",
      align: "right",
      interactive: true,
      // PI and CO_PI are upstream-owned; the portal never edits or removes them.
      cell: (row) =>
        row.role !== "PI" && row.role !== "CO_PI" ? (
          <div className="flex justify-end gap-2">
            <GuardedButton
              variant="ghost"
              size="sm"
              allowed={canEditRoles}
              reason={requires("projectsWrite")}
              onClick={() => setEditing(row)}
              aria-label={`Edit ${row.display_name ?? row.user_id}`}
            >
              Edit
            </GuardedButton>
            <GuardedButton
              variant="destructive"
              size="sm"
              allowed={canEditAccess}
              reason={requires("allocationsWrite")}
              onClick={() =>
                removeMutation.mutate(row.id, { onError: (err) => toast.error(err.message) })
              }
              aria-label={`Remove ${row.display_name ?? row.user_id}`}
              disabled={removeMutation.isPending}
            >
              Remove
            </GuardedButton>
          </div>
        ) : null,
    },
  ];

  function handleEditSubmit(payload: MemberEditPayload) {
    if (!editing) return;
    roleMutation.mutate(
      { userId: editing.user_id, payload },
      { onSuccess: () => setEditing(null), onError: (err) => toast.error(err.message) },
    );
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-3">
        <h2 className="font-display text-lg font-semibold text-foreground">Memberships</h2>
        {headerCta}
      </div>
      <DataTable
        columns={columns}
        rows={members}
        rowKey={(row) => row.id}
        onRowClick={(row) => setViewing(row.user_id)}
        empty="No members yet."
      />
      <UserDetailsDrawer userId={viewing} onClose={() => setViewing(null)} />
      <MemberEditDialog
        member={editing}
        onClose={() => setEditing(null)}
        onSubmit={handleEditSubmit}
        isPending={roleMutation.isPending}
      />
      <AddMemberDialog
        open={adding}
        onOpenChange={setAdding}
        allocationId={allocation.id}
        defaultEndTime={allocation.end_time}
        onSubmit={handleAdd}
        isPending={addMutation.isPending}
        canSetRole={canEditRoles}
      />
    </div>
  );
}
