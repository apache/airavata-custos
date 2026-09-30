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
import { cn } from "@/lib/utils";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { Input } from "@/shared/ui/input";
import { Skeleton } from "@/shared/ui/skeleton";
import { Search } from "lucide-react";
import * as React from "react";
import { hasOptionalCount, statusMeta, utcCalendarDaysBetween } from "../lib";
import { useUserActivity, useUserActivityAnalytics } from "../queries";
import type { ActivitySortKey, ActivityView, SortDirection, UserActivityRow } from "../schemas";
import { LoginTrend, StatusComposition, SummaryCard } from "./ActivityOverview";
import { ActivityTable } from "./ActivityTable";
import { DaysRangePicker } from "./DaysRangePicker";
import { UserActivityDrawer } from "./UserActivityDrawer";

export function ActivityPage() {
  const ability = useAbility();
  if (ability.cannot("read", "UserActivity")) return <ErrorState message="Not permitted." />;
  return <ActivityDashboard />;
}
function ActivityDashboard() {
  const [windowDays, setWindowDays] = React.useState(30);
  const [view, setView] = React.useState<ActivityView>("all");
  const [query, setQuery] = React.useState("");
  const [search, setSearch] = React.useState("");
  const [page, setPage] = React.useState(1);
  const [pageSize, setPageSize] = React.useState(10);
  const [sort, setSort] = React.useState<ActivitySortKey>("last_login");
  const [direction, setDirection] = React.useState<SortDirection>("desc");
  const [selectedUser, setSelectedUser] = React.useState<UserActivityRow | null>(null);
  React.useEffect(() => {
    if (query.trim() === search) return;
    const timer = setTimeout(() => {
      setSearch(query.trim());
      setPage(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [query, search]);
  const users = useUserActivity({
    window: windowDays,
    status: view,
    query: search,
    limit: pageSize,
    offset: (page - 1) * pageSize,
    sort,
    direction,
  });
  const analytics = useUserActivityAnalytics(windowDays);
  const data = analytics.data;
  const counts = data
    ? {
        active: data.active_users,
        dormant: data.users_ever_logged_in - data.active_users,
        never: data.total_users - data.users_ever_logged_in,
      }
    : null;
  const activeChange =
    data && hasOptionalCount(data.prior_active_users)
      ? data.active_users - data.prior_active_users
      : undefined;
  const oldestAccountDays =
    data?.oldest_never_created_at != null
      ? utcCalendarDaysBetween(data.generated_at, data.oldest_never_created_at)
      : null;
  const showDormantOver90 =
    data !== undefined &&
    hasOptionalCount(data.dormant_over_90_days) &&
    data.dormant_over_90_days > 0 &&
    windowDays < 90;
  function changeView(value: ActivityView) {
    setView(value);
    setPage(1);
  }
  return (
    <div className="space-y-4">
      <section aria-labelledby="activity-overview-heading">
        <div className="flex flex-wrap items-baseline gap-3">
          <h2 id="activity-overview-heading" className="font-display text-xl font-semibold">
            User activity overview
          </h2>
          <span className="font-mono text-xs uppercase tracking-wide text-muted-foreground">
            OIDC-linked identities
          </span>
          <span className="text-xs text-muted-foreground">
            {process.env.NEXT_PUBLIC_PORTAL_USE_MSW === "true" ? "Demo data · " : ""}Refreshes every
            minute
          </span>
        </div>
        <p className="text-sm text-muted-foreground">
          Sign-in activity for every linked identity. Pick a status below to filter the audit table.
        </p>
      </section>
      <Card>
        <CardContent className="flex flex-wrap items-center justify-between gap-4 px-4 py-3">
          <div className="flex flex-wrap items-center gap-3">
            <span className="font-mono text-xs uppercase text-muted-foreground">
              Activity window
            </span>
            <DaysRangePicker
              value={windowDays}
              onChange={(days) => {
                setWindowDays(days);
                setPage(1);
                setSelectedUser(null);
              }}
              label="Activity window"
            />
          </div>
          <p className="max-w-sm text-xs text-muted-foreground">
            Active means a sign-in within the last {windowDays} user-local calendar days, including
            today. Statuses and charts follow this window; All time is lifetime activity.
          </p>
        </CardContent>
      </Card>
      {analytics.isPending ? (
        <Skeleton aria-label="Loading activity overview" className="h-64 w-full" />
      ) : analytics.error ? (
        <ErrorState
          message="We couldn't load user activity summaries."
          onRetry={() => analytics.refetch()}
        />
      ) : data && counts ? (
        <>
          <fieldset
            className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4"
            aria-label="User activity summaries"
          >
            <SummaryCard
              title="All identities"
              value={data.total_users}
              detail="Provisioned through OIDC"
              color="var(--custos-blue-500)"
              selected={view === "all"}
              onClick={() => changeView("all")}
            />
            {(Object.keys(counts) as (keyof typeof counts)[]).map((status) => (
              <SummaryCard
                key={status}
                title={statusMeta[status].label}
                value={counts[status]}
                percent={
                  data.total_users ? Math.round((counts[status] / data.total_users) * 100) : 0
                }
                detail={
                  status === "active" ? (
                    <>
                      Signed in within {windowDays} days
                      {activeChange !== undefined ? (
                        <>
                          {" "}
                          ·{" "}
                          <span
                            className={
                              activeChange < 0
                                ? "text-[color:var(--tone-error-fg)]"
                                : activeChange > 0
                                  ? "text-[color:var(--tone-ok-fg)]"
                                  : undefined
                            }
                          >
                            {activeChange > 0 ? "+" : ""}
                            {activeChange}
                          </span>{" "}
                          vs prior {windowDays} days
                        </>
                      ) : null}
                    </>
                  ) : status === "dormant" ? (
                    <>
                      No sign-in for {windowDays}+ days
                      {showDormantOver90 ? ` · ${data.dormant_over_90_days} over 90 days` : ""}
                    </>
                  ) : (
                    <>
                      No recorded sign-in
                      {oldestAccountDays !== null
                        ? ` · oldest account created ${oldestAccountDays} days ago`
                        : ""}
                    </>
                  )
                }
                color={statusMeta[status].color}
                selected={view === status}
                onClick={() => changeView(status)}
              />
            ))}
          </fieldset>
          <StatusComposition counts={counts} />
          <LoginTrend analytics={data} />
        </>
      ) : null}
      <Card className="gap-0 pb-0" aria-labelledby="activity-audit-heading">
        <CardHeader className="flex flex-row flex-wrap items-center justify-between gap-3 border-b pb-4">
          <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
            <CardTitle className="text-lg">
              <h2 id="activity-audit-heading">User activity audit</h2>
            </CardTitle>
            <p className="text-xs text-muted-foreground">
              {users.error
                ? "Activity unavailable"
                : users.data
                  ? `${users.data.total} matching identities`
                  : "Loading identities"}
            </p>
          </div>
          <div className="flex w-full flex-wrap gap-3 sm:w-auto">
            <select
              className="h-9 min-w-36 rounded-md border bg-background px-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              aria-label="Filter users by status"
              value={view}
              onChange={(event) => changeView(event.target.value as ActivityView)}
            >
              <option value="all">All</option>
              <option value="active">Active</option>
              <option value="dormant">Dormant</option>
              <option value="never">Never signed in</option>
            </select>
            <div className="relative min-w-0 flex-1 sm:w-72">
              <Search
                aria-hidden="true"
                className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
              />
              <Input
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Search by name or email"
                aria-label="Search users"
                className="h-9 w-full pl-9"
              />
            </div>
          </div>
        </CardHeader>
        <CardContent className="px-0">
          {users.isPending || query.trim() !== search ? (
            <Skeleton aria-label="Loading activity table" className="h-64 w-full" />
          ) : users.error ? (
            <ErrorState
              message="We couldn't load the selected users."
              onRetry={() => users.refetch()}
            />
          ) : users.data ? (
            // Held-over rows stay readable but dimmed, so a slow page change
            // never looks like fresh data for the page being requested.
            <div
              aria-busy={users.isPlaceholderData}
              className={cn("transition-opacity", users.isPlaceholderData && "opacity-60")}
            >
              <ActivityTable
                rows={users.data.items}
                total={users.data.total}
                windowDays={windowDays}
                page={page}
                pageSize={pageSize}
                sort={sort}
                direction={direction}
                busy={users.isPlaceholderData}
                onSort={(key) => {
                  setSort(key);
                  setDirection(sort === key && direction === "asc" ? "desc" : "asc");
                  setPage(1);
                }}
                onPage={setPage}
                onPageSize={(size) => {
                  setPageSize(size);
                  setPage(1);
                }}
                onSelect={setSelectedUser}
              />
            </div>
          ) : null}
        </CardContent>
      </Card>
      {selectedUser ? (
        <UserActivityDrawer
          key={`${selectedUser.user_id}-${windowDays}`}
          user={selectedUser}
          initialWindow={windowDays}
          onClose={() => setSelectedUser(null)}
        />
      ) : null}
    </div>
  );
}
