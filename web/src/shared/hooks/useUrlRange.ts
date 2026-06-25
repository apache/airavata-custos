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

import { useRouter, useSearchParams } from "next/navigation";
import * as React from "react";

const DAY_MS = 24 * 60 * 60 * 1000;

export const RANGE_PRESETS = ["24h", "7d", "30d", "90d", "custom"] as const;
export type RangePreset = (typeof RANGE_PRESETS)[number];

export type AnalyticsRange = {
  from: Date;
  to: Date;
  preset: RangePreset;
};

const PRESET_DAYS: Record<Exclude<RangePreset, "custom">, number> = {
  "24h": 1,
  "7d": 7,
  "30d": 30,
  "90d": 90,
};

function isRangePreset(value: string | null): value is RangePreset {
  return value !== null && (RANGE_PRESETS as readonly string[]).includes(value);
}

function parseIso(input: string | null): Date | null {
  if (!input) return null;
  const ms = Date.parse(input);
  if (Number.isNaN(ms)) return null;
  return new Date(ms);
}

export function rangeFromPreset(preset: Exclude<RangePreset, "custom">, now = Date.now()): AnalyticsRange {
  return { from: new Date(now - PRESET_DAYS[preset] * DAY_MS), to: new Date(now), preset };
}

export function defaultRange(now = Date.now()): AnalyticsRange {
  return rangeFromPreset("30d", now);
}

export type UseUrlRangeOptions = {
  paramPrefix?: string;
  now?: () => number;
};

// `?from=&to=&preset=` <-> { from, to, preset }. Default = last 30 days.
// Custom ranges require both `from` and `to`; missing/bad values fall back to default.
export function useUrlRange(options: UseUrlRangeOptions = {}) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const nowFn = options.now ?? Date.now;

  const range = React.useMemo<AnalyticsRange>(() => {
    const preset = searchParams.get("preset");
    if (preset === "custom") {
      const from = parseIso(searchParams.get("from"));
      const to = parseIso(searchParams.get("to"));
      if (from && to && from.getTime() <= to.getTime()) {
        return { from, to, preset: "custom" };
      }
      return defaultRange(nowFn());
    }
    if (isRangePreset(preset) && preset !== "custom") {
      return rangeFromPreset(preset, nowFn());
    }
    return defaultRange(nowFn());
  }, [searchParams, nowFn]);

  const setRange = React.useCallback(
    (next: AnalyticsRange) => {
      const params = new URLSearchParams(searchParams.toString());
      params.set("preset", next.preset);
      if (next.preset === "custom") {
        params.set("from", next.from.toISOString());
        params.set("to", next.to.toISOString());
      } else {
        params.delete("from");
        params.delete("to");
      }
      const query = params.toString();
      router.replace(query ? `?${query}` : "?", { scroll: false });
    },
    [router, searchParams],
  );

  return { range, setRange };
}
