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
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import type * as React from "react";
import {
  Bar,
  CartesianGrid,
  Cell,
  ComposedChart,
  Line,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { completeActivityTrend, statusMeta } from "../lib";
import type { UserActivityAnalytics } from "../schemas";

export function SummaryCard({
  title,
  value,
  percent,
  detail,
  color,
  selected,
  onClick,
}: {
  title: string;
  value: number;
  percent?: number;
  detail: React.ReactNode;
  color: string;
  selected: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={selected}
      className={cn(
        "min-h-36 rounded-lg border bg-card p-4 text-left shadow-sm transition-colors hover:bg-muted/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        selected ? "border-brand" : "border-border",
      )}
    >
      <span
        className={cn(
          "font-mono",
          "inline-flex items-center gap-2 text-xs font-normal uppercase tracking-wide text-muted-foreground",
        )}
      >
        <span
          aria-hidden="true"
          className="size-2 rounded-full"
          style={{ backgroundColor: color }}
        />
        {title}
      </span>
      <span className="mt-2 flex items-baseline gap-2">
        <span className="font-display text-3xl font-bold tabular-nums">{value}</span>
        {percent !== undefined ? (
          <span className="text-base text-muted-foreground tabular-nums">{percent}%</span>
        ) : null}
      </span>
      <span className="mt-1 block text-xs text-muted-foreground">{detail}</span>
    </button>
  );
}

export function StatusComposition({ counts }: { counts: Record<keyof typeof statusMeta, number> }) {
  const total = counts.active + counts.dormant + counts.never;
  const entries = (Object.keys(statusMeta) as (keyof typeof statusMeta)[]).map((status) => ({
    status,
    count: counts[status],
    percent: total ? (counts[status] / total) * 100 : 0,
  }));

  return (
    <Card>
      <CardHeader className="pb-3">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <CardTitle
            className={cn(
              "font-mono",
              "text-xs font-normal uppercase tracking-wide text-muted-foreground",
            )}
          >
            Status composition · {total} identities
          </CardTitle>
          <span className="text-xs text-muted-foreground">
            Mutually exclusive — every identity sits in exactly one band
          </span>
        </div>
      </CardHeader>
      <CardContent>
        <div
          className="flex h-3 overflow-hidden rounded-full"
          role="img"
          aria-label={entries
            .map(({ status, count }) => `${statusMeta[status].label} ${count}`)
            .join(", ")}
        >
          {entries.map(({ status, percent }) => (
            <span
              key={status}
              style={{ width: `${percent}%`, backgroundColor: statusMeta[status].color }}
            />
          ))}
        </div>
        <div className="mt-3 flex flex-wrap gap-x-6 gap-y-2 text-sm">
          {entries.map(({ status, count, percent }) => (
            <span key={status} className="inline-flex items-center gap-2">
              <span
                className="h-2.5 w-2.5 rounded-full"
                style={{ backgroundColor: statusMeta[status].color }}
              />
              <span>{statusMeta[status].label}</span>
              <span className="text-muted-foreground">
                {count} · {Math.round(percent)}%
              </span>
            </span>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

export function LoginTrend({ analytics }: { analytics: UserActivityAnalytics }) {
  const windowDays = analytics.window_days;
  const data = completeActivityTrend(analytics).map((point) => ({
    date: point.date,
    label: point.date.slice(5),
    loginSessions: point.login_count,
    activeUsers: point.active_users,
  }));
  const sessions = data.reduce((total, point) => total + point.loginSessions, 0);
  const peak = Math.max(0, ...data.map((point) => point.loginSessions));
  const peakPoint = peak > 0 ? data.find((point) => point.loginSessions === peak) : undefined;
  const labelInterval = Math.max(0, Math.ceil(windowDays / 7) - 1);

  return (
    <Card>
      <CardHeader className="flex flex-row flex-wrap justify-between gap-3 pb-2">
        <div>
          <CardTitle className="text-lg">Sign-in trend</CardTitle>
          <p className="text-sm text-muted-foreground">
            {sessions} sign-ins over {windowDays} days · {(sessions / windowDays).toFixed(1)} per
            day
            {peakPoint ? ` · peak ${peak} on ${peakPoint.label}` : ""}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-5 text-xs text-muted-foreground">
          <span className="inline-flex items-center gap-2">
            <span className="h-2.5 w-2.5 rounded-full bg-[color:var(--custos-blue-500)]" />
            Sign-ins per day (weekends lighter)
          </span>
          <span className="inline-flex items-center gap-2">
            <span className="h-0.5 w-5 bg-[color:var(--custos-green-500)]" />
            Distinct users
          </span>
        </div>
      </CardHeader>
      <CardContent>
        <p className="mb-2 text-xs text-muted-foreground">
          Pale baseline marks indicate 0 sign-ins.
        </p>
        <div role="img" aria-label={`Sign-in trend for the last ${windowDays} days`}>
          {sessions === 0 ? (
            <p className="flex h-40 items-center justify-center text-sm text-muted-foreground">
              No sign-ins in this period.
            </p>
          ) : (
            <ResponsiveContainer width="100%" height={280}>
              <ComposedChart data={data} margin={{ top: 8, right: 8, left: -12, bottom: 0 }}>
                <CartesianGrid stroke="var(--border)" vertical={false} />
                <XAxis
                  dataKey="date"
                  interval={labelInterval}
                  tickFormatter={(label: string) =>
                    new Intl.DateTimeFormat("en-US", {
                      month: "short",
                      day: "numeric",
                      timeZone: "UTC",
                    }).format(new Date(`${label}T00:00:00Z`))
                  }
                  tickLine={false}
                  tickMargin={8}
                  tick={{ fontSize: 11, fontFamily: "monospace" }}
                  stroke="var(--muted-foreground)"
                />
                <YAxis
                  allowDecimals={false}
                  axisLine={false}
                  tickLine={false}
                  tick={{ fontSize: 11, fontFamily: "monospace" }}
                  stroke="var(--muted-foreground)"
                />
                <Tooltip labelFormatter={(_, payload) => payload[0]?.payload.date ?? ""} />
                <Bar
                  isAnimationActive={false}
                  dataKey="loginSessions"
                  name="Sign-ins per day"
                  fill="var(--custos-blue-500)"
                  minPointSize={3}
                  radius={[3, 3, 0, 0]}
                >
                  {data.map((point) => {
                    const day = new Date(`${point.date}T00:00:00Z`).getUTCDay();
                    return (
                      <Cell
                        key={point.date}
                        fillOpacity={
                          point.loginSessions === 0 ? 0.2 : day === 0 || day === 6 ? 0.5 : 1
                        }
                      />
                    );
                  })}
                </Bar>
                <Line
                  isAnimationActive={false}
                  type="linear"
                  dataKey="activeUsers"
                  name="Distinct users"
                  stroke="var(--custos-green-500)"
                  strokeWidth={2}
                  dot={false}
                />
              </ComposedChart>
            </ResponsiveContainer>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
