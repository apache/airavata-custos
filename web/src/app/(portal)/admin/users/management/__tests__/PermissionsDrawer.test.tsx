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
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { PermissionsDrawer } from "../PermissionsDrawer";

const queryMocks = vi.hoisted(() => ({
  mutateAsync: vi.fn(),
}));

vi.mock("@/features/core/users/queries", () => ({
  useDirectPrivileges: () => ({ data: [], isLoading: false, isError: false }),
  useRoleDetails: () => ({ roles: [], isLoading: false, isError: false }),
  useUpdateUserRoles: () => ({
    mutateAsync: queryMocks.mutateAsync,
    isPending: false,
  }),
}));

vi.mock("../RoleAssignMenu", () => ({
  RoleAssignMenu: ({
    onSave,
    error,
  }: {
    onSave: (roleIds: string[], reason?: string) => Promise<boolean>;
    error: string | null;
  }) => (
    <div>
      <button type="button" onClick={() => void onSave([], "Onboarding")}>
        Save mocked roles
      </button>
      {error ? <span role="alert">{error}</span> : null}
    </div>
  ),
}));

const user: UserManagementRow = {
  id: "user-1",
  email: "user@example.org",
  first_name: "Example",
  last_name: "User",
  roles: [{ id: "role-1", name: "Administrator" }],
  identities: [],
  rolesLoading: false,
  identitiesLoading: false,
  rolesError: false,
  identitiesError: false,
};

describe("PermissionsDrawer", () => {
  it("clears a role assignment error when the drawer closes", async () => {
    queryMocks.mutateAsync.mockRejectedValueOnce(new Error("role update failed"));
    const onClose = vi.fn();
    const props = {
      rolesCatalog: [{ id: "role-1", name: "Administrator" }],
      canManageRoles: true,
      canReadDirectPrivileges: false,
      currentUserEmail: "another@example.org",
      onClose,
    };
    const view = render(<PermissionsDrawer {...props} user={user} />);

    fireEvent.click(screen.getByRole("button", { name: "Save mocked roles" }));
    await waitFor(() => expect(screen.getByRole("alert")).toHaveTextContent("role update failed"));

    fireEvent.click(screen.getByRole("button", { name: "Close" }));
    expect(onClose).toHaveBeenCalled();

    view.rerender(<PermissionsDrawer {...props} user={null} />);
    view.rerender(<PermissionsDrawer {...props} user={user} />);
    expect(screen.queryByRole("alert")).toBeNull();
  });
});
