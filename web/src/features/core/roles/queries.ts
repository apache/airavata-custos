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

import { identityKeys } from "@/features/core/identity/queries";
import { userKeys } from "@/features/core/users/queries";
import {
  deleteRolesById,
  deleteRolesByIdPrivilegesByKey,
  deleteUsersByIdRolesByRoleId,
  getPrivilegesCatalog,
  getRoles,
  getRolesById,
  getRolesByIdHolders,
  postRoles,
  postRolesByIdPrivileges,
  postUsersByIdRoles,
  putRolesById,
} from "@/generated/core/sdk.gen";
import type { CreateRoleRequest, PrivilegeKey, Role } from "@/generated/core/types.gen";
import { type QueryClient, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

// A role joined with its privilege bundle and holders for the role cards.
export type RoleRow = Role & { privileges: PrivilegeKey[]; holderIds: string[] };

type RoleInput = CreateRoleRequest & { privileges: PrivilegeKey[]; memberUserIds: string[] };

export const roleKeys = {
  all: ["roles"] as const,
  rows: () => [...roleKeys.all, "rows"] as const,
  privileges: () => [...roleKeys.all, "privileges"] as const,
};

async function listRoleRows(): Promise<RoleRow[]> {
  const roles = await getRoles();
  return Promise.all(
    roles.map(async (role) => {
      const path = { id: role.id };
      const [detail, holders] = await Promise.all([
        getRolesById({ path }),
        getRolesByIdHolders({ path }),
      ]);
      return {
        ...role,
        privileges: detail.privileges ?? [],
        holderIds: holders.map((holder) => holder.user_id),
      };
    }),
  );
}

// Thrown once the role row is saved but some privilege or member changes
// failed; `role` holds what was applied, so a retry diffs against it.
export class RoleSaveError extends Error {
  constructor(
    public readonly role: RoleRow,
    failed: string[],
  ) {
    super(`Role saved, but these changes failed: ${failed.join("; ")}`);
    this.name = "RoleSaveError";
  }
}

// Applies the form's privilege and member sets one call at a time, attempting
// every step and reporting the failures together.
async function reconcile(current: RoleRow, input: RoleInput): Promise<RoleRow> {
  const id = current.id;
  let { privileges, holderIds } = current;
  const failed: string[] = [];
  async function attempt(label: string, call: () => Promise<unknown>): Promise<boolean> {
    try {
      await call();
      return true;
    } catch (err) {
      failed.push(`${label} (${err instanceof Error ? err.message : String(err)})`);
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
    if (await attempt(`add member ${userId}`, () => postUsersByIdRoles(options)))
      holderIds = [...holderIds, userId];
  }
  for (const userId of current.holderIds.filter((u) => !input.memberUserIds.includes(u))) {
    const options = { path: { id: userId, roleId: id } };
    if (await attempt(`remove member ${userId}`, () => deleteUsersByIdRolesByRoleId(options)))
      holderIds = holderIds.filter((u) => u !== userId);
  }
  const saved = { ...current, privileges, holderIds };
  if (failed.length > 0) throw new RoleSaveError(saved, failed);
  return saved;
}

// Role edits change holders' roles and effective privileges, the caller's included.
function invalidateRoleData(client: QueryClient) {
  return Promise.all([
    client.invalidateQueries({ queryKey: roleKeys.all }),
    client.invalidateQueries({ queryKey: userKeys.all }),
    client.invalidateQueries({ queryKey: identityKeys.all }),
  ]);
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
      const role = await postRoles({ body: { name: input.name, description: input.description } });
      return reconcile({ ...role, privileges: [], holderIds: [] }, input);
    },
    onSettled: () => invalidateRoleData(client),
  });
}

export function useDeleteRole() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteRolesById({ path: { id } }),
    onSettled: () => invalidateRoleData(client),
  });
}

export function useUpdateRole() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: async ({ role, input }: { role: RoleRow; input: RoleInput }) => {
      const updated = await putRolesById({
        path: { id: role.id },
        body: { name: input.name, description: input.description },
      });
      return reconcile({ ...role, ...updated }, input);
    },
    onSettled: () => invalidateRoleData(client),
  });
}
