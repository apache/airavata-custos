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

import { useClusters } from "@/features/core/clusters/queries";
import type { ClusterAccountResponse, User } from "@/generated/core/zod.gen";
import { useMutation, useQueries, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  approveClusterAccount,
  denyClusterAccount,
  listClusterAccounts,
  listUserClusterAccounts,
} from "./api";
import type { ClusterAccountListParams } from "./types";

export const clusterAccountKeys = {
  all: ["cluster-accounts"] as const,
  list: (params: ClusterAccountListParams = {}) =>
    [...clusterAccountKeys.all, "list", params] as const,
  user: (userId: string) => [...clusterAccountKeys.all, "user", userId] as const,
};

const DEFAULTS = {
  staleTime: 30_000,
  gcTime: 300_000,
  refetchOnWindowFocus: false,
} as const;

export function useClusterAccounts(
  params: ClusterAccountListParams = {},
  options: { enabled?: boolean } = {},
) {
  return useQuery({
    queryKey: clusterAccountKeys.list(params),
    queryFn: () => listClusterAccounts(params),
    enabled: options.enabled ?? true,
    ...DEFAULTS,
  });
}

// Each user's cluster accounts, filled in with the cluster and user names the per-user route omits.
export function useUserClusterAccounts(users: User[], enabled: boolean) {
  const clusters = useClusters(enabled).data ?? [];
  const names = new Map(clusters.map((c) => [c.id, c.name]));
  const byUser = useQueries({
    queries: users.map((user) => ({
      queryKey: clusterAccountKeys.user(user.id),
      queryFn: () => listUserClusterAccounts(user.id),
      enabled,
      ...DEFAULTS,
    })),
    combine: (results) =>
      new Map(
        users.map((user, i): [string, ClusterAccountResponse[]] => [
          user.id,
          (results[i]?.data ?? []).map((a) => ({
            ...a,
            cluster_name: names.get(a.compute_cluster_id) ?? a.compute_cluster_id,
            display_name: [user.first_name, user.last_name].filter(Boolean).join(" "),
            email: user.email,
            // The per-user route has no request date.
            requested_at: "",
          })),
        ]),
      ),
  });
  return { clusters, byUser };
}

export function useApproveClusterAccount() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => approveClusterAccount(id),
    onSuccess: () => client.invalidateQueries({ queryKey: clusterAccountKeys.all }),
  });
}

// One call per account so a failure does not stop the rest.
export function useApproveClusterAccounts() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: (ids: string[]) => Promise.allSettled(ids.map((id) => approveClusterAccount(id))),
    onSettled: () => client.invalidateQueries({ queryKey: clusterAccountKeys.all }),
  });
}

export function useDenyClusterAccount() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({ id, note }: { id: string; note: string }) => denyClusterAccount(id, note),
    onSuccess: () => client.invalidateQueries({ queryKey: clusterAccountKeys.all }),
  });
}
