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

import { getRolesById } from "@/generated/core/sdk.gen";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { type ReactNode, createElement, useState } from "react";
import { describe, expect, it, vi } from "vitest";
import { RoleSaveError, useCreateRole, useDeleteRole, useRoleRow } from "../queries";

vi.mock("@/generated/core/sdk.gen", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/generated/core/sdk.gen")>()),
  postRoles: vi.fn(async () => ({ id: "role-new", name: "Reviewer" })),
  postRolesByIdPrivileges: vi.fn(async ({ body }: { body: { privilege: string } }) => {
    if (body.privilege === "core:users:write") throw new Error("forbidden");
  }),
  postUsersByIdRoles: vi.fn(async ({ path }: { path: { id: string } }) => {
    if (path.id === "u2") throw new Error("user is merged");
  }),
  getUsersById: vi.fn(async () => ({ id: "u2", first_name: "Ada", last_name: "Lovelace" })),
  getRolesById: vi.fn(async () => ({ role: { id: "r1", name: "Reviewer" }, privileges: [] })),
  getRolesByIdHolders: vi.fn(async () => []),
  deleteRolesById: vi.fn(async () => undefined),
}));

function wrapper({ children }: { children: ReactNode }) {
  return createElement(QueryClientProvider, { client: new QueryClient() }, children);
}

describe("useCreateRole", () => {
  it("reports failed steps by name with the created role so a retry edits it", async () => {
    const { result } = renderHook(() => useCreateRole(), { wrapper });
    const error = await act(() =>
      result.current
        .mutateAsync({
          name: "Reviewer",
          description: "",
          privileges: ["core:users:read", "core:users:write"],
          memberUserIds: ["u1", "u2"],
        })
        .then(
          () => null,
          (err: Error) => err,
        ),
    );
    expect(error).toBeInstanceOf(RoleSaveError);
    expect(error?.message).toContain("grant core:users:write (forbidden)");
    expect(error?.message).toContain("add Ada Lovelace (user is merged)");
    expect(error instanceof RoleSaveError && error.role).toMatchObject({
      id: "role-new",
      privileges: ["core:users:read"],
      holderIds: ["u1"],
    });
  });
});

describe("useDeleteRole", () => {
  it("closes the dialog and drops the deleted role's row before refreshing, so it is not refetched", async () => {
    const { result } = renderHook(
      () => {
        const [open, setOpen] = useState(true);
        return {
          row: useRoleRow(open ? "r1" : undefined),
          remove: useDeleteRole(() => setOpen(false)),
        };
      },
      { wrapper },
    );
    await waitFor(() => expect(result.current.row.isSuccess).toBe(true));
    await act(() => result.current.remove.mutateAsync("r1"));
    expect(getRolesById).toHaveBeenCalledTimes(1);
  });
});
