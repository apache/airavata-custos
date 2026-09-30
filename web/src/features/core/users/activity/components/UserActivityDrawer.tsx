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
import { ErrorState } from "@/shared/ui/ErrorState";
import { KpiCard } from "@/shared/ui/KpiCard";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
} from "@/shared/ui/drawer";
import { Skeleton } from "@/shared/ui/skeleton";
import { Activity, CalendarDays, Flame, LogIn, X } from "lucide-react";
import * as React from "react";
import { activityActionLabel, formatLastLogin } from "../lib";
import { useUserActivityAnalytics } from "../queries";
import type { UserActivityRow } from "../schemas";
import { LoginTrend } from "./ActivityOverview";
import { ActivityStatus } from "./ActivityStatus";
import { DaysRangePicker } from "./DaysRangePicker";

export function UserActivityDrawer({
  user,
  initialWindow,
  onClose,
}: { user: UserActivityRow; initialWindow: number; onClose: () => void }) {
  const [windowDays, setWindowDays] = React.useState(initialWindow);
  const analytics = useUserActivityAnalytics(windowDays, user.user_id);
  const reviewAccess = activityActionLabel(user, initialWindow) === "Review access";
  return (
    <Drawer
      open
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      direction="right"
    >
      <DrawerContent className="w-[min(94vw,760px)] overflow-y-auto sm:max-w-[760px]">
        <DrawerHeader className="border-b pr-16">
          <DrawerTitle>{reviewAccess ? "Review access" : "User engagement audit"}</DrawerTitle>
          <DrawerDescription>
            {user.name} · {user.email}
            {reviewAccess
              ? ". Sign-in activity only; this does not change roles or cluster access."
              : ""}
          </DrawerDescription>
          <DrawerClose
            aria-label="Close user activity"
            className="absolute right-4 top-4 rounded-md p-2 hover:bg-muted"
          >
            <X size={16} />
          </DrawerClose>
        </DrawerHeader>
        <div className="space-y-5 p-5">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <p className="font-display text-xl font-semibold">{user.name}</p>
            <ActivityStatus user={user} windowDays={initialWindow} />
          </div>
          <DaysRangePicker
            value={windowDays}
            onChange={setWindowDays}
            label="User analytics date range"
          />
          {analytics.isPending ? (
            <Skeleton aria-label="Loading individual activity" className="h-64 w-full" />
          ) : analytics.error ? (
            <ErrorState
              message="We couldn't load this user's activity."
              onRetry={() => analytics.refetch()}
            />
          ) : analytics.data ? (
            <>
              <div className="grid gap-3 sm:grid-cols-2">
                <KpiCard title="Last login" value={formatLastLogin(user)} icon={CalendarDays} />
                <KpiCard
                  title={`Logins · ${windowDays}d`}
                  value={analytics.data.window_login_count}
                  icon={LogIn}
                />
                <KpiCard
                  title={`Active days · ${windowDays}d`}
                  value={analytics.data.window_active_days}
                  icon={Activity}
                />
                <KpiCard
                  title="Current login streak"
                  value={`${user.current_streak}d`}
                  icon={Flame}
                />
              </div>
              <LoginTrend analytics={analytics.data} />
              <Card>
                <CardHeader>
                  <CardTitle>Activity summary</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="grid grid-cols-[1fr_auto] gap-3 text-sm tabular-nums">
                    <dt>Lifetime sign-ins</dt>
                    <dd>{analytics.data.lifetime_login_count}</dd>
                    <dt>Lifetime active days</dt>
                    <dd>{analytics.data.lifetime_active_days}</dd>
                    <dt>Dashboard activity window</dt>
                    <dd>{initialWindow} days</dd>
                  </dl>
                </CardContent>
              </Card>
            </>
          ) : null}
        </div>
      </DrawerContent>
    </Drawer>
  );
}
