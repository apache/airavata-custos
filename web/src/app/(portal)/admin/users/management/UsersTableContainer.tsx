"use client";

import * as React from "react";
import { useCurrentUser } from "@/features/core/identity/queries";
import { useRolesCatalog, useUserPageDetails, useUsers } from "@/features/core/users/queries";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { UsersTable } from "./UsersTable";

const PAGE_SIZE = 25;

export function UsersTableContainer() {
  const ability = useAbility();
  const canReadUsers = ability.can("read", "User");
  const canManageRoles = ability.can("manage", "Role");
  const canReadDirectPrivileges = ability.can("manage", "PrivilegeGrant");
  const { user: currentUser } = useCurrentUser();
  const [page, setPage] = React.useState(1);
  const usersQuery = useUsers(
    { limit: PAGE_SIZE, offset: (page - 1) * PAGE_SIZE },
    { enabled: canReadUsers },
  );
  const rolesQuery = useRolesCatalog(canManageRoles);
  const roleReadsEnabled = canManageRoles && rolesQuery.isSuccess;
  const hydratedRows = useUserPageDetails(
    usersQuery.data?.items ?? [],
    rolesQuery.data ?? [],
    roleReadsEnabled,
  ).map((row) => ({
    ...row,
    rolesLoading: canManageRoles && (rolesQuery.isLoading || row.rolesLoading),
    rolesError: canManageRoles && (rolesQuery.isError || row.rolesError),
  }));

  if (!canReadUsers) {
    return <ErrorState heading="Not permitted" message="You cannot view users." />;
  }
  if (usersQuery.isLoading) {
    return <TableSkeleton rows={8} columns={canManageRoles ? 5 : 4} />;
  }
  if (usersQuery.isError) {
    return (
      <ErrorState
        message={
          usersQuery.error instanceof Error ? usersQuery.error.message : "Could not load users."
        }
        onRetry={() => void usersQuery.refetch()}
      />
    );
  }

  return (
    <UsersTable
      users={hydratedRows}
      rolesCatalog={rolesQuery.data ?? []}
      currentUserEmail={currentUser?.email}
      canManageRoles={canManageRoles}
      canReadDirectPrivileges={canReadDirectPrivileges}
      page={page}
      pageSize={PAGE_SIZE}
      total={usersQuery.data?.total ?? 0}
      onPageChange={setPage}
    />
  );
}
