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

import type { ClusterView, Conflict } from "@/generated/slurm/zod.gen";
import { http, HttpResponse } from "msw";

const CLUSTER_VIEW = "*/api/v1/connectors/slurm/clusters/:id";

// Every cluster reads in sync, with no partitions or system accounts.
const inSync: ClusterView = { partitions: [], system_accounts: [], conflicts: [] };

// Overrides parts of the cluster view; for server.use.
export const slurmView = (view: Partial<ClusterView>) =>
  http.get(CLUSTER_VIEW, () => HttpResponse.json({ ...inSync, ...view }));

export const slurmConflicts = (conflicts: Conflict[]) => slurmView({ conflicts });

export const slurmHandlers = [
  http.get(CLUSTER_VIEW, () => HttpResponse.json(inSync)),
  http.post(`${CLUSTER_VIEW}/conflicts/resolve`, () => new HttpResponse(null, { status: 204 })),
];
