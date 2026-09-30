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
  type ActivitySortKey,
  type ActivityView,
  type SortDirection,
  userActivityAnalyticsSchema,
  userActivityListSchema,
} from "./schemas";

export type UserActivityParams = {
  window: number;
  status: ActivityView;
  query: string;
  limit: number;
  offset: number;
  sort: ActivitySortKey;
  direction: SortDirection;
};
export async function getUserActivity(params: UserActivityParams) {
  const search = new URLSearchParams(
    Object.entries(params).map(([key, value]) => [key, String(value)]),
  );
  const data = userActivityListSchema.parse(await apiFetch(`/users/activity?${search}`));
  // Older backends ignore unknown filters. Never display unfiltered rows as a filtered result.
  if (
    data.window_days !== params.window ||
    data.status !== params.status ||
    data.offset !== params.offset ||
    data.limit !== params.limit
  ) {
    throw new Error("The activity response does not match the requested filters.");
  }
  return data;
}
export async function getUserActivityAnalytics(windowDays: number, userID?: string) {
  const path = userID
    ? `/users/${encodeURIComponent(userID)}/activity/analytics`
    : "/users/activity/analytics";
  const data = userActivityAnalyticsSchema.parse(await apiFetch(`${path}?window=${windowDays}`));
  if (data.window_days !== windowDays)
    throw new Error("The activity response does not match the requested window.");
  return data;
}
