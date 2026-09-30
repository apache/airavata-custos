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

import { activityStatus } from "@/features/core/users/activity/lib";
import { http, HttpResponse } from "msw";

// Fixed UTC fixture clock; demo data stays deterministic across machines and dates.
const NOW = "2026-09-16T12:00:00Z";
const dateAt = (days: number) => new Date(Date.parse(NOW) - days * 86_400_000).toISOString();
export const activityUsers = Array.from({ length: 225 }, (_, index) => {
  const days = [0, 1, 6, 7, 29, 30, 89, 120, null][index % 9] ?? null;
  const events =
    days === null
      ? []
      : Array.from({ length: 12 }, (_, offset) => {
          // Vary earlier visits across weekdays while preserving each user's last sign-in.
          const eventDays = days + offset * 7 - (offset === 0 ? 0 : (index + offset) % 5);
          const weekday = new Date(dateAt(eventDays)).getUTCDay();
          return {
            days: eventDays,
            count: weekday === 0 || weekday === 6 ? 1 : 1 + (index % 3),
          };
        });
  return {
    user_id: `activity-${index}`,
    name: `Activity User ${String(index + 1).padStart(3, "0")}`,
    email: `activity${index + 1}@example.org`,
    role_names: [["Staff"], ["Student"], ["Researcher"]][index % 3],
    created_at: dateAt(100 + index),
    last_login: days === null ? null : dateAt(days),
    inactive_days: days,
    login_count: events.reduce((sum, event) => sum + event.count, 0),
    login_day_count: events.length,
    current_streak: days !== null && days <= 1 ? 1 : 0,
    events,
  };
});
export function activityAnalytics(windowDays: number, userID?: string) {
  const users = userID ? activityUsers.filter((user) => user.user_id === userID) : activityUsers;
  const trend = Array.from({ length: windowDays }, (_, offset) => {
    const days = windowDays - offset - 1;
    return {
      date: dateAt(days).slice(0, 10),
      active_users: users.filter((user) => user.events.some((event) => event.days === days)).length,
      login_count: users.reduce(
        (sum, user) => sum + (user.events.find((event) => event.days === days)?.count ?? 0),
        0,
      ),
    };
  });
  return {
    generated_at: NOW,
    prior_active_users: users.filter((user) =>
      user.events.some((event) => event.days >= windowDays && event.days < 2 * windowDays),
    ).length,
    dormant_over_90_days: users.filter(
      (user) =>
        user.inactive_days !== null && user.inactive_days >= windowDays && user.inactive_days > 90,
    ).length,
    oldest_never_created_at:
      users
        .filter((user) => user.last_login === null)
        .map((user) => user.created_at)
        .sort()[0] ?? null,
    window_days: windowDays,
    total_users: users.length,
    users_ever_logged_in: users.filter((user) => user.last_login !== null).length,
    active_users: users.filter(
      (user) => user.inactive_days !== null && user.inactive_days < windowDays,
    ).length,
    lifetime_login_count: users.reduce((sum, user) => sum + user.login_count, 0),
    lifetime_active_days: users.reduce((sum, user) => sum + user.login_day_count, 0),
    window_login_count: trend.reduce((sum, point) => sum + point.login_count, 0),
    window_active_days: trend.reduce((sum, point) => sum + point.active_users, 0),
    trend,
  };
}
export function activityList(url: URL) {
  const windowDays = Number(url.searchParams.get("window") ?? 30);
  const status = url.searchParams.get("status") ?? "all";
  const query = (url.searchParams.get("query") ?? "").toLowerCase();
  const sort = url.searchParams.get("sort") ?? "last_login";
  const direction = url.searchParams.get("direction") === "asc" ? 1 : -1;
  const limit = Number(url.searchParams.get("limit") ?? 10);
  const offset = Number(url.searchParams.get("offset") ?? 0);
  const items = activityUsers
    .map(({ events, ...user }) => ({
      ...user,
      window_login_count: events
        .filter((event) => event.days < windowDays)
        .reduce((sum, event) => sum + event.count, 0),
    }))
    .filter(
      (user) =>
        (status === "all" || activityStatus(user, windowDays) === status) &&
        `${user.name} ${user.email}`.toLowerCase().includes(query),
    )
    .sort((a, b) => {
      const key = sort === "name" || sort === "login_count" ? sort : "last_login";
      const left = a[key];
      const right = b[key];
      if (left === right) return a.user_id.localeCompare(b.user_id);
      if (left === null) return 1;
      if (right === null) return -1;
      return (left < right ? -1 : 1) * direction;
    });
  return {
    items: items.slice(offset, offset + limit),
    total: items.length,
    limit,
    offset,
    window_days: windowDays,
    status,
  };
}
export const userActivityHandlers = [
  http.get("*/api/v1/users/activity", ({ request }) =>
    HttpResponse.json(activityList(new URL(request.url))),
  ),
  http.get("*/api/v1/users/activity/analytics", ({ request }) =>
    HttpResponse.json(
      activityAnalytics(Number(new URL(request.url).searchParams.get("window") ?? 30)),
    ),
  ),
  http.get("*/api/v1/users/:id/activity/analytics", ({ request, params }) => {
    if (!activityUsers.some((user) => user.user_id === params.id))
      return HttpResponse.json({ error: "Not found" }, { status: 404 });
    return HttpResponse.json(
      activityAnalytics(
        Number(new URL(request.url).searchParams.get("window") ?? 30),
        String(params.id),
      ),
    );
  }),
];
