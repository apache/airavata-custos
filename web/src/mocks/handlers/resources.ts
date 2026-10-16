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

import ratesFixture from "@/features/core/resources/__fixtures__/rates.json";
import resourcesFixture from "@/features/core/resources/__fixtures__/resources.json";
import type {
  ComputeAllocationResource,
  ComputeAllocationResourceRate,
  GetComputeAllocationResourceRatesByIdData,
  GetComputeAllocationResourcesByIdData,
  GetComputeAllocationResourcesByIdRatesData,
  PostComputeAllocationResourceRatesData,
  PostComputeAllocationResourcesData,
} from "@/generated/core/types.gen";
import {
  zComputeAllocationResource,
  zComputeAllocationResourceRate,
} from "@/generated/core/zod.gen";
import { http, HttpResponse } from "msw";
import { z } from "zod";
import { notFound } from "../paging";

const resources: ComputeAllocationResource[] = z
  .array(zComputeAllocationResource)
  .parse(resourcesFixture);
const rates: ComputeAllocationResourceRate[] = z
  .array(zComputeAllocationResourceRate)
  .parse(ratesFixture);

function ratesFor(resourceId: string): ComputeAllocationResourceRate[] {
  return rates.filter((r) => r.compute_allocation_resource_id === resourceId);
}

// Allocation aggregates per resource; rate_count derives from live rate state.
const ALLOCATION_AGGREGATES: Record<
  string,
  { allocation_count: number; total_allocated: number; total_used_su: number }
> = {
  "21000001-0000-4000-8000-000021000001": {
    allocation_count: 3,
    total_allocated: 500000,
    total_used_su: 312400,
  },
  "21000002-0000-4000-8000-000021000002": {
    allocation_count: 2,
    total_allocated: 40000,
    total_used_su: 45100,
  },
};

export const resourcesHandlers = [
  http.get("*/api/v1/compute-allocation-resources", () => HttpResponse.json(resources)),

  http.post<never, PostComputeAllocationResourcesData["body"]>(
    "*/api/v1/compute-allocation-resources",
    async ({ request }) => {
      const created = zComputeAllocationResource.parse({
        ...(await request.json()),
        id: crypto.randomUUID(),
      });
      resources.push(created);
      return HttpResponse.json(created, { status: 201 });
    },
  ),

  http.get("*/api/v1/compute-allocation-resources/summary", () =>
    HttpResponse.json(
      resources.map((r) => ({
        ...r,
        ...(ALLOCATION_AGGREGATES[r.id] ?? {
          allocation_count: 0,
          total_allocated: 0,
          total_used_su: 0,
        }),
        rate_count: ratesFor(r.id).length,
      })),
    ),
  ),

  http.get<GetComputeAllocationResourcesByIdData["path"]>(
    "*/api/v1/compute-allocation-resources/:id",
    ({ params }) => {
      const found = resources.find((r) => r.id === params.id);
      return found ? HttpResponse.json(found) : notFound("compute allocation resource");
    },
  ),

  http.get("*/api/v1/compute-allocation-resources/:id/allocations", () => HttpResponse.json([])),

  http.get("*/api/v1/compute-allocation-resources/:id/membership-overrides", () =>
    HttpResponse.json([]),
  ),

  http.get<GetComputeAllocationResourcesByIdRatesData["path"]>(
    "*/api/v1/compute-allocation-resources/:id/rates",
    ({ params }) => HttpResponse.json(ratesFor(params.id)),
  ),

  http.get<GetComputeAllocationResourceRatesByIdData["path"]>(
    "*/api/v1/compute-allocation-resource-rates/:id",
    ({ params }) => {
      const found = rates.find((r) => r.id === params.id);
      return found ? HttpResponse.json(found) : notFound("compute allocation resource rate");
    },
  ),

  http.post<never, PostComputeAllocationResourceRatesData["body"]>(
    "*/api/v1/compute-allocation-resource-rates",
    async ({ request }) => {
      const created = zComputeAllocationResourceRate.parse({
        ...(await request.json()),
        id: crypto.randomUUID(),
      });
      rates.push(created);
      return HttpResponse.json(created, { status: 201 });
    },
  ),
];
