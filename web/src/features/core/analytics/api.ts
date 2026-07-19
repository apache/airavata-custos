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
import {
  type AllocationJobs,
  allocationJobsSchema,
  type AnalyticsContext,
  analyticsContextsSchema,
  type UsageSummary,
  usageSummarySchema,
} from "./schemas";

export async function getAnalyticsContexts(): Promise<AnalyticsContext[]> {
  const raw = await apiFetch("/connectors/analytics/contexts");
  return analyticsContextsSchema.parse(raw);
}

export async function getUsageSummary(allocationId: string): Promise<UsageSummary> {
  const raw = await apiFetch(`/connectors/analytics/allocations/${allocationId}/usage-summary`);
  return usageSummarySchema.parse(raw);
}

export type JobsParams = { mine?: boolean; limit?: number; offset?: number };

export async function getAllocationJobs(
  allocationId: string,
  params: JobsParams = {},
): Promise<AllocationJobs> {
  const search = new URLSearchParams();
  if (params.mine) search.set("mine", "true");
  if (typeof params.limit === "number") search.set("limit", String(params.limit));
  if (typeof params.offset === "number") search.set("offset", String(params.offset));
  const qs = search.toString();
  const raw = await apiFetch(
    `/connectors/analytics/allocations/${allocationId}/jobs${qs ? `?${qs}` : ""}`,
  );
  return allocationJobsSchema.parse(raw);
}
