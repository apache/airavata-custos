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
import { afterEach, describe, expect, it, vi } from "vitest";
import { identityKeys } from "@/features/core/identity/queries";
import { applyUserRoleChanges, useUpdateUserRoles } from "../queries";

const apiMocks = vi.hoisted(() => ({
  assignUserRole: vi.fn(),
  removeUserRole: vi.fn(),
}));

vi.mock("@/generated/core/sdk.gen", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/generated/core/sdk.gen")>()),
  postUsersByIdRoles: apiMocks.assignUserRole,
  deleteUsersByIdRolesByRoleId: apiMocks.removeUserRole,
}));

afterEach(() => {
  apiMocks.assignUserRole.mockReset();
  apiMocks.removeUserRole.mockReset();
});

type AssignOptions = { body: { role_id: string; reason?: string } };
type RemoveOptions = { path: { roleId: string } };

// Records each call as "assign:<role>" or "remove:<role>"; `fail` rejects matching calls.
function recordCalls(fail: (call: string, attempt: number) => boolean = () => false) {
  const calls: string[] = [];
  const record = async (call: string) => {
    const attempt = calls.filter((c) => c === call).length;
    calls.push(call);
    if (fail(call, attempt)) throw new Error(`${call} failed`);
  };
  apiMocks.assignUserRole.mockImplementation(({ body }: AssignOptions) =>
    record(`assign:${body.role_id}`),
  );
  apiMocks.removeUserRole.mockImplementation(({ path }: RemoveOptions) =>
    record(`remove:${path.roleId}`),
  );
  return calls;
}

describe("applyUserRoleChanges", () => {
  it("applies additions before removals in a deterministic order", async () => {
    const calls = recordCalls();
    await applyUserRoleChanges({
      userId: "user-1",
      currentRoleIds: ["old-role"],
      desiredRoleIds: ["new-role"],
    });
    expect(calls).toEqual(["assign:new-role", "remove:old-role"]);
  });

  it("rolls back completed changes in reverse order when a later change fails", async () => {
    const calls = recordCalls((call, attempt) => call === "remove:old-role-2" && attempt === 0);
    await expect(
      applyUserRoleChanges({
        userId: "user-1",
        currentRoleIds: ["old-role-1", "old-role-2"],
        desiredRoleIds: ["new-role"],
      }),
    ).rejects.toThrow("completed changes were rolled back");
    expect(calls).toEqual([
      "assign:new-role",
      "remove:old-role-1",
      "remove:old-role-2",
      "assign:old-role-1",
      "remove:new-role",
    ]);
  });

  it("warns when rollback cannot fully restore the previous roles", async () => {
    recordCalls((call) => call === "remove:old-role" || call === "remove:new-role");
    await expect(
      applyUserRoleChanges({
        userId: "user-1",
        currentRoleIds: ["old-role"],
        desiredRoleIds: ["new-role"],
      }),
    ).rejects.toThrow("rollback could not fully restore");
  });
});

describe("useUpdateUserRoles", () => {
  it("invalidates current access after role changes", async () => {
    apiMocks.assignUserRole.mockResolvedValue({
      user_id: "user-1",
      role_id: "new-role",
    });
    const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    client.setQueryData(identityKeys.privileges(), ["core:roles:manage"]);
    client.setQueryData([...identityKeys.access("user-1"), true], {
      roles: [],
      direct: [],
      privileges: ["core:roles:manage"],
      provenance: true,
    });
    const wrapper = ({ children }: { children: ReactNode }) =>
      createElement(QueryClientProvider, { client }, children);
    const { result } = renderHook(() => useUpdateUserRoles(), { wrapper });

    await act(() =>
      result.current.mutateAsync({
        userId: "user-1",
        currentRoleIds: [],
        desiredRoleIds: ["new-role"],
      }),
    );

    expect(client.getQueryState(identityKeys.privileges())?.isInvalidated).toBe(true);
    expect(client.getQueryState([...identityKeys.access("user-1"), true])?.isInvalidated).toBe(
      true,
    );
  });
});
