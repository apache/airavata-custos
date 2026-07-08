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
import type {
  CreateOrganizationPayload,
  Organization,
} from "@/features/core/organizations/schemas";
import organizationsFixture from "@/features/core/organizations/__fixtures__/organizations.json";

const organizations: Organization[] = (organizationsFixture as Organization[]).map((o) => ({
  ...o,
}));

export const organizationsHandlers = [
  http.get("*/api/v1/organizations", ({ request }) => {
    const url = new URL(request.url);
    const limit = Number(url.searchParams.get("limit") ?? organizations.length);
    const offset = Number(url.searchParams.get("offset") ?? 0);
    const items = organizations.slice(offset, offset + limit);
    return HttpResponse.json({ items, total: organizations.length });
  }),

  http.get("*/api/v1/organizations/:id", ({ params }) => {
    const id = String(params.id);
    const found = organizations.find((o) => o.id === id);
    if (!found) return HttpResponse.json({ error: "organization not found" }, { status: 404 });
    return HttpResponse.json(found);
  }),

  http.post("*/api/v1/organizations", async ({ request }) => {
    const payload = (await request.json()) as CreateOrganizationPayload;
    const organization: Organization = {
      id: `org-${Date.now()}`,
      name: payload.name,
      originated_id: payload.originated_id ?? "",
    };
    organizations.unshift(organization);
    return HttpResponse.json(organization, { status: 201 });
  }),
];
