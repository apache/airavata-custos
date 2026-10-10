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

import { AllocationName } from "@/features/core/allocations/components/AllocationName";
import * as React from "react";
import { CreateTempAccountDialog } from "@/features/connectors/temp-account/components/CreateTempAccountDialog";
import { TempMembershipDialog } from "@/features/connectors/temp-account/components/TempMembershipDialog";
import { useRemoveTempAccount, useTempMemberships } from "@/features/connectors/temp-account/queries";
import type { ComputeAllocationMembership } from "@/generated/temp-account/types.gen";
import { ApiError } from "@/shared/api/client";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { formatDate } from "@/shared/format";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";

export function TempAccountsContainer() {
  const ability = useAbility();
  const canRead = ability.can("read", "TempAccount");
  const canWrite = ability.can("write", "TempAccount");

  const [draft, setDraft] = React.useState("");
  const [userId, setUserId] = React.useState<string>();
  // undefined: closed; null: assigning; a membership: editing it.
  const [editing, setEditing] = React.useState<ComputeAllocationMembership | null>();

  const memberships = useTempMemberships(canRead ? userId : undefined);
  const remove = useRemoveTempAccount();

  if (!canRead && !canWrite) {
    return <ErrorState heading="Not permitted" message="You cannot manage temporary accounts." />;
  }

  function select(id: string | undefined) {
    setUserId(id);
    setDraft(id ?? "");
  }

  function handleRemove(id: string) {
    confirmToast(`Remove temporary account ${id}?`, "Remove", () =>
      remove.mutate(
        { path: { user_id: id } },
        toastOnSuccess("Temporary account removed", () => select(undefined)),
      ),
    );
  }

  const columns: DataTableColumn<ComputeAllocationMembership>[] = [
    {
      key: "allocation",
      header: "Compute allocation",
      cell: (m) => <AllocationName id={m.compute_allocation_id} />,
    },
    { key: "start", header: "Starts", cell: (m) => formatDate(m.start_time) },
    { key: "end", header: "Ends", cell: (m) => formatDate(m.end_time) },
    { key: "status", header: "Status", cell: (m) => m.membership_status },
    {
      key: "actions",
      header: "",
      align: "right",
      cell: (m) =>
        canWrite ? (
          <Button type="button" variant="ghost" size="sm" onClick={() => setEditing(m)}>
            Edit
          </Button>
        ) : null,
    },
  ];

  // The connector answers 404 when the account holds no membership.
  const notFound = memberships.error instanceof ApiError && memberships.error.status === 404;

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="font-display text-[28px] font-bold leading-tight">Temporary accounts</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            VIRTUAL users created outside ACCESS, with their allocation memberships.
          </p>
        </div>
        {canWrite ? <CreateTempAccountDialog onCreated={select} /> : null}
      </div>

      <form
        className="flex flex-wrap items-end gap-3 rounded-md border bg-card p-4"
        onSubmit={(e) => {
          e.preventDefault();
          setUserId(draft.trim() || undefined);
        }}
      >
        <div className="flex flex-col gap-1">
          <Label htmlFor="temp-user-id">User ID</Label>
          <Input
            id="temp-user-id"
            value={draft}
            onChange={(e) => setDraft(e.currentTarget.value)}
            className="w-80 font-mono"
          />
        </div>
        <Button type="submit" variant="outline">
          Open
        </Button>
        {userId && canWrite ? (
          <div className="ml-auto flex gap-2">
            <Button type="button" variant="outline" onClick={() => setEditing(null)}>
              Assign allocation
            </Button>
            <Button type="button" variant="outline" onClick={() => handleRemove(userId)}>
              Remove account
            </Button>
          </div>
        ) : null}
      </form>

      {!userId || !canRead ? null : memberships.isLoading ? (
        <TableSkeleton />
      ) : notFound ? (
        <EmptyState heading="No allocation memberships" description={memberships.error?.message} />
      ) : memberships.error ? (
        <ErrorState message={memberships.error.message} onRetry={() => memberships.refetch()} />
      ) : (
        <DataTable
          columns={columns}
          rows={memberships.data ?? []}
          rowKey={(m) => m.id ?? ""}
        />
      )}

      {canWrite && userId ? (
        <TempMembershipDialog
          userId={userId}
          membership={editing}
          onClose={() => setEditing(undefined)}
        />
      ) : null}
    </div>
  );
}
