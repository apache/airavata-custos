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
import { type ReactNode, createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import { RoleSaveError, useCreateRole } from "../queries";

vi.mock("@/generated/core/sdk.gen", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/generated/core/sdk.gen")>()),
  postRoles: vi.fn(async () => ({ id: "role-new", name: "Reviewer" })),
  postRolesByIdPrivileges: vi.fn(async ({ body }: { body: { privilege: string } }) => {
    if (body.privilege === "core:users:write") throw new Error("forbidden");
  }),
  postUsersByIdRoles: vi.fn(async () => ({})),
}));

function wrapper({ children }: { children: ReactNode }) {
  return createElement(QueryClientProvider, { client: new QueryClient() }, children);
}

describe("useCreateRole", () => {
  it("reports failed steps with the created role so a retry edits it", async () => {
    const { result } = renderHook(() => useCreateRole(), { wrapper });
    const error = await act(() =>
      result.current
        .mutateAsync({
          name: "Reviewer",
          description: "",
          privileges: ["core:users:read", "core:users:write"],
          memberUserIds: ["u1"],
        })
        .then(
          () => null,
          (err: Error) => err,
        ),
    );
    expect(error).toBeInstanceOf(RoleSaveError);
    expect(error?.message).toContain("grant core:users:write (forbidden)");
    expect(error instanceof RoleSaveError && error.role).toMatchObject({
      id: "role-new",
      privileges: ["core:users:read"],
      holderIds: ["u1"],
    });
  });
});
