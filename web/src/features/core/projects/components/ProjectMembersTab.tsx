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

import Link from "next/link";
import * as React from "react";
import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Label } from "@/shared/ui/label";
import { TableSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import {
  useProjectMembers,
  useRemoveProjectMember,
  useUpdateProjectMember,
} from "../queries";
import type { ProjectMember, ProjectMemberAllocation, ProjectMemberRole } from "../schemas";

const CHIP_ROLE_SUFFIX: Record<ProjectMemberAllocation["role"], string> = {
  PI: "P",
  CO_PI: "C",
  ALLOCATION_MANAGER: "M",
  MEMBER: "M",
};

const CHIP_ROLE_TINT: Record<ProjectMemberAllocation["role"], string> = {
  PI: "bg-indigo-50 text-indigo-700",
  CO_PI: "bg-sky-50 text-sky-700",
  ALLOCATION_MANAGER: "bg-slate-100 text-slate-600",
  MEMBER: "bg-slate-100 text-slate-600",
};

function AllocationsCell({ allocations }: { allocations: ProjectMemberAllocation[] }) {
  if (allocations.length === 0) return <span className="text-xs text-muted-foreground">—</span>;
  return (
    <div className="flex flex-wrap items-center gap-1">
      {allocations.map((a) => (
        <Link
          key={a.id}
          href={`/allocations/${a.id}`}
          className={`inline-flex h-6 max-w-[14ch] items-center gap-1 truncate rounded px-1.5 text-xs ${CHIP_ROLE_TINT[a.role]}`}
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

function initialsFrom(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length >= 2) return `${parts[0]?.[0] ?? ""}${parts[1]?.[0] ?? ""}`.toUpperCase();
  return name.slice(0, 2).toUpperCase();
}

const ROLE_LABELS: Record<ProjectMemberRole, string> = {
  PI: "PI",
  CO_PI: "Co-PI",
  ALLOCATION_MANAGER: "Allocation Manager",
  MEMBER: "Member",
};

export function ProjectMembersTab({ projectId, canManage }: ProjectMembersTabProps) {
  const query = useProjectMembers(projectId);
  const updateMutation = useUpdateProjectMember(projectId);
  const removeMutation = useRemoveProjectMember(projectId);
  const [editing, setEditing] = React.useState<ProjectMember | null>(null);

  if (query.isLoading) return <TableSkeleton rows={4} columns={4} />;
  if (query.error) {
    return <ErrorState message={(query.error as Error).message} onRetry={() => query.refetch()} />;
  }
  const members = query.data ?? [];
  if (members.length === 0) {
    return (
      <EmptyState
        heading="No members yet"
        description="Add members to grant access to this project's allocations."
      />
    );
  }

  const columns: Array<DataTableColumn<ProjectMember>> = [
    {
      key: "member",
      header: "Member",
      cell: (row) => (
        <div className="flex items-center gap-3">
          <Avatar className="size-8">
            <AvatarFallback>{initialsFrom(row.display_name)}</AvatarFallback>
          </Avatar>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-medium text-foreground">{row.display_name}</span>
              {row.type === "VIRTUAL" && (
                <Badge variant="outline" className="text-[10px] uppercase tracking-wide">
                  Virtual
                </Badge>
              )}
            </div>
            <div className="text-xs text-muted-foreground">{row.email}</div>
          </div>
        </div>
      ),
    },
    {
      key: "role",
      header: "Role",
      cell: (row) => <span className="text-sm">{ROLE_LABELS[row.role]}</span>,
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
          variant={row.status === "ACTIVE" ? "active" : row.status === "INACTIVE" ? "inactive" : "deleted"}
          label={row.status}
        />
      ),
    },
    {
      key: "actions",
      header: "",
      align: "right",
      interactive: true,
      // PI and CO_PI come from the upstream allocation; never edited or removed here.
      cell: (row) =>
        canManage && row.role !== "PI" && row.role !== "CO_PI" ? (
          <div className="flex justify-end gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setEditing(row)}
              aria-label={`Edit ${row.display_name}`}
            >
              Edit
            </Button>
            <Button
              variant="destructive"
              size="sm"
              onClick={() => removeMutation.mutate(row.user_id)}
              aria-label={`Remove ${row.display_name}`}
              disabled={removeMutation.isPending}
            >
              Remove
            </Button>
          </div>
        ) : null,
    },
  ];

  return (
    <div className="space-y-3">
      <p className="text-xs text-muted-foreground">
        {members.length} member{members.length === 1 ? "" : "s"}
      </p>
      <DataTable columns={columns} rows={members} rowKey={(row) => row.id} />
      <EditMemberDialog
        member={editing}
        onClose={() => setEditing(null)}
        onSubmit={(payload) => {
          if (!editing) return;
          updateMutation.mutate(
            { userId: editing.user_id, payload },
            { onSuccess: () => setEditing(null) },
          );
        }}
        isPending={updateMutation.isPending}
      />
    </div>
  );
}

type EditMemberDialogProps = {
  member: ProjectMember | null;
  onClose: () => void;
  onSubmit: (payload: { role?: ProjectMemberRole }) => void;
  isPending: boolean;
};

function EditMemberDialog({ member, onClose, onSubmit, isPending }: EditMemberDialogProps) {
  const [role, setRole] = React.useState<ProjectMemberRole>("MEMBER");

  React.useEffect(() => {
    if (member) setRole(member.role === "ALLOCATION_MANAGER" ? "ALLOCATION_MANAGER" : "MEMBER");
  }, [member]);

  return (
    <Dialog open={Boolean(member)} onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit member</DialogTitle>
          <DialogDescription>{member?.display_name}</DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            onSubmit({ role });
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="member-role">Role</Label>
            <select
              id="member-role"
              value={role}
              onChange={(e) => setRole(e.target.value as ProjectMemberRole)}
              className="h-9 w-full rounded-md border bg-background px-3 text-sm"
            >
              <option value="ALLOCATION_MANAGER">Allocation Manager</option>
              <option value="MEMBER">Member</option>
            </select>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending ? "Saving…" : "Save"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
