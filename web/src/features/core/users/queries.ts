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

import { clusterAccountKeys } from "@/features/core/cluster-accounts/queries";
import { clusterKeys } from "@/features/core/clusters/queries";
import { identityKeys } from "@/features/core/identity/queries";
import { roleKeys } from "@/features/core/roles/queries";
import {
  deleteUserIdentitiesById,
  deleteUsersByIdPrivilegesByKey,
  deleteUsersByIdRolesByRoleId,
  getPrivilegesByKeyHolders,
  getRoles,
  getRolesById,
  getUserIdentitiesById,
  getUserIdentitiesOidcSubjectsByOidcSub,
  getUserIdentitiesSourcesBySourceExternalByExternalId,
  getUsers,
  getUsersById,
  getUsersByIdChangeRequests,
  getUsersByIdComputeAllocationMemberships,
  getUsersByIdComputeAllocationUsages,
  getUsersByIdComputeClusterUsers,
  getUsersByIdPrivileges,
  getUsersByIdRoles,
  getUsersByIdUserIdentities,
  postUserIdentities,
  postUsers,
  postUsersByIdPrivileges,
  postUsersByIdRoles,
  postUsersMerge,
  putUserIdentitiesById,
  putUsersByIdStatus,
} from "@/generated/core/sdk.gen";
import type {
  CreateUserRequest,
  GetRolesByIdResponse,
  GetUsersData,
  GrantPrivilegeRequest,
  MergeUsersRequest,
  PrivilegeKey,
  Role,
  User,
  UserIdentity,
} from "@/generated/core/types.gen";
import { useInvalidating } from "@/shared/api/useInvalidating";
import {
  skipToken,
  useInfiniteQuery,
  useMutation,
  useQueries,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";

export type UserManagementRow = User & {
  roles: Role[];
  identities: UserIdentity[];
  rolesLoading: boolean;
  identitiesLoading: boolean;
  rolesError: boolean;
  identitiesError: boolean;
};

type UpdateUserRolesInput = {
  userId: string;
  currentRoleIds: string[];
  desiredRoleIds: string[];
  reason?: string;
};

export const userKeys = {
  all: ["users"] as const,
  roles: (userId: string) => [...userKeys.all, "roles", userId] as const,
  identities: (userId: string) => [...userKeys.all, "identities", userId] as const,
};

// The three ways an identity resolves: its own id, an OIDC subject, or a source's external id.
export type IdentityLookup =
  | { by: "id"; id: string }
  | { by: "oidc"; oidcSub: string }
  | { by: "external"; source: string; externalId: string };

async function assign(userId: string, roleId: string, reason?: string) {
  await postUsersByIdRoles({
    path: { id: userId },
    body: { role_id: roleId, reason: reason?.trim() || undefined },
  });
}

async function remove(userId: string, roleId: string) {
  await deleteUsersByIdRolesByRoleId({ path: { id: userId, roleId } });
}

export async function applyUserRoleChanges({
  userId,
  currentRoleIds,
  desiredRoleIds,
  reason,
}: UpdateUserRolesInput): Promise<void> {
  const current = new Set(currentRoleIds);
  const desired = new Set(desiredRoleIds);
  const changes = [
    ...[...desired]
      .filter((roleId) => !current.has(roleId))
      .map((roleId) => ({
        apply: () => assign(userId, roleId, reason),
        rollback: () => remove(userId, roleId),
      })),
    ...[...current]
      .filter((roleId) => !desired.has(roleId))
      .map((roleId) => ({
        apply: () => remove(userId, roleId),
        rollback: () => assign(userId, roleId),
      })),
  ];
  const applied: typeof changes = [];

  try {
    for (const change of changes) {
      await change.apply();
      applied.push(change);
    }
  } catch (error) {
    let rollbackFailed = false;
    for (const change of [...applied].reverse()) {
      await change.rollback().catch(() => {
        rollbackFailed = true;
      });
    }
    const message = error instanceof Error ? error.message : String(error);
    if (rollbackFailed) {
      throw new Error(
        `Role update partially failed, and rollback could not fully restore the previous roles. Refresh to see the current assignments. Original error: ${message}`,
      );
    }
    if (applied.length > 0) {
      throw new Error(`Role update failed; completed changes were rolled back. ${message}`);
    }
    throw error;
  }
}

export function useUsers(query: GetUsersData["query"], enabled: boolean) {
  return useQuery({
    queryKey: [...userKeys.all, "list", query],
    queryFn: () => getUsers({ query }),
    enabled,
  });
}

// GET /users has no search and caps a page at 200, so pickers page through it and filter locally.
export function useUserPages(enabled: boolean) {
  return useInfiniteQuery({
    queryKey: [...userKeys.all, "pages"],
    queryFn: ({ pageParam }) => getUsers({ query: { limit: 50, offset: pageParam } }),
    initialPageParam: 0,
    getNextPageParam: (last, pages) => {
      const loaded = pages.reduce((n, page) => n + page.items.length, 0);
      return loaded < last.total ? loaded : undefined;
    },
    enabled,
  });
}

export function useRolesCatalog(enabled: boolean) {
  return useQuery({
    queryKey: [...userKeys.all, "roles-catalog"],
    queryFn: () => getRoles(),
    enabled,
  });
}

export function useDirectPrivileges(userId: string | undefined, enabled: boolean) {
  return useQuery({
    queryKey: [...userKeys.all, "direct-privileges", userId],
    queryFn:
      userId && enabled
        ? () => getUsersByIdPrivileges({ path: { id: userId } })
        : skipToken,
  });
}

export function useRoleDetails(
  roleIds: string[],
  enabled: boolean,
): { details: GetRolesByIdResponse[]; isLoading: boolean; isError: boolean } {
  const results = useQueries({
    queries: roleIds.map((roleId) => ({
      queryKey: [...userKeys.all, "role-detail", roleId],
      queryFn: () => getRolesById({ path: { id: roleId } }),
      enabled,
    })),
  });
  return {
    details: results.flatMap((result) => (result.data ? [result.data] : [])),
    isLoading: enabled && results.some((result) => result.isLoading),
    isError: enabled && results.some((result) => result.isError),
  };
}

export function useUserPageDetails(
  users: User[],
  rolesCatalog: Role[],
  canManageRoles: boolean,
): UserManagementRow[] {
  const identityResults = useQueries({
    queries: users.map((user) => ({
      queryKey: userKeys.identities(user.id),
      queryFn: () => getUsersByIdUserIdentities({ path: { id: user.id } }),
    })),
  });
  const roleResults = useQueries({
    queries: users.map((user) => ({
      queryKey: userKeys.roles(user.id),
      queryFn: () => getUsersByIdRoles({ path: { id: user.id } }),
      enabled: canManageRoles,
    })),
  });
  const roleById = new Map(rolesCatalog.map((role) => [role.id, role]));

  return users.map((user, index) => {
    const identitiesQuery = identityResults[index];
    const rolesQuery = roleResults[index];
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
    mutationFn: applyUserRoleChanges,
    // Role grants also show as role holders and privilege holders on the roles page.
    onSettled: (_data, _error, variables) =>
      Promise.all([
        client.invalidateQueries({ queryKey: userKeys.roles(variables.userId) }),
        client.invalidateQueries({ queryKey: [...userKeys.all, "privilege-holders"] }),
        client.invalidateQueries({ queryKey: roleKeys.all }),
        client.invalidateQueries({ queryKey: identityKeys.privileges() }),
        client.invalidateQueries({ queryKey: identityKeys.access(variables.userId) }),
      ]),
  });
}

export function useUser(userId: string | undefined) {
  return useQuery({
    queryKey: [...userKeys.all, "detail", userId],
    queryFn: userId ? () => getUsersById({ path: { id: userId } }) : skipToken,
  });
}

// A user who may sit outside the loaded page, such as one an identity lookup resolved.
export function useUserRow(
  userId: string | undefined,
  rolesCatalog: Role[],
  canManageRoles: boolean,
): UserManagementRow | undefined {
  const query = useUser(userId);
  return useUserPageDetails(query.data ? [query.data] : [], rolesCatalog, canManageRoles)[0];
}

// Onboarding also creates the user's cluster account and, for a portal admin, the admin role grant.
export function useCreateUser() {
  return useInvalidating(
    (body: CreateUserRequest) => postUsers({ body }),
    userKeys.all,
    clusterAccountKeys.all,
    [...clusterKeys.all, "users"],
    roleKeys.all,
  );
}

export function useUpdateUserStatus() {
  return useInvalidating(putUsersByIdStatus<true>, userKeys.all);
}

// A merge moves identities, cluster accounts, projects and memberships, so every read refreshes.
export function useMergeUsers() {
  return useInvalidating((body: MergeUsersRequest) => postUsersMerge({ body }), []);
}

// Direct grants change effective privileges, the caller's included.
export function useGrantPrivilege(userId: string) {
  return useInvalidating(
    (body: GrantPrivilegeRequest) => postUsersByIdPrivileges({ path: { id: userId }, body }),
    userKeys.all,
    identityKeys.all,
  );
}

export function useRevokePrivilege(userId: string) {
  return useInvalidating(
    (key: PrivilegeKey) => deleteUsersByIdPrivilegesByKey({ path: { id: userId, key } }),
    userKeys.all,
    identityKeys.all,
  );
}

export function usePrivilegeHolders(key: PrivilegeKey | undefined) {
  return useQuery({
    queryKey: [...userKeys.all, "privilege-holders", key],
    queryFn: key
      ? () => getPrivilegesByKeyHolders({ path: { key } })
      : skipToken,
  });
}

export function useSaveIdentity(userId: string) {
  return useInvalidating(
    (body: UserIdentity) =>
      body.id
        ? putUserIdentitiesById({ path: { id: body.id }, body })
        : postUserIdentities({ body: { ...body, user_id: userId } }),
    userKeys.identities(userId),
    identityKeys.identities(userId),
  );
}

export function useDeleteIdentity(userId: string) {
  return useInvalidating(
    (id: string) => deleteUserIdentitiesById({ path: { id } }),
    userKeys.identities(userId),
    identityKeys.identities(userId),
  );
}

export function lookupIdentity(lookup: IdentityLookup): Promise<UserIdentity> {
  switch (lookup.by) {
    case "id":
      return getUserIdentitiesById({ path: { id: lookup.id } });
    case "oidc":
      return getUserIdentitiesOidcSubjectsByOidcSub({ path: { oidcSub: lookup.oidcSub } });
    case "external":
      return getUserIdentitiesSourcesBySourceExternalByExternalId({
        path: { source: lookup.source, externalId: lookup.externalId },
      });
  }
}

export function useUserClusterAccounts(userId: string) {
  return useQuery({
    queryKey: [...userKeys.all, "cluster-accounts", userId],
    queryFn: () => getUsersByIdComputeClusterUsers({ path: { id: userId } }),
  });
}

export function useUserMemberships(userId: string) {
  return useQuery({
    queryKey: [...userKeys.all, "memberships", userId],
    queryFn: () => getUsersByIdComputeAllocationMemberships({ path: { id: userId } }),
  });
}

export function useUserUsages(userId: string) {
  return useQuery({
    queryKey: [...userKeys.all, "usages", userId],
    queryFn: () => getUsersByIdComputeAllocationUsages({ path: { id: userId } }),
  });
}

export function useUserChangeRequests(userId: string) {
  return useQuery({
    queryKey: [...userKeys.all, "change-requests", userId],
    queryFn: () => getUsersByIdChangeRequests({ path: { id: userId } }),
  });
}
