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
import { afterEach, describe, expect, it, vi } from "vitest";
import { useAmieAction } from "../queries";

const mocks = vi.hoisted(() => ({
  retry: vi.fn(),
  toast: { success: vi.fn() },
}));

vi.mock("@/generated/amie/sdk.gen", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/generated/amie/sdk.gen")>()),
  postConnectorsAmiePacketsByIdRetry: mocks.retry,
}));
vi.mock("sonner", () => ({ toast: mocks.toast }));

afterEach(() => vi.clearAllMocks());

function renderRetry() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  const wrapper = ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client }, children);
  return renderHook(() => useAmieAction("retryPacket"), { wrapper }).result;
}

describe("useAmieAction", () => {
  it("attempts every id and reports a partial failure with the backend message", async () => {
    mocks.retry.mockImplementation(({ path }: { path: { id: string } }) =>
      path.id === "b" ? Promise.reject(new Error("packet retry not supported")) : Promise.resolve(),
    );
    const result = renderRetry();
    result.current.mutate(["a", "b"]);
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(mocks.retry).toHaveBeenCalledTimes(2);
    expect(result.current.error?.message).toBe("1 of 2 failed: packet retry not supported");
    expect(mocks.toast.success).not.toHaveBeenCalled();
  });
});
