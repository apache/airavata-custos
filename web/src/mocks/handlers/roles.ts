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
import type { PrivilegeKey, Role, UserRole } from "@/generated/core/types.gen";

const PRIVILEGES: PrivilegeKey[] = [
  "core:allocations:read",
  "core:allocations:write",
  "core:clusters:read",
  "core:clusters:write",
  "core:organizations:read",
  "core:organizations:write",
  "core:privileges:grant",
  "core:projects:read",
  "core:projects:write",
  "core:roles:manage",
  "core:traces:read",
  "core:users:read",
  "core:users:write",
];

type MockRole = Role & { privileges: PrivilegeKey[]; holders: UserRole[] };

let nextRoleId = 4;

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
  ];

const roles = new Map<string, MockRole>(initialRoles.map((role) => [role.id ?? "", role]));

function publicRole(role: MockRole): Role {
  const { privileges: _privileges, holders: _holders, ...rest } = role;
  return rest;
}

function roleById(id: string): MockRole | undefined {
  return roles.get(id);
}

export const rolesHandlers = [
  http.get("*/api/v1/privileges/catalog", () => HttpResponse.json(PRIVILEGES)),

  http.get("*/api/v1/roles", () => HttpResponse.json(Array.from(roles.values()).map(publicRole))),

  http.post("*/api/v1/roles", async ({ request }) => {
    const body = (await request.json()) as { name?: string; description?: string };
    const name = body.name?.trim();
    if (!name) return HttpResponse.json({ error: "role name is required" }, { status: 400 });
    const duplicate = Array.from(roles.values()).some(
      (role) => role.name?.toLowerCase() === name.toLowerCase(),
    );
    if (duplicate) return HttpResponse.json({ error: "role name already exists" }, { status: 409 });

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

  http.get("*/api/v1/roles/:roleId", ({ params }) => {
    const role = roleById(String(params.roleId));
    if (!role) return HttpResponse.json({ error: "not found" }, { status: 404 });
    return HttpResponse.json({ role: publicRole(role), privileges: role.privileges });
  }),

  http.put("*/api/v1/roles/:roleId", async ({ params, request }) => {
    const role = roleById(String(params.roleId));
    if (!role) return HttpResponse.json({ error: "not found" }, { status: 404 });
    if (role.is_system) {
      return HttpResponse.json({ error: "system roles cannot be renamed" }, { status: 400 });
    }

    const body = (await request.json()) as { name?: string; description?: string };
    const name = body.name?.trim();
    if (!name) return HttpResponse.json({ error: "role name is required" }, { status: 400 });
    const duplicate = Array.from(roles.values()).some(
      (candidate) =>
        candidate.id !== role.id && candidate.name?.toLowerCase() === name.toLowerCase(),
    );
    if (duplicate) return HttpResponse.json({ error: "role name already exists" }, { status: 409 });

    role.name = name;
    role.description = body.description?.trim() ?? "";
    return HttpResponse.json(publicRole(role));
  }),

  http.get("*/api/v1/roles/:roleId/holders", ({ params }) => {
    const role = roleById(String(params.roleId));
    if (!role) return HttpResponse.json({ error: "not found" }, { status: 404 });
    return HttpResponse.json(role.holders);
  }),

  http.post("*/api/v1/users/:userId/roles", async ({ params, request }) => {
    const body = (await request.json()) as { role_id?: string };
    const role = body.role_id ? roleById(body.role_id) : undefined;
    if (!role) return HttpResponse.json({ error: "not found" }, { status: 404 });
    const userId = String(params.userId);
    if (role.holders.some((holder) => holder.user_id === userId)) {
      return HttpResponse.json({ error: "user already holds that role" }, { status: 409 });
    }
    const holder = {
      user_id: userId,
      role_id: role.id,
      granted_at: new Date().toISOString(),
    };
    role.holders = [...role.holders, holder];
    return HttpResponse.json(holder, { status: 201 });
  }),

  http.delete("*/api/v1/users/:userId/roles/:roleId", ({ params }) => {
    const role = roleById(String(params.roleId));
    if (!role) return HttpResponse.json({ error: "not found" }, { status: 404 });
    const userId = String(params.userId);
    if (!role.holders.some((holder) => holder.user_id === userId)) {
      return HttpResponse.json({ error: "user does not hold that role" }, { status: 404 });
    }
    role.holders = role.holders.filter((holder) => holder.user_id !== userId);
    return new HttpResponse(null, { status: 204 });
  }),

  http.post("*/api/v1/roles/:roleId/privileges", async ({ params, request }) => {
    const role = roleById(String(params.roleId));
    if (!role) return HttpResponse.json({ error: "not found" }, { status: 404 });
    const body = (await request.json()) as { privilege?: PrivilegeKey };
    const privilege = body.privilege;
    if (!privilege || !PRIVILEGES.includes(privilege)) {
      return HttpResponse.json({ error: "unknown privilege" }, { status: 400 });
    }
    if (role.privileges.includes(privilege)) {
      return HttpResponse.json({ error: "role already carries that privilege" }, { status: 409 });
    }
    role.privileges = [...role.privileges, privilege].sort();
    return new HttpResponse(null, { status: 204 });
  }),

  http.delete("*/api/v1/roles/:roleId/privileges/:key", ({ params }) => {
    const role = roleById(String(params.roleId));
    if (!role) return HttpResponse.json({ error: "not found" }, { status: 404 });
    const key = String(params.key) as PrivilegeKey;
    if (!role.privileges.includes(key)) {
      return HttpResponse.json({ error: "role does not carry that privilege" }, { status: 404 });
    }
    role.privileges = role.privileges.filter((privilege) => privilege !== key);
    return new HttpResponse(null, { status: 204 });
  }),
];
