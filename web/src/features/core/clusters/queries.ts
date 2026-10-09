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
  getComputeClusters,
  getComputeClustersById,
  getComputeClustersByIdUsers,
  getComputeClustersByIdUsersByUserId,
  postComputeClusters,
} from "@/generated/core/sdk.gen";
import { useInvalidating } from "@/shared/api/useInvalidating";
import { skipToken, useQuery } from "@tanstack/react-query";

export const clusterKeys = {
  all: ["clusters"] as const,
  list: () => [...clusterKeys.all, "list"] as const,
  detail: (id: string) => [...clusterKeys.all, "detail", id] as const,
  users: (id: string) => [...clusterKeys.all, "users", id] as const,
  user: (id: string, userId: string) => [...clusterKeys.users(id), userId] as const,
};

export function useClusters(enabled = true) {
  return useQuery({
    queryKey: clusterKeys.list(),
    queryFn: () => getComputeClusters(),
    enabled,
  });
}

// Resolves a cluster id to its name, falling back to the id.
export function useClusterName() {
  const clusters = useClusters().data;
  return (id = "") => clusters?.find((c) => c.id === id)?.name ?? id;
}

export function useClusterUsers(id: string | undefined) {
  return useQuery({
    queryKey: clusterKeys.users(id ?? ""),
    queryFn: id ? () => getComputeClustersByIdUsers({ path: { id } }) : skipToken,
  });
}

export function useCluster(id: string | undefined) {
  return useQuery({
    queryKey: clusterKeys.detail(id ?? ""),
    queryFn: id ? () => getComputeClustersById({ path: { id } }) : skipToken,
  });
}

export function useClusterUser(id: string | undefined, userId: string | undefined) {
  return useQuery({
    queryKey: clusterKeys.user(id ?? "", userId ?? ""),
    queryFn:
      id && userId
        ? () => getComputeClustersByIdUsersByUserId({ path: { id, userId } })
        : skipToken,
    retry: false,
  });
}

export function useCreateCluster() {
  return useInvalidating(postComputeClusters<true>, clusterKeys.list());
}
