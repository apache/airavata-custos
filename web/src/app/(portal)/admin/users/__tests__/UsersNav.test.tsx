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

import type { PrivilegeKey } from "@/generated/core/types.gen";
import { defineAbilitiesFor } from "@/shared/casl/abilities";
import { isNavItemVisible, portalNav } from "@/shared/layout/nav";
import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { UsersNav } from "../UsersNav";

let currentPrivileges: PrivilegeKey[] = [];

vi.mock("next/navigation", () => ({
  usePathname: () => "/admin/users/management",
}));

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => defineAbilitiesFor(currentPrivileges),
}));

const pendingMocks = vi.hoisted(() => ({ total: 0 }));

vi.mock("@/features/core/cluster-accounts/queries", () => ({
  useClusterAccounts: () => ({ data: { items: [], total: pendingMocks.total } }),
}));

describe("UsersNav", () => {
  beforeEach(() => {
    currentPrivileges = [];
    pendingMocks.total = 0;
  });

  // Make sure the approvals tab is only there for someone who can read
  // clusters, and that it carries the number of accounts waiting.
  it("shows Cluster accounts with the pending count for cluster readers", () => {
    currentPrivileges = ["core:users:read", "core:clusters:read"];
    pendingMocks.total = 3;
    render(<UsersNav />);
    const tab = screen.getByRole("link", { name: /Cluster accounts/ });
    expect(tab).toHaveTextContent("3");
  });

  it("skips the pending count for a cluster writer who cannot read clusters", () => {
    currentPrivileges = ["core:clusters:write"];
    pendingMocks.total = 3;
    render(<UsersNav />);
    expect(screen.getByRole("link", { name: "Cluster accounts" })).toHaveTextContent(
      /^Cluster accounts$/,
    );
  });

  it("hides Cluster accounts without clusters:read", () => {
    currentPrivileges = ["core:users:read"];
    render(<UsersNav />);
    expect(screen.queryByRole("link", { name: /Cluster accounts/ })).toBeNull();
  });

  it("hides Role management without roles:manage", () => {
    currentPrivileges = ["core:users:read"];
    render(<UsersNav />);
    expect(screen.getByRole("link", { name: "User management" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Role management" })).toBeNull();
  });

  it("shows Role management with roles:manage", () => {
    currentPrivileges = ["core:users:read", "core:roles:manage"];
    render(<UsersNav />);
    expect(screen.getByRole("link", { name: "User management" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Role management" })).toBeInTheDocument();
  });

  it("offers a cluster reader only the tab their privilege opens", () => {
    currentPrivileges = ["core:clusters:read"];
    render(<UsersNav />);
    expect(screen.getAllByRole("link").map((link) => link.textContent)).toEqual([
      "Cluster accounts",
    ]);
    const item = portalNav.filter((i) => i.href === "/admin/users");
    expect(item.filter((i) => isNavItemVisible(i, defineAbilitiesFor(currentPrivileges)))).toEqual(
      item,
    );
  });
});
