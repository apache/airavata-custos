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

import type { UsageDailyBucket, UsageResource } from "./schemas";

// Categorical series colors; chart-5 is the neutral used for "Other".
export const CHART_SERIES_COLORS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
];
export const CHART_OTHER_COLOR = "var(--chart-5)";

export type ResourceSeries = { id: string; name: string; color: string; nativeUnit: string };

// The (up to four) largest resources as named series. Anything beyond is folded
// into a neutral "Other" segment by the widgets, keeping stacks to four colors.
export function buildResourceSeries(byResource: UsageResource[]): ResourceSeries[] {
  return [...byResource]
    .sort((a, b) => b.used - a.used)
    .slice(0, CHART_SERIES_COLORS.length)
    .map((r, i) => ({
      id: r.resource_id,
      name: r.name,
      color: CHART_SERIES_COLORS[i] ?? CHART_OTHER_COLOR,
      nativeUnit: r.native_unit,
    }));
}

// Urgency thresholds. Amber warns, red alarms; both live only in the tile pills.
export const AMBER_DAYS = 30;
export const AMBER_PCT_LEFT = 20;
export const RED_DAYS = 7;
export const RED_PCT_LEFT = 5;

export type UrgencyBand = "ok" | "amber" | "red";

const DAY_MS = 24 * 60 * 60 * 1000;

export function pctRemaining(total: number, used: number): number {
  if (total <= 0) return 0;
  const left = ((total - used) / total) * 100;
  return Math.max(0, Math.min(100, left));
}

// Whole days from now until the given time, floored at zero.
export function daysUntil(iso: string, now: Date): number {
  const end = new Date(iso).getTime();
  const diff = end - now.getTime();
  if (diff <= 0) return 0;
  return Math.ceil(diff / DAY_MS);
}

// Band for a days-remaining tile (allocation/project end).
export function daysBand(daysLeft: number): UrgencyBand {
  if (daysLeft <= RED_DAYS) return "red";
  if (daysLeft <= AMBER_DAYS) return "amber";
  return "ok";
}

// Band for the credits-left tile, from the raw percentage remaining.
export function creditsBand(pctLeft: number): UrgencyBand {
  if (pctLeft <= RED_PCT_LEFT) return "red";
  if (pctLeft <= AMBER_PCT_LEFT) return "amber";
  return "ok";
}

export function bucketTotal(bucket: UsageDailyBucket): number {
  return Object.values(bucket.by_resource).reduce((a, v) => a + v, 0);
}

export const DAY_PERIODS = 30;
export const WEEK_PERIODS = 12;

export type UsagePeriod = {
  label: string;
  bySeriesId: Record<string, number>;
  total: number;
  // The trailing period is still in progress; the chart renders it faded.
  isPartial: boolean;
};

// Buckets the daily series into the last 30 days or 12 weeks. Week buckets are
// 7-day chunks aligned to the most recent day, so the last chunk is the current
// (partial) week.
export function buildPeriods(daily: UsageDailyBucket[], mode: "day" | "week"): UsagePeriod[] {
  if (mode === "day") {
    const days = daily.slice(-DAY_PERIODS);
    return days.map((b, i) => ({
      label: b.date,
      bySeriesId: { ...b.by_resource },
      total: bucketTotal(b),
      isPartial: i === days.length - 1,
    }));
  }

  const weeks: UsageDailyBucket[][] = [];
  for (let i = daily.length; i > 0; i -= 7) {
    weeks.unshift(daily.slice(Math.max(0, i - 7), i));
  }
  const recent = weeks.slice(-WEEK_PERIODS);
  return recent.map((chunk, i) => {
    const bySeriesId: Record<string, number> = {};
    for (const b of chunk) {
      for (const [id, v] of Object.entries(b.by_resource)) {
        bySeriesId[id] = (bySeriesId[id] ?? 0) + v;
      }
    }
    const total = Object.values(bySeriesId).reduce((a, v) => a + v, 0);
    return { label: chunk[0]?.date ?? "", bySeriesId, total, isPartial: i === recent.length - 1 };
  });
}

// Compact credits for tiles, axes, and bars ("1.22M", "84.2K", "72K").
export function formatCredits(n: number): string {
  return new Intl.NumberFormat(undefined, {
    notation: "compact",
    maximumFractionDigits: 2,
  }).format(n);
}

// Full grouped credits for tooltips and sublines.
export function formatCreditsFull(n: number): string {
  return new Intl.NumberFormat(undefined, { maximumFractionDigits: 0 }).format(n);
}

export function formatNative(n: number): string {
  return new Intl.NumberFormat(undefined, { maximumFractionDigits: 0 }).format(n);
}

// "Aug 1, 2026"
export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

// "Jul 14, 09:12"
export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}
