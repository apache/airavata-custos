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

import { renderHook } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { useCanonicalParams, useShallowSearchParams } from "../useShallowSearchParams";

let searchString = "q=a";

vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(searchString),
}));

it("re-syncs when a same-route Next Link changes the query", () => {
  window.history.replaceState(null, "", "/admin/traces?q=a");
  const { result, rerender } = renderHook(() => useShallowSearchParams());
  expect(result.current.get("q")).toBe("a");
  searchString = "q=b";
  rerender();
  expect(result.current.get("q")).toBe("b");
});

// Make sure a list drops the filters it ignored, empty ones too, from the address bar and leaves other keys alone.
it("rewrites the owned keys to the applied filters", () => {
  window.history.replaceState(
    null,
    "",
    "/admin/traces?source=bogus&page=2abc&trace=ab&status=done",
  );
  renderHook(() => useCanonicalParams({ source: [], page: undefined, status: ["done"] }));
  expect(window.location.search).toBe("?trace=ab&status=done");
  window.history.replaceState(null, "", "/admin/traces?q=&status=done");
  renderHook(() => useCanonicalParams({ q: undefined, status: ["done"] }));
  expect(window.location.search).toBe("?status=done");
});
