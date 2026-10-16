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
import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import { PermissionsDrawer } from "../PermissionsDrawer";

const access = vi.hoisted(() => ({
  role: [] as string[],
  direct: [] as Array<{ id: string; privilege: string }>,
  denied: [] as string[],
}));

vi.mock("@/features/core/users/queries", () => ({
  useDirectPrivileges: () => ({ data: access.direct, isLoading: false, isError: false }),
  useRoleDetails: () => ({
    details: [{ id: "r-1", privileges: access.role }],
    isLoading: false,
    isError: false,
  }),
  useUpdateUserRoles: () => ({ mutate: vi.fn(), isPending: false }),
}));
vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({ can: (_: string, subject: string) => !access.denied.includes(subject) }),
}));
vi.mock("@/shared/ui/SideDrawer", () => ({
  SideDrawer: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));
vi.mock("../UserStatusSection", () => ({
  UserStatusSection: ({ canWrite }: { canWrite: boolean }) =>
    canWrite ? <p>status writes</p> : null,
}));
vi.mock("../UserActivitySections", () => ({ UserActivitySections: () => null }));
vi.mock("../EditUserNameDialog", () => ({ EditUserNameDialog: () => null }));
vi.mock("../RoleAssignMenu", () => ({
  RoleAssignMenu: () => <button type="button">Assign</button>,
}));
vi.mock("../IdentitiesSection", () => ({
  IdentitiesSection: ({ canWrite }: { canWrite: boolean }) =>
    canWrite ? <p>identity writes</p> : null,
}));
vi.mock("../DirectPrivilegesSection", () => ({
  DirectPrivilegesSection: () => <p>direct grants</p>,
}));

const user: UserManagementRow = {
  id: "u-1",
  email: "u1@example.org",
  first_name: "U",
  last_name: "One",
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

const renderDrawer = (canReadDirectPrivileges = true) =>
  render(
    <PermissionsDrawer
      user={user}
      rolesCatalog={[]}
      canManageRoles
      canReadDirectPrivileges={canReadDirectPrivileges}
      currentUserId="admin"
      onClose={vi.fn()}
    />,
  );

describe("PermissionsDrawer", () => {
  it("gives a users:write holder the status and identity writes, and the role menu", () => {
    renderDrawer();
    expect(screen.getByText("status writes")).toBeInTheDocument();
    expect(screen.getByText("identity writes")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Assign" })).toBeInTheDocument();
  });

  it("keeps status and identities read-only without users:write", () => {
    access.denied = ["User"];
    renderDrawer();
    expect(screen.queryByText("status writes")).not.toBeInTheDocument();
    expect(screen.queryByText("identity writes")).not.toBeInTheDocument();
    access.denied = [];
  });

  it("lists role and direct privileges together, and direct grants only to a grant holder", () => {
    access.role = ["core:users:read"];
    access.direct = [{ id: "g-1", privilege: "core:traces:read" }];
    const { unmount } = renderDrawer();
    expect(screen.getByText("Effective privileges")).toBeInTheDocument();
    expect(screen.getByText("core:users")).toBeInTheDocument();
    expect(screen.getByText("core:traces")).toBeInTheDocument();
    expect(screen.getByText("direct grants")).toBeInTheDocument();
    unmount();

    renderDrawer(false);
    expect(screen.getByText("Role-derived privileges")).toBeInTheDocument();
    expect(screen.queryByText("direct grants")).not.toBeInTheDocument();
  });
});
