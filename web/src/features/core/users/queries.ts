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

import { useMutation, useQueries, useQuery, useQueryClient } from "@tanstack/react-query";
import { identityKeys } from "@/features/core/identity/queries";
import {
  assignUserRole,
  getRoleDetail,
  listDirectPrivileges,
  listRolesCatalog,
  listRolesForUser,
  listUserIdentities,
  listUsers,
  removeUserRole,
} from "./api";
import type { Role, User, UserRole } from "./schemas";
import type {
  RoleWithPrivileges,
  UpdateUserRolesInput,
  UserListParams,
  UserManagementRow,
} from "./types";

export const userKeys = {
  all: ["users"] as const,
  list: (params: UserListParams = {}) => [...userKeys.all, "list", params] as const,
  rolesCatalog: () => [...userKeys.all, "roles-catalog"] as const,
  roles: (userId: string) => [...userKeys.all, "roles", userId] as const,
  identities: (userId: string) => [...userKeys.all, "identities", userId] as const,
  directPrivileges: (userId: string) => [...userKeys.all, "direct-privileges", userId] as const,
  roleDetail: (roleId: string) => [...userKeys.all, "role-detail", roleId] as const,
};

const DEFAULTS = {
  staleTime: 30_000,
  gcTime: 300_000,
  refetchOnWindowFocus: false,
} as const;

type RoleUpdateOperations = {
  assign: (userId: string, roleId: string, reason?: string) => Promise<unknown>;
  remove: (userId: string, roleId: string) => Promise<unknown>;
};

const defaultRoleUpdateOperations: RoleUpdateOperations = {
  assign: assignUserRole,
  remove: removeUserRole,
};

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

export async function applyUserRoleChanges(
  { userId, currentRoleIds, desiredRoleIds, reason }: UpdateUserRolesInput,
  operations: RoleUpdateOperations = defaultRoleUpdateOperations,
): Promise<void> {
  const current = new Set(currentRoleIds);
  const desired = new Set(desiredRoleIds);
  const changes = [
    ...[...desired]
      .filter((roleId) => !current.has(roleId))
      .map((roleId) => ({
        apply: () => operations.assign(userId, roleId, reason),
        rollback: () => operations.remove(userId, roleId),
      })),
    ...[...current]
      .filter((roleId) => !desired.has(roleId))
      .map((roleId) => ({
        apply: () => operations.remove(userId, roleId),
        rollback: () => operations.assign(userId, roleId),
      })),
  ];
  const applied: typeof changes = [];

  try {
    for (const change of changes) {
      await change.apply();
      applied.push(change);
    }
  } catch (error) {
    const rollbackFailures: unknown[] = [];
    for (const change of [...applied].reverse()) {
      try {
        await change.rollback();
      } catch (rollbackError) {
        rollbackFailures.push(rollbackError);
      }
    }

    if (rollbackFailures.length > 0) {
      throw new Error(
        `Role update partially failed, and rollback could not fully restore the previous roles. Refresh to see the current assignments. Original error: ${errorMessage(error)}`,
      );
    }
    if (applied.length > 0) {
      throw new Error(
        `Role update failed; completed changes were rolled back. ${errorMessage(error)}`,
      );
    }
    throw error;
  }
}

export function useUsers(params: UserListParams = {}, options: { enabled?: boolean } = {}) {
  return useQuery({
    queryKey: userKeys.list(params),
    queryFn: () => listUsers(params),
    enabled: options.enabled ?? true,
    ...DEFAULTS,
  });
}

export function useRolesCatalog(enabled: boolean) {
  return useQuery({
    queryKey: userKeys.rolesCatalog(),
    queryFn: listRolesCatalog,
    enabled,
    ...DEFAULTS,
  });
}

export function useUserRoles(userId: string | undefined, enabled: boolean) {
  return useQuery({
    queryKey: userId ? userKeys.roles(userId) : [...userKeys.all, "roles", "none"],
    queryFn: () => listRolesForUser(userId as string),
    enabled: Boolean(userId) && enabled,
    ...DEFAULTS,
  });
}

