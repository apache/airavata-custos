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

import { http, HttpResponse } from "msw";
import settingsFixture from "@/features/core/identity/__fixtures__/settings.json";
import type {
  CallerRoleGrant,
  DeleteRolesByIdPrivilegesByKeyData,
  DeleteUsersByIdRolesByRoleIdData,
  GetRolesByIdData,
  GetRolesByIdHoldersData,
  GetUsersByIdRolesData,
  PostRolesByIdPrivilegesData,
  PostRolesData,
  PostUsersByIdRolesData,
  PrivilegeKey,
  PutRolesByIdData,
  Role,
  UserRole,
} from "@/generated/core/types.gen";
import { zGetRolesByIdResponse, zPrivilegeKey, zUserRole } from "@/generated/core/zod.gen";
import { notFound } from "../paging";
import { z } from "zod";
import { CONNECTOR_PRIVILEGES } from "./privileges";

const PRIVILEGES: PrivilegeKey[] = zPrivilegeKey.options;
// The backend catalog also lists connector-registered keys.
const CATALOG: string[] = [...PRIVILEGES, ...CONNECTOR_PRIVILEGES];

type MockRole = Role & { privileges: PrivilegeKey[]; holders: UserRole[] };

let nextRoleId = 4;

const settingsGrants: UserRole[] = z.array(zUserRole).parse(settingsFixture.roles);

const initialRoles: MockRole[] = [
    {
      id: "role-super-admin",
      name: "Super Admin",
      description: "Full administrative access across the portal.",
      is_system: true,
      created_at: "2026-01-15T09:00:00Z",
      privileges: PRIVILEGES,
      holders: [
        { user_id: "u1", role_id: "role-super-admin", granted_at: "2026-01-15T09:00:00Z" },
      ],
    },
    {
      id: "role-operator",
      name: "Operator",
      description: "Day-to-day allocation and project operations.",
      is_system: false,
      created_at: "2026-01-15T09:00:00Z",
      privileges: [
        "core:allocations:read",
        "core:allocations:write",
        "core:projects:read",
        "core:projects:write",
        "core:clusters:read",
      ],
      holders: [
        { user_id: "u2", role_id: "role-operator", granted_at: "2026-01-15T09:00:00Z" },
        { user_id: "u4", role_id: "role-operator", granted_at: "2026-01-15T09:00:00Z" },
      ],
    },
    {
      id: "role-auditor",
      name: "Auditor",
      description: "Read-only access across allocations, projects, and tracing.",
      is_system: false,
      created_at: "2026-01-15T09:00:00Z",
      privileges: ["core:allocations:read", "core:projects:read", "core:traces:read"],
      holders: [
        { user_id: "u3", role_id: "role-auditor", granted_at: "2026-01-15T09:00:00Z" },
      ],
    },
    ...z
      .array(zGetRolesByIdResponse)
      .parse(Object.values(settingsFixture.roleDetails))
      .map(({ role, privileges }) => ({
        ...role,
        privileges: privileges ?? [],
        holders: settingsGrants.filter((g) => g.role_id === role.id),
      })),
  ];

const roles = new Map<string, MockRole>(initialRoles.map((role) => [role.id, role]));

function publicRole(role: MockRole): Role {
  const { privileges: _privileges, holders: _holders, ...rest } = role;
  return rest;
}

const nameTaken = (name: string | undefined, exceptId?: string) =>
  Array.from(roles.values()).some(
    (role) => role.id !== exceptId && role.name?.toLowerCase() === name?.toLowerCase(),
  );

function grantsHeldBy(userId: string): Array<{ role: MockRole; grant: UserRole }> {
  return Array.from(roles.values()).flatMap((role) =>
    role.holders.filter((grant) => grant.user_id === userId).map((grant) => ({ role, grant })),
  );
}

// Shared with /me so role assignments stay consistent across handlers.
export function callerRoleGrants(userId: string): CallerRoleGrant[] {
  return grantsHeldBy(userId).map(({ role, grant }) => ({
    role: publicRole(role),
    privileges: role.privileges,
    granted_at: grant.granted_at,
  }));
}

