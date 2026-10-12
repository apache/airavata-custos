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
import { renderHook } from "@testing-library/react";
import type * as React from "react";
import { describe, expect, it, vi } from "vitest";

const post = vi.hoisted(() => vi.fn());
const get = vi.hoisted(() => vi.fn());
vi.mock("@/generated/core/sdk.gen", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/generated/core/sdk.gen")>()),
  getEventsDeliveriesById: get,
  postEventsDeliveriesByIdRetry: post,
}));

import { ApiError } from "@/shared/api/client";
import { eventKeys, useRetryDelivery } from "../queries";

const id = "6f1c2a4e-8d3b-4c7a-9e21-5b0d3f7a9c18";

function setup() {
  const client = new QueryClient();
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return { client, result: renderHook(() => useRetryDelivery(), { wrapper }).result };
}

describe("useRetryDelivery", () => {
  // Make sure a retry of a delivery that settled since it was shown is refused before posting.
  it("refuses a delivery that is no longer failed and refetches", async () => {
    get.mockResolvedValueOnce({ status: "SUCCEEDED" });
    const { client, result } = setup();
    const invalidate = vi.spyOn(client, "invalidateQueries");
    await expect(result.current.mutateAsync({ path: { id } })).rejects.toThrow(
      "This delivery already succeeded.",
    );
    expect(post).not.toHaveBeenCalled();
    expect(invalidate).toHaveBeenCalledWith({ queryKey: eventKeys.all });
  });

  // Make sure a retry of a delivery that was deleted resets it, so the drawer shows it is gone.
  it("resets a delivery that no longer exists", async () => {
    get.mockRejectedValueOnce(new ApiError(404, "/events/deliveries", null));
    const { client, result } = setup();
    const reset = vi.spyOn(client, "resetQueries");
    await expect(result.current.mutateAsync({ path: { id } })).rejects.toThrow();
    expect(reset).toHaveBeenCalledWith({ queryKey: eventKeys.delivery(id) });
  });
});
