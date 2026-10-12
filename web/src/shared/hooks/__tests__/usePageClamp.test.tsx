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
import { afterEach, describe, expect, it, vi } from "vitest";

vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(window.location.search),
}));

import { pageParam, useStatusSearchList } from "../usePageClamp";

describe("useStatusSearchList", () => {
  afterEach(() => vi.useRealTimers());

  it("reads its page from ?page= and writes search trimmed, back on page 1", () => {
    vi.useFakeTimers();
    window.history.replaceState(null, "", "/projects?page=3");
    const useList = vi.fn(() => ({ data: { items: [], total: 0 } }) as never);
    const { result } = renderHook(() => useStatusSearchList(["ACTIVE"], useList));
    expect(useList).toHaveBeenLastCalledWith({
      limit: 50,
      offset: 100,
      status: undefined,
      q: undefined,
    });
    expect(result.current.pagination.page).toBe(3);

    act(() => result.current.onSearchChange("  genomics "));
    expect(result.current.search).toBe("  genomics ");
    act(() => vi.advanceTimersByTime(300));
    expect(window.location.search).toBe("?q=genomics");
    expect(useList).toHaveBeenLastCalledWith({
      limit: 50,
      offset: 0,
      status: undefined,
      q: "genomics",
    });
  });

  it("rewrites ignored filters and page to the ones it applied", () => {
    window.history.replaceState(null, "", "/projects?status=bogus&page=2abc&q=%20genomics%20&x=1");
    const useList = vi.fn(() => ({ data: { items: [], total: 0 } }) as never);
    renderHook(() => useStatusSearchList(["ACTIVE"], useList));
    expect(Object.fromEntries(new URLSearchParams(window.location.search))).toEqual({
      q: "genomics",
      x: "1",
    });
  });

  it("drops empty filters and keeps a UUID search lowercased in the URL", () => {
    const id = "0B9F2C4E-6A1D-4E8B-9C3F-1A2B3C4D5E6F";
    const useList = vi.fn(() => ({ data: { items: [], total: 0 } }) as never);
    window.history.replaceState(null, "", "/projects?q=&status=");
    renderHook(() => useStatusSearchList(["ACTIVE"], useList));
    expect(window.location.search).toBe("");
    window.history.replaceState(null, "", `/projects?q=${id}`);
    renderHook(() => useStatusSearchList(["ACTIVE"], useList));
    expect(window.location.search).toBe(`?q=${id.toLowerCase()}`);
  });
});

describe("pageParam", () => {
  // Make sure only a plain whole number from 1 selects a page; anything else is page 1.
  it("reads strict whole page numbers", () => {
    const read = (page: string) => pageParam(new URLSearchParams({ page }));
    expect(read("3")).toBe(3);
    for (const bad of ["2abc", "2.5", " 2 ", "0", "-1", "abc", ""]) expect(read(bad)).toBe(1);
  });
});
