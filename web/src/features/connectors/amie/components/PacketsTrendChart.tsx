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
import { StackedAreaUsage } from "@/shared/charts/StackedAreaUsage";
import type { PacketStatBucket, PacketStatus } from "../types";

const STATUS_ORDER: PacketStatus[] = ["PROCESSED", "DECODED", "NEW", "FAILED"];
// Recharts paints area fill AND tooltip label in the same color on white.
// Use 700-step so tooltip labels clear WCAG AA 4.5:1.
const STATUS_COLORS: Record<PacketStatus, string> = {
  PROCESSED: "var(--custos-green-700)",
  DECODED: "var(--custos-amber-700)",
  NEW: "var(--custos-blue-700)",
  FAILED: "var(--custos-red-700)",
};

export type PacketsTrendChartProps = {
  buckets: PacketStatBucket[];
  height?: number;
};

export function PacketsTrendChart({ buckets, height = 220 }: PacketsTrendChartProps) {
  const byDay = React.useMemo(() => {
    const days = new Map<string, Record<PacketStatus, number> & { date: string }>();
    for (const b of buckets) {
      const row =
        days.get(b.date) ??
        ({ date: b.date, NEW: 0, DECODED: 0, PROCESSED: 0, FAILED: 0 } as Record<
          PacketStatus,
          number
        > & { date: string });
      row[b.status] = (row[b.status] ?? 0) + b.count;
      days.set(b.date, row);
    }
    return Array.from(days.values()).sort((a, b) => a.date.localeCompare(b.date));
  }, [buckets]);

  if (byDay.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">No packet activity in the selected window.</p>
    );
  }

  return (
    <div className="rounded-md border bg-card p-4">
      <header className="mb-3 flex items-baseline justify-between">
        <h2 className="font-heading text-sm font-semibold">Packets per day</h2>
        <p className="text-xs text-muted-foreground">last {byDay.length} days · by status</p>
      </header>
      <StackedAreaUsage
        data={byDay}
        seriesKeys={STATUS_ORDER}
        colors={STATUS_ORDER.map((s) => STATUS_COLORS[s])}
        height={height}
        ariaLabel="AMIE packets per day grouped by status"
      />
      <ul className="mt-2 flex flex-wrap gap-3 text-xs text-muted-foreground">
        {STATUS_ORDER.map((s) => (
          <li key={s} className="flex items-center gap-1">
            <span
              aria-hidden="true"
              className="inline-block size-2 rounded-full"
              style={{ background: STATUS_COLORS[s] }}
            />
            {s}
          </li>
        ))}
      </ul>
    </div>
  );
}
