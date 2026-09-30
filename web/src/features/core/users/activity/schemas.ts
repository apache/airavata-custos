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

const count = z.number().int().nonnegative();
const optionalCount = count.nullish();
export const activityViewSchema = z.enum(["all", "active", "dormant", "never"]);
export type ActivityView = z.infer<typeof activityViewSchema>;
export type ActivitySortKey = "name" | "last_login" | "login_count";
export type SortDirection = "asc" | "desc";
export const userActivityRowSchema = z.object({
  user_id: z.string(),
  name: z.string(),
  email: z.string(),
  role_names: z
    .array(z.string().min(1))
    .nullable()
    .optional()
    .transform((names) => names ?? undefined),
  last_login: z.string().datetime({ offset: true }).nullable(),
  inactive_days: z.number().int().nullable(),
  login_count: count,
  window_login_count: count,
  login_day_count: count,
  current_streak: count,
});
export type UserActivityRow = z.infer<typeof userActivityRowSchema>;
export const userActivityListSchema = z.object({
  items: z
    .array(userActivityRowSchema)
    .nullable()
    .transform((items) => items ?? []),
  total: count,
  limit: z.number().int().positive(),
  offset: count,
  window_days: z.number().int().min(1).max(365),
  status: activityViewSchema,
});
export type UserActivityList = z.infer<typeof userActivityListSchema>;
export const userActivityAnalyticsSchema = z
  .object({
    generated_at: z.string().datetime({ offset: true }),
    window_days: z.number().int().min(1).max(365),
    total_users: count,
    users_ever_logged_in: count,
    active_users: count,
    prior_active_users: optionalCount,
    dormant_over_90_days: optionalCount,
    oldest_never_created_at: z.string().datetime({ offset: true }).nullish(),
    lifetime_login_count: count,
    lifetime_active_days: count,
    window_login_count: count,
    window_active_days: count,
    trend: z.array(z.object({ date: z.string().date(), active_users: count, login_count: count })),
  })
  .refine(
    (data) =>
      data.active_users <= data.users_ever_logged_in &&
      data.users_ever_logged_in <= data.total_users,
    {
      message: "Activity counts must describe mutually exclusive populations",
    },
  );
export type UserActivityAnalytics = z.infer<typeof userActivityAnalyticsSchema>;
