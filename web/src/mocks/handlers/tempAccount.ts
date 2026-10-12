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

import type {
  ComputeAllocationMembership,
  DeleteConnectorsTempAccountRemoveByUserIdData,
  GetConnectorsTempAccountMembershipByUserIdData,
  PostConnectorsTempAccountAssignAllocationData,
  PostConnectorsTempAccountCreateData,
  PostConnectorsTempAccountUpdateAllocationData,
  User,
} from "@/generated/core/types.gen";
import { zComputeAllocationMembership, zUser } from "@/generated/core/zod.gen";
import { http, HttpResponse } from "msw";
import { notFound } from "../paging";

const taraId = "5f0c2a7e-3b9d-4e61-8a24-c1d7e9b0f342";
const users: User[] = [
  {
    id: taraId,
    organization_id: "14000001-0000-4000-8000-000014000001",
    first_name: "Tara",
    middle_name: "",
    last_name: "Temp",
    email: "tara.temp@sample.example.edu",
    status: "ACTIVE",
    type: "VIRTUAL",
  },
];

const memberships: ComputeAllocationMembership[] = [1, 2].map((n) => ({
  id: `000001${n}0-0000-4000-8000-0990000001${n}0`,
  user_id: taraId,
  compute_allocation_id: `0100000${n}-0000-4000-8000-00000100000${n}`,
  start_time: "2026-01-01T00:00:00Z",
  end_time: "2026-12-31T00:00:00Z",
  membership_status: "ACTIVE",
}));

const badRequest = (error: string) => HttpResponse.json({ error }, { status: 400 });
const notVirtual = () => badRequest("User type must be VIRTUAL");
const findUser = (id: unknown) => users.find((u) => u.id === id);

// The connector decodes core users and memberships, so its bodies are those shapes.
export const tempAccountHandlers = [
  http.get("*/api/v1/users/:id", ({ params }) => {
    const user = findUser(params.id);
    return user ? HttpResponse.json(user) : undefined;
  }),

  http.post<never, PostConnectorsTempAccountCreateData["body"]>(
    "*/api/v1/connectors/temp-account/create",
    async ({ request }) => {
      const body = await request.json();
      if (body.type !== "VIRTUAL") return notVirtual();
      if (!body.email) return badRequest("user email is required");
      if (!body.organization_id) return badRequest("user organization_id is required");
      if (users.some((u) => u.email === body.email)) {
        return HttpResponse.json({ error: "user already exists" }, { status: 409 });
      }
      const user = zUser.parse({
        ...body,
        id: crypto.randomUUID(),
        status: body.status || "PENDING",
      });
      users.push(user);
      return HttpResponse.json(user, { status: 201 });
    },
  ),

  http.delete<DeleteConnectorsTempAccountRemoveByUserIdData["path"]>(
    "*/api/v1/connectors/temp-account/remove/:user_id",
    ({ params }) => {
      const user = findUser(params.user_id);
      if (!user) return notFound("user");
      if (user.type !== "VIRTUAL") return notVirtual();
      users.splice(users.indexOf(user), 1);
      return new HttpResponse(null, { status: 204 });
    },
  ),

  // An empty list answers 404, as the connector does.
  http.get<GetConnectorsTempAccountMembershipByUserIdData["path"]>(
    "*/api/v1/connectors/temp-account/membership/:user_id",
    ({ params }) => {
      const user = findUser(params.user_id);
      if (!user) return notFound("user");
      if (user.type !== "VIRTUAL") return notVirtual();
      const rows = memberships.filter((m) => m.user_id === user.id);
      return rows.length ? HttpResponse.json(rows) : notFound("membership");
    },
  ),

  http.post<never, PostConnectorsTempAccountAssignAllocationData["body"]>(
    "*/api/v1/connectors/temp-account/assign-allocation",
    async ({ request }) => {
      const body = await request.json();
      if (!body.user_id || !body.compute_allocation_id) {
        return badRequest("UserID and ComputeAllocationID must be provided");
      }
      const user = findUser(body.user_id);
      if (!user) return notFound("user");
      if (user.type !== "VIRTUAL") return notVirtual();
      const created = zComputeAllocationMembership.parse({ ...body, id: crypto.randomUUID() });
      memberships.push(created);
      return HttpResponse.json(created);
    },
  ),

  // Blank fields keep the stored value, as in the core membership update.
  http.post<never, PostConnectorsTempAccountUpdateAllocationData["body"]>(
    "*/api/v1/connectors/temp-account/update-allocation",
    async ({ request }) => {
      const body = await request.json();
      if (!body.user_id || !body.compute_allocation_id) {
        return badRequest("UserID and ComputeAllocationID must be provided");
      }
      if (!body.id) return badRequest("compute allocation membership id is required");
      const found = memberships.find((m) => m.id === body.id);
      if (!found) return notFound("membership");
      for (const [key, value] of Object.entries(body)) {
        if (value) Object.assign(found, { [key]: value });
      }
      return HttpResponse.json(found);
    },
  ),

  // A temporary membership is a core membership; other ids fall through to the core handler.
  http.delete("*/api/v1/compute-allocation-memberships/:id", ({ params }) => {
    const idx = memberships.findIndex((m) => m.id === params.id);
    if (idx === -1) return;
    memberships.splice(idx, 1);
    return new HttpResponse(null, { status: 204 });
  }),
];
