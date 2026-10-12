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

const sdk = vi.hoisted(() => ({
  postComputeAllocationMembershipResourceOverrides: vi.fn(async () => undefined),
}));
vi.mock("@/generated/core/sdk.gen", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/generated/core/sdk.gen")>()),
  ...sdk,
}));

import { ApiError } from "@/shared/api/client";
import { allocationKeys, useCreateOverride } from "../queries";

const ALLOCATION = "01000005-0000-4000-8000-000001000005";
const MEMBERSHIP = "02000005-0000-4000-8000-000002000005";
const RESOURCE = "03000005-0000-4000-8000-000003000005";

// A resource another tab detached leaves the override picker after a refused add on it.
describe("useCreateOverride", () => {
  it("refreshes the allocation's resources when the resource was detached", async () => {
    sdk.postComputeAllocationMembershipResourceOverrides.mockRejectedValueOnce(
      new ApiError(404, "", null, "resource is not attached to this allocation"),
    );
    const client = new QueryClient();
    client.setQueryData(allocationKeys.resources(ALLOCATION), [{ id: RESOURCE }]);
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
    const { result } = renderHook(() => useCreateOverride(ALLOCATION, MEMBERSHIP), { wrapper });
    await expect(
      act(() =>
        result.current.mutateAsync({
          body: {
            compute_allocation_membership_id: MEMBERSHIP,
            compute_allocation_resource_id: RESOURCE,
            override_resource_amount: 1,
            override_resource_time: 1,
          },
        }),
      ),
    ).rejects.toThrow();
    expect(client.getQueryState(allocationKeys.resources(ALLOCATION))?.isInvalidated).toBe(true);
  });
});
