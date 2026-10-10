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
import type {
  GetUsersByIdPrivilegesData,
  GetUsersByIdUserIdentitiesData,
  PutUsersByIdData,
  User,
} from "@/generated/core/types.gen";
import { zUser, zUserIdentity, zUserPrivilege } from "@/generated/core/zod.gen";
import { z } from "zod";
import { effectivePrivileges } from "./privileges";
import { callerRoleGrants } from "./roles";

// Per-run mutable copy so PUT /users/{id} name edits are visible on the next
// /me read within a session.
let user: User = zUser.parse(fixture.user);
const identities = z.array(zUserIdentity).parse(fixture.identities);
const direct = z.array(zUserPrivilege).parse(fixture.direct);

export const identityHandlers = [
  http.get("*/api/v1/me", () =>
    HttpResponse.json({
      user,
      privileges: effectivePrivileges(),
      roles: callerRoleGrants(user.id),
    }),
  ),
  http.get<GetUsersByIdUserIdentitiesData["path"]>(
    "*/api/v1/users/:id/user-identities",
    ({ params }) => HttpResponse.json(params.id === user.id ? identities : []),
  ),
  http.get<GetUsersByIdPrivilegesData["path"]>("*/api/v1/users/:id/privileges", ({ params }) =>
    HttpResponse.json(params.id === user.id ? direct : []),
  ),
  http.put<PutUsersByIdData["path"], PutUsersByIdData["body"]>(
    "*/api/v1/users/:id",
    async ({ request }) => {
      // Blank names keep the stored value, as in the backend.
      const { first_name, middle_name, last_name } = await request.json();
      user = {
        ...user,
        first_name: first_name || user.first_name,
        middle_name: middle_name || user.middle_name,
        last_name: last_name || user.last_name,
      };
      return HttpResponse.json(user);
    },
  ),
];
