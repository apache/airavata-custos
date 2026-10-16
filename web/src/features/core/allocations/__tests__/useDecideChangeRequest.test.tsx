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

const ID = "04000005-0000-4000-8000-000004000005";
const APPROVER = "05000005-0000-4000-8000-000005000005";
const stored = vi.hoisted(() => ({
  current: { id: "", requested_su_amount: 1500, reason: "x", change_status: "PENDING" },
}));
const put = vi.hoisted(() => vi.fn(async () => ({})));
vi.mock("@/generated/core/sdk.gen", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/generated/core/sdk.gen")>()),
  getComputeAllocationChangeRequestsById: async () => stored.current,
  putComputeAllocationChangeRequestsById: put,
}));

import { allocationKeys, useDecideChangeRequest } from "../queries";

function renderDecide() {
  const client = new QueryClient();
  const invalidate = vi.spyOn(client, "invalidateQueries");
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return { invalidate, ...renderHook(() => useDecideChangeRequest(), { wrapper }) };
}

describe("useDecideChangeRequest", () => {
  it("sends the request as the server holds it with the decision and approver", async () => {
    stored.current = { ...stored.current, id: ID, change_status: "PENDING" };
    const { result, invalidate } = renderDecide();
    await act(() =>
      result.current.mutateAsync({ id: ID, change_status: "APPROVED", approver_id: APPROVER }),
    );
    expect(put).toHaveBeenCalledWith({
      path: { id: ID },
      body: { ...stored.current, change_status: "APPROVED", approver_id: APPROVER },
    });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: allocationKeys.all });
  });

  it("refuses a request decided since it was shown, and refreshes it", async () => {
    stored.current = { ...stored.current, id: ID, change_status: "APPROVED" };
    put.mockClear();
    const { result, invalidate } = renderDecide();
    await expect(
      act(() =>
        result.current.mutateAsync({ id: ID, change_status: "REJECTED", approver_id: APPROVER }),
      ),
    ).rejects.toThrow("This request was already approved.");
    expect(put).not.toHaveBeenCalled();
    expect(invalidate).toHaveBeenCalledWith({ queryKey: allocationKeys.all });
  });
});
