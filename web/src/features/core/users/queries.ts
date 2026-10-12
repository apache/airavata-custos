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
import { identityKeys, refreshAfterPrivilegeChange } from "@/features/core/identity/queries";
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
  putUsersById,
  putUsersByIdStatus,
} from "@/generated/core/sdk.gen";
import type {
  CreateUserRequest,
  GetUsersData,
  GrantPrivilegeRequest,
  MergeUsersRequest,
  PrivilegeKey,
  Role,
  User,
  UserIdentity,
  UserIdentityInput,
  UserStatus,
} from "@/generated/core/types.gen";
import { type KnownError, errorCopy } from "@/shared/api/client";
import { listOf } from "@/shared/api/listOf";
import { paged } from "@/shared/api/paged";
import { useInvalidating, useInvalidatingWith } from "@/shared/api/useInvalidating";
import { isGone } from "@/shared/api/useOnGone";
import { endSession } from "@/shared/auth/useSignOut";
import { useAbility } from "@/shared/casl/AbilityProvider";
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

export type IdentityLookup =
  | { id: string }
  | { oidc_sub: string }
  | { source: string; external_id: string };

async function assign(userId: string, roleId: string, reason?: string) {
  await postUsersByIdRoles({
    path: { id: userId },
    body: { role_id: roleId, reason: reason?.trim() || undefined },
  });
}

async function remove(userId: string, roleId: string) {
  await deleteUsersByIdRolesByRoleId({ path: { id: userId, roleId } });
}

const USER_ROLE_ERRORS: KnownError[] = [
  { status: 409, when: /already holds role/, copy: "This user already holds that role." },
  { status: 404, when: /does not hold role/, copy: "This user no longer holds that role." },
  { status: 404, when: /role "[^"]*" does not exist/, copy: "That role no longer exists." },
  {
    status: 400,
    when: /would leave no holder/,
    copy: "This change would leave a privilege with no holder.",
  },
];

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
    const message = errorCopy(error, USER_ROLE_ERRORS);
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
    ...paged,
  });
}

// GET /users has no search and caps a page at 200, so pickers page through it and filter locally.
export function useUserPages(enabled: boolean) {
  return useInfiniteQuery({
    queryKey: [...userKeys.all, "pages"],
    queryFn: ({ pageParam }) => getUsers({ query: { limit: 50, offset: pageParam } }),
    initialPageParam: 0,
    getNextPageParam: (last, pages) => {
      const loaded = pages.reduce((n, page) => n + listOf(page.items).length, 0);
      return loaded < last.total ? loaded : undefined;
    },
    enabled,
  });
}

export function useRolesCatalog(enabled: boolean) {
  return useQuery({
    queryKey: [...userKeys.all, "roles-catalog"],
    queryFn: () => getRoles().then(listOf),
    enabled,
  });
}

// Gated here so every observer of this key shares one queryFn and enabled rule.
export function useDirectPrivileges(userId: string | undefined) {
  const canRead = useAbility().can("write", "PrivilegeGrant");
  return useQuery({
    queryKey: [...userKeys.all, "direct-privileges", userId],
    queryFn: userId
      ? () => getUsersByIdPrivileges({ path: { id: userId } }).then(listOf)
      : skipToken,
    enabled: canRead,
  });
}

export function useRoleDetails(roleIds: string[], enabled: boolean) {
  const results = useQueries({
    queries: roleIds.map((roleId) => ({
      queryKey: [...userKeys.all, "role-detail", roleId],
      queryFn: () => getRolesById({ path: { id: roleId } }),
      enabled,
    })),
  });
  return {
    details: results.flatMap(({ data }) =>
      data ? [{ id: data.role.id, privileges: listOf(data.privileges) }] : [],
    ),
    isLoading: enabled && results.some((result) => result.isLoading),
    isError: enabled && results.some((result) => result.isError),
  };
}

export const userIdentities = (userId: string) => ({
  queryKey: userKeys.identities(userId),
  queryFn: () => getUsersByIdUserIdentities({ path: { id: userId } }).then(listOf),
});

