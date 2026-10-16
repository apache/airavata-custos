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

import type { UserManagementRow } from "@/features/core/users/queries";
import { QueryClient, QueryClientProvider, useQuery } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { UserStatusSection } from "../UserStatusSection";

const apiMocks = vi.hoisted(() => ({
  merge: vi.fn(),
  status: vi.fn(),
  toast: vi.fn(),
  confirm: vi.fn(),
  endSession: vi.fn(),
}));

vi.mock("@/shared/auth/useSignOut", () => ({ endSession: apiMocks.endSession }));

vi.mock("@/generated/core/sdk.gen", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/generated/core/sdk.gen")>()),
  getUsers: async () => ({
    items: [
      {
        id: "survivor",
        email: "keep@example.org",
        first_name: "Dup",
        last_name: "User",
        status: "ACTIVE",
      },
      { id: "admin", email: "admin@example.org", status: "ACTIVE" },
      { id: "gone", email: "gone@example.org", status: "MERGED" },
      { id: "banned", email: "banned@example.org", status: "SUSPENDED" },
    ],
    total: 4,
  }),
  postUsersMerge: apiMocks.merge,
  putUsersByIdStatus: apiMocks.status,
}));

vi.mock("sonner", () => ({ toast: { success: apiMocks.toast, dismiss: vi.fn() } }));

vi.mock("@/shared/ui/sonner", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/shared/ui/sonner")>()),
  confirmToast: (message: string, _label: string, onConfirm: () => void) => {
    apiMocks.confirm(message);
    onConfirm();
  },
}));

const user: UserManagementRow = {
  id: "retiring",
  email: "dup@example.org",
  first_name: "Dup",
  last_name: "User",
  organization_id: "org-1",
  status: "ACTIVE",
  type: "VIRTUAL",
  roles: [],
  identities: [],
  rolesLoading: false,
  identitiesLoading: false,
  rolesError: false,
  identitiesError: false,
};

// The drawer's user refetches as MERGED, unmounting the dialog, while the
// merge still awaits a slower invalidated query.
function Drawer({ onMerged }: { onMerged: () => void }) {
  const merged = () => apiMocks.merge.mock.calls.length > 0;
  const { data } = useQuery({
    queryKey: ["drawer-user"],
    queryFn: async () => ({ ...user, status: merged() ? ("MERGED" as const) : user.status }),
  });
  useQuery({
    queryKey: ["slow"],
    queryFn: () => new Promise((resolve) => setTimeout(() => resolve(1), merged() ? 50 : 0)),
  });
  return data ? (
    <UserStatusSection user={data} canWrite isCurrentUser={false} onMerged={onMerged} />
  ) : null;
}

describe("UserStatusSection", () => {
  it.each([
    {
      name: "a PENDING user",
      row: { ...user, status: "PENDING" as const },
      canWrite: true,
      select: true,
      merge: true,
    },
    {
      name: "a MERGED user",
      row: { ...user, status: "MERGED" as const },
      canWrite: true,
      select: true,
      merge: false,
    },
    {
      name: "a caller without users:write",
      row: user,
      canWrite: false,
      select: false,
      merge: false,
    },
  ])(
    "offers the status and merge controls the server allows on $name",
    ({ row, canWrite, select, merge }) => {
      render(
        <QueryClientProvider client={new QueryClient()}>
          <UserStatusSection
            user={row}
            canWrite={canWrite}
            isCurrentUser={false}
            onMerged={vi.fn()}
          />
        </QueryClientProvider>,
      );
      expect(Boolean(screen.queryByRole("combobox", { name: "User status" }))).toBe(select);
      expect(Boolean(screen.queryByRole("button", { name: "Merge into…" }))).toBe(merge);
    },
  );

  it.each([
    {
      name: "a pending user",
      row: { ...user, status: "PENDING" as const },
      self: false,
      message: "Suspend Dup User? Dup User can no longer sign in.",
    },
    {
      name: "yourself",
      row: user,
      self: true,
      message: "Suspend yourself? You will be signed out and can no longer sign in.",
    },
  ])("asks before suspending $name, since that ends sign-in", async ({ row, self, message }) => {
    apiMocks.status.mockResolvedValue({ data: undefined });
    render(
      <QueryClientProvider client={new QueryClient()}>
        <UserStatusSection user={row} canWrite isCurrentUser={self} onMerged={vi.fn()} />
      </QueryClientProvider>,
    );
    fireEvent.click(screen.getByRole("combobox", { name: "User status" }));
    const option = await screen.findByRole("option", { name: "Suspended" });
    fireEvent.pointerDown(option);
    fireEvent.mouseDown(option);
    fireEvent.pointerUp(option);
    fireEvent.mouseUp(option);
    fireEvent.click(option);
    expect(apiMocks.confirm).toHaveBeenCalledWith(message);
    await waitFor(() => expect(apiMocks.status).toHaveBeenCalled());
    // Suspending yourself signs out at once instead of refetching under the ended session.
    await waitFor(() => expect(apiMocks.endSession).toHaveBeenCalledTimes(self ? 1 : 0));
  });

  // The survivor shares the retiring user's name, so both read with their email.
  it("merges the drawer user into the picked mergeable survivor and confirms it", async () => {
    apiMocks.merge.mockResolvedValue({
      id: "survivor",
      email: "keep@example.org",
      first_name: "Dup",
      last_name: "User",
    });
    const onMerged = vi.fn();
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={client}>
        <Drawer onMerged={onMerged} />
      </QueryClientProvider>,
    );

    fireEvent.click(await screen.findByRole("button", { name: "Merge into…" }));
    expect(await screen.findAllByRole("radio")).toHaveLength(2);
    fireEvent.click(screen.getByRole("radio", { name: /keep@example\.org/ }));
    fireEvent.click(screen.getByRole("button", { name: "Merge" }));

    await waitFor(() =>
      expect(apiMocks.toast).toHaveBeenCalledWith("Merged into Dup User (keep@example.org)"),
    );
    expect(onMerged).toHaveBeenCalled();
    expect(apiMocks.confirm).toHaveBeenCalledWith(
      expect.stringContaining(
        "Merge Dup User (dup@example.org) into Dup User (keep@example.org)? Identities, cluster accounts, PI projects and memberships move over",
      ),
    );
    expect(apiMocks.merge).toHaveBeenCalledWith({
      body: { surviving_user_id: "survivor", retiring_user_id: "retiring" },
    });
  });
});
