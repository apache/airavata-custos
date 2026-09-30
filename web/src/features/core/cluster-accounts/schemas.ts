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

import { z } from "zod";
import {
  zClusterAccountApproval,
  zClusterAccountResponse,
  zComputeClusterUser,
} from "@/generated/core/zod.gen";

export const clusterAccountApprovalSchema = zClusterAccountApproval;
export type ClusterAccountApproval = z.infer<typeof clusterAccountApprovalSchema>;

// The generated schema marks every field optional and provisioned_at is null
// until the account exists on the cluster.
export const clusterAccountSchema = zClusterAccountResponse
  .extend({ provisioned_at: z.string().nullish() })
  .required({
    id: true,
    compute_cluster_id: true,
    user_id: true,
    local_username: true,
    approval_status: true,
    display_name: true,
    email: true,
    cluster_name: true,
  });
export type ClusterAccount = z.infer<typeof clusterAccountSchema>;

export const clusterAccountListResponseSchema = z.object({
  items: z.array(clusterAccountSchema),
  total: z.number().int().nonnegative(),
});
export type ClusterAccountListResponse = z.infer<typeof clusterAccountListResponseSchema>;

export const reviewedClusterUserSchema = zComputeClusterUser
  .extend({ provisioned_at: z.string().nullish() })
  .required({ id: true, approval_status: true });
export type ReviewedClusterUser = z.infer<typeof reviewedClusterUserSchema>;
