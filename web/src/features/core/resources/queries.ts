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

import { clusterKeys } from "@/features/core/clusters/queries";
import {
  getComputeAllocationResourceRatesById,
  getComputeAllocationResources,
  getComputeAllocationResourcesById,
  getComputeAllocationResourcesByIdAllocations,
  getComputeAllocationResourcesByIdMembershipOverrides,
  getComputeAllocationResourcesByIdRates,
  getComputeAllocationResourcesSummary,
  postComputeAllocationResourceRates,
  postComputeAllocationResources,
} from "@/generated/core/sdk.gen";
import type { ComputeAllocationResourceRate } from "@/generated/core/types.gen";
import { type KnownError, errorCopy } from "@/shared/api/client";
import { listOf } from "@/shared/api/listOf";
import {
  queryOptions,
  skipToken,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { toast } from "sonner";

export const resourceKeys = {
  all: ["resources"] as const,
  list: () => [...resourceKeys.all, "list"] as const,
  catalog: () => [...resourceKeys.all, "catalog"] as const,
  detail: (id: string) => [...resourceKeys.all, "detail", id] as const,
  allocations: (id: string) => [...resourceKeys.detail(id), "allocations"] as const,
  overrides: (id: string) => [...resourceKeys.detail(id), "overrides"] as const,
  rate: (id: string) => [...resourceKeys.all, "rate", id] as const,
  rates: (id: string) => [...resourceKeys.all, "rates", id] as const,
};

export function useResourceSummaries() {
  return useQuery({
    queryKey: resourceKeys.list(),
    queryFn: () => getComputeAllocationResourcesSummary().then(listOf),
  });
}

// Mappings and usage rows carry only a resource id; the catalog names them.
export function useResources({ enabled = true } = {}) {
  return useQuery({
    queryKey: resourceKeys.catalog(),
    queryFn: () => getComputeAllocationResources().then(listOf),
    enabled,
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
    queryFn: id
      ? () => getComputeAllocationResourcesByIdAllocations({ path: { id } }).then(listOf)
      : skipToken,
  });
}

export function useResourceOverrides(id: string | undefined) {
  return useQuery({
    queryKey: resourceKeys.overrides(id ?? ""),
    queryFn: id
      ? () => getComputeAllocationResourcesByIdMembershipOverrides({ path: { id } }).then(listOf)
      : skipToken,
  });
}

export function useResourceRate(id: string | undefined) {
  return useQuery({
    queryKey: resourceKeys.rate(id ?? ""),
    queryFn: id ? () => getComputeAllocationResourceRatesById({ path: { id } }) : skipToken,
  });
}

// Master answers 400 "not found" when a create names a missing cluster or resource.
const CLUSTER_GONE: KnownError = {
  status: 400,
  when: /not found/,
  copy: "This cluster no longer exists.",
};
const RESOURCE_GONE: KnownError = {
  status: 400,
  when: /not found/,
  copy: "This resource no longer exists.",
};

// typedCluster: the cluster came from a typed ID, not the cluster list.
export function useCreateResource(typedCluster = false) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: postComputeAllocationResources<true>,
    meta: { inlineError: true },
    onError: (err, { body }) => {
      const clusterGone = errorCopy(err, [CLUSTER_GONE]) === CLUSTER_GONE.copy;
      // The picked cluster is gone; the picker must drop it.
      if (clusterGone && !typedCluster) client.invalidateQueries({ queryKey: clusterKeys.list() });
      toast.error(
        clusterGone
          ? typedCluster
            ? "No cluster has this ID."
            : CLUSTER_GONE.copy
          : errorCopy(err, [
              { status: 409, copy: `This cluster already has a resource named ${body.name}.` },
            ]),
      );
    },
    onSettled: () => client.invalidateQueries({ queryKey: resourceKeys.all }),
  });
}

const ratesQuery = (id: string | undefined) =>
  queryOptions({
    queryKey: resourceKeys.rates(id ?? ""),
    queryFn: id
      ? () => getComputeAllocationResourcesByIdRates({ path: { id } }).then(listOf)
      : skipToken,
  });

export function useResourceRates(id: string | undefined) {
  return useQuery(ratesQuery(id));
}

// The rate in force at now, ordered as the server picks it: latest start, earliest end, lowest id.
export function effectiveRate(
  rates: ComputeAllocationResourceRate[],
  now = Date.now(),
): ComputeAllocationResourceRate | null {
  const [best] = rates
    .filter((r) => Date.parse(r.start_time) <= now && Date.parse(r.end_time) > now)
    .sort(
      (a, b) =>
        Date.parse(b.start_time) - Date.parse(a.start_time) ||
        Date.parse(a.end_time) - Date.parse(b.end_time) ||
        (a.id < b.id ? -1 : 1),
    );
  return best ?? null;
}

export type RateStatus =
  | { kind: "ACTIVE" | "NEVER" | "EXPIRED" }
  | { kind: "SCHEDULED"; from: string }
  | { kind: "OVERRIDDEN"; until: string | undefined };

// Labels a rate by the effective-rate rule. A rate only ever wins at a window boundary, so a
// future rate takes effect at the first boundary it wins at (never when there is none before its
// end), and an overridden rate is overridden until the first later boundary it wins at.
export function rateStatus(
  rate: ComputeAllocationResourceRate,
  rates: ComputeAllocationResourceRate[],
  now = Date.now(),
): RateStatus {
  const start = Date.parse(rate.start_time);
  const end = Date.parse(rate.end_time);
  const firstWin = (from: number) =>
    rates
      .flatMap((r) => [r.start_time, r.end_time])
      .filter((b) => Date.parse(b) >= from && Date.parse(b) < end)
      .sort((a, b) => Date.parse(a) - Date.parse(b))
      .find((b) => effectiveRate(rates, Date.parse(b))?.id === rate.id);
  if (start > now) {
    const from = firstWin(start);
    return from ? { kind: "SCHEDULED", from } : { kind: "NEVER" };
  }
  if (end <= now) return { kind: "EXPIRED" };
  const effective = effectiveRate(rates, now);
  if (!effective || effective.id === rate.id) return { kind: "ACTIVE" };
  return { kind: "OVERRIDDEN", until: firstWin(now) };
}

// Derived from the rate list: the effective endpoint answers "no rate" with a 404 the browser logs.
export function useEffectiveRate(id: string | undefined) {
  return useQuery({ ...ratesQuery(id), select: (rates) => effectiveRate(rates) });
}

export function useCreateResourceRate(resourceId: string) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: postComputeAllocationResourceRates<true>,
    meta: { errors: [RESOURCE_GONE] },
    onError: (err) => {
      // Resetting the gone resource lets the drawer's gone state take over.
      if (errorCopy(err, [RESOURCE_GONE]) === RESOURCE_GONE.copy)
        void client.resetQueries({ queryKey: resourceKeys.detail(resourceId) });
    },
    onSettled: () =>
      Promise.all(
        [resourceKeys.rates(resourceId), resourceKeys.list()].map((queryKey) =>
          client.invalidateQueries({ queryKey }),
        ),
      ),
  });
}
