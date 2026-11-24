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
  PostComputeAllocationResourceRatesData,
  PostComputeAllocationResourcesData,
} from "@/generated/core/types.gen";
import {
  type ComputeAllocationResource,
  type ComputeAllocationResourceRate,
  type ComputeAllocationResourceSummary,
  zGetComputeAllocationResourcesByIdRatesEffectiveResponse,
  zGetComputeAllocationResourcesByIdRatesResponse,
  zGetComputeAllocationResourcesSummaryResponse,
  zPostComputeAllocationResourceRatesResponse,
  zPostComputeAllocationResourcesResponse,
  zPutComputeAllocationResourcesByIdResponse,
} from "@/generated/core/zod.gen";
import { apiFetch } from "@/shared/api/client";

export async function listResourceSummaries(): Promise<ComputeAllocationResourceSummary[]> {
  const raw = await apiFetch("/compute-allocation-resources/summary");
  return zGetComputeAllocationResourcesSummaryResponse.parse(raw) ?? [];
}

export async function listResourceRates(id: string): Promise<ComputeAllocationResourceRate[]> {
  const raw = await apiFetch(`/compute-allocation-resources/${id}/rates`);
  return zGetComputeAllocationResourcesByIdRatesResponse.parse(raw) ?? [];
}

export async function getEffectiveRate(
  id: string,
  at?: string,
): Promise<ComputeAllocationResourceRate> {
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

export async function createResourceRate(
  payload: CreateResourceRatePayload,
): Promise<ComputeAllocationResourceRate> {
  const raw = await apiFetch("/compute-allocation-resource-rates", {
    method: "POST",
    body: payload,
  });
  return zPostComputeAllocationResourceRatesResponse.parse(raw);
}

export type CreateResourcePayload = Required<
  Pick<PostComputeAllocationResourcesData["body"], "name" | "resource_type" | "compute_cluster_id">
>;

export async function createResource(
  payload: CreateResourcePayload,
): Promise<ComputeAllocationResource> {
  const raw = await apiFetch("/compute-allocation-resources", { method: "POST", body: payload });
  return zPostComputeAllocationResourcesResponse.parse(raw);
}

export async function updateResource(
  id: string,
  payload: CreateResourcePayload,
): Promise<ComputeAllocationResource> {
  const raw = await apiFetch(`/compute-allocation-resources/${id}`, {
    method: "PUT",
    body: payload,
  });
  return zPutComputeAllocationResourcesByIdResponse.parse(raw);
}

export async function deleteResource(id: string): Promise<void> {
  await apiFetch(`/compute-allocation-resources/${id}`, { method: "DELETE" });
}
