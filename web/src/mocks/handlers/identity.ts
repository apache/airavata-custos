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
import fixture from "@/features/core/identity/__fixtures__/settings.json";
import { effectivePrivileges } from "./privileges";

type RoleDetail = (typeof fixture.roleDetails)[keyof typeof fixture.roleDetails];

// Per-run mutable copy so PUT /users/{id} name edits are visible on the next
// /me read within a session.
let user = { ...fixture.user };

export const identityHandlers = [
  http.get("*/api/v1/me", () => HttpResponse.json({ user, privileges: effectivePrivileges() })),
  http.get("*/api/v1/users/:id/user-identities", ({ params }) =>
    HttpResponse.json(String(params.id) === fixture.user.id ? fixture.identities : []),
  ),
  http.get("*/api/v1/users/:id/roles", ({ params }) =>
    HttpResponse.json(String(params.id) === fixture.user.id ? fixture.roles : []),
  ),
  http.get("*/api/v1/roles/:roleId", ({ params }) => {
    const detail = (fixture.roleDetails as Record<string, RoleDetail | undefined>)[
      String(params.roleId)
    ];
    if (!detail) return HttpResponse.json({ error: "not found" }, { status: 404 });
    return HttpResponse.json(detail);
  }),
  http.get("*/api/v1/users/:id/privileges", ({ params }) =>
    HttpResponse.json(String(params.id) === fixture.user.id ? fixture.direct : []),
  ),
  http.put("*/api/v1/users/:id", async ({ request }) => {
    const body = (await request.json()) as {
      first_name?: string;
      middle_name?: string;
      last_name?: string;
    };
    user = { ...user, ...body };
    return HttpResponse.json(user);
  }),
];
