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

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createResourceRate, getEffectiveRate, listResourceRates, listResourceSummaries } from "./api";

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
    queryFn: () => listResourceRates(id as string),
    enabled: Boolean(id),
    ...DEFAULTS,
  });
}

export function useEffectiveRate(id: string | undefined) {
  return useQuery({
    queryKey: id ? resourceKeys.effective(id) : [...resourceKeys.all, "effective", "none"],
    queryFn: () => getEffectiveRate(id as string),
    enabled: Boolean(id),
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
