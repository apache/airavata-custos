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
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import type * as React from "react";
import { describe, expect, it, vi } from "vitest";

const list = vi.hoisted(() => vi.fn());
vi.mock("@/generated/core/sdk.gen", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/generated/core/sdk.gen")>()),
  getComputeAllocations: list,
}));
vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => false }) }));

import { useAllocations, useAllocationsByProject } from "../queries";

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <QueryClientProvider client={new QueryClient()}>{children}</QueryClientProvider>
);

// Without the read privilege the server answers with the caller's allocations whole.
list.mockResolvedValue({
  items: [
    { id: "a-1", name: "GPU pool", status: "ACTIVE", project_id: "p-1" },
    { id: "a-2", name: "CPU pool", status: "ACTIVE", project_id: "p-2" },
    { id: "a-3", name: "GPU legacy", status: "INACTIVE", project_id: "p-1" },
  ],
  total: 3,
});

describe("scoped allocation lists", () => {
  it("apply the search, status and page the server ignores", async () => {
    const { result } = renderHook(
      () => useAllocations({ q: "gpu", status: "ACTIVE", limit: 50, offset: 0 }),
      { wrapper },
    );
    await waitFor(() =>
      expect(result.current.data).toEqual({
        items: [expect.objectContaining({ id: "a-1" })],
        total: 1,
      }),
    );
    expect(list).toHaveBeenCalledWith({ query: undefined });
  });

  it("keep only the project's allocations", async () => {
    const { result } = renderHook(() => useAllocationsByProject("p-1"), { wrapper });
    await waitFor(() => expect(result.current.data?.map((a) => a.id)).toEqual(["a-1", "a-3"]));
  });
});
