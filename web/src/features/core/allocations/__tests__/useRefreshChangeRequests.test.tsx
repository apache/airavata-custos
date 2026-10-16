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
import { ApiError } from "@/shared/api/client";
import { isGone, useOnGone } from "@/shared/api/useOnGone";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import type * as React from "react";
import { describe, expect, it, vi } from "vitest";

const read = vi.hoisted(() => vi.fn());
vi.mock("@/generated/core/sdk.gen", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/generated/core/sdk.gen")>()),
  getComputeAllocationChangeRequestsById: read,
}));

import { useChangeRequest, useRefreshChangeRequests } from "../queries";

describe("useRefreshChangeRequests", () => {
  it("refreshes the lists behind a gone request without refetching its 404", async () => {
    read.mockRejectedValue(new ApiError(404, "/compute-allocation-change-requests/cr-1", {}));
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
    const { result } = renderHook(
      () => {
        const query = useChangeRequest("cr-1");
        useOnGone(isGone(query.error), useRefreshChangeRequests());
        return query;
      },
      { wrapper },
    );
    await waitFor(() => expect(result.current.error).toBeTruthy());
    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(read).toHaveBeenCalledTimes(1);
  });
});
