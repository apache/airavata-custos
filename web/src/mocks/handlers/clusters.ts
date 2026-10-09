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

import { http, HttpResponse } from "msw";
import clusterAccountsFixture from "@/features/core/cluster-accounts/__fixtures__/cluster-accounts.json";
import clusterUsersFixture from "@/features/core/clusters/__fixtures__/cluster-users.json";
import clustersFixture from "@/features/core/clusters/__fixtures__/clusters.json";
import type {
  ClusterAccountResponse,
  ComputeCluster,
  ComputeClusterUser,
  GetComputeClustersByIdData,
  GetComputeClustersByIdUsersData,
  GetUsersByIdComputeClusterUsersData,
  PostComputeClusterUsersByIdApproveData,
  PostComputeClusterUsersByIdDenyData,
} from "@/generated/core/types.gen";
import {
  zClusterAccountResponse,
  zComputeCluster,
  zComputeClusterUser,
} from "@/generated/core/zod.gen";
import { z } from "zod";
import { notFound, page } from "../paging";

const clusters: ComputeCluster[] = z.array(zComputeCluster).parse(clustersFixture);
const clusterUsers: ComputeClusterUser[] = z.array(zComputeClusterUser).parse(clusterUsersFixture);
export const clusterAccounts: ClusterAccountResponse[] = z
  .array(zClusterAccountResponse)
  .parse(clusterAccountsFixture);

export const clustersHandlers = [
  http.get("*/api/v1/compute-clusters", () => HttpResponse.json(clusters)),

  http.get("*/api/v1/compute-cluster-users", ({ request }) => {
    const url = new URL(request.url);
    const status = url.searchParams.get("approval_status");
    const rows = status
      ? clusterAccounts.filter((a) => a.approval_status === status)
      : clusterAccounts;
    const { items, total } = page(url, rows);
    return HttpResponse.json({ items, total });
  }),

  http.post<PostComputeClusterUsersByIdApproveData["path"]>(
    "*/api/v1/compute-cluster-users/:id/approve",
    ({ params }) => {
      const found = clusterAccounts.find((a) => a.id === params.id);
      if (!found) return notFound("cluster user");
      if (found.approval_status === "APPROVED") {
        return HttpResponse.json({ error: "already approved" }, { status: 409 });
      }
      found.approval_status = "APPROVED";
      found.reviewed_at = new Date().toISOString();
      found.reviewed_by = "user-admin";
      found.review_note = undefined;
      return HttpResponse.json(found);
    },
  ),

  http.post<
    PostComputeClusterUsersByIdDenyData["path"],
    PostComputeClusterUsersByIdDenyData["body"]
  >("*/api/v1/compute-cluster-users/:id/deny", async ({ params, request }) => {
    const found = clusterAccounts.find((a) => a.id === params.id);
    if (!found) return notFound("cluster user");
    if (found.approval_status !== "PENDING") {
      return HttpResponse.json({ error: "already reviewed" }, { status: 409 });
    }
    const note = (await request.json()).note?.trim();
    if (!note) {
      return HttpResponse.json({ error: "a reason is required to deny" }, { status: 400 });
    }
    found.approval_status = "DENIED";
    found.reviewed_at = new Date().toISOString();
    found.reviewed_by = "user-admin";
    found.review_note = note;
    return HttpResponse.json(found);
  }),

  http.get<GetComputeClustersByIdData["path"]>("*/api/v1/compute-clusters/:id", ({ params }) => {
    const found = clusters.find((c) => c.id === params.id);
    if (!found) return notFound("cluster");
    return HttpResponse.json(found);
  }),

  http.get<GetComputeClustersByIdUsersData["path"]>(
    "*/api/v1/compute-clusters/:id/users",
    ({ params }) =>
      HttpResponse.json(clusterUsers.filter((u) => u.compute_cluster_id === params.id)),
  ),

  http.get<GetUsersByIdComputeClusterUsersData["path"]>(
    "*/api/v1/users/:id/compute-cluster-users",
    ({ params }) => HttpResponse.json(clusterUsers.filter((u) => u.user_id === params.id)),
  ),
];
