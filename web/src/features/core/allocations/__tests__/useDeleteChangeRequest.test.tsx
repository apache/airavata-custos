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
import { act, renderHook } from "@testing-library/react";
import type * as React from "react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/generated/core/sdk.gen", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/generated/core/sdk.gen")>()),
  deleteComputeAllocationChangeRequestsById: async () => undefined,
}));

import { allocationKeys, useDeleteChangeRequest } from "../queries";

describe("useDeleteChangeRequest", () => {
  it("drops the deleted request's cached detail and timeline once its page unmounts", async () => {
    const client = new QueryClient();
    client.setQueryData(allocationKeys.changeRequestDetail("cr-1"), { change_status: "PENDING" });
    client.setQueryData(allocationKeys.latestChangeRequestEvent("cr-1"), {});
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
    const { result, unmount } = renderHook(() => useDeleteChangeRequest(), { wrapper });
    await act(() => result.current.mutateAsync("cr-1"));
    expect(client.getQueryData(allocationKeys.changeRequestDetail("cr-1"))).toBeDefined();
    unmount();
    expect(client.getQueryData(allocationKeys.changeRequestDetail("cr-1"))).toBeUndefined();
    expect(client.getQueryData(allocationKeys.latestChangeRequestEvent("cr-1"))).toBeUndefined();
  });
});
