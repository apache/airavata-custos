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
import { ApiError } from "@/shared/api/client";
import { defineAbilitiesFor } from "@/shared/casl/abilities";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { UsersTable } from "../UsersTable";

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => defineAbilitiesFor([]),
}));

const userReads = vi.hoisted(() => vi.fn());
vi.mock("@/generated/core/sdk.gen", async (original) => ({
  ...(await original<object>()),
  getUsersById: async () => {
    userReads();
    throw new ApiError(404, "/users/00000099-0000-4000-8000-000000000099", {
      error: "user not found",
    });
  },
}));

vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(window.location.search),
}));

const visit = (search: string) => window.history.replaceState(null, "", `/admin/users${search}`);

const row: UserManagementRow = {
  id: "00000001-0000-4000-8000-000000000001",
  email: "user@example.org",
  first_name: "Example",
  last_name: "User",
  organization_id: "org-1",
  status: "ACTIVE",
  type: "VIRTUAL",
  roles: [
    {
      id: "role-1",
      name: "Administrator",
      description: null,
      is_system: false,
      created_at: "2026-01-01T00:00:00Z",
    },
  ],
  identities: [
    {
      id: "identity-1",
      user_id: "00000001-0000-4000-8000-000000000001",
      source: "cilogon",
      external_id: "e-1",
      created_at: "2026-01-01T00:00:00Z",
    },
  ],
  rolesLoading: false,
  identitiesLoading: false,
  rolesError: false,
  identitiesError: false,
};

function renderTable(canManageRoles: boolean, users: UserManagementRow[] = [row]) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const view = render(
    <QueryClientProvider client={client}>
      <UsersTable
        users={users}
        rolesCatalog={[
          {
            id: "role-1",
            name: "Administrator",
            description: null,
            is_system: false,
            created_at: "2026-01-01T00:00:00Z",
          },
        ]}
        rolesReady
        currentUserId="00000001-0000-4000-8000-000000000001"
        canManageRoles={canManageRoles}
        canReadDirectPrivileges={false}
        page={1}
        pageSize={25}
        total={users.length}
      />
    </QueryClientProvider>,
  );
  return view;
}

describe("UsersTable", () => {
  beforeEach(() => {
    visit("");
  });

  it("renders a hydrated user and marks the current user", () => {
    renderTable(true);
    expect(screen.getByText("Example User")).toBeInTheDocument();
    expect(screen.getByText("(You)")).toBeInTheDocument();
    expect(screen.getAllByText("Administrator")).toHaveLength(1);
    expect(screen.getAllByText("CILogon")).toHaveLength(1);
  });

  it("hides all role UI without roles:manage", () => {
    renderTable(false);
    expect(screen.queryByText("Roles")).toBeNull();
    expect(screen.queryByText("Administrator")).toBeNull();
    expect(screen.queryByRole("combobox", { name: /role/i })).toBeNull();
  });

  it("keeps rows visible while filtered details are loading or unavailable", () => {
    visit("?role=role-1&identity=cilogon");
    renderTable(true, [
      {
        ...row,
        roles: [],
        identities: [],
        rolesLoading: true,
        identitiesError: true,
      },
    ]);

    expect(screen.getByText("Example User")).toBeInTheDocument();
    expect(screen.getByText("Loading roles…")).toBeInTheDocument();
    expect(screen.getByText("Identities unavailable")).toBeInTheDocument();
  });

  it("keeps the open user in ?user= while typing in search, and drops it on close", () => {
    renderTable(true);
    fireEvent.click(screen.getByText("Example User"));
    fireEvent.change(screen.getByRole("searchbox"), { target: { value: "exam" } });

    expect(screen.getByText("Edit roles")).toBeInTheDocument();
    expect(window.location.search).toBe("?user=00000001-0000-4000-8000-000000000001");
    fireEvent.click(screen.getByRole("button", { name: /close/i }));
    expect(window.location.search).toBe("");
  });

  it("opens a not-found drawer for a ?user= id that does not resolve", async () => {
    visit("?user=00000099-0000-4000-8000-000000000099");
    renderTable(true);
    expect(
      await screen.findByRole("heading", { name: "This user no longer exists" }),
    ).toBeInTheDocument();
  });

  it("rewrites the URL to the filters it applied", async () => {
    const id = "0b5c2d4e-1f2a-4b3c-8d9e-0f1a2b3c4d5e";
    visit(`?q=%20exam%20&role=bogus&identity=nope&user=%20${id.toUpperCase()}%20`);
    renderTable(true);
    await waitFor(() => expect(window.location.search).toBe(`?q=exam&user=${id}`));
  });

  it("drops a malformed ?user= before any request", async () => {
    userReads.mockClear();
    visit("?user=not-a-uuid");
    renderTable(true);
    await waitFor(() => expect(window.location.search).toBe(""));
    expect(userReads).not.toHaveBeenCalled();
  });
});
