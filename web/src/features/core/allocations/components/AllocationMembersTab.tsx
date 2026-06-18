"use client";

import * as React from "react";
import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import {
  useAddMember,
  useAllocationMembers,
  useRemoveMember,
  useUpdateMember,
} from "../queries";
import type {
  AllocationMembership,
  ComputeAllocation,
  CreateMembershipPayload,
} from "../schemas";
import { AddMemberDialog } from "./AddMemberDialog";
import { MemberEditDialog, type MemberEditPayload } from "./MemberEditDialog";

export type AllocationMembersTabProps = {
  allocation: ComputeAllocation;
  canManage: boolean;
};

const ROLE_LABELS: Record<string, string> = {
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

export function AllocationMembersTab({ allocation, canManage }: AllocationMembersTabProps) {
  const query = useAllocationMembers(allocation.id);
  const addMutation = useAddMember(allocation.id);
  const updateMutation = useUpdateMember(allocation.id);
  const removeMutation = useRemoveMember(allocation.id);
  const [editing, setEditing] = React.useState<AllocationMembership | null>(null);
  const [adding, setAdding] = React.useState(false);

  if (query.isLoading) return <TableSkeleton rows={4} columns={4} />;
  if (query.error) {
    return <ErrorState message={(query.error as Error).message} onRetry={() => query.refetch()} />;
  }
  const members = query.data ?? [];

  const headerCta = canManage ? (
    <Button size="sm" onClick={() => setAdding(true)}>
      + Add member
    </Button>
  ) : null;

  if (members.length === 0) {
    return (
      <div className="space-y-3">
        <div className="flex justify-end">{headerCta}</div>
        <EmptyState
          heading="No members yet"
          description="Add a user to grant them access to this allocation."
        />
        <AddMemberDialog
          open={adding}
          onOpenChange={setAdding}
          allocationId={allocation.id}
          defaultEndTime={allocation.end_time}
          onSubmit={(payload: CreateMembershipPayload) =>
            addMutation.mutate(payload, { onSuccess: () => setAdding(false) })
          }
          isPending={addMutation.isPending}
        />
      </div>
    );
  }

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
      cell: (row) => (
        <span className="text-sm">{row.role ? ROLE_LABELS[row.role] : "—"}</span>
      ),
    },
    {
      key: "status",
      header: "Status",
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
      align: "right",
      interactive: true,
      cell: (row) =>
        canManage ? (
          <div className="flex justify-end gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setEditing(row)}
              aria-label={`Edit ${row.display_name ?? row.user_id}`}
            >
              Edit
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => removeMutation.mutate(row.id)}
              aria-label={`Remove ${row.display_name ?? row.user_id}`}
              disabled={removeMutation.isPending}
            >
              Remove
            </Button>
          </div>
        ) : null,
    },
  ];

  function handleEditSubmit(payload: MemberEditPayload) {
    if (!editing) return;
    updateMutation.mutate(
      { id: editing.id, patch: payload },
      { onSuccess: () => setEditing(null) },
    );
  }

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          {members.length} member{members.length === 1 ? "" : "s"}
        </p>
        {headerCta}
      </div>
      <DataTable columns={columns} rows={members} rowKey={(row) => row.id} />
      <MemberEditDialog
        member={editing}
        onClose={() => setEditing(null)}
        onSubmit={handleEditSubmit}
        isPending={updateMutation.isPending}
      />
      <AddMemberDialog
        open={adding}
        onOpenChange={setAdding}
        allocationId={allocation.id}
        defaultEndTime={allocation.end_time}
        onSubmit={(payload: CreateMembershipPayload) =>
          addMutation.mutate(payload, { onSuccess: () => setAdding(false) })
        }
        isPending={addMutation.isPending}
      />
    </div>
  );
}
