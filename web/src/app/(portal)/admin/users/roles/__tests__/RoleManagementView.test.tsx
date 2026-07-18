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

import { defineAbilitiesFor } from "@/shared/casl/abilities";
import type { Privilege } from "@/features/core/identity/types";
import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { RoleManagementView } from "../RoleManagementView";

let currentPrivileges: Privilege[] = [];

vi.mock("next/navigation", () => ({
  usePathname: () => "/admin/users/roles",
}));

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => defineAbilitiesFor(currentPrivileges),
}));

vi.mock("../RolesGrid", () => ({
  RolesGrid: () => <div>Roles grid</div>,
}));

vi.mock("../RoleFormDialog", () => ({
  RoleFormDialog: ({ triggerContent }: { triggerContent: string }) => (
    <button type="button">{triggerContent}</button>
  ),
}));

describe("RoleManagementView", () => {
  beforeEach(() => {
    currentPrivileges = [];
  });

  it("blocks the roles page without roles:manage and does not mount the grid", () => {
    currentPrivileges = ["core:users:read"];
    render(<RoleManagementView />);
    expect(screen.getByRole("alert")).toHaveTextContent(/do not have permission to manage roles/i);
    expect(screen.queryByText("Roles grid")).toBeNull();
    expect(screen.queryByRole("button", { name: "Create role" })).toBeNull();
    expect(screen.queryByRole("link", { name: "Role Management" })).toBeNull();
  });

  it("renders role management UI with roles:manage", () => {
    currentPrivileges = ["core:roles:manage"];
    render(<RoleManagementView />);
    expect(screen.getByText("Roles grid")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create role" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Role Management" })).toBeInTheDocument();
  });
});
