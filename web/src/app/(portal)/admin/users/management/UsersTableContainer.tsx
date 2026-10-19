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
import { toast } from "sonner";
import { useAllocations } from "@/features/core/allocations/queries";
import {
  type ReviewAction,
  ReviewClusterAccountDialog,
} from "@/features/core/cluster-accounts/components/ReviewClusterAccountDialog";
import {
  useApproveClusterAccount,
  useClusterAccounts,
  useDenyClusterAccount,
  useUserClusterAccounts,
} from "@/features/core/cluster-accounts/queries";
import { useCurrentUser } from "@/features/core/identity/queries";
import {
  useCreateUser,
  useRolesCatalog,
  useUserPageDetails,
  useUsers,
  useUsersById,
} from "@/features/core/users/queries";
import type { CreateUserPayload } from "@/features/core/users/schemas";
import type { ClusterAccountResponse } from "@/generated/core/zod.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { Button } from "@/shared/ui/button";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { AddUserDialog } from "./AddUserDialog";
import {
  type AccountStatusFilter,
  toAccountStatus,
  UsersTable,
} from "./UsersTable";

const PAGE_SIZE = 25;

// All pages through users; a status pages through cluster accounts in that status, grouped by user.
export function UsersTableContainer() {
  const ability = useAbility();
  const canReadUsers = ability.can("read", "User");
  const canManageRoles = ability.can("manage", "Role");
  const canReadDirectPrivileges = ability.can("manage", "PrivilegeGrant");
  const canCreateUsers = ability.can("manage", "User");
  const canReadAccounts = ability.can("read", "Cluster");
  const canReviewAccounts = ability.can("manage", "Cluster");
  const { user: currentUser } = useCurrentUser();
  const params = useShallowSearchParams();
  const filter = canReadAccounts ? toAccountStatus(params.get("status")) : "ALL";
  const byStatus = filter !== "ALL";
  const [page, setPage] = React.useState(1);
  const [addDialogOpen, setAddDialogOpen] = React.useState(false);
  const [addError, setAddError] = React.useState<string | null>(null);
  const [review, setReview] = React.useState<{
    account: ClusterAccountResponse;
    action: ReviewAction;
  } | null>(null);
  const [reviewError, setReviewError] = React.useState<string | null>(null);
  const createMutation = useCreateUser();
  const approve = useApproveClusterAccount();
  const deny = useDenyClusterAccount();
  const allocationsQuery = useAllocations({ status: "ACTIVE" });

  async function handleAddUser(payload: CreateUserPayload) {
    setAddError(null);
    try {
      await createMutation.mutateAsync(payload);
      toast.success(payload.portal_admin ? "Admin user added" : "User added");
      setAddDialogOpen(false);
    } catch (err) {
      setAddError(err instanceof Error ? err.message : "Failed to add user");
    }
  }
  const usersQuery = useUsers(
    { limit: PAGE_SIZE, offset: (page - 1) * PAGE_SIZE },
    { enabled: canReadUsers && !byStatus },
  );
  const accountsQuery = useClusterAccounts(
    { ...(filter === "ALL" ? {} : { approval_status: filter }), limit: PAGE_SIZE, offset: (page - 1) * PAGE_SIZE },
    { enabled: canReadUsers && byStatus },
  );
  const accountUsers = useUsersById(
    [...new Set((accountsQuery.data?.items ?? []).map((a) => a.user_id))],
    canReadUsers && byStatus,
  );
  const pageUsers = byStatus ? accountUsers.users : (usersQuery.data?.items ?? []);
  const { clusters, byUser } = useUserClusterAccounts(pageUsers, canReadAccounts);
  const rolesQuery = useRolesCatalog(canManageRoles);
  const roleReadsEnabled = canManageRoles && rolesQuery.isSuccess;
  const hydratedRows = useUserPageDetails(
    pageUsers,
    rolesQuery.data ?? [],
    roleReadsEnabled,
  ).map((row) => ({
    ...row,
    rolesLoading: canManageRoles && (rolesQuery.isLoading || row.rolesLoading),
    rolesError: canManageRoles && (rolesQuery.isError || row.rolesError),
    accounts: byUser.get(row.id) ?? [],
  }));

  function setFilter(next: AccountStatusFilter) {
    const nextParams = new URLSearchParams(params);
    if (next === "ALL") nextParams.delete("status");
    else nextParams.set("status", next);
    replaceShallowSearchParams(nextParams);
    setPage(1);
  }

  function openReview(action: ReviewAction) {
    return (account: ClusterAccountResponse) => {
      setReviewError(null);
      setReview({ account, action });
    };
  }

  async function handleReview(note: string) {
    if (!review) return;
    const { account, action } = review;
    setReviewError(null);
    try {
      if (action === "approve") await approve.mutateAsync(account.id);
      else await deny.mutateAsync({ id: account.id, note });
      toast.success(
        `Cluster account ${action === "approve" ? "approved" : "denied"} for ${account.local_username} on ${account.cluster_name}`,
      );
      setReview(null);
    } catch (err) {
      setReviewError(err instanceof Error ? err.message : "The review did not go through.");
    }
  }

  if (!canReadUsers) {
    return <ErrorState heading="Not permitted" message="You cannot view users." />;
  }
  const listQuery = byStatus ? accountsQuery : usersQuery;
  const error = listQuery.error ?? accountUsers.error;

  return (
    <>
      <UsersTable
        users={hydratedRows}
        rolesCatalog={rolesQuery.data ?? []}
        currentUserEmail={currentUser?.email}
        canManageRoles={canManageRoles}
        canReadDirectPrivileges={canReadDirectPrivileges}
        canReadAllocations={ability.can("read", "Allocation")}
        page={page}
        pageSize={PAGE_SIZE}
        total={listQuery.data?.total ?? 0}
        onPageChange={setPage}
        placeholder={
          listQuery.isLoading || accountUsers.isLoading ? (
            <TableSkeleton rows={8} columns={canManageRoles ? 7 : 6} />
          ) : error ? (
            <ErrorState message={error.message} onRetry={() => void listQuery.refetch()} />
          ) : undefined
        }
        clusterAccounts={
          canReadAccounts
            ? {
                status: filter,
                clusters,
                onStatusChange: setFilter,
                onApprove: canReviewAccounts ? openReview("approve") : undefined,
                onDeny: canReviewAccounts ? openReview("deny") : undefined,
              }
            : undefined
        }
        toolbarAction={
          canCreateUsers ? (
            <Button
              className="sm:ml-auto"
              onClick={() => {
                setAddError(null);
                setAddDialogOpen(true);
              }}
            >
              + Add user
            </Button>
          ) : null
        }
      />
      {canCreateUsers ? (
        <AddUserDialog
          open={addDialogOpen}
          onOpenChange={setAddDialogOpen}
          onSubmit={handleAddUser}
          isPending={createMutation.isPending}
          error={addError}
          canCreateAdmins={canManageRoles}
          allocations={allocationsQuery.data?.items ?? []}
        />
      ) : null}
      {canReviewAccounts ? (
        <ReviewClusterAccountDialog
          account={review?.account ?? null}
          action={review?.action ?? "approve"}
          onOpenChange={(open) => {
            if (!open) setReview(null);
          }}
          onConfirm={handleReview}
          isPending={approve.isPending || deny.isPending}
          error={reviewError}
        />
      ) : null}
    </>
  );
}
