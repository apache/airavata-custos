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

import { apiFetch } from "@/shared/api/client";
import {
  type PrivilegeKey,
  type Role,
  type RoleInput,
  type RoleRow,
  privilegeCatalogResponseSchema,
  roleDetailResponseSchema,
  roleHoldersResponseSchema,
  roleInputSchema,
  roleSchema,
  rolesResponseSchema,
} from "./schemas";

export async function listRoles(): Promise<Role[]> {
  return rolesResponseSchema.parse(await apiFetch("/roles"));
}

export async function getRolePrivileges(roleId: string): Promise<PrivilegeKey[]> {
  const detail = roleDetailResponseSchema.parse(await apiFetch(`/roles/${roleId}`));
  return detail.privileges;
}

export async function getRoleMemberCount(roleId: string): Promise<number> {
  const holders = roleHoldersResponseSchema.parse(await apiFetch(`/roles/${roleId}/holders`));
  return holders.length;
}

export async function listPrivilegeCatalog(): Promise<PrivilegeKey[]> {
  return privilegeCatalogResponseSchema.parse(await apiFetch("/privileges/catalog"));
}

export async function listRoleRows(): Promise<RoleRow[]> {
  const roles = await listRoles();
  return Promise.all(
    roles.map(async (role) => {
      const roleId = role.id ?? "";
      if (!roleId) return { ...role, privileges: [], memberCount: 0 };
      const [privileges, memberCount] = await Promise.all([
        getRolePrivileges(roleId),
        getRoleMemberCount(roleId),
      ]);
      return { ...role, privileges, memberCount };
    }),
  );
}

export async function createRole(input: RoleInput): Promise<Role> {
  const payload = roleInputSchema.parse(input);
  return roleSchema.parse(
    await apiFetch("/roles", {
      method: "POST",
      body: { name: payload.name, description: payload.description },
    }),
  );
}

export async function updateRole(roleId: string, input: RoleInput): Promise<Role> {
  const payload = roleInputSchema.parse(input);
  return roleSchema.parse(
    await apiFetch(`/roles/${roleId}`, {
      method: "PUT",
      body: { name: payload.name, description: payload.description },
    }),
  );
}

export async function addPrivilegeToRole(
  roleId: string,
  privilege: PrivilegeKey,
): Promise<void> {
  await apiFetch(`/roles/${roleId}/privileges`, {
    method: "POST",
    body: { privilege },
  });
}

export async function removePrivilegeFromRole(
  roleId: string,
  privilege: PrivilegeKey,
): Promise<void> {
  await apiFetch(`/roles/${roleId}/privileges/${encodeURIComponent(privilege)}`, {
    method: "DELETE",
  });
}

export async function reconcileRolePrivileges(
  roleId: string,
  current: PrivilegeKey[],
  next: PrivilegeKey[],
): Promise<void> {
  const currentSet = new Set(current);
  const nextSet = new Set(next);
  const toAdd = next.filter((key) => !currentSet.has(key));
  const toRemove = current.filter((key) => !nextSet.has(key));

  for (const privilege of toAdd) {
    await addPrivilegeToRole(roleId, privilege);
  }
  for (const privilege of toRemove) {
    await removePrivilegeFromRole(roleId, privilege);
  }
}
