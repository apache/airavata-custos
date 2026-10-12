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

import {
  zAllocation,
  zAllocationJobs,
  zJob,
  zProjectContext,
  zUsageDailyBucket,
  zUsageMember,
  zUsageResource,
  zUsageSummary,
} from "@/generated/analytics/zod.gen";
import { z } from "zod";

const listOf = <T extends z.ZodType>(item: T) =>
  z
    .array(item)
    .nullable()
    .transform((v) => v ?? []);

export const analyticsRoleSchema = z.enum(["PI", "CO_PI", "ALLOCATION_MANAGER", "MEMBER"]);
export type AnalyticsRole = z.infer<typeof analyticsRoleSchema>;

export const analyticsAllocationSchema = zAllocation;
export type AnalyticsAllocation = z.infer<typeof analyticsAllocationSchema>;

export const analyticsContextSchema = zProjectContext.extend({
  role: analyticsRoleSchema,
  allocations: listOf(analyticsAllocationSchema),
});
export type AnalyticsContext = z.infer<typeof analyticsContextSchema>;

export const analyticsContextsSchema = z.array(analyticsContextSchema);

export const usageDailyBucketSchema = zUsageDailyBucket.extend({
  by_resource: zUsageDailyBucket.shape.by_resource.transform((v) => v ?? {}),
});
export type UsageDailyBucket = z.infer<typeof usageDailyBucketSchema>;

// cap is null in v1: per-resource caps are time-varying and not yet computed.
export const usageResourceSchema = zUsageResource;
export type UsageResource = z.infer<typeof usageResourceSchema>;

export const usageMemberSchema = zUsageMember;
export type UsageMember = z.infer<typeof usageMemberSchema>;

export const usageSummarySchema = zUsageSummary.extend({
  daily: listOf(usageDailyBucketSchema),
  by_resource: listOf(usageResourceSchema),
  // null means the caller may not see per-member data; [] means a manager on an
  // allocation with no usage yet. The distinction drives whether the widget
  // renders at all.
  by_member: z.array(usageMemberSchema).nullable(),
});
export type UsageSummary = z.infer<typeof usageSummarySchema>;

export const analyticsJobSchema = zJob;
export type AnalyticsJob = z.infer<typeof analyticsJobSchema>;

export const allocationJobsSchema = zAllocationJobs.extend({
  jobs: listOf(analyticsJobSchema),
});
export type AllocationJobs = z.infer<typeof allocationJobsSchema>;
