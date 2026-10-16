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
import type { ProjectContext } from "@/generated/analytics/types.gen";
import { renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import contexts from "../__fixtures__/contexts.json";
import { useAnalyticsSelection } from "../useAnalyticsSelection";

vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(window.location.search),
}));

describe("useAnalyticsSelection", () => {
  it("drops an allocation the caller cannot see and says so", async () => {
    window.history.replaceState(null, "", "/analytics?allocation=bogus");
    const { result } = renderHook(() => useAnalyticsSelection(contexts as ProjectContext[], true));
    await waitFor(() => expect(window.location.search).toBe(""));
    expect(result.current.unavailable).toBe(true);
    expect(result.current.allocation?.id).toBe("01000007-0000-4000-8000-000001000007");
  });

  it("drops a foreign project and rewrites an accepted allocation to its normalised form", async () => {
    window.history.replaceState(
      null,
      "",
      "/analytics?project=foreign&allocation=%2001000006-0000-4000-8000-000001000006%20",
    );
    const { result } = renderHook(() => useAnalyticsSelection(contexts as ProjectContext[], true));
    await waitFor(() =>
      expect(window.location.search).toBe(
        "?project=18000001-0000-4000-8000-000018000001&allocation=01000006-0000-4000-8000-000001000006",
      ),
    );
    expect(result.current.unavailable).toBe(true);
    expect(result.current.allocation?.id).toBe("01000006-0000-4000-8000-000001000006");
  });
});
