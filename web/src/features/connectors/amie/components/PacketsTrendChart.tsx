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
import type { PacketStatBucketResponse } from "@/generated/amie/types.gen";
import { StackedAreaUsage } from "@/shared/charts/StackedAreaUsage";
import { packetStatusLabel } from "../utils";

// Recharts paints area fill AND tooltip label in the same color on white.
// Use 700-step so tooltip labels clear WCAG AA 4.5:1.
const STATUS_COLORS: Record<PacketStatBucketResponse["status"], string> = {
  PROCESSED: "var(--custos-green-700)",
  DECODED: "var(--custos-amber-700)",
  NEW: "var(--custos-blue-700)",
  WAITING_APPROVAL: "var(--custos-purple-700)",
  REFUSED: "var(--custos-gray-500)",
  FAILED: "var(--custos-red-700)",
};
// Stacking order, bottom to top.
const STATUS_ORDER = Object.keys(STATUS_COLORS);

// Matches the inbox's 30d stats window.
const DAYS = 30;

type DayRow = { date: string } & Record<string, string | number>;

// One row per UTC day of the window ending today, zero-filled; buckets outside it are dropped.
export function fillDays(buckets: PacketStatBucketResponse[], now = Date.now()) {
  const rows = new Map<string, DayRow>();
  for (let i = DAYS - 1; i >= 0; i--) {
    const date = new Date(now - i * 86_400_000).toISOString().slice(0, 10);
    rows.set(date, { date, ...Object.fromEntries(STATUS_ORDER.map((s) => [s, 0])) });
  }
  for (const { date, status, count } of buckets) {
    const row = rows.get(date);
    if (row) row[status] = Number(row[status]) + count;
  }
  return Array.from(rows.values());
}

export function PacketsTrendChart({ buckets }: { buckets: PacketStatBucketResponse[] }) {
  const byDay = React.useMemo(() => fillDays(buckets), [buckets]);

  if (buckets.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">No packet activity in the selected window.</p>
    );
  }

  return (
    <div className="rounded-md border bg-card p-4">
      <header className="mb-3 flex items-baseline justify-between">
        <h2 className="font-heading text-sm font-semibold">Packets per day</h2>
        <p className="text-xs text-muted-foreground">last {DAYS} days · by status</p>
      </header>
      <StackedAreaUsage
        data={byDay}
        seriesKeys={STATUS_ORDER}
        colors={Object.values(STATUS_COLORS)}
        height={220}
        ariaLabel="AMIE packets per day grouped by status"
      />
      <ul className="mt-2 flex flex-wrap gap-3 text-xs text-muted-foreground">
        {Object.entries(STATUS_COLORS).map(([s, color]) => (
          <li key={s} className="flex items-center gap-1">
            <span
              aria-hidden="true"
              className="inline-block size-2 rounded-full"
              style={{ background: color }}
            />
            {packetStatusLabel(s)}
          </li>
        ))}
      </ul>
    </div>
  );
}
