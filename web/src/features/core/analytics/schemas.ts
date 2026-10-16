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

export const analyticsContextSchema = zProjectContext.extend({
  role: analyticsRoleSchema,
  allocations: listOf(zAllocation),
});
export type AnalyticsContext = z.infer<typeof analyticsContextSchema>;

export const analyticsContextsSchema = z.array(analyticsContextSchema);

export const usageDailyBucketSchema = zUsageDailyBucket.extend({
  by_resource: zUsageDailyBucket.shape.by_resource.transform((v) => v ?? {}),
});
export type UsageDailyBucket = z.infer<typeof usageDailyBucketSchema>;

export const usageSummarySchema = zUsageSummary.extend({
  daily: listOf(usageDailyBucketSchema),
  by_resource: listOf(zUsageResource),
  // null means the caller may not see per-member data; [] means a manager on an
  // allocation with no usage yet. The distinction drives whether the widget
  // renders at all.
  by_member: z.array(zUsageMember).nullable(),
});
export type UsageSummary = z.infer<typeof usageSummarySchema>;

export const allocationJobsSchema = zAllocationJobs.extend({
  jobs: listOf(zJob),
});
export type AllocationJobs = z.infer<typeof allocationJobsSchema>;
