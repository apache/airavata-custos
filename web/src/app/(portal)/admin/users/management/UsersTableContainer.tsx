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