export function useUserPageDetails(
  page: User[],
  rolesCatalog: Role[],
  canManageRoles: boolean,
): UserManagementRow[] {
  const identityResults = useQueries({
    queries: page.map((user) => userIdentities(user.id)),
  });
  const roleResults = useQueries({
    queries: page.map((user) => ({
      queryKey: userKeys.roles(user.id),
      queryFn: () => getUsersByIdRoles({ path: { id: user.id } }).then(listOf),
      enabled: canManageRoles,
    })),
  });
  const roleById = new Map(rolesCatalog.map((role) => [role.id, role]));

  return page.map((user, index) => {
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
    meta: { errors: USER_ROLE_ERRORS },
    // Role grants also show as role holders and privilege holders on the roles page.
    onSettled: (_data, _error, { userId }) =>
      refreshAfterPrivilegeChange(
        client,
        identityKeys.access(userId),
        userKeys.roles(userId),
        [...userKeys.all, "privilege-holders"],
        roleKeys.all,
      ),
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

const CREATE_USER_ERRORS: KnownError[] = [
  { status: 409, when: /user with email/, copy: "A user with this email already exists." },
  { status: 409, when: /is taken on cluster/, copy: "This username is taken on that cluster." },
  { status: 404, when: /lookup allocation/, copy: "No allocation has that ID." },
  {
    status: 400,
    when: /allocation "[^"]*" is on cluster/,
    copy: "This allocation is on another cluster.",
  },
  { status: 400, when: /allocation "[^"]*" is [A-Z]+$/, copy: "This allocation is not active." },
  { status: 400, when: /pick a cluster, 0 are/, copy: "No cluster is registered yet." },
  { status: 400, when: /pick a cluster/, copy: "Several clusters are registered; choose one." },
  { status: 400, when: /compute cluster "[^"]*" not found/, copy: "No cluster has that ID." },
  { status: 400, when: /longer than/, copy: "A cluster username has at most 32 characters." },
];

// Onboarding also creates the user's cluster account and, for a portal admin, the admin role grant.
export function useCreateUser() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateUserRequest) => postUsers({ body }),
    meta: { errors: CREATE_USER_ERRORS },
    onSettled: () =>
      Promise.all(
        [userKeys.all, clusterAccountKeys.all, [...clusterKeys.all, "users"], roleKeys.all].map(
          (queryKey) => client.invalidateQueries({ queryKey }),
        ),
      ),
  });
}

// A status that ends the caller's own sign-in leaves the portal at once, before any refetch meets
// the 403s of the ended session.
export function useUpdateUserStatus(leaves: (status: UserStatus) => boolean) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: putUsersByIdStatus<true>,
    onSettled: (_data, error, { body }) =>
      !error && body.status && leaves(body.status)
        ? endSession("/sign-in?error=user_disabled")
        : client.invalidateQueries({ queryKey: userKeys.all }),
  });
}

// The signed-in user's own name shows in their profile too.
export function useUpdateUserName() {
  return useInvalidatingWith(
    { errors: [{ status: 404, copy: "This user no longer exists." }] },
    putUsersById<true>,
    userKeys.all,
    identityKeys.me(),
  );
}

// A merge moves identities, cluster accounts, projects and memberships, so every read refreshes.
export function useMergeUsers() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (body: MergeUsersRequest) => postUsersMerge({ body }),
    meta: {
      errors: [
        { status: 409, when: /already merged/, copy: "This user is already merged." },
        {
          status: 400,
          when: /must be ACTIVE or PENDING/,
          copy: "Only active or pending users can be merged.",
        },
        { status: 400, when: /does not exist/, copy: "One of these users no longer exists." },
      ],
    },
    onSettled: () => client.invalidateQueries(),
  });
}

type PrivilegeGrantInput = GrantPrivilegeRequest & { userId: string };

