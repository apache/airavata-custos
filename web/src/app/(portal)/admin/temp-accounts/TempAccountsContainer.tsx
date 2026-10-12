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

import { CreateTempAccountDialog } from "@/features/connectors/temp-account/components/CreateTempAccountDialog";
import { TempMembershipDialog } from "@/features/connectors/temp-account/components/TempMembershipDialog";
import {
  useRemoveTempAccount,
  useTempMemberships,
} from "@/features/connectors/temp-account/queries";
import { AllocationName } from "@/features/core/allocations/components/AllocationName";
import { membershipStatusText } from "@/features/core/allocations/components/AllocationStatusLabel";
import { useAllocation } from "@/features/core/allocations/queries";
import { useUser } from "@/features/core/users/queries";
import type { ComputeAllocationMembership } from "@/generated/core/types.gen";
import { ApiError } from "@/shared/api/client";
import { isGone } from "@/shared/api/useOnGone";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { formatDay } from "@/shared/format";
import {
  setSearchParam,
  useCanonicalParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { ROUTE_LABELS, queryUuid } from "@/shared/layout/nav";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState, NotPermitted } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { displayName, emailBeside } from "@/shared/users-admin/displayName";
import { idInput } from "@/shared/users-admin/idInput";
import * as React from "react";

