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

import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useUrlRange } from "../useUrlRange";

const replace = vi.fn();
let searchString = "";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace }),
  useSearchParams: () => new URLSearchParams(searchString),
}));

const FIXED_NOW = Date.parse("2026-05-01T00:00:00Z");
const now = () => FIXED_NOW;
const DAY_MS = 24 * 60 * 60 * 1000;

beforeEach(() => {
  replace.mockReset();
  searchString = "";
});

describe("useUrlRange", () => {
  it("defaults to last 30 days when no params are present", () => {
    const { result } = renderHook(() => useUrlRange({ now }));
    expect(result.current.range.preset).toBe("30d");
    expect(result.current.range.to.getTime()).toBe(FIXED_NOW);
    expect(result.current.range.from.getTime()).toBe(FIXED_NOW - 30 * DAY_MS);
  });

  it("parses a known preset from the URL", () => {
    searchString = "preset=7d";
    const { result } = renderHook(() => useUrlRange({ now }));
    expect(result.current.range.preset).toBe("7d");
    expect(result.current.range.from.getTime()).toBe(FIXED_NOW - 7 * DAY_MS);
  });

  it("parses a custom from/to when both ISO timestamps validate", () => {
    const from = new Date(FIXED_NOW - 10 * DAY_MS).toISOString();
    const to = new Date(FIXED_NOW - 1 * DAY_MS).toISOString();
    searchString = `preset=custom&from=${from}&to=${to}`;
    const { result } = renderHook(() => useUrlRange({ now }));
    expect(result.current.range.preset).toBe("custom");
    expect(result.current.range.from.toISOString()).toBe(from);
    expect(result.current.range.to.toISOString()).toBe(to);
  });

  it("falls back to default when custom range is missing endpoints", () => {
    searchString = "preset=custom&from=invalid";
    const { result } = renderHook(() => useUrlRange({ now }));
    expect(result.current.range.preset).toBe("30d");
  });

  it("setRange to preset clears from/to and routes via replace", () => {
    const { result } = renderHook(() => useUrlRange({ now }));
    act(() => {
      result.current.setRange({
        from: new Date(FIXED_NOW - 7 * DAY_MS),
        to: new Date(FIXED_NOW),
        preset: "7d",
      });
    });
    expect(replace).toHaveBeenCalledTimes(1);
    const args = replace.mock.calls[0] as [string, { scroll: boolean }];
    expect(args[0]).toContain("preset=7d");
    expect(args[0]).not.toContain("from=");
    expect(args[1]).toEqual({ scroll: false });
  });

  it("setRange to custom writes from and to ISO strings", () => {
    const { result } = renderHook(() => useUrlRange({ now }));
    const from = new Date(FIXED_NOW - 5 * DAY_MS);
    const to = new Date(FIXED_NOW);
    act(() => {
      result.current.setRange({ from, to, preset: "custom" });
    });
    const url = replace.mock.calls[0]?.[0] as string;
    expect(url).toContain("preset=custom");
    expect(url).toContain(`from=${encodeURIComponent(from.toISOString())}`);
    expect(url).toContain(`to=${encodeURIComponent(to.toISOString())}`);
  });

  it("round-trips a preset chip through ?preset= (write then read)", () => {
    // Write — setRange with a preset should drop from/to and only carry ?preset=.
    const { result, rerender } = renderHook(() => useUrlRange({ now }));
    act(() => {
      result.current.setRange({
        from: new Date(FIXED_NOW - 7 * DAY_MS),
        to: new Date(FIXED_NOW),
        preset: "7d",
      });
    });
    const writtenUrl = replace.mock.calls[0]?.[0] as string;
    expect(writtenUrl).toBe("?preset=7d");

    // Read — feeding that URL back through the hook resolves to the same preset.
    searchString = "preset=7d";
    rerender();
    expect(result.current.range.preset).toBe("7d");
    expect(result.current.range.to.getTime()).toBe(FIXED_NOW);
    expect(result.current.range.from.getTime()).toBe(FIXED_NOW - 7 * DAY_MS);
  });

  it("round-trips a custom window through ?preset=custom&from=&to= (write then read)", () => {
    const from = new Date(FIXED_NOW - 12 * DAY_MS);
    const to = new Date(FIXED_NOW - 2 * DAY_MS);
    const { result, rerender } = renderHook(() => useUrlRange({ now }));
    act(() => {
      result.current.setRange({ from, to, preset: "custom" });
    });
    const writtenUrl = replace.mock.calls[0]?.[0] as string;
    expect(writtenUrl).toContain("preset=custom");
    expect(writtenUrl).toContain(`from=${encodeURIComponent(from.toISOString())}`);
    expect(writtenUrl).toContain(`to=${encodeURIComponent(to.toISOString())}`);

    searchString = writtenUrl.replace(/^\?/, "");
    rerender();
    expect(result.current.range.preset).toBe("custom");
    expect(result.current.range.from.toISOString()).toBe(from.toISOString());
    expect(result.current.range.to.toISOString()).toBe(to.toISOString());
  });
});
