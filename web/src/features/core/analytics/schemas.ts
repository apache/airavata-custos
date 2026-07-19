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

// Credits are integer SUs well inside the safe-integer range, so plain numbers
// (not the generated int64 bigint) keep the charts simple.
export const analyticsRoleSchema = z.enum(["PI", "CO_PI", "ALLOCATION_MANAGER", "MEMBER"]);
export type AnalyticsRole = z.infer<typeof analyticsRoleSchema>;

export const analyticsAllocationSchema = z.object({
  id: z.string(),
  name: z.string(),
  status: z.string(),
  initial_su_amount: z.number(),
  used_su_amount: z.number(),
  start_time: z.string(),
  end_time: z.string(),
});
export type AnalyticsAllocation = z.infer<typeof analyticsAllocationSchema>;

export const analyticsContextSchema = z.object({
  project_id: z.string(),
  project_name: z.string(),
  role: analyticsRoleSchema,
  allocations: z
    .array(analyticsAllocationSchema)
    .nullish()
    .transform((v) => v ?? []),
});
export type AnalyticsContext = z.infer<typeof analyticsContextSchema>;

export const analyticsContextsSchema = z.array(analyticsContextSchema);

export const usageDailyBucketSchema = z.object({
  date: z.string(),
  by_resource: z
    .record(z.string(), z.number())
    .nullish()
    .transform((v) => v ?? {}),
});
export type UsageDailyBucket = z.infer<typeof usageDailyBucketSchema>;

export const usageResourceSchema = z.object({
  resource_id: z.string(),
  name: z.string(),
  resource_type: z.string(),
  used: z.number(),
  // Null in v1: per-resource caps are time-varying and not yet computed.
  cap: z.number().nullish(),
  used_native: z.number(),
  native_unit: z.string(),
  used_by_caller: z.number(),
});
export type UsageResource = z.infer<typeof usageResourceSchema>;

export const usageMemberSchema = z.object({
  user_id: z.string(),
  name: z.string(),
  used: z.number(),
});
export type UsageMember = z.infer<typeof usageMemberSchema>;

export const usageSummarySchema = z.object({
  total: z.number(),
  used: z.number(),
  daily: z
    .array(usageDailyBucketSchema)
    .nullish()
    .transform((v) => v ?? []),
  by_resource: z
    .array(usageResourceSchema)
    .nullish()
    .transform((v) => v ?? []),
  // null means the caller may not see per-member data; [] means a manager on an
  // allocation with no usage yet. The distinction drives whether the widget
  // renders at all.
  by_member: z
    .array(usageMemberSchema)
    .nullish()
    .transform((v) => v ?? null),
});
export type UsageSummary = z.infer<typeof usageSummarySchema>;

export const analyticsJobSchema = z.object({
  id: z.string(),
  job_id: z.string(),
  calculated_time: z.string(),
  user_id: z.string(),
  user_name: z.string(),
  resource_id: z.string(),
  resource_name: z.string(),
  resource_type: z.string(),
  used_raw: z.number(),
  native_unit: z.string(),
  used: z.number(),
});
export type AnalyticsJob = z.infer<typeof analyticsJobSchema>;

export const allocationJobsSchema = z.object({
  jobs: z
    .array(analyticsJobSchema)
    .nullish()
    .transform((v) => v ?? []),
  total: z.number(),
});
export type AllocationJobs = z.infer<typeof allocationJobsSchema>;
