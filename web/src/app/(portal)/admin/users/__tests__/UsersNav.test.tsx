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
import { UsersNav } from "../UsersNav";

let currentPrivileges: Privilege[] = [];

vi.mock("next/navigation", () => ({
  usePathname: () => "/admin/users/management",
}));

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => defineAbilitiesFor(currentPrivileges),
}));

describe("UsersNav", () => {
  beforeEach(() => {
    currentPrivileges = [];
  });

  it("hides Role Management without roles:manage", () => {
    currentPrivileges = ["core:users:read"];
    render(<UsersNav />);
    expect(screen.getByRole("link", { name: "User Management" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "Role Management" })).toBeNull();
  });

  it("shows Role Management with roles:manage", () => {
    currentPrivileges = ["core:users:read", "core:roles:manage"];
    render(<UsersNav />);
    expect(screen.getByRole("link", { name: "User Management" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Role Management" })).toBeInTheDocument();
  });
});
