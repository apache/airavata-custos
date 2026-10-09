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

"use client";

import {
  getComputeAllocationResourceRatesById,
  getComputeAllocationResources,
  getComputeAllocationResourcesById,
  getComputeAllocationResourcesByIdAllocations,
  getComputeAllocationResourcesByIdMembershipOverrides,
  getComputeAllocationResourcesByIdRates,
  getComputeAllocationResourcesByIdRatesEffective,
  getComputeAllocationResourcesSummary,
  postComputeAllocationResourceRates,
  postComputeAllocationResources,
} from "@/generated/core/sdk.gen";
import { useInvalidating } from "@/shared/api/useInvalidating";
import { skipToken, useQuery } from "@tanstack/react-query";

export const resourceKeys = {
  all: ["resources"] as const,
  list: () => [...resourceKeys.all, "list"] as const,
  catalog: () => [...resourceKeys.all, "catalog"] as const,
  detail: (id: string) => [...resourceKeys.all, "detail", id] as const,
  allocations: (id: string) => [...resourceKeys.detail(id), "allocations"] as const,
  overrides: (id: string) => [...resourceKeys.detail(id), "overrides"] as const,
  rate: (id: string) => [...resourceKeys.all, "rate", id] as const,
  rates: (id: string) => [...resourceKeys.all, "rates", id] as const,
  effective: (id: string) => [...resourceKeys.all, "effective", id] as const,
};

export function useResourceSummaries() {
  return useQuery({
    queryKey: resourceKeys.list(),
    queryFn: () => getComputeAllocationResourcesSummary(),
  });
}

export function useResources() {
  return useQuery({
    queryKey: resourceKeys.catalog(),
    queryFn: () => getComputeAllocationResources(),
  });
}

export function useResource(id: string | undefined) {
  return useQuery({
    queryKey: resourceKeys.detail(id ?? ""),
    queryFn: id ? () => getComputeAllocationResourcesById({ path: { id } }) : skipToken,
  });
}

export function useResourceAllocations(id: string | undefined) {
  return useQuery({
    queryKey: resourceKeys.allocations(id ?? ""),
    queryFn: id ? () => getComputeAllocationResourcesByIdAllocations({ path: { id } }) : skipToken,
  });
}

export function useResourceOverrides(id: string | undefined) {
  return useQuery({
    queryKey: resourceKeys.overrides(id ?? ""),
    queryFn: id
      ? () => getComputeAllocationResourcesByIdMembershipOverrides({ path: { id } })
      : skipToken,
  });
}

export function useResourceRate(id: string | undefined) {
  return useQuery({
    queryKey: resourceKeys.rate(id ?? ""),
    queryFn: id ? () => getComputeAllocationResourceRatesById({ path: { id } }) : skipToken,
  });
}

export function useCreateResource() {
  return useInvalidating(postComputeAllocationResources<true>, resourceKeys.all);
}

export function useResourceRates(id: string | undefined) {
  return useQuery({
    queryKey: resourceKeys.rates(id ?? ""),
    queryFn: id ? () => getComputeAllocationResourcesByIdRates({ path: { id } }) : skipToken,
  });
}

export function useEffectiveRate(id: string | undefined) {
  return useQuery({
    queryKey: resourceKeys.effective(id ?? ""),
    queryFn: id
      ? () => getComputeAllocationResourcesByIdRatesEffective({ path: { id } })
      : skipToken,
    retry: false,
  });
}

export function useCreateResourceRate(resourceId: string) {
  return useInvalidating(
    postComputeAllocationResourceRates<true>,
    resourceKeys.rates(resourceId),
    resourceKeys.effective(resourceId),
    resourceKeys.list(),
  );
}
