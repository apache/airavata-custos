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

import fixture from "@/features/core/identity/__fixtures__/settings.json";
import { zPostUsersBody, zPostUsersByIdRolesBody } from "@/generated/core/zod.gen";
import { http, HttpResponse } from "msw";

const roles = Object.values(fixture.roleDetails).map((detail) => detail.role);

// Mirrors the server's role_id and identity_source filters; only the signed-in user has grants.
export function matchesDirectoryFilter(userId: string, params: URLSearchParams): boolean {
  const isFixtureUser = userId === fixture.user.id;
  const roleId = params.get("role_id");
  const source = params.get("identity_source");
  return (
    (!roleId || (isFixtureUser && fixture.roles.some((r) => r.role_id === roleId))) &&
    (!source || (isFixtureUser && fixture.identities.some((i) => i.source === source)))
  );
}

export const usersHandlers = [
  http.get("*/api/v1/users", ({ request }) => {
    const url = new URL(request.url);
    const limit = Number(url.searchParams.get("limit") ?? 25);
    const offset = Number(url.searchParams.get("offset") ?? 0);
    const q = url.searchParams.get("q")?.toLowerCase() ?? "";
    // Mirrors the server's q filter.
    const users = [fixture.user].filter(
      (u) =>
        [u.email, u.first_name, u.last_name, `${u.first_name} ${u.last_name}`].some((field) =>
          field.toLowerCase().includes(q),
        ) && matchesDirectoryFilter(u.id, url.searchParams),
    );
    return HttpResponse.json({
      items: users.slice(offset, offset + limit),
      total: users.length,
    });
  }),
  http.post("*/api/v1/users", async ({ request }) => {
    const body = zPostUsersBody.parse(await request.json());
    return HttpResponse.json(
      {
        id: "user-created-1",
        organization_id: "system",
        email: body.email,
        first_name: body.first_name ?? "",
        last_name: body.last_name ?? "",
        status: "PENDING",
        type: body.portal_admin ? "SYSTEM" : "CLUSTER_LOCAL",
      },
      { status: 201 },
    );
  }),
  http.get("*/api/v1/roles", () => HttpResponse.json(roles)),
  http.post<{ id: string }>("*/api/v1/users/:id/roles", async ({ params, request }) => {
    const body = zPostUsersByIdRolesBody.parse(await request.json());
    return HttpResponse.json(
      {
        user_id: String(params.id),
        role_id: body.role_id,
        granted_at: new Date().toISOString(),
        granted_by: null,
        reason: body.reason ?? null,
      },
      { status: 201 },
    );
  }),
  http.delete("*/api/v1/users/:id/roles/:roleId", () => new HttpResponse(null, { status: 204 })),
];
