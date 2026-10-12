import { defineAbilitiesFor } from "@/shared/casl/abilities";
import { fireEvent, render, screen } from "@testing-library/react";
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
import type * as React from "react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => defineAbilitiesFor(["core:privileges:grant"]),
}));
vi.mock("next-auth/react", () => ({ useSession: () => ({ data: null }) }));
vi.mock("@/shared/ui/sonner", () => ({ confirmToast: vi.fn(), toastOnSuccess: vi.fn() }));
vi.mock("@/features/core/roles/queries", () => ({
  usePrivilegeCatalog: () => ({ data: ["core:users:read"] }),
}));
const grant = vi.hoisted(() => vi.fn(async () => ({})));
vi.mock("@/features/core/users/queries", () => ({
  usePrivilegeHolders: (key?: string) => ({
    data: key && [
      { id: "g-1", user_id: "00000001-0000-4000-8000-000000000001", privilege: key },
      { id: "g-2", user_id: "00000002-0000-4000-8000-000000000002", privilege: key },
    ],
  }),
  useGrantPrivilege: () => ({ isPending: false, mutateAsync: grant }),
  useRevokePrivilege: () => ({ isPending: false }),
  useUser: () => ({}),
}));
vi.mock("@/shared/ui/select", () => ({
  Select: ({
    value,
    onValueChange,
    children,
  }: {
    value: string;
    onValueChange: (v: string) => void;
    children: React.ReactNode;
  }) => (
    <select aria-label="Privilege" value={value} onChange={(e) => onValueChange(e.target.value)}>
      <option value="" />
      {children}
    </select>
  ),
  SelectTrigger: () => null,
  SelectValue: () => null,
  SelectContent: ({ children }: { children: React.ReactNode }) => <>{children}</>,
  SelectItem: ({ value, children }: { value: string; children: React.ReactNode }) => (
    <option value={value}>{children}</option>
  ),
}));

import { PrivilegeHoldersCard } from "../PrivilegeHoldersCard";

describe("<PrivilegeHoldersCard />", () => {
  it("counts holders for a grant-only caller and grants by user ID", () => {
    render(<PrivilegeHoldersCard />);
    fireEvent.change(screen.getByLabelText("Privilege"), { target: { value: "core:users:read" } });
    expect(screen.getByText("2 users hold core:users:read directly.")).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("User ID"), {
      target: { value: " 00000003-0000-4000-8000-000000000003 " },
    });
    fireEvent.click(screen.getByRole("button", { name: "Grant" }));
    expect(grant).toHaveBeenCalledWith({
      userId: "00000003-0000-4000-8000-000000000003",
      privilege: "core:users:read",
      reason: undefined,
    });
    expect(screen.getByRole("button", { name: "Revoke" })).toBeInTheDocument();
  });
});