export const rolesHandlers = [
  http.get("*/api/v1/privileges/catalog", () => HttpResponse.json(CATALOG)),

  http.get("*/api/v1/roles", () => HttpResponse.json(Array.from(roles.values()).map(publicRole))),

  http.post<never, PostRolesData["body"]>("*/api/v1/roles", async ({ request }) => {
    const body = await request.json();
    const name = body.name?.trim();
    if (!name) return HttpResponse.json({ error: "role name is required" }, { status: 400 });
    if (nameTaken(name))
      return HttpResponse.json({ error: "role name already exists" }, { status: 409 });

    const id = `role-custom-${nextRoleId++}`;
    const role: MockRole = {
      id,
      name,
      description: body.description?.trim() ?? "",
      is_system: false,
      created_at: new Date().toISOString(),
      privileges: [],
      holders: [],
    };
    roles.set(id, role);
    return HttpResponse.json(publicRole(role), { status: 201 });
  }),

  http.get<GetRolesByIdData["path"]>("*/api/v1/roles/:id", ({ params }) => {
    const role = roles.get(params.id);
    if (!role) return notFound("role");
    return HttpResponse.json({ role: publicRole(role), privileges: role.privileges });
  }),

  http.put<PutRolesByIdData["path"], PutRolesByIdData["body"]>(
    "*/api/v1/roles/:id",
    async ({ params, request }) => {
      const role = roles.get(params.id);
      if (!role) return notFound("role");
      // Blank fields keep the stored value; only system-role renames are refused.
      const body = await request.json();
      const name = body.name?.trim() || role.name;
      if (role.is_system && name !== role.name) {
        return HttpResponse.json({ error: "cannot rename system role" }, { status: 400 });
      }
      if (nameTaken(name, role.id))
        return HttpResponse.json({ error: "role name already exists" }, { status: 409 });

      role.name = name;
      role.description = body.description?.trim() || role.description;
      return HttpResponse.json(publicRole(role));
    },
  ),

  http.get<GetRolesByIdHoldersData["path"]>("*/api/v1/roles/:id/holders", ({ params }) => {
    const role = roles.get(params.id);
    if (!role) return notFound("role");
    return HttpResponse.json(role.holders);
  }),

  http.get<GetUsersByIdRolesData["path"]>("*/api/v1/users/:id/roles", ({ params }) =>
    HttpResponse.json(grantsHeldBy(params.id).map(({ grant }) => grant)),
  ),

  http.post<PostUsersByIdRolesData["path"], PostUsersByIdRolesData["body"]>(
    "*/api/v1/users/:id/roles",
    async ({ params, request }) => {
      const body = await request.json();
      const role = roles.get(body.role_id ?? "");
      if (!role) return notFound("role");
      const userId = params.id;
      if (role.holders.some((holder) => holder.user_id === userId)) {
        return HttpResponse.json({ error: "user already holds that role" }, { status: 409 });
      }
      const holder: UserRole = {
        user_id: userId,
        role_id: role.id,
        granted_at: new Date().toISOString(),
      };
      role.holders = [...role.holders, holder];
      return HttpResponse.json(holder, { status: 201 });
    },
  ),

  http.delete<DeleteUsersByIdRolesByRoleIdData["path"]>(
    "*/api/v1/users/:id/roles/:roleId",
    ({ params }) => {
      const role = roles.get(params.roleId);
      if (!role) return notFound("role");
      const userId = params.id;
      if (!role.holders.some((holder) => holder.user_id === userId)) {
        return HttpResponse.json({ error: "user does not hold that role" }, { status: 404 });
      }
      role.holders = role.holders.filter((holder) => holder.user_id !== userId);
      return new HttpResponse(null, { status: 204 });
    },
  ),

  http.post<PostRolesByIdPrivilegesData["path"], PostRolesByIdPrivilegesData["body"]>(
    "*/api/v1/roles/:id/privileges",
    async ({ params, request }) => {
      const role = roles.get(params.id);
      if (!role) return notFound("role");
      const { privilege } = await request.json();
      if (!privilege || !CATALOG.includes(privilege)) {
        return HttpResponse.json({ error: "unknown privilege" }, { status: 400 });
      }
      if (role.privileges.includes(privilege)) {
        return HttpResponse.json({ error: "role already carries that privilege" }, { status: 409 });
      }
      role.privileges = [...role.privileges, privilege].sort();
      return new HttpResponse(null, { status: 204 });
    },
  ),

  http.delete<DeleteRolesByIdPrivilegesByKeyData["path"]>(
    "*/api/v1/roles/:id/privileges/:key",
    ({ params: { id, key } }) => {
      const role = roles.get(id);
      if (!role) return notFound("role");
      if (!role.privileges.includes(key)) {
        return HttpResponse.json({ error: "role does not carry that privilege" }, { status: 404 });
      }
      role.privileges = role.privileges.filter((privilege) => privilege !== key);
      return new HttpResponse(null, { status: 204 });
    },
  ),
];
