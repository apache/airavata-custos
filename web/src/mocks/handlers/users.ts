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
import type { PostUsersData, User } from "@/generated/core/types.gen";
import { zUser } from "@/generated/core/zod.gen";
import { http, HttpResponse } from "msw";
import { page } from "../paging";

const users = [zUser.parse(fixture.user)];

export const usersHandlers = [
  http.get("*/api/v1/users", ({ request }) => {
    const { items, total } = page(new URL(request.url), users);
    return HttpResponse.json({ items, total });
  }),
  http.post<never, PostUsersData["body"]>("*/api/v1/users", async ({ request }) => {
    const { email = "", first_name = "", last_name = "", portal_admin } = await request.json();
    if (users.some((u) => u.email === email)) {
      return HttpResponse.json({ error: "user already exists" }, { status: 409 });
    }
    const user: User = {
      id: crypto.randomUUID(),
      organization_id: fixture.user.organization_id,
      email,
      first_name,
      last_name,
      status: "PENDING",
      type: portal_admin ? "SYSTEM" : "CLUSTER_LOCAL",
    };
    users.push(user);
    return HttpResponse.json(user, { status: 201 });
  }),
];
