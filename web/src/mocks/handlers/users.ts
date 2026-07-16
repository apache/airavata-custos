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

const roles = Object.values(fixture.roleDetails).map((detail) => detail.role);

export const usersHandlers = [
  http.get("*/api/v1/users", ({ request }) => {
    const url = new URL(request.url);
    const limit = Number(url.searchParams.get("limit") ?? 25);
    const offset = Number(url.searchParams.get("offset") ?? 0);
    const users = [fixture.user];
    return HttpResponse.json({
      items: users.slice(offset, offset + limit),
      total: users.length,
    });
  }),
  http.get("*/api/v1/roles", () => HttpResponse.json(roles)),
  http.post("*/api/v1/users/:id/roles", async ({ params, request }) => {
    const body = (await request.json()) as { role_id?: string };
    return HttpResponse.json(
      {
        user_id: String(params.id),
        role_id: body.role_id,
        granted_at: new Date().toISOString(),
      },
      { status: 201 },
    );
  }),
  http.delete("*/api/v1/users/:id/roles/:roleId", () => new HttpResponse(null, { status: 204 })),
];
