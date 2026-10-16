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
  postComputeClusters,
} from "@/generated/core/sdk.gen";
import { errorCopy } from "@/shared/api/client";
import { listOf } from "@/shared/api/listOf";
import { skipToken, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

export const clusterKeys = {
  all: ["clusters"] as const,
  list: () => [...clusterKeys.all, "list"] as const,
  detail: (id: string) => [...clusterKeys.all, "detail", id] as const,
  users: (id: string) => [...clusterKeys.all, "users", id] as const,
};

export function useClusters(enabled = true) {
  return useQuery({
    queryKey: clusterKeys.list(),
    queryFn: () => getComputeClusters().then(listOf),
    enabled,
  });
}

export function useClusterUsers(id: string | undefined) {
  return useQuery({
    queryKey: clusterKeys.users(id ?? ""),
    queryFn: id ? () => getComputeClustersByIdUsers({ path: { id } }).then(listOf) : skipToken,
  });
}

export function useCluster(id: string | undefined) {
  return useQuery({
    queryKey: clusterKeys.detail(id ?? ""),
    queryFn: id ? () => getComputeClustersById({ path: { id } }) : skipToken,
  });
}

export function useCreateCluster() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: postComputeClusters<true>,
    meta: { inlineError: true },
    onError: (err, { body }) =>
      toast.error(
        errorCopy(err, [{ status: 409, copy: `A cluster named ${body.name} already exists.` }]),
      ),
    onSettled: () => client.invalidateQueries({ queryKey: clusterKeys.list() }),
  });
}
