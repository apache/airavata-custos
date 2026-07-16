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

import { describe, expect, it } from "vitest";
import {
  buildPeriods,
  buildResourceSeries,
  creditsBand,
  daysBand,
  daysUntil,
  formatCredits,
  pctRemaining,
} from "../lib";
import type { UsageDailyBucket, UsageResource } from "../schemas";

const NOW = new Date("2026-07-14T12:00:00Z");

function daysFromNow(n: number): string {
  return new Date(NOW.getTime() + n * 86_400_000).toISOString();
}

function resource(id: string, used: number, name = id): UsageResource {
  return {
    resource_id: id,
    name,
    resource_type: "GPU_HOURS",
    used,
    cap: null,
    used_native: used,
    native_unit: "GPU-hours",
    used_by_caller: 0,
  };
}

// Build a daily series ending today: `perDay` credits on a single resource for
// the last `activeDays`, preceded by `zeroDays` empty days.
function series(zeroDays: number, activeDays: number, perDay: number): UsageDailyBucket[] {
  const out: UsageDailyBucket[] = [];
  const total = zeroDays + activeDays;
  for (let i = 0; i < total; i++) {
    const date = new Date(NOW.getTime() - (total - 1 - i) * 86_400_000).toISOString().slice(0, 10);
    const active = i >= zeroDays;
    out.push({ date, by_resource: active ? { r1: perDay } : {} });
  }
  return out;
}

describe("pctRemaining", () => {
  it("computes remaining percentage", () => {
    expect(pctRemaining(1000, 250)).toBe(75);
  });
  it("returns 0 for non-positive total", () => {
    expect(pctRemaining(0, 0)).toBe(0);
  });
  it("clamps to [0, 100]", () => {
    expect(pctRemaining(1000, 1200)).toBe(0);
    expect(pctRemaining(1000, -100)).toBe(100);
  });
});

describe("daysUntil", () => {
  it("counts whole days ahead", () => {
    expect(daysUntil(daysFromNow(18), NOW)).toBe(18);
  });
  it("floors past dates at zero", () => {
    expect(daysUntil(daysFromNow(-3), NOW)).toBe(0);
  });
});

describe("daysBand", () => {
  it("red at or below 7 days", () => {
    expect(daysBand(7)).toBe("red");
    expect(daysBand(0)).toBe("red");
  });
  it("amber between 8 and 30", () => {
    expect(daysBand(8)).toBe("amber");
    expect(daysBand(30)).toBe("amber");
  });
  it("ok above 30", () => {
    expect(daysBand(31)).toBe("ok");
  });
});

describe("creditsBand", () => {
  it("red at or below 5% left", () => {
    expect(creditsBand(5)).toBe("red");
  });
  it("amber between 6% and 20%", () => {
    expect(creditsBand(6)).toBe("amber");
    expect(creditsBand(20)).toBe("amber");
  });
  it("ok above 20%", () => {
    expect(creditsBand(21)).toBe("ok");
  });
});

describe("buildPeriods", () => {
  it("day mode keeps the last 30 buckets and marks the trailing one partial", () => {
    const daily = series(0, 40, 100); // 40 days
    const periods = buildPeriods(daily, "day");
    expect(periods).toHaveLength(30);
    expect(periods[periods.length - 1]?.isPartial).toBe(true);
    expect(periods[0]?.isPartial).toBe(false);
  });

  it("week mode aggregates into 7-day chunks with the current week partial", () => {
    const daily = series(0, 21, 100); // 3 weeks of 700/week
    const periods = buildPeriods(daily, "week");
    expect(periods).toHaveLength(3);
    expect(periods[0]?.total).toBe(700);
    expect(periods[periods.length - 1]?.isPartial).toBe(true);
  });
});

describe("buildResourceSeries", () => {
  it("keeps the four largest as named series in descending order", () => {
    const s = buildResourceSeries([
      resource("a", 10),
      resource("b", 50),
      resource("c", 30),
      resource("d", 20),
      resource("e", 5),
    ]);
    expect(s.map((x) => x.id)).toEqual(["b", "c", "d", "a"]);
    expect(s).toHaveLength(4);
    expect(s[0]?.color).toBe("var(--chart-1)");
  });
});

describe("formatCredits", () => {
  it("uses compact notation", () => {
    expect(formatCredits(1_220_000)).toBe("1.22M");
    expect(formatCredits(84_200)).toBe("84.2K");
    expect(formatCredits(72_000)).toBe("72K");
  });
});
