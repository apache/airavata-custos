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

import { allocationKeys } from "@/features/core/allocations/queries";
import { skipToken, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  type CreateResourcePayload,
  createResource,
  createResourceRate,
  deleteResource,
  getEffectiveRate,
  listResourceRates,
  listResourceSummaries,
  updateResource,
} from "./api";

export const resourceKeys = {
  all: ["resources"] as const,
  list: () => [...resourceKeys.all, "list"] as const,
  rates: (id: string) => [...resourceKeys.all, "rates", id] as const,
  effective: (id: string) => [...resourceKeys.all, "effective", id] as const,
};

const DEFAULTS = {
  staleTime: 30_000,
  gcTime: 300_000,
  refetchOnWindowFocus: false,
} as const;

export function useResourceSummaries() {
  return useQuery({
    queryKey: resourceKeys.list(),
    queryFn: listResourceSummaries,
    ...DEFAULTS,
  });
}

export function useResourceRates(id: string | undefined) {
  return useQuery({
    queryKey: id ? resourceKeys.rates(id) : [...resourceKeys.all, "rates", "none"],
    queryFn: id ? () => listResourceRates(id) : skipToken,
    ...DEFAULTS,
  });
}

export function useEffectiveRate(id: string | undefined) {
  return useQuery({
    queryKey: id ? resourceKeys.effective(id) : [...resourceKeys.all, "effective", "none"],
    queryFn: id ? () => getEffectiveRate(id) : skipToken,
    retry: false,
    ...DEFAULTS,
  });
}

export function useCreateResourceRate(resourceId: string) {
  const client = useQueryClient();
  return useMutation({
    mutationFn: createResourceRate,
    onSuccess: () => {
      client.invalidateQueries({ queryKey: resourceKeys.rates(resourceId) });
      client.invalidateQueries({ queryKey: resourceKeys.effective(resourceId) });
    },
  });
}

type SaveResourceInput = {
  id: string | undefined;
  resource: CreateResourcePayload;
  rate: number | undefined;
};

// A new rate supersedes the current one from now.
export function useSaveResource() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, resource, rate }: SaveResourceInput) => {
      const saved = id ? await updateResource(id, resource) : await createResource(resource);
      if (rate !== undefined) {
        const start = new Date();
        const end = new Date(start);
        end.setUTCFullYear(end.getUTCFullYear() + 1);
        await createResourceRate({
          compute_allocation_resource_id: saved.id,
          rate,
          start_time: start.toISOString(),
          end_time: end.toISOString(),
        });
      }
      return saved;
    },
    onSuccess: () => client.invalidateQueries({ queryKey: resourceKeys.all }),
  });
}

export function useDeleteResource() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: deleteResource,
    onSuccess: () => {
      client.invalidateQueries({ queryKey: resourceKeys.list() });
      // The resource's grants go with it.
      client.invalidateQueries({ queryKey: allocationKeys.all });
    },
  });
}
