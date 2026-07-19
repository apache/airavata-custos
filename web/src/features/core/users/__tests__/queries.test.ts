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
import { applyUserRoleChanges, useUpdateUserRoles, userKeys } from "../queries";
import type { UserRole } from "../schemas";

const apiMocks = vi.hoisted(() => ({
  assignUserRole: vi.fn(),
  removeUserRole: vi.fn(),
}));

vi.mock("../api", async (importOriginal) => ({
  ...(await importOriginal<typeof import("../api")>()),
  assignUserRole: apiMocks.assignUserRole,
  removeUserRole: apiMocks.removeUserRole,
}));

afterEach(() => {
  apiMocks.assignUserRole.mockReset();
  apiMocks.removeUserRole.mockReset();
});

describe("userKeys", () => {
  it("uses one shared role-grant key per user", () => {
    expect(userKeys.roles("user-1")).toEqual(["users", "roles", "user-1"]);
  });

  it("keeps role detail separate from user grants", () => {
    expect(userKeys.roleDetail("role-1")).toEqual(["users", "role-detail", "role-1"]);
  });

  it("carries pagination parameters in list keys", () => {
    const params = { limit: 25, offset: 25 };
    expect(userKeys.list(params)).toEqual(["users", "list", params]);
  });
});

describe("applyUserRoleChanges", () => {
  it("applies additions before removals in a deterministic order", async () => {
    const calls: string[] = [];
    const operations = {
      assign: vi.fn(async (_userId: string, roleId: string) => {
        calls.push(`assign:${roleId}`);
      }),
      remove: vi.fn(async (_userId: string, roleId: string) => {
        calls.push(`remove:${roleId}`);
      }),
    };

    await applyUserRoleChanges(
      {
        userId: "user-1",
        currentRoleIds: ["old-role"],
        desiredRoleIds: ["new-role"],
      },
      operations,
    );

    expect(calls).toEqual(["assign:new-role", "remove:old-role"]);
  });

  it("forwards the reason to assign operations", async () => {
    const assignArgs: { roleId: string; reason?: string }[] = [];
    const operations = {
      assign: vi.fn(async (_userId: string, roleId: string, reason?: string) => {
        assignArgs.push({ roleId, reason });
      }),
      remove: vi.fn(async () => undefined),
    };

    await applyUserRoleChanges(
      {
        userId: "user-1",
        currentRoleIds: [],
        desiredRoleIds: ["new-role"],
        reason: "Onboarding",
      },
      operations,
    );

    expect(assignArgs).toEqual([{ roleId: "new-role", reason: "Onboarding" }]);
  });

  it("rolls back completed changes in reverse order when a later change fails", async () => {
    const calls: string[] = [];
    let oldRoleRemovalAttempts = 0;
    const operations = {
      assign: vi.fn(async (_userId: string, roleId: string) => {
        calls.push(`assign:${roleId}`);
      }),
      remove: vi.fn(async (_userId: string, roleId: string) => {
        calls.push(`remove:${roleId}`);
        if (roleId === "old-role-2" && oldRoleRemovalAttempts++ === 0) {
          throw new Error("backend rejected removal");
        }
      }),
    };

    await expect(
      applyUserRoleChanges(
        {
          userId: "user-1",
          currentRoleIds: ["old-role-1", "old-role-2"],
          desiredRoleIds: ["new-role"],
        },
        operations,
      ),
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
    let newRoleRemovalAttempts = 0;
    const operations = {
      assign: vi.fn(async () => undefined),
      remove: vi.fn(async (_userId: string, roleId: string) => {
        if (roleId === "old-role") throw new Error("update failed");
        if (roleId === "new-role" && newRoleRemovalAttempts++ === 0) {
          throw new Error("rollback failed");
        }
      }),
    };

    await expect(
      applyUserRoleChanges(
        {
          userId: "user-1",
          currentRoleIds: ["old-role"],
          desiredRoleIds: ["new-role"],
        },
        operations,
      ),
    ).rejects.toThrow("rollback could not fully restore");
  });
});

describe("useUpdateUserRoles", () => {
  it("updates the cached role baseline before the mutation resolves", async () => {
    apiMocks.assignUserRole.mockResolvedValue({
      user_id: "user-1",
      role_id: "new-role",
    });
    apiMocks.removeUserRole.mockResolvedValue(undefined);
    const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    client.setQueryData<UserRole[]>(userKeys.roles("user-1"), [
      { user_id: "user-1", role_id: "old-role" },
    ]);
    const wrapper = ({ children }: { children: ReactNode }) =>
      createElement(QueryClientProvider, { client }, children);
    const { result } = renderHook(() => useUpdateUserRoles(), { wrapper });

    await act(() =>
      result.current.mutateAsync({
        userId: "user-1",
        currentRoleIds: ["old-role"],
        desiredRoleIds: ["new-role"],
      }),
    );

    expect(client.getQueryData<UserRole[]>(userKeys.roles("user-1"))).toEqual([
      { user_id: "user-1", role_id: "new-role" },
    ]);
  });

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