export function TempAccountsContainer() {
  const ability = useAbility();
  const canRead = ability.can("read", "TempAccount");
  const canWrite = ability.can("write", "TempAccount");

  const params = useShallowSearchParams();
  const userId = queryUuid(params.get("user"));
  // A user found not to be temporary leaves the address; its id stays in the field.
  const [refused, setRefused] = React.useState<string>();
  useCanonicalParams({ user: userId === refused ? undefined : userId });
  const setUserId = (id?: string) => setSearchParam(params, "user", id);
  const [draft, setDraft] = React.useState(userId ?? "");
  const [editing, setEditing] = React.useState<ComputeAllocationMembership | null>();
  const userInput = React.useRef<HTMLInputElement>(null);

  const canReadUser = ability.can("read", "User");
  const canReadAllocation = ability.can("read", "Allocation");
  const userQuery = useUser((canRead || canWrite) && canReadUser ? userId : undefined);
  const user = userQuery.data;
  const userMissing = isGone(userQuery.error);
  // A readable user is asked for memberships only once known VIRTUAL; the connector refuses others.
  const memberships = useTempMemberships(
    canRead && (!canReadUser || user?.type === "VIRTUAL") ? userId : undefined,
  );
  const remove = useRemoveTempAccount();

  const status = memberships.error instanceof ApiError ? memberships.error.status : undefined;
  // The connector answers 404 both for a missing user and for one without memberships.
  const noMemberships = status === 404;
  const notFound = userMissing || (noMemberships && !canReadUser);
  const notTemporary = status === 400 || (user !== undefined && user.type !== "VIRTUAL");
  const isVirtual = user?.type === "VIRTUAL" || memberships.isSuccess;
  React.useEffect(() => {
    if (userId && notTemporary) setRefused(userId);
  }, [userId, notTemporary]);

  if (!canRead && !canWrite) {
    return <NotPermitted title={ROUTE_LABELS["/admin/temp-accounts"]} />;
  }

  function open(id: string | undefined) {
    setRefused(undefined);
    setUserId(id);
  }

  function select(id: string | undefined) {
    open(id);
    setDraft(id ?? "");
  }

  function handleRemove(id: string) {
    const who = user ? displayName(user) : "this user";
    confirmToast(`Remove the temporary account of ${who}?`, "Remove", () =>
      toastOnSuccess(
        remove.mutateAsync({ path: { user_id: id } }),
        "Temporary account removed",
        () => {
          select(undefined);
          userInput.current?.focus();
        },
      ),
    );
  }

  const columns: DataTableColumn<ComputeAllocationMembership>[] = [
    ...(canReadAllocation
      ? [
          {
            key: "allocation",
            header: "Compute allocation",
            cell: (m: ComputeAllocationMembership) => (
              <AllocationName id={m.compute_allocation_id} />
            ),
          },
        ]
      : []),
    { key: "start", header: "Starts", cell: (m) => formatDay(m.start_time) },
    { key: "end", header: "Ends", cell: (m) => formatDay(m.end_time) },
    { key: "status", header: "Status", cell: (m) => membershipStatusText(m.membership_status) },
    ...(canWrite
      ? [
          {
            key: "actions",
            header: "",
            align: "right" as const,
            cell: (m: ComputeAllocationMembership) => (
              <EditMembershipButton
                membership={m}
                canReadAllocation={canReadAllocation}
                onEdit={() => setEditing(m)}
              />
            ),
          },
        ]
      : []),
  ];

  const rows = memberships.data ?? [];

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="font-display text-[28px] font-bold leading-tight">
            {ROUTE_LABELS["/admin/temp-accounts"]}
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Temporary users created outside ACCESS, with their allocation memberships.
          </p>
        </div>
        {canWrite ? <CreateTempAccountDialog onCreated={select} /> : null}
      </div>

      <form
        className="flex flex-wrap items-end gap-3 rounded-md border bg-card p-4"
        onSubmit={(e) => {
          e.preventDefault();
          open(draft.trim() || undefined);
        }}
      >
        <div className="flex flex-col gap-1">
          <Label htmlFor="temp-user-id">User ID</Label>
          <Input
            ref={userInput}
            id="temp-user-id"
            value={draft}
            onChange={(e) => setDraft(e.currentTarget.value)}
            {...idInput("user")}
            className="w-80 font-mono"
          />
        </div>
        <Button type="submit" variant="outline">
          Open
        </Button>
        {userId && canWrite && !notFound && !notTemporary ? (
          <div className="ml-auto flex gap-2">
            <Button type="button" variant="outline" onClick={() => setEditing(null)}>
              Assign allocation
            </Button>
            {isVirtual ? (
              <Button type="button" variant="outline" onClick={() => handleRemove(userId)}>
                Remove account
              </Button>
            ) : null}
          </div>
        ) : null}
      </form>

      {user && userId && !notFound && !notTemporary ? (
        <div className="min-w-0">
          <h2 className="font-display text-lg font-semibold [overflow-wrap:anywhere]">
            {displayName(user)}
          </h2>
          {emailBeside(displayName(user), user.email) ? (
            <p className="text-sm text-muted-foreground [overflow-wrap:anywhere]">{user.email}</p>
          ) : null}
        </div>
      ) : null}

      {notTemporary || (!userId && refused) ? (
        <EmptyState heading="Not a temporary account" />
      ) : !userId || !canRead ? null : memberships.isLoading || userQuery.isLoading ? (
        <TableSkeleton />
      ) : userMissing ? (
        <EmptyState heading="No user has that ID" />
      ) : notFound ? (
        <EmptyState heading="No user has that ID, or the user has no allocation memberships" />
      ) : memberships.error && !noMemberships ? (
        <ErrorState error={memberships.error} onRetry={() => memberships.refetch()} />
      ) : rows.length === 0 ? (
        <EmptyState heading="No allocation memberships" />
      ) : (
        <DataTable
          caption="Allocation memberships"
          columns={columns}
          rows={rows}
          rowKey={(m) => m.id}
        />
      )}

      {canWrite && userId ? (
        <TempMembershipDialog
          userId={userId}
          membership={editing}
          memberships={canRead ? memberships : undefined}
          onClose={() => setEditing(undefined)}
        />
      ) : null}
    </div>
  );
}

// Rows are told apart by their allocation, or by their dates for a caller who cannot read it.
function EditMembershipButton({
  membership: m,
  canReadAllocation,
  onEdit,
}: {
  membership: ComputeAllocationMembership;
  canReadAllocation: boolean;
  onEdit: () => void;
}) {
  const name = useAllocation(canReadAllocation ? m.compute_allocation_id : undefined).data?.name;
  return (
    <Button
      type="button"
      variant="ghost"
      size="sm"
      aria-label={
        name
          ? `Edit ${name} membership`
          : `Edit membership from ${formatDay(m.start_time)} to ${formatDay(m.end_time)}`
      }
      onClick={onEdit}
    >
      Edit
    </Button>
  );
}
