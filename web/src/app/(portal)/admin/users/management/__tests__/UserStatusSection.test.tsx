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
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { UserStatusSection } from "../UserStatusSection";

const apiMocks = vi.hoisted(() => ({ merge: vi.fn() }));

vi.mock("@/generated/core/sdk.gen", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/generated/core/sdk.gen")>()),
  getUsers: async () => ({ items: [{ id: "survivor", email: "keep@example.org" }], total: 1 }),
  postUsersMerge: apiMocks.merge,
}));

vi.mock("@/shared/ui/sonner", () => ({
  confirmToast: (_message: string, _label: string, onConfirm: () => void) => onConfirm(),
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

describe("UserStatusSection", () => {
  it("merges the drawer user into the picked survivor, never the reverse", async () => {
    apiMocks.merge.mockResolvedValue({ id: "survivor", email: "keep@example.org" });
    const onMerged = vi.fn();
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={client}>
        <UserStatusSection user={user} canWrite onMerged={onMerged} />
      </QueryClientProvider>,
    );

    fireEvent.click(screen.getByRole("button", { name: "Merge into…" }));
    fireEvent.click(await screen.findByRole("radio"));
    fireEvent.click(screen.getByRole("button", { name: "Merge" }));

    await waitFor(() => expect(onMerged).toHaveBeenCalled());
    expect(apiMocks.merge).toHaveBeenCalledWith({
      body: { surviving_user_id: "survivor", retiring_user_id: "retiring" },
    });
  });
});
