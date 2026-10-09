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

import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { RoleAssignMenu } from "../RoleAssignMenu";

const { confirmToast } = vi.hoisted(() => ({ confirmToast: vi.fn() }));
vi.mock("@/shared/ui/sonner", () => ({ confirmToast }));

const roleDetailsMock = vi.hoisted(() => ({
  details: [{ role: { id: "role-1", name: "Administrator", is_system: false, created_at: "2026-01-01T00:00:00Z" }, privileges: ["core:roles:manage"] }],
  isLoading: false,
  isError: false,
}));

vi.mock("@/features/core/users/queries", () => ({
  useRoleDetails: () => roleDetailsMock,
}));

function openAndUnassign(isCurrentUser: boolean) {
  const onSave = vi.fn();
  render(
    <RoleAssignMenu
      roles={[{ id: "role-1", name: "Administrator", is_system: false, created_at: "2026-01-01T00:00:00Z" }]}
      heldRoleIds={new Set(["role-1"])}
      onSave={onSave}
      triggerLabel="Manage user roles"
      isCurrentUser={isCurrentUser}
      isPending={false}
    />,
  );
  fireEvent.click(screen.getByRole("button", { name: /edit roles/i }));
  fireEvent.click(screen.getByRole("button", { name: "Unassign" }));
  return onSave;
}

describe("RoleAssignMenu", () => {
  beforeEach(() => {
    confirmToast.mockClear();
    roleDetailsMock.details = [
      { role: { id: "role-1", name: "Administrator", is_system: false, created_at: "2026-01-01T00:00:00Z" }, privileges: ["core:roles:manage"] },
    ];
    roleDetailsMock.isError = false;
  });

  it("submits the typed reason with the role set", () => {
    const onSave = openAndUnassign(false);
    fireEvent.change(screen.getByLabelText(/reason/i), {
      target: { value: "Onboarding new team member" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(onSave).toHaveBeenCalledWith([], "Onboarding new team member", expect.any(Function));
  });

  it("confirms before self-removing a role-manager role", () => {
    const onSave = openAndUnassign(true);
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(confirmToast).toHaveBeenCalled();
    expect(onSave).not.toHaveBeenCalled();
  });

  it("uses a conservative confirmation when current-user privileges are unavailable", () => {
    roleDetailsMock.details = [];
    roleDetailsMock.isError = true;
    const onSave = openAndUnassign(true);
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(confirmToast).toHaveBeenCalledWith(
      expect.stringContaining("privileges are unavailable"),
      "Continue",
      expect.any(Function),
    );
    expect(onSave).not.toHaveBeenCalled();
  });
});