export function useUserIdentities(userId: string | undefined, enabled = true) {
  return useQuery({
    queryKey: userId ? userKeys.identities(userId) : [...userKeys.all, "identities", "none"],
    queryFn: () => listUserIdentities(userId as string),
    enabled: Boolean(userId) && enabled,
    ...DEFAULTS,
  });
}

export function useDirectPrivileges(userId: string | undefined, enabled: boolean) {
  return useQuery({
    queryKey: userId
      ? userKeys.directPrivileges(userId)
      : [...userKeys.all, "direct-privileges", "none"],
    queryFn: () => listDirectPrivileges(userId as string),
    enabled: Boolean(userId) && enabled,
    ...DEFAULTS,
  });
}

export function useRoleDetails(
  roleIds: string[],
  enabled: boolean,
): { roles: RoleWithPrivileges[]; isLoading: boolean; isError: boolean } {
  const stableIds = Array.from(new Set(roleIds.filter(Boolean))).sort();
  const results = useQueries({
    queries: stableIds.map((roleId) => ({
      queryKey: userKeys.roleDetail(roleId),
      queryFn: () => getRoleDetail(roleId),
      enabled,
      ...DEFAULTS,
    })),
  });

  const roles = results.flatMap((result) => {
    if (!result.data?.role) return [];
    return [{ ...result.data.role, privileges: result.data.privileges }];
  });
  return {
    roles,
    isLoading: enabled && results.some((result) => result.isLoading),
    isError: enabled && results.some((result) => result.isError),
  };
}

export function useUserPageDetails(
  users: User[],
  rolesCatalog: Role[],
  canManageRoles: boolean,
): UserManagementRow[] {
  const usersWithIds = users.filter((user): user is User & { id: string } => Boolean(user.id));
  const identityResults = useQueries({
    queries: usersWithIds.map((user) => ({
      queryKey: userKeys.identities(user.id),
      queryFn: () => listUserIdentities(user.id),
      ...DEFAULTS,
    })),
  });
  const roleResults = useQueries({
    queries: usersWithIds.map((user) => ({
      queryKey: userKeys.roles(user.id),
      queryFn: () => listRolesForUser(user.id),
      enabled: canManageRoles,
      ...DEFAULTS,
    })),
  });
  const roleById = new Map(rolesCatalog.map((role) => [role.id, role]));
  const indexById = new Map(usersWithIds.map((user, index) => [user.id, index]));

  return users.map((user) => {
    const index = indexById.get(user.id);
    const identitiesQuery = index === undefined ? undefined : identityResults[index];
    const rolesQuery = index === undefined ? undefined : roleResults[index];
    const roles = canManageRoles
      ? (rolesQuery?.data ?? []).flatMap((grant) => {
          const role = roleById.get(grant.role_id);
          return role ? [role] : [];
        })
      : [];
    return {
      ...user,
      roles,
      identities: identitiesQuery?.data ?? [],
      rolesLoading: canManageRoles && Boolean(rolesQuery?.isLoading),
      identitiesLoading: Boolean(identitiesQuery?.isLoading),
      rolesError: canManageRoles && Boolean(rolesQuery?.isError),
      identitiesError: Boolean(identitiesQuery?.isError),
    };
  });
}

export function useUpdateUserRoles() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateUserRolesInput) => applyUserRoleChanges(input),
    onSuccess: (_data, { userId, desiredRoleIds }) => {
      client.setQueryData<UserRole[]>(userKeys.roles(userId), (currentGrants = []) =>
        desiredRoleIds.map(
          (roleId) =>
            currentGrants.find((grant) => grant.role_id === roleId) ?? {
              user_id: userId,
              role_id: roleId,
            },
        ),
      );
    },
    onSettled: (_data, _error, variables) =>
      Promise.all([
        client.invalidateQueries({ queryKey: userKeys.roles(variables.userId) }),
        client.invalidateQueries({ queryKey: identityKeys.privileges() }),
        client.invalidateQueries({ queryKey: identityKeys.access(variables.userId) }),
      ]).then(() => undefined),
  });
}
