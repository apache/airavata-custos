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

import { useQuery } from "@tanstack/react-query";
import { getCluster, listClusterUsers, listClusters } from "./api";

export const clusterKeys = {
  all: ["clusters"] as const,
  list: () => [...clusterKeys.all, "list"] as const,
  detail: (id: string) => [...clusterKeys.all, "detail", id] as const,
  users: (id: string) => [...clusterKeys.all, "users", id] as const,
};

const DEFAULTS = {
  staleTime: 30_000,
  gcTime: 300_000,
  refetchOnWindowFocus: false,
} as const;

export function useClusters() {
  return useQuery({
    queryKey: clusterKeys.list(),
    queryFn: listClusters,
    ...DEFAULTS,
  });
}

export function useCluster(id: string | undefined) {
  return useQuery({
    queryKey: id ? clusterKeys.detail(id) : [...clusterKeys.all, "detail", "none"],
    queryFn: () => getCluster(id as string),
    enabled: Boolean(id),
    ...DEFAULTS,
  });
}

export function useClusterUsers(id: string | undefined) {
  return useQuery({
    queryKey: id ? clusterKeys.users(id) : [...clusterKeys.all, "users", "none"],
    queryFn: () => listClusterUsers(id as string),
    enabled: Boolean(id),
    ...DEFAULTS,
  });
}
