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

import clusterAccountsFixture from "@/features/core/cluster-accounts/__fixtures__/cluster-accounts.json";
import clusterUsersFixture from "@/features/core/clusters/__fixtures__/cluster-users.json";
import clustersFixture from "@/features/core/clusters/__fixtures__/clusters.json";
import type {
  ClusterAccountResponse,
  ComputeCluster,
  ComputeClusterUser,
  DeleteComputeClusterUsersByIdData,
  GetComputeClusterUsersByIdData,
  GetComputeClustersByIdData,
  GetComputeClustersByIdUsersData,
  GetUsersByIdComputeClusterUsersData,
  PostComputeClusterUsersByIdApproveData,
  PostComputeClusterUsersByIdDenyData,
  PostComputeClusterUsersData,
  PutComputeClusterUsersByIdData,
} from "@/generated/core/types.gen";
import {
  zClusterAccountResponse,
  zComputeCluster,
  zComputeClusterUser,
} from "@/generated/core/zod.gen";
import { http, HttpResponse } from "msw";
import { z } from "zod";
import { listJson, notFound, page } from "../paging";

const clusters: ComputeCluster[] = z.array(zComputeCluster).parse(clustersFixture);
const clusterUsers: ComputeClusterUser[] = z.array(zComputeClusterUser).parse(clusterUsersFixture);
// The approval queue's rows, which the list endpoint joins with user and cluster names.
export const clusterAccounts: ClusterAccountResponse[] = z
  .array(zClusterAccountResponse)
  .parse(clusterAccountsFixture);

const allAccounts = (): ComputeClusterUser[] => [...clusterAccounts, ...clusterUsers];
const findAccount = (id: string) => allAccounts().find((a) => a.id === id);
// Every response but the queue list is the bare mapping, without the joined names.
const bare = (a: ComputeClusterUser) => zComputeClusterUser.parse(a);

export const clustersHandlers = [
  http.get("*/api/v1/compute-clusters", () => listJson(clusters)),

  http.get("*/api/v1/compute-cluster-users", ({ request }) => {
    const url = new URL(request.url);
    const status = url.searchParams.get("approval_status");
    const rows = status
      ? clusterAccounts.filter((a) => a.approval_status === status)
      : clusterAccounts;
    const { items, total } = page(url, rows);
    return HttpResponse.json({ items, total });
  }),

  http.post<never, PostComputeClusterUsersData["body"]>(
    "*/api/v1/compute-cluster-users",
    async ({ request }) => {
      const { compute_cluster_id, user_id, local_username } = await request.json();
      if (!compute_cluster_id || !user_id || !local_username) {
        return HttpResponse.json({ error: "invalid input" }, { status: 400 });
      }
      if (!clusters.some((c) => c.id === compute_cluster_id)) {
        return HttpResponse.json({ error: "compute cluster not found" }, { status: 400 });
      }
      const now = new Date().toISOString();
      const created: ComputeClusterUser = {
        id: crypto.randomUUID(),
        compute_cluster_id,
        user_id,
        local_username,
        access_level: "USER",
        approval_status: "APPROVED",
        provisioned_at: null,
        reviewed_by: "24000005-0000-4000-8000-000024000005",
        reviewed_at: now,
      };
      clusterUsers.push(created);
      return HttpResponse.json(created, { status: 201 });
    },
  ),

  http.get<GetComputeClusterUsersByIdData["path"]>(
    "*/api/v1/compute-cluster-users/:id",
    ({ params }) => {
      const found = findAccount(params.id);
      return found ? HttpResponse.json(bare(found)) : notFound("compute cluster user");
    },
  ),

  // Blank fields keep the stored value, as in the backend; access level is not written.
  http.put<PutComputeClusterUsersByIdData["path"], PutComputeClusterUsersByIdData["body"]>(
    "*/api/v1/compute-cluster-users/:id",
    async ({ params, request }) => {
      const found = findAccount(params.id);
      if (!found) return notFound("compute cluster user");
      const body = await request.json();
      found.compute_cluster_id = body.compute_cluster_id || found.compute_cluster_id;
      found.user_id = body.user_id || found.user_id;
      found.local_username = body.local_username || found.local_username;
      const { compute_cluster_id, user_id, local_username } = found;
      return HttpResponse.json({ id: params.id, compute_cluster_id, user_id, local_username });
    },
  ),

  http.delete<DeleteComputeClusterUsersByIdData["path"]>(
    "*/api/v1/compute-cluster-users/:id",
    ({ params }) => {
      for (const list of [clusterAccounts, clusterUsers]) {
        const index = list.findIndex((a) => a.id === params.id);
        if (index >= 0) {
          list.splice(index, 1);
          return new HttpResponse(null, { status: 204 });
        }
      }
      return notFound("compute cluster user");
    },
  ),

  http.post<PostComputeClusterUsersByIdApproveData["path"]>(
    "*/api/v1/compute-cluster-users/:id/approve",
    ({ params }) => {
      const found = findAccount(params.id);
      if (!found) return notFound("compute cluster user");
      if (found.approval_status === "APPROVED") {
        return HttpResponse.json({ error: "already approved" }, { status: 409 });
      }
      found.approval_status = "APPROVED";
      found.reviewed_at = new Date().toISOString();
      found.reviewed_by = "24000005-0000-4000-8000-000024000005";
      found.review_note = undefined;
      return HttpResponse.json(bare(found));
    },
  ),

  http.post<
    PostComputeClusterUsersByIdDenyData["path"],
    PostComputeClusterUsersByIdDenyData["body"]
  >("*/api/v1/compute-cluster-users/:id/deny", async ({ params, request }) => {
    const found = findAccount(params.id);
    if (!found) return notFound("compute cluster user");
    const note = (await request.json()).note?.trim();
    if (!note) {
      return HttpResponse.json({ error: "a reason is required to deny" }, { status: 400 });
    }
    if (found.approval_status !== "PENDING") {
      return HttpResponse.json({ error: "already reviewed" }, { status: 409 });
    }
    found.approval_status = "DENIED";
    found.reviewed_at = new Date().toISOString();
    found.reviewed_by = "24000005-0000-4000-8000-000024000005";
    found.review_note = note;
    return HttpResponse.json(bare(found));
  }),

  http.get<GetComputeClustersByIdData["path"]>("*/api/v1/compute-clusters/:id", ({ params }) => {
    const found = clusters.find((c) => c.id === params.id);
    if (!found) return notFound("cluster");
    return HttpResponse.json(found);
  }),

  http.get<GetComputeClustersByIdUsersData["path"]>(
    "*/api/v1/compute-clusters/:id/users",
    ({ params }) =>
      listJson(
        allAccounts()
          .filter((u) => u.compute_cluster_id === params.id)
          .map(bare),
      ),
  ),

  http.get<GetUsersByIdComputeClusterUsersData["path"]>(
    "*/api/v1/users/:id/compute-cluster-users",
    ({ params }) =>
      listJson(
        allAccounts()
          .filter((u) => u.user_id === params.id)
          .map(bare),
      ),
  ),
];
