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
import { getMe, getUserPrivileges } from "@/generated/core/sdk.gen";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook } from "@testing-library/react";
import { type ReactNode, createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import { useMe } from "../queries";

// A refused sign-in: the session holds the refusal and no Custos user.
vi.mock("next-auth/react", () => ({
  useSession: () => ({ status: "authenticated", data: { authError: "user_disabled" } }),
}));
vi.mock("@/generated/core/sdk.gen", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/generated/core/sdk.gen")>()),
  getUserPrivileges: vi.fn(async () => ({ privileges: [] })),
  getMe: vi.fn(async () => ({})),
}));

describe("useMe", () => {
  it("makes no identity lookup for a refused sign-in", async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: { children: ReactNode }) =>
      createElement(QueryClientProvider, { client }, children);
    renderHook(() => useMe(), { wrapper });
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(getUserPrivileges).not.toHaveBeenCalled();
    expect(getMe).not.toHaveBeenCalled();
  });
});
