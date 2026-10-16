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

import { QueryClient, QueryClientProvider, useQuery } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";

vi.mock("next-auth/react", () => ({
  useSession: () => ({ data: { custosUserId: "u-1" }, status: "authenticated" }),
}));

import { identityKeys, refreshAfterPrivilegeChange, usePrivileges } from "../queries";

const server = setupServer(
  http.get("*/api/v1/user/privileges", () =>
    HttpResponse.json({ privileges: ["core:allocations:read", "core:clusters:read"] }),
  ),
);

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterAll(() => server.close());
afterEach(() => server.resetHandlers());

function wrapper({ children }: { children: React.ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

describe("usePrivileges", () => {
  it("fetches and validates the caller privileges payload", async () => {
    const { result } = renderHook(() => usePrivileges(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(["core:allocations:read", "core:clusters:read"]);
  });
});

describe("refreshAfterPrivilegeChange", () => {
  it("does not refetch a read the caller's refreshed privileges no longer allow", async () => {
    const client = new QueryClient();
    let held = ["core:roles:manage"];
    const roles = vi.fn(async () => []);
    renderHook(
      () => {
        const privileges = useQuery({
          queryKey: identityKeys.privileges(),
          queryFn: async () => held,
        });
        useQuery({
          queryKey: ["roles"],
          queryFn: roles,
          enabled: privileges.data?.includes("core:roles:manage") ?? false,
        });
      },
      {
        wrapper: ({ children }) => (
          <QueryClientProvider client={client}>{children}</QueryClientProvider>
        ),
      },
    );
    await waitFor(() => expect(roles).toHaveBeenCalledTimes(1));
    held = [];
    await act(() => refreshAfterPrivilegeChange(client, ["roles"]));
    expect(roles).toHaveBeenCalledTimes(1);
  });
});
