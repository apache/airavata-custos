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
import { userKeys } from "@/features/core/users/queries";
import {
  deleteComputeClusterUsersById,
  getComputeClusterUsers,
  getComputeClusterUsersById,
  postComputeClusterUsers,
  postComputeClusterUsersByIdApprove,
  postComputeClusterUsersByIdDeny,
  putComputeClusterUsersById,
} from "@/generated/core/sdk.gen";
import type { GetComputeClusterUsersData } from "@/generated/core/types.gen";
import { useInvalidating } from "@/shared/api/useInvalidating";
import { skipToken, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export const clusterAccountKeys = {
  all: ["cluster-accounts"] as const,
  list: (query: GetComputeClusterUsersData["query"]) =>
    [...clusterAccountKeys.all, "list", query] as const,
  detail: (id: string) => [...clusterAccountKeys.all, "detail", id] as const,
};

export function useClusterAccounts(query: GetComputeClusterUsersData["query"], enabled = true) {
  return useQuery({
    queryKey: clusterAccountKeys.list(query),
    queryFn: () => getComputeClusterUsers({ query }),
    enabled,
  });
}

export function useClusterAccount(id: string | undefined) {
  return useQuery({
    queryKey: clusterAccountKeys.detail(id ?? ""),
    queryFn: id ? () => getComputeClusterUsersById({ path: { id } }) : skipToken,
  });
}

// A review changes the account rows the cluster users and user drawers list too.
const accountKeys = () => [
  clusterAccountKeys.all,
  [...clusterKeys.all, "users"],
  [...userKeys.all, "cluster-accounts"],
];

export function useApproveClusterAccount() {
  return useInvalidating(postComputeClusterUsersByIdApprove<true>, ...accountKeys());
}

export function useDenyClusterAccount() {
  return useInvalidating(postComputeClusterUsersByIdDeny<true>, ...accountKeys());
}

export function useCreateClusterAccount() {
  return useInvalidating(postComputeClusterUsers<true>, ...accountKeys());
}

export function useUpdateClusterAccount() {
  return useInvalidating(putComputeClusterUsersById<true>, ...accountKeys());
}

export function useDeleteClusterAccount() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: deleteComputeClusterUsersById<true>,
    onSuccess: (_data, { path }) => {
      client.removeQueries({ queryKey: clusterAccountKeys.detail(path.id) });
      for (const queryKey of accountKeys()) client.invalidateQueries({ queryKey });
    },
  });
}
