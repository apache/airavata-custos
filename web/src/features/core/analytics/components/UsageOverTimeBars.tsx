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

import * as React from "react";
import { cn } from "@/lib/utils";
import {
  buildPeriods,
  buildResourceSeries,
  CHART_OTHER_COLOR,
  formatCredits,
  formatCreditsFull,
  formatDate,
  type ResourceSeries,
} from "../lib";
import type { UsageSummary } from "../schemas";

const PLOT_HEIGHT = 160;

export function UsageOverTimeBars({ summary }: { summary: UsageSummary }) {
  const [mode, setMode] = React.useState<"day" | "week">("day");
  const series = React.useMemo(() => buildResourceSeries(summary.by_resource), [summary]);
  const periods = React.useMemo(() => buildPeriods(summary.daily, mode), [summary, mode]);
  const maxTotal = Math.max(1, ...periods.map((p) => p.total));

  const hasUsage = periods.some((p) => p.total > 0);

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
        <div>
          <h2 className="text-sm font-semibold">Usage over time</h2>
          <p className="text-xs text-muted-foreground">
            {mode === "day" ? "last 30 days" : "last 12 weeks"} · faded bar is the current period
          </p>
        </div>
        <Toggle mode={mode} onChange={setMode} />
      </div>

      <Legend series={series} hasOther={hasOther(periods, series)} />

      {hasUsage ? (
        <div className="mt-3 flex items-end gap-[3px]" style={{ height: PLOT_HEIGHT }}>
          {periods.map((p) => (
            <Bar
              key={p.label}
              period={p}
              series={series}
              maxTotal={maxTotal}
              plotHeight={PLOT_HEIGHT}
            />
          ))}
        </div>
      ) : (
        <p className="mt-6 text-sm text-muted-foreground">No usage recorded in this window yet.</p>
      )}
    </div>
  );
}

function otherValue(
  period: { total: number; bySeriesId: Record<string, number> },
  series: ResourceSeries[],
): number {
  const named = series.reduce((a, s) => a + (period.bySeriesId[s.id] ?? 0), 0);
  return Math.max(0, period.total - named);
}

function hasOther(periods: ReturnType<typeof buildPeriods>, series: ResourceSeries[]): boolean {
  return periods.some((p) => otherValue(p, series) > 0);
}

function Bar({
  period,
  series,
  maxTotal,
  plotHeight,
}: {
  period: ReturnType<typeof buildPeriods>[number];
  series: ResourceSeries[];
  maxTotal: number;
  plotHeight: number;
}) {
  const other = otherValue(period, series);
  // Segments bottom-to-top: named series then Other on top.
  const segments = [
    ...series.map((s) => ({
      id: s.id,
      name: s.name,
      color: s.color,
      value: period.bySeriesId[s.id] ?? 0,
    })),
    ...(other > 0
      ? [{ id: "__other__", name: "Other", color: CHART_OTHER_COLOR, value: other }]
      : []),
  ].filter((seg) => seg.value > 0);

  return (
    <div
      className={cn(
        "group relative flex flex-1 flex-col-reverse justify-start",
        period.isPartial && "opacity-60",
      )}
      style={{ height: plotHeight }}
      role="img"
      aria-label={`${formatDate(period.label)}${period.isPartial ? " so far" : ""}: ${formatCreditsFull(period.total)} credits`}
    >
      {segments.map((seg, i) => (
        <div
          key={seg.id}
          className={cn(
            i === segments.length - 1 && "rounded-t-[3px]",
            period.isPartial &&
              i === segments.length - 1 &&
              "border-t border-dashed border-foreground",
          )}
          style={{ height: `${(seg.value / maxTotal) * plotHeight}px`, background: seg.color }}
        />
      ))}
      <BarTooltip period={period} segments={segments} />
    </div>
  );
}

function BarTooltip({
  period,
  segments,
}: {
  period: ReturnType<typeof buildPeriods>[number];
  segments: Array<{ id: string; name: string; color: string; value: number }>;
}) {
  return (
    <div className="pointer-events-none absolute bottom-full left-1/2 z-10 mb-1 hidden w-44 -translate-x-1/2 rounded-md border border-border bg-popover p-2 text-xs shadow-md group-hover:block">
      <div className="mb-1 font-medium">
        {formatDate(period.label)}
        {period.isPartial ? " (so far)" : ""}
      </div>
      <ul className="space-y-0.5">
        {segments.map((seg) => (
          <li key={seg.id} className="flex items-center justify-between gap-2">
            <span className="flex items-center gap-1.5">
              <span
                className="h-2 w-2 rounded-sm"
                style={{ background: seg.color }}
                aria-hidden="true"
              />
              {seg.name}
            </span>
            <span className="tabular-nums">{formatCredits(seg.value)}</span>
          </li>
        ))}
      </ul>
      <div className="mt-1 flex items-center justify-between border-t border-border pt-1 font-medium">
        <span>Total</span>
        <span className="tabular-nums">{formatCreditsFull(period.total)}</span>
      </div>
    </div>
  );
}

function Legend({ series, hasOther }: { series: ResourceSeries[]; hasOther: boolean }) {
  const items = [
    ...series.map((s) => ({ name: s.name, color: s.color })),
    ...(hasOther ? [{ name: "Other", color: CHART_OTHER_COLOR }] : []),
  ];
  if (items.length < 2) return null;
  return (
    <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
      {items.map((it) => (
        <span key={it.name} className="flex items-center gap-1.5">
          <span
            className="h-2.5 w-2.5 rounded-sm"
            style={{ background: it.color }}
            aria-hidden="true"
          />
          {it.name}
        </span>
      ))}
    </div>
  );
}

function Toggle({
  mode,
  onChange,
}: { mode: "day" | "week"; onChange: (m: "day" | "week") => void }) {
  return (
    <div className="inline-flex rounded-md border border-border p-0.5 text-xs">
      {(["day", "week"] as const).map((m) => (
        <button
          key={m}
          type="button"
          onClick={() => onChange(m)}
          className={cn(
            "rounded px-2.5 py-1 font-medium capitalize",
            mode === m ? "bg-accent text-accent-foreground" : "text-muted-foreground",
          )}
        >
          {m}
        </button>
      ))}
    </div>
  );
}
