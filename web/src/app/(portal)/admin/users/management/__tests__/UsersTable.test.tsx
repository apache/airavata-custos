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

import type { UserManagementRow } from "@/features/core/users/types";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { UsersTable } from "../UsersTable";

const searchParamMocks = vi.hoisted(() => ({
  params: new URLSearchParams(),
  replace: vi.fn(),
}));

vi.mock("@/shared/hooks/useShallowSearchParams", () => ({
  useShallowSearchParams: () => searchParamMocks.params,
  replaceShallowSearchParams: searchParamMocks.replace,
}));

const row: UserManagementRow = {
  id: "user-1",
  email: "user@example.org",
  first_name: "Example",
  last_name: "User",
  roles: [{ id: "role-1", name: "Administrator" }],
  identities: [{ id: "identity-1", user_id: "user-1", source: "cilogon" }],
  rolesLoading: false,
  identitiesLoading: false,
  rolesError: false,
  identitiesError: false,
};

function renderTable(canManageRoles: boolean, users: UserManagementRow[] = [row]) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const onPageChange = vi.fn();
  const view = render(
    <QueryClientProvider client={client}>
      <UsersTable
        users={users}
        rolesCatalog={[{ id: "role-1", name: "Administrator" }]}
        currentUserEmail="user@example.org"
        canManageRoles={canManageRoles}
        canReadDirectPrivileges={false}
        page={1}
        pageSize={25}
        total={users.length}
        onPageChange={onPageChange}
      />
    </QueryClientProvider>,
  );
  return { ...view, onPageChange };
}

describe("UsersTable", () => {
  beforeEach(() => {
    searchParamMocks.params = new URLSearchParams();
    searchParamMocks.replace.mockReset();
  });

  it("renders a hydrated user and marks the current user", () => {
    renderTable(true);
    expect(screen.getByText("Example User")).toBeInTheDocument();
    expect(screen.getByText("(You)")).toBeInTheDocument();
    expect(screen.getAllByText("Administrator")).toHaveLength(2);
    expect(screen.getAllByText("CILogon")).toHaveLength(2);
  });

  it("hides all role UI without roles:manage", () => {
    renderTable(false);
    expect(screen.queryByText("Roles")).toBeNull();
    expect(screen.queryByText("Administrator")).toBeNull();
    expect(screen.queryByRole("combobox", { name: /role/i })).toBeNull();
  });

  it("keeps rows visible while filtered details are loading or unavailable", () => {
    searchParamMocks.params = new URLSearchParams([
      ["role", "role-1"],
      ["identity", "cilogon"],
    ]);
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

  it("keeps the permissions drawer open while typing in search", () => {
    const { onPageChange } = renderTable(true);
    fireEvent.click(screen.getByText("Example User"));
    expect(screen.getByText("Edit roles")).toBeInTheDocument();

    fireEvent.change(screen.getByRole("searchbox"), { target: { value: "exam" } });

    expect(screen.getByText("Edit roles")).toBeInTheDocument();
    expect(onPageChange).not.toHaveBeenCalled();
  });
});
