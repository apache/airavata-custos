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

import { identityKeys, refreshAfterPrivilegeChange } from "@/features/core/identity/queries";
import { userKeys } from "@/features/core/users/queries";
import {
  deleteRolesById,
  deleteRolesByIdPrivilegesByKey,
  deleteUsersByIdRolesByRoleId,
  getPrivilegesCatalog,
  getRoles,
  getRolesById,
  getRolesByIdHolders,
  getUsersById,
  postRoles,
  postRolesByIdPrivileges,
  postUsersByIdRoles,
  putRolesById,
} from "@/generated/core/sdk.gen";
import type { CreateRoleRequest, PrivilegeKey, Role } from "@/generated/core/types.gen";
import { type KnownError, errorCopy } from "@/shared/api/client";
import { listOf } from "@/shared/api/listOf";
import { edited } from "@/shared/ui/FormDialog";
import { displayName } from "@/shared/users-admin/displayName";
import {
  type QueryClient,
  skipToken,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";

export type RoleRow = Role & { privileges: PrivilegeKey[]; holderIds: string[] };

type RoleInput = CreateRoleRequest & { privileges: PrivilegeKey[]; memberUserIds: string[] };

export const roleKeys = {
  all: ["roles"] as const,
  rows: () => [...roleKeys.all, "rows"] as const,
  row: (id: string | undefined) => [...roleKeys.all, "row", id] as const,
  privileges: () => [...roleKeys.all, "privileges"] as const,
};

async function roleRow(id: string): Promise<RoleRow> {
  const [{ role, privileges }, holders] = await Promise.all([
    getRolesById({ path: { id } }),
    getRolesByIdHolders({ path: { id } }).then(listOf),
  ]);
  return {
    ...role,
    privileges: listOf(privileges),
    holderIds: holders.map((holder) => holder.user_id),
  };
}

async function listRoleRows(): Promise<RoleRow[]> {
  const roles = listOf(await getRoles());
  return Promise.all(roles.map((role) => roleRow(role.id)));
}

// A role save is several requests (the role, its privileges, its holders); any of them can refuse.
const ROLE_ERRORS: KnownError[] = [
  {
    status: 409,
    when: /role "[^"]*" already exists/,
    copy: "A role with this name already exists.",
  },
  { status: 400, when: /system role/, copy: "System roles keep their name and cannot be deleted." },
  {
    status: 400,
    when: /last source of|would leave no holder/,
    copy: "This change would leave a privilege with no holder.",
  },
  { status: 404, when: /role "[^"]*" does not exist/, copy: "This role no longer exists." },
];

// Thrown once the role row is saved but some privilege or member changes
// failed; `role` holds what was applied, so a retry diffs against it.
export class RoleSaveError extends Error {
  constructor(
    public readonly role: RoleRow,
    failed: string[],
  ) {
    super(`Role saved, but these changes failed: ${failed.join("; ")}`);
  }
}

// Fetched only for a failure message, so the message names the user rather than an id.
function userLabel(id: string): Promise<string> {
  return getUsersById({ path: { id } }).then(displayName, () => id);
}

// Applies the form's privilege and member sets one call at a time, attempting
// every step and reporting the failures together.
async function reconcile(current: RoleRow, input: RoleInput): Promise<RoleRow> {
  const id = current.id;
  let { privileges, holderIds } = current;
  const failed: string[] = [];
  async function attempt(
    label: string | (() => Promise<string>),
    call: () => Promise<unknown>,
  ): Promise<boolean> {
    try {
      await call();
      return true;
    } catch (err) {
      const text = typeof label === "string" ? label : await label();
      failed.push(`${text} (${errorCopy(err, ROLE_ERRORS)})`);
      return false;
    }
  }
  for (const privilege of input.privileges.filter((p) => !current.privileges.includes(p))) {
    const body = { privilege };
    if (await attempt(`grant ${privilege}`, () => postRolesByIdPrivileges({ path: { id }, body })))
      privileges = [...privileges, privilege];
  }
  for (const key of current.privileges.filter((p) => !input.privileges.includes(p))) {
    if (await attempt(`revoke ${key}`, () => deleteRolesByIdPrivilegesByKey({ path: { id, key } })))
      privileges = privileges.filter((p) => p !== key);
  }
  for (const userId of input.memberUserIds.filter((u) => !current.holderIds.includes(u))) {
    const options = { path: { id: userId }, body: { role_id: id } };
    if (
      await attempt(
        async () => `add ${await userLabel(userId)}`,
        () => postUsersByIdRoles(options),
      )
    )
      holderIds = [...holderIds, userId];
  }
  for (const userId of current.holderIds.filter((u) => !input.memberUserIds.includes(u))) {
    const options = { path: { id: userId, roleId: id } };
    if (
      await attempt(
        async () => `remove ${await userLabel(userId)}`,
        () => deleteUsersByIdRolesByRoleId(options),
      )
    )
      holderIds = holderIds.filter((u) => u !== userId);
  }
  const saved = { ...current, privileges, holderIds };
  if (failed.length > 0) throw new RoleSaveError(saved, failed);
  return saved;
}

// Role edits change holders' roles and effective privileges, the caller's included.
function invalidateRoleData(client: QueryClient) {
  return refreshAfterPrivilegeChange(client, roleKeys.all, userKeys.all, identityKeys.all);
}

export function useRoleRow(id: string | undefined) {
  return useQuery({
    queryKey: roleKeys.row(id),
    queryFn: id ? () => roleRow(id) : skipToken,
  });
}

export function useRoleRows() {
  return useQuery({ queryKey: roleKeys.rows(), queryFn: listRoleRows });
}

// GET /privileges/catalog is gated on core:privileges:grant, not core:roles:manage.
export function usePrivilegeCatalog(enabled: boolean) {
  return useQuery({
    queryKey: roleKeys.privileges(),
    queryFn: () => getPrivilegesCatalog(),
    enabled,
  });
}

export function useCreateRole() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: async (input: RoleInput) => {
      const role = await postRoles({
        body: { name: input.name, description: input.description },
      });
      return reconcile({ ...role, privileges: [], holderIds: [] }, input);
    },
    meta: { errors: ROLE_ERRORS },
    onSettled: () => invalidateRoleData(client),
  });
}

// The open dialog closes and the role's row query goes before the refresh, which would 404 on it.
export function useDeleteRole(onDeleted: () => void) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteRolesById({ path: { id } }),
    onSuccess: (_data, id) => {
      onDeleted();
      client.removeQueries({ queryKey: roleKeys.row(id) });
    },
    meta: { errors: ROLE_ERRORS },
    onSettled: () => invalidateRoleData(client),
  });
}

export function useUpdateRole() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: async ({ role, input }: { role: RoleRow; input: RoleInput }) => {
      // PUT replaces the name and description together; a blank description keeps the stored one.
      const description = role.description ?? "";
      const body = edited(
        { name: role.name, description },
        { name: input.name, description: input.description || description },
      );
      const updated = body ? await putRolesById({ path: { id: role.id }, body }) : {};
      return reconcile({ ...role, ...updated, id: role.id }, input);
    },
    meta: { errors: ROLE_ERRORS },
    onSettled: () => invalidateRoleData(client),
  });
}
