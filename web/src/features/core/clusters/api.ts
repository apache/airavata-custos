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
import { apiFetch } from "@/shared/api/client";
import {
  type ComputeCluster,
  type ComputeClusterUser,
  computeClusterSchema,
  computeClusterUserSchema,
} from "./schemas";

export async function listClusters(): Promise<ComputeCluster[]> {
  const raw = await apiFetch("/compute-clusters");
  return z.array(computeClusterSchema).parse(raw);
}

export async function getCluster(id: string): Promise<ComputeCluster> {
  const raw = await apiFetch(`/compute-clusters/${id}`);
  return computeClusterSchema.parse(raw);
}

export async function listClusterUsers(id: string): Promise<ComputeClusterUser[]> {
  const raw = await apiFetch(`/compute-clusters/${id}/users`);
  return z.array(computeClusterUserSchema).parse(raw);
}
