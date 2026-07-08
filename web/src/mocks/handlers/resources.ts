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
import ratesFixture from "@/features/core/resources/__fixtures__/rates.json";
import resourcesFixture from "@/features/core/resources/__fixtures__/resources.json";
import type { ComputeAllocationResource, Rate } from "@/features/core/resources/schemas";

const resources = resourcesFixture as ComputeAllocationResource[];
const rates = ratesFixture as Rate[];

function ratesFor(resourceId: string): Rate[] {
  return rates.filter((r) => r.compute_allocation_resource_id === resourceId);
}

export const resourcesHandlers = [
  http.get("*/api/v1/compute-allocation-resources", () => HttpResponse.json(resources)),

  http.get("*/api/v1/compute-allocation-resources/:id/rates/effective", ({ params, request }) => {
    const id = String(params.id);
    const at = new URL(request.url).searchParams.get("at");
    const now = at ? new Date(at).getTime() : Date.now();
    const active = ratesFor(id).find(
      (r) => now >= new Date(r.start_time).getTime() && now < new Date(r.end_time).getTime(),
    );
    if (!active) return HttpResponse.json({ error: "no effective rate" }, { status: 404 });
    return HttpResponse.json(active);
  }),

  http.get("*/api/v1/compute-allocation-resources/:id/rates", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(ratesFor(id));
  }),

  http.post("*/api/v1/compute-allocation-resource-rates", async ({ request }) => {
    const body = (await request.json()) as Partial<Rate>;
    const created: Rate = {
      id: `rate-${Date.now()}`,
      compute_allocation_resource_id: String(body.compute_allocation_resource_id),
      rate: Number(body.rate),
      start_time: String(body.start_time),
      end_time: String(body.end_time),
    };
    rates.push(created);
    return HttpResponse.json(created, { status: 201 });
  }),
];
