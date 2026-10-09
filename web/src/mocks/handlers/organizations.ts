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
import organizationsFixture from "@/features/core/organizations/__fixtures__/organizations.json";
import type {
  GetOrganizationsByIdData,
  Organization,
  PostOrganizationsData,
} from "@/generated/core/types.gen";
import { zOrganization } from "@/generated/core/zod.gen";
import { z } from "zod";
import { notFound, page } from "../paging";

const organizations: Organization[] = z.array(zOrganization).parse(organizationsFixture);

export const organizationsHandlers = [
  http.get("*/api/v1/organizations", ({ request }) => {
    const { items, total } = page(new URL(request.url), organizations);
    return HttpResponse.json({ items, total });
  }),

  http.get<GetOrganizationsByIdData["path"]>("*/api/v1/organizations/:id", ({ params }) => {
    const found = organizations.find((o) => o.id === params.id);
    if (!found) return notFound("organization");
    return HttpResponse.json(found);
  }),

  http.post<never, PostOrganizationsData["body"]>("*/api/v1/organizations", async ({ request }) => {
    const { name, originated_id = "" } = await request.json();
    const organization: Organization = { id: `org-${Date.now()}`, name, originated_id };
    organizations.unshift(organization);
    return HttpResponse.json(organization, { status: 201 });
  }),
];
