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

import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { getAllocationJobs, getAnalyticsContexts, getUsageSummary, type JobsParams } from "./api";

export const analyticsKeys = {
  all: ["analytics"] as const,
  contexts: () => [...analyticsKeys.all, "contexts"] as const,
  summary: (allocationId: string) => [...analyticsKeys.all, "summary", allocationId] as const,
  jobs: (allocationId: string, params: JobsParams) =>
    [...analyticsKeys.all, "jobs", allocationId, params] as const,
};

const DEFAULTS = {
  staleTime: 30_000,
  gcTime: 300_000,
  refetchOnWindowFocus: false,
} as const;

export function useAnalyticsContexts() {
  const { status } = useSession();
  return useQuery({
    queryKey: analyticsKeys.contexts(),
    queryFn: getAnalyticsContexts,
    enabled: status === "authenticated",
    ...DEFAULTS,
  });
}

export function useUsageSummary(allocationId: string | undefined) {
  return useQuery({
    queryKey: allocationId
      ? analyticsKeys.summary(allocationId)
      : [...analyticsKeys.all, "summary", "none"],
    queryFn: () => getUsageSummary(allocationId as string),
    enabled: Boolean(allocationId),
    ...DEFAULTS,
  });
}

export function useAllocationJobs(allocationId: string | undefined, params: JobsParams) {
  return useQuery({
    queryKey: allocationId
      ? analyticsKeys.jobs(allocationId, params)
      : [...analyticsKeys.all, "jobs", "none"],
    queryFn: () => getAllocationJobs(allocationId as string, params),
    enabled: Boolean(allocationId),
    ...DEFAULTS,
  });
}
