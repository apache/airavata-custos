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
  getConnectorsAnalyticsAllocationsByIdJobs,
  getConnectorsAnalyticsAllocationsByIdUsageSummary,
  getConnectorsAnalyticsContexts,
} from "@/generated/analytics/sdk.gen";
import type { GetConnectorsAnalyticsAllocationsByIdJobsData } from "@/generated/analytics/types.gen";
import { skipToken, useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";

export const analyticsKeys = {
  all: ["analytics"] as const,
  contexts: () => [...analyticsKeys.all, "contexts"] as const,
  summary: (allocationId: string) => [...analyticsKeys.all, "summary", allocationId] as const,
  jobs: (
    allocationId: string,
    query: GetConnectorsAnalyticsAllocationsByIdJobsData["query"],
  ) => [...analyticsKeys.all, "jobs", allocationId, query] as const,
};

export function useAnalyticsContexts() {
  const { status } = useSession();
  return useQuery({
    queryKey: analyticsKeys.contexts(),
    queryFn: () => getConnectorsAnalyticsContexts(),
    enabled: status === "authenticated",
  });
}

export function useUsageSummary(allocationId: string | undefined) {
  return useQuery({
    queryKey: analyticsKeys.summary(allocationId ?? ""),
    queryFn: allocationId
      ? () => getConnectorsAnalyticsAllocationsByIdUsageSummary({ path: { id: allocationId } })
      : skipToken,
  });
}

export function useAllocationJobs(
  allocationId: string | undefined,
  query: GetConnectorsAnalyticsAllocationsByIdJobsData["query"],
) {
  return useQuery({
    queryKey: analyticsKeys.jobs(allocationId ?? "", query),
    queryFn: allocationId
      ? () => getConnectorsAnalyticsAllocationsByIdJobs({ path: { id: allocationId }, query })
      : skipToken,
  });
}
