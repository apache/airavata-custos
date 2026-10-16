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

import { amieKeys } from "@/features/connectors/amie/queries";
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
import type {
  ClusterAccountListResponse,
  GetComputeClusterUsersData,
} from "@/generated/core/types.gen";
import type { KnownError } from "@/shared/api/client";
import { paged } from "@/shared/api/paged";
import { isGone } from "@/shared/api/useOnGone";
import {
  type QueryClient,
  skipToken,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { markReviewed } from "./review";

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
    ...paged,
  });
}

export function useClusterAccount(id: string | undefined) {
  return useQuery({
    queryKey: clusterAccountKeys.detail(id ?? ""),
    queryFn: id ? () => getComputeClusterUsersById({ path: { id } }) : skipToken,
  });
}

// The account's user as a loaded list row names them, which needs no users:read.
export function useListedAccountName(id: string): string | undefined {
  const row = useQueryClient()
    .getQueriesData<ClusterAccountListResponse>({ queryKey: [...clusterAccountKeys.all, "list"] })
    .flatMap(([, data]) => data?.items ?? [])
    .find((item) => item.id === id);
  return row && (row.display_name || row.email);
}

// Every list holding account rows: the account lists, and the cluster users and user drawers.
export const clusterAccountListKeys = () => [
  [...clusterAccountKeys.all, "list"],
  [...clusterKeys.all, "users"],
  [...userKeys.all, "cluster-accounts"],
];

// A review changes the account rows the cluster users and user drawers list too.
const accountKeys = () => [
  clusterAccountKeys.all,
  [...clusterKeys.all, "users"],
  [...userKeys.all, "cluster-accounts"],
];

// A write answered 404 drops the stale account, so the drawer shows it gone.
const resetIfGone = (client: QueryClient, error: unknown, options: { path?: { id: string } }) =>
  isGone(error) && options.path
    ? client.resetQueries({ queryKey: clusterAccountKeys.detail(options.path.id) })
    : undefined;

// A review answers the packets held on the account.
function useReview<O extends { path: { id: string } }, R>(mutationFn: (options: O) => Promise<R>) {
  const client = useQueryClient();
  return useMutation({
    mutationFn,
    meta: { errors: REVIEW_ERRORS },
    onSuccess: markReviewed,
    onError: (error, options) => resetIfGone(client, error, options),
    onSettled: () =>
      Promise.all(
        [...accountKeys(), amieKeys.all].map((queryKey) => client.invalidateQueries({ queryKey })),
      ),
  });
}

export const useApproveClusterAccount = () => useReview(postComputeClusterUsersByIdApprove<true>);

export const useDenyClusterAccount = () => useReview(postComputeClusterUsersByIdDeny<true>);

// The username refusal follows the pair one, since master words a taken username only as the name.
const ACCOUNT_ERRORS: KnownError[] = [
  { status: 400, when: /compute cluster "[^"]*" not found/, copy: "No cluster has that ID." },
  { status: 400, when: /user "[^"]*" not found/, copy: "No user has that ID." },
  {
    status: 409,
    when: /already mapped/,
    copy: "This user already has an account on that cluster.",
  },
  { status: 409, copy: "This username is taken on that cluster." },
  { status: 404, copy: "This cluster account no longer exists." },
];

const REVIEW_ERRORS: KnownError[] = [
  { status: 409, when: /already approved/, copy: "This account is already approved." },
  { status: 409, when: /already denied/, copy: "This account is already denied." },
  { status: 404, copy: "This cluster account no longer exists." },
];

function useAccountWrite<O extends { path?: { id: string } }, R>(
  mutationFn: (options: O) => Promise<R>,
) {
  const client = useQueryClient();
  return useMutation({
    mutationFn,
    meta: { errors: ACCOUNT_ERRORS },
    onError: (error, options) => resetIfGone(client, error, options),
    onSettled: () =>
      Promise.all(accountKeys().map((queryKey) => client.invalidateQueries({ queryKey }))),
  });
}

export const useCreateClusterAccount = () => useAccountWrite(postComputeClusterUsers<true>);

export const useUpdateClusterAccount = () => useAccountWrite(putComputeClusterUsersById<true>);

export function useDeleteClusterAccount() {
  const client = useQueryClient();
  return useMutation({
    mutationFn: deleteComputeClusterUsersById<true>,
    onError: (error, options) => resetIfGone(client, error, options),
    onSuccess: (_data, { path }) => {
      client.removeQueries({ queryKey: clusterAccountKeys.detail(path.id) });
      for (const queryKey of accountKeys()) client.invalidateQueries({ queryKey });
    },
  });
}
