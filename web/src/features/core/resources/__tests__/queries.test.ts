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
import { effectiveRate, rateStatus } from "../queries";

describe("effectiveRate", () => {
  const rate = (id: string, start: string, end: string) => ({
    id,
    compute_allocation_resource_id: "res-1",
    rate: 1,
    start_time: start,
    end_time: end,
  });
  const now = Date.parse("2026-06-01T00:00:00Z");

  it("picks the in-force rate with the latest start, or null when none applies", () => {
    const rates = [
      rate("old", "2026-01-01T00:00:00Z", "2027-01-01T00:00:00Z"),
      rate("new", "2026-05-01T00:00:00Z", "2026-12-01T00:00:00Z"),
      rate("future", "2026-07-01T00:00:00Z", "2027-07-01T00:00:00Z"),
    ];
    expect(effectiveRate(rates, now)?.id).toBe("new");
    expect(effectiveRate([rates[2] as (typeof rates)[number]], now)).toBeNull();
  });

  it("breaks a tied start as the server does: earliest end, then lowest id", () => {
    const start = "2026-05-01T00:00:00Z";
    const tied = [
      rate("b", start, "2026-12-01T00:00:00Z"),
      rate("c", start, "2026-09-01T00:00:00Z"),
      rate("a", start, "2026-09-01T00:00:00Z"),
    ];
    expect(effectiveRate(tied, now)?.id).toBe("a");
  });
});

describe("rateStatus", () => {
  const rate = (id: string, start: string, end: string) => ({
    id,
    compute_allocation_resource_id: "res-1",
    rate: 1,
    start_time: start,
    end_time: end,
  });

  it("dates an override by the first boundary the rate wins at, through stacked rates", () => {
    const long = rate("long", "2026-01-01T00:00:00Z", "2027-01-01T00:00:00Z");
    const mid = rate("mid", "2026-03-01T00:00:00Z", "2026-12-01T00:00:00Z");
    const short = rate("short", "2026-05-01T00:00:00Z", "2026-07-01T00:00:00Z");
    const now = Date.parse("2026-06-01T00:00:00Z");
    expect(rateStatus(long, [long, mid, short], now)).toEqual({
      kind: "OVERRIDDEN",
      until: mid.end_time,
    });
    expect(rateStatus(mid, [long, mid, short], now)).toEqual({
      kind: "OVERRIDDEN",
      until: short.end_time,
    });
    const tail = rate("tail", "2026-06-15T00:00:00Z", "2027-06-01T00:00:00Z");
    expect(rateStatus(long, [long, tail], Date.parse("2026-07-01T00:00:00Z"))).toEqual({
      kind: "OVERRIDDEN",
      until: undefined,
    });
  });

  it("schedules a future rate from the first boundary it wins at, or never when there is none", () => {
    const now = Date.parse("2026-06-01T00:00:00Z");
    const never = rate("never", "2027-01-01T00:00:00Z", "2027-12-01T00:00:00Z");
    const shorter = rate("shorter", "2027-01-01T00:00:00Z", "2027-03-01T00:00:00Z");
    const later = rate("later", "2027-03-01T00:00:00Z", "2028-01-01T00:00:00Z");
    expect(rateStatus(never, [never, shorter, later], now)).toEqual({ kind: "NEVER" });
    expect(rateStatus(never, [never, shorter], now)).toEqual({
      kind: "SCHEDULED",
      from: shorter.end_time,
    });
  });
});
