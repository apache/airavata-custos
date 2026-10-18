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

import { apiFetch } from "@/shared/api/client";
import type { PostComputeClusterUsersByIdDenyData } from "@/generated/core/types.gen";
import {
  type ClusterAccountListResponse,
  type ComputeClusterUser,
  zGetComputeClusterUsersResponse,
  zGetUsersByIdComputeClusterUsersResponse,
  zPostComputeClusterUsersByIdApproveResponse,
  zPostComputeClusterUsersByIdDenyResponse,
} from "@/generated/core/zod.gen";
import type { ClusterAccountListParams } from "./types";

export async function listClusterAccounts(
  params: ClusterAccountListParams = {},
): Promise<ClusterAccountListResponse> {
  const search = new URLSearchParams();
  if (params.approval_status) search.set("approval_status", params.approval_status);
  if (typeof params.limit === "number") search.set("limit", String(params.limit));
  if (typeof params.offset === "number") search.set("offset", String(params.offset));
  const query = search.toString();
  return zGetComputeClusterUsersResponse.parse(
    await apiFetch(`/compute-cluster-users${query ? `?${query}` : ""}`),
  );
}

export async function listUserClusterAccounts(userId: string): Promise<ComputeClusterUser[]> {
  return (
    zGetUsersByIdComputeClusterUsersResponse.parse(
      await apiFetch(`/users/${userId}/compute-cluster-users`),
    ) ?? []
  );
}

export async function approveClusterAccount(id: string): Promise<ComputeClusterUser> {
  return zPostComputeClusterUsersByIdApproveResponse.parse(
    await apiFetch(`/compute-cluster-users/${id}/approve`, { method: "POST" }),
  );
}

export async function denyClusterAccount(id: string, note: string): Promise<ComputeClusterUser> {
  const body: PostComputeClusterUsersByIdDenyData["body"] = { note };
  return zPostComputeClusterUsersByIdDenyResponse.parse(
    await apiFetch(`/compute-cluster-users/${id}/deny`, { method: "POST", body }),
  );
}