// Direct grants change effective privileges, the caller's included.
export function useGrantPrivilege() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ userId, ...body }: PrivilegeGrantInput) =>
      postUsersByIdPrivileges({ path: { id: userId }, body }),
    meta: {
      errors: [
        { status: 409, copy: "This user already holds that privilege directly." },
        // An unknown user fails the grant's foreign key, which master words only as this.
        { status: 400, when: /^invalid input$/, copy: "No user has that ID." },
      ],
    },
    onSettled: () => refreshAfterPrivilegeChange(client, userKeys.all, identityKeys.all),
  });
}

export function useRevokePrivilege() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ userId, key }: { userId: string; key: PrivilegeKey }) =>
      deleteUsersByIdPrivilegesByKey({ path: { id: userId, key } }),
    meta: {
      errors: [
        { status: 404, copy: "This user does not hold that privilege directly." },
        { status: 400, when: /self-revoke/, copy: "You cannot revoke this from yourself." },
        {
          status: 400,
          when: /last active/,
          copy: "Someone must keep the privilege to grant privileges.",
        },
      ],
    },
    onSettled: () => refreshAfterPrivilegeChange(client, userKeys.all, identityKeys.all),
  });
}

export function usePrivilegeHolders(key: PrivilegeKey | undefined) {
  return useQuery({
    queryKey: [...userKeys.all, "privilege-holders", key],
    queryFn: key ? () => getPrivilegesByKeyHolders({ path: { key } }).then(listOf) : skipToken,
    select: (rows) => rows.toSorted((a, b) => b.granted_at.localeCompare(a.granted_at)),
  });
}

// An edit's identity is the whole record, since PUT replaces it.
export function useSaveIdentity(userId: string) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (identity: UserIdentityInput) =>
      identity.id
        ? putUserIdentitiesById({ path: { id: identity.id }, body: identity })
        : postUserIdentities({ body: { ...identity, user_id: userId } }),
    meta: {
      errors: [
        { status: 409, copy: "This identity is already recorded." },
        { status: 404, copy: "This identity no longer exists." },
      ],
    },
    onSettled: () =>
      Promise.all(
        [userKeys.identities(userId), identityKeys.identities(userId)].map((queryKey) =>
          client.invalidateQueries({ queryKey }),
        ),
      ),
  });
}

export function useDeleteIdentity(userId: string) {
  return useInvalidating(
    (id: string) => deleteUserIdentitiesById({ path: { id } }),
    userKeys.identities(userId),
    identityKeys.identities(userId),
  );
}

export const lookupIdentity = (lookup: IdentityLookup): Promise<UserIdentity | undefined> =>
  ("id" in lookup
    ? getUserIdentitiesById({ path: { id: lookup.id } })
    : "oidc_sub" in lookup
      ? getUserIdentitiesOidcSubjectsByOidcSub({ path: { oidcSub: lookup.oidc_sub } })
      : getUserIdentitiesSourcesBySourceExternalByExternalId({
          path: { source: lookup.source, externalId: lookup.external_id },
        })
  ).catch((error: unknown) => {
    if (isGone(error)) return undefined;
    throw error;
  });

export function useUserClusterAccounts(userId: string) {
  return useQuery({
    queryKey: [...userKeys.all, "cluster-accounts", userId],
    queryFn: () => getUsersByIdComputeClusterUsers({ path: { id: userId } }).then(listOf),
  });
}

export function useUserMemberships(userId: string) {
  return useQuery({
    queryKey: [...userKeys.all, "memberships", userId],
    queryFn: () => getUsersByIdComputeAllocationMemberships({ path: { id: userId } }).then(listOf),
  });
}

export function useUserUsages(userId: string) {
  return useQuery({
    queryKey: [...userKeys.all, "usages", userId],
    queryFn: () => getUsersByIdComputeAllocationUsages({ path: { id: userId } }).then(listOf),
  });
}

export function useUserChangeRequests(userId: string) {
  return useQuery({
    queryKey: [...userKeys.all, "change-requests", userId],
    queryFn: () => getUsersByIdChangeRequests({ path: { id: userId } }).then(listOf),
  });
}
