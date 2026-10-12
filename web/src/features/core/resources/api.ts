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

import type { PostComputeAllocationResourceRatesData } from "@/generated/core/types.gen";
import {
  zGetComputeAllocationResourcesByIdRatesEffectiveResponse,
  zGetComputeAllocationResourcesByIdRatesResponse,
  zGetComputeAllocationResourcesSummaryResponse,
  zPostComputeAllocationResourceRatesResponse,
} from "@/generated/core/zod.gen";
import { apiFetch } from "@/shared/api/client";
import type { Rate, ResourceSummary } from "./schemas";

// The backend serializes empty lists as JSON null (Go nil slice).
const summaryListSchema = zGetComputeAllocationResourcesSummaryResponse.nullable();
const rateListSchema = zGetComputeAllocationResourcesByIdRatesResponse.nullable();

export async function listResourceSummaries(): Promise<ResourceSummary[]> {
  const raw = await apiFetch("/compute-allocation-resources/summary");
  return summaryListSchema.parse(raw) ?? [];
}

export async function listResourceRates(id: string): Promise<Rate[]> {
  const raw = await apiFetch(`/compute-allocation-resources/${id}/rates`);
  return rateListSchema.parse(raw) ?? [];
}

export async function getEffectiveRate(id: string, at?: string): Promise<Rate> {
  const qs = at ? `?at=${encodeURIComponent(at)}` : "";
  const raw = await apiFetch(`/compute-allocation-resources/${id}/rates/effective${qs}`);
  return zGetComputeAllocationResourcesByIdRatesEffectiveResponse.parse(raw);
}

export type CreateResourceRatePayload = Required<
  Pick<
    PostComputeAllocationResourceRatesData["body"],
    "compute_allocation_resource_id" | "rate" | "start_time" | "end_time"
  >
>;

export async function createResourceRate(payload: CreateResourceRatePayload): Promise<Rate> {
  const raw = await apiFetch("/compute-allocation-resource-rates", {
    method: "POST",
    body: payload,
  });
  return zPostComputeAllocationResourceRatesResponse.parse(raw);
}
