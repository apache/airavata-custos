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

import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Privilege } from "@/features/core/identity/types";
import { defineAbilitiesFor } from "@/shared/casl/abilities";
import { Sidebar } from "../Sidebar";

let currentPrivileges: Privilege[] = [];

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => defineAbilitiesFor(currentPrivileges),
}));

vi.mock("next/navigation", () => ({
  usePathname: () => "/",
}));

const NEW_ADMIN_LABELS = ["Organizations", "Resources", "Users & Permissions"];

const FULL_PRIVILEGES: Privilege[] = [
  "core:organizations:read",
  "core:clusters:read",
  "core:users:read",
];

beforeEach(() => {
  currentPrivileges = [];
});

describe("Sidebar admin entries", () => {
  it("shows the merged admin entries for a full-privilege ability", () => {
    currentPrivileges = FULL_PRIVILEGES;
    render(<Sidebar />);
    for (const label of NEW_ADMIN_LABELS) {
      expect(screen.getByRole("link", { name: label })).toBeInTheDocument();
    }
  });

  it("hides them when the ability lacks the mapped privileges", () => {
    currentPrivileges = [];
    render(<Sidebar />);
    for (const label of NEW_ADMIN_LABELS) {
      expect(screen.queryByRole("link", { name: label })).toBeNull();
    }
  });
});
