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
  DEFAULT_FILTERS,
  type ListFilters,
  bannerBounds,
  hasActiveFilters,
  parseFilters,
  serializeFilters,
  statusFiltersToApi,
  windowToFromTo,
} from "@/features/core/audit/components/traceListUrlState";

function p(qs: string): URLSearchParams {
  return new URLSearchParams(qs);
}

describe("traceListUrlState — parseFilters", () => {
  it("returns defaults when params are empty", () => {
    expect(parseFilters(p(""))).toEqual(DEFAULT_FILTERS);
  });

  it("drops bogus status / source / window values", () => {
    const params = p("status=bogus&source=evil&window=42y&pageSize=7");
    const f = parseFilters(params);
    expect(f.status).toEqual([]);
    expect(f.source).toEqual([]);
    expect(f.window).toBe("30d");
    expect(f.pageSize).toBe(50);
  });

  it("falls back page < 1 to 1, page invalid to 1", () => {
    expect(parseFilters(p("page=0")).page).toBe(1);
    expect(parseFilters(p("page=-3")).page).toBe(1);
    expect(parseFilters(p("page=abc")).page).toBe(1);
    expect(parseFilters(p("page=4")).page).toBe(4);
  });

  it("accepts only 25/50/100 for pageSize", () => {
    expect(parseFilters(p("pageSize=25")).pageSize).toBe(25);
    expect(parseFilters(p("pageSize=100")).pageSize).toBe(100);
    expect(parseFilters(p("pageSize=37")).pageSize).toBe(50);
  });

  it("preserves the order of repeated source values", () => {
    const f = parseFilters(p("source=amie&source=slurm"));
    expect(f.source).toEqual(["amie", "slurm"]);
  });

  it("trims the q query", () => {
    expect(parseFilters(p("q=%20alice%20")).q).toBe("alice");
  });

  it("defaults failingOver24h to false; accepts =1", () => {
    expect(parseFilters(p("")).failingOver24h).toBe(false);
    expect(parseFilters(p("failingOver24h=1")).failingOver24h).toBe(true);
    expect(parseFilters(p("failingOver24h=0")).failingOver24h).toBe(false);
  });
});

describe("traceListUrlState — serializeFilters", () => {
  it("omits all defaults", () => {
    expect(serializeFilters(DEFAULT_FILTERS).toString()).toBe("");
  });

  it("emits status entries when changed", () => {
    const filters: ListFilters = { ...DEFAULT_FILTERS, status: ["ok", "error"] };
    const qs = serializeFilters(filters).toString();
    expect(qs).toContain("status=ok");
    expect(qs).toContain("status=error");
  });

  it("does not emit status when it equals the default", () => {
    const qs = serializeFilters({ ...DEFAULT_FILTERS, status: ["error"] }).toString();
    expect(qs).toBe("");
  });

  it("emits window/page/pageSize only when off-default", () => {
    const filters: ListFilters = {
      ...DEFAULT_FILTERS,
      window: "24h",
      page: 3,
      pageSize: 100,
    };
    const qs = serializeFilters(filters);
    expect(qs.get("window")).toBe("24h");
    expect(qs.get("page")).toBe("3");
    expect(qs.get("pageSize")).toBe("100");
  });

  it("round-trips failingOver24h=1", () => {
    const filters: ListFilters = { ...DEFAULT_FILTERS, failingOver24h: true };
    const qs = serializeFilters(filters);
    expect(qs.get("failingOver24h")).toBe("1");
    expect(parseFilters(qs).failingOver24h).toBe(true);
  });

  it("omits failingOver24h when false", () => {
    expect(serializeFilters(DEFAULT_FILTERS).has("failingOver24h")).toBe(false);
  });
});

describe("traceListUrlState — round-trip", () => {
  it("parse(serialize(filters)) === filters for typical edits", () => {
    const f: ListFilters = {
      status: ["error", "ok"],
      source: ["amie", "comanage"],
      window: "7d",
      q: "alice",
      page: 2,
      pageSize: 25,
      failingOver24h: false,
    };
    const out = parseFilters(serializeFilters(f));
    expect(new Set(out.status)).toEqual(new Set(f.status));
    expect(new Set(out.source)).toEqual(new Set(f.source));
    expect(out.window).toBe(f.window);
    expect(out.q).toBe(f.q);
    expect(out.page).toBe(f.page);
    expect(out.pageSize).toBe(f.pageSize);
  });
});

describe("traceListUrlState — hasActiveFilters", () => {
  it("is false for defaults", () => {
    expect(hasActiveFilters(DEFAULT_FILTERS)).toBe(false);
  });

  it("is true when status differs, source set, window changed, or q present", () => {
    expect(hasActiveFilters({ ...DEFAULT_FILTERS, status: ["ok"] })).toBe(true);
    expect(hasActiveFilters({ ...DEFAULT_FILTERS, source: ["amie"] })).toBe(true);
    expect(hasActiveFilters({ ...DEFAULT_FILTERS, window: "24h" })).toBe(true);
    expect(hasActiveFilters({ ...DEFAULT_FILTERS, q: "x" })).toBe(true);
  });

  it("is true when failingOver24h is on", () => {
    expect(hasActiveFilters({ ...DEFAULT_FILTERS, failingOver24h: true })).toBe(true);
  });
});

describe("traceListUrlState — windowToFromTo", () => {
  it("anchors to/from at now and now − N days", () => {
    const now = Date.parse("2026-06-05T12:00:00.000Z");
    const r24 = windowToFromTo("24h", now);
    expect(r24.to).toBe("2026-06-05T12:00:00.000Z");
    expect(r24.from).toBe("2026-06-04T12:00:00.000Z");
    const r7 = windowToFromTo("7d", now);
    expect(r7.from).toBe("2026-05-29T12:00:00.000Z");
    const r30 = windowToFromTo("30d", now);
    expect(r30.from).toBe("2026-05-06T12:00:00.000Z");
  });
});

describe("traceListUrlState — bannerBounds", () => {
  it("anchors from at 30d ago and to at 24h ago", () => {
    const now = Date.parse("2026-06-05T12:00:00.000Z");
    const { from, to } = bannerBounds(now);
    expect(to).toBe("2026-06-04T12:00:00.000Z");
    expect(from).toBe("2026-05-06T12:00:00.000Z");
  });
});

describe("traceListUrlState — statusFiltersToApi", () => {
  it("maps ok→0, error→1, orphaned→3", () => {
    const { apiStatus, inProgressOnly } = statusFiltersToApi(["ok", "error", "orphaned"]);
    expect(new Set(apiStatus)).toEqual(new Set([0, 1, 3]));
    expect(inProgressOnly).toBe(false);
  });

  it("strips in-progress from the wire and flags it when sole filter", () => {
    const sole = statusFiltersToApi(["in-progress"]);
    expect(sole.apiStatus).toEqual([]);
    expect(sole.inProgressOnly).toBe(true);

    const mixed = statusFiltersToApi(["in-progress", "error"]);
    expect(mixed.apiStatus).toEqual([1]);
    expect(mixed.inProgressOnly).toBe(false);
  });
});
