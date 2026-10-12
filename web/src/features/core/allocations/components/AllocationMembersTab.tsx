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

import { roleLabel } from "@/features/core/projects/components/ProjectMembersTab";
import { UserPicker } from "@/features/core/users/components/UserPicker";
import type { AllocationMembershipResponse, ComputeAllocation } from "@/generated/core/types.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { pluralize } from "@/shared/format";
import { DataTable, type DataTableColumn, NameCell } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Field, FormDialog, idText } from "@/shared/ui/FormDialog";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import { Button } from "@/shared/ui/button";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { emailBeside, initials } from "@/shared/users-admin/displayName";
import * as React from "react";
import { useAddMember, useAllocationMembers, useRemoveMember } from "../queries";
import { AllocationStatusLabel, knownDay } from "./AllocationStatusLabel";
import { memberName } from "./MemberName";
import { MembershipDrawer } from "./MembershipDrawer";
import { useDrawerParam } from "./useDrawerParam";

// Today, or an ended allocation's last day, since a membership may not end before it starts.
function memberStart(endTime: string): string {
  const today = `${new Date().toISOString().slice(0, 10)}T00:00:00Z`;
  return Date.parse(endTime) < Date.parse(today) ? endTime : today;
}

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
  const { id: allocationId, end_time: endTime } = allocation;
  const query = useAllocationMembers(allocationId);
  const addMutation = useAddMember(allocationId);
  const removeMutation = useRemoveMember(allocationId);
  const [selectedId, setSelectedId] = useDrawerParam("membership");
  const canReadUsers = useAbility().can("read", "User");
  const [adding, setAdding] = React.useState(false);
  const [picked, setPicked] = React.useState<string>();

  if (query.isLoading) return <TableSkeleton rows={4} columns={4} />;
  if (query.error) {
    return <ErrorState onRetry={() => query.refetch()} />;
  }
  const members = query.data ?? [];

  const headerCta = canManage ? (
    <FormDialog
      trigger={<Button size="sm">+ Add member</Button>}
      open={adding}
      onOpenChange={(open) => {
        setAdding(open);
        setPicked(undefined);
      }}
      title="Add member"
      description="Add a user to this allocation. They will see it under their allocations."
      submitLabel="Add member"
      isPending={addMutation.isPending}
      submitDisabled={canReadUsers && !picked}
      onSubmit={(form, close) =>
        toastOnSuccess(
          addMutation.mutateAsync({
            body: {
              compute_allocation_id: allocationId,
              user_id: canReadUsers ? (picked ?? "") : idText(form, "user"),
              start_time: memberStart(endTime),
              end_time: endTime,
            },
          }),
          "Member added",
          close,
        )
      }
    >
      {canReadUsers ? (
        <UserPicker
          id="member-user"
          label="User"
          enabled={adding}
          single
          selected={new Set(picked ? [picked] : [])}
          onToggle={setPicked}
          eligible={(u) => !members.some((m) => m.user_id === u.id)}
        />
      ) : (
        <Field
          label="User ID"
          name="user"
          required
          pattern="\s*[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}\s*"
          title="A user ID, e.g. 0b7c2f1e-4a6d-4e8b-9c3a-5f2d1e0a9b7c"
        />
      )}
    </FormDialog>
  ) : null;

  const columns: Array<DataTableColumn<AllocationMembershipResponse>> = [
    {
      key: "member",
      header: "Member",
      cell: (row) => {
        const name = memberName(row);
        return (
          <div className="flex items-center gap-3">
            <Avatar className="size-8">
              <AvatarFallback>{initials(name)}</AvatarFallback>
            </Avatar>
            <div className="min-w-0">
              <NameCell text={name} className="font-medium text-foreground" />
              {!row.email ? (
                <NameCell text={row.user_id} className="font-mono text-xs text-muted-foreground" />
              ) : emailBeside(name, row.email) ? (
                <NameCell text={row.email} className="text-xs text-muted-foreground" />
              ) : null}
            </div>
          </div>
        );
      },
    },
    {
      key: "role",
      header: "Role",
      cell: (row) => <span className="text-sm">{row.role ? roleLabel(row.role) : "—"}</span>,
    },
    {
      key: "start",
      header: "Start",
      cell: (row) => (
        <span className="text-sm text-muted-foreground">{knownDay(row.start_time)}</span>
      ),
    },
    {
      key: "end",
      header: "End",
      cell: (row) => (
        <span className="text-sm text-muted-foreground">{knownDay(row.end_time)}</span>
      ),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => (
        <AllocationStatusLabel
          status={row.membership_status}
          startTime={row.start_time}
          endTime={row.end_time}
        />
      ),
    },
  ];
  if (canManage) {
    columns.push({
      key: "actions",
      header: "",
      align: "right",
      interactive: true,
      cell: (row) => (
        <div className="flex justify-end gap-2">
          <Button
            variant="destructive"
            size="sm"
            onClick={() =>
              confirmToast(`Remove ${memberName(row)}?`, "Remove", () =>
                toastOnSuccess(
                  removeMutation.mutateAsync({ path: { id: row.id } }),
                  "Member removed",
                ),
              )
            }
            aria-label={`Remove ${memberName(row)}`}
            pending={removeMutation.isPending && removeMutation.variables?.path.id === row.id}
          >
            Remove
          </Button>
        </div>
      ),
    });
  }

  const selected = members.find((m) => m.id === selectedId);
  return (
    <div className="space-y-3">
      {members.length === 0 ? (
        <>
          <div className="flex justify-end">{headerCta}</div>
          <EmptyState
            heading="No members yet"
            description="Add a user to grant them access to this allocation."
          />
        </>
      ) : (
        <>
          <div className="flex items-center justify-between">
            <p className="text-xs text-muted-foreground">
              {members.length} {pluralize("member", members.length)}
            </p>
            {headerCta}
          </div>
          <DataTable
            caption="Members"
            columns={columns}
            rows={members}
            rowKey={(row) => row.id}
            onRowClick={canRead ? (row) => setSelectedId(row.id) : undefined}
          />
        </>
      )}
      <MembershipDrawer
        allocationId={allocationId}
        membershipId={canRead ? selectedId : undefined}
        name={selected && memberName(selected)}
        canManage={canManage}
        onClose={() => setSelectedId()}
      />
    </div>
  );
}
