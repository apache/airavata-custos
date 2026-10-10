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
import { type ReactNode, createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import { ApiError } from "@/shared/api/client";
import { useCurrentSuAmount } from "../queries";

const latestDiff = vi.hoisted(() => vi.fn());

vi.mock("@/generated/core/sdk.gen", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/generated/core/sdk.gen")>()),
  getComputeAllocationsByIdDiffsLatest: latestDiff,
}));

function currentSu() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const wrapper = ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client }, children);
  return renderHook(() => useCurrentSuAmount({ id: "alloc-1", initial_su_amount: 1000 }), {
    wrapper,
  }).result;
}

describe("useCurrentSuAmount", () => {
  it("takes the latest diff's SU amount", async () => {
    latestDiff.mockResolvedValue({ new_su_amount: 1500 });
    const result = currentSu();
    await waitFor(() => expect(result.current.data).toBe(1500));
  });

  it("falls back to the initial amount before any diff exists", async () => {
    latestDiff.mockRejectedValue(new ApiError(404, "/diffs/latest", null));
    const result = currentSu();
    await waitFor(() => expect(result.current.data).toBe(1000));
  });
});
