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

import * as React from "react";
import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import { Button } from "@/shared/ui/button";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { StatusBadge, statusBadgeVariantFromAllocationStatus } from "@/shared/ui/StatusBadge";
import { Field, FormDialog, text } from "@/shared/ui/FormDialog";
import { useAddMember, useAllocationMembers, useRemoveMember } from "../queries";
import type {
  AllocationMembershipResponse,
  ComputeAllocation,
} from "@/generated/core/types.gen";
import { initialsFrom, ROLE_LABELS } from "@/features/core/projects/components/ProjectMembersTab";
import { MembershipDrawer } from "./MembershipDrawer";

export type AllocationMembersTabProps = {
  allocation: ComputeAllocation;
  canManage: boolean;
  // Membership detail requires the allocations read privilege.
  canRead: boolean;
};

export function AllocationMembersTab({
  allocation,
  canManage,
  canRead,
}: AllocationMembersTabProps) {
  const { id: allocationId = "", end_time: endTime = "" } = allocation;
  const query = useAllocationMembers(allocationId);
  const addMutation = useAddMember(allocationId);
  const removeMutation = useRemoveMember(allocationId);
  const [selectedId, setSelectedId] = React.useState<string>();

  if (query.isLoading) return <TableSkeleton rows={4} columns={4} />;
  if (query.error) {
    return <ErrorState message={query.error.message} onRetry={() => query.refetch()} />;
  }
  const members = query.data ?? [];

  const headerCta = canManage ? (
    <FormDialog
      trigger={<Button size="sm">+ Add member</Button>}
      title="Add member"
      description="Add a user to this allocation. They will see it under their allocations."
      submitLabel="Add member"
      isPending={addMutation.isPending}
      onSubmit={(form, close) =>
        addMutation.mutate(
          {
            body: {
              compute_allocation_id: allocationId,
              user_id: text(form, "user"),
              start_time: new Date().toISOString(),
              end_time: endTime,
            },
          },
          toastOnSuccess("Member added", close),
        )
      }
    >
      <Field label="User ID" name="user" placeholder="user-123" required />
    </FormDialog>
  ) : null;

  if (members.length === 0) {
    return (
      <div className="space-y-3">
        <div className="flex justify-end">{headerCta}</div>
        <EmptyState
          heading="No members yet"
          description="Add a user to grant them access to this allocation."
        />
      </div>
    );
  }

  const columns: Array<DataTableColumn<AllocationMembershipResponse>> = [
    {
      key: "member",
      header: "Member",
      cell: (row) => {
        const name = row.display_name ?? row.user_id ?? "";
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
        <span className="text-sm">{row.role ? (ROLE_LABELS[row.role] ?? row.role) : "—"}</span>
      ),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => (
        <StatusBadge
          variant={statusBadgeVariantFromAllocationStatus(row.membership_status)}
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
              variant="destructive"
              size="sm"
              onClick={() =>
                confirmToast(`Remove ${row.display_name ?? row.user_id}?`, "Remove", () =>
                  removeMutation.mutate(
                    { path: { id: row.id ?? "" } },
                    toastOnSuccess("Member removed"),
                  ),
                )
              }
              aria-label={`Remove ${row.display_name ?? row.user_id}`}
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
      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          {members.length} member{members.length === 1 ? "" : "s"}
        </p>
        {headerCta}
      </div>
      <DataTable
        columns={columns}
        rows={members}
        rowKey={(row) => row.id ?? ""}
        onRowClick={canRead ? (row) => setSelectedId(row.id) : undefined}
      />
      <MembershipDrawer
        allocationId={allocationId}
        membershipId={selectedId}
        canManage={canManage}
        onClose={() => setSelectedId(undefined)}
      />
    </div>
  );
}
