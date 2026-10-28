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

import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { ProjectMember } from "../schemas";

function member(over: Partial<ProjectMember>): ProjectMember {
  return {
    id: "pm-x",
    project_id: "p-1",
    user_id: "u-x",
    email: "x@sample.example.edu",
    display_name: "Member X",
    role: "MEMBER",
    status: "ACTIVE",
    added_time: "2026-01-01T00:00:00.000Z",
    allocations: [],
    ...over,
  };
}

const members = [
  member({ id: "pm-1", user_id: "u-1", display_name: "Ada Pi", role: "PI" }),
  member({ id: "pm-3", user_id: "u-3", display_name: "Cira Member", role: "MEMBER" }),
];
const allocations = [
  { id: "a-1", name: "Alpha", end_time: "2027-01-01T00:00:00Z" },
  { id: "a-2", name: "Beta", end_time: "2027-01-01T00:00:00Z" },
];
const ciraInAlpha = {
  id: "m-1",
  compute_allocation_id: "a-1",
  user_id: "u-3",
  membership_status: "ACTIVE",
};

const roleMutation = { mutateAsync: vi.fn(async () => undefined), isPending: false };
const toggle = { mutateAsync: vi.fn(async () => undefined), isPending: false };
vi.mock("@/features/core/users/components/UserDetailsDrawer", () => ({
  UserDetailsDrawer: () => null,
}));
vi.mock("../queries", () => ({
  useProjectMembers: () => ({ data: members, isLoading: false, error: null, refetch: vi.fn() }),
  useUpdateProjectMember: () => roleMutation,
}));
vi.mock("@/features/core/allocations/queries", () => ({
  useAllocationsByProject: () => ({ data: { items: allocations }, isLoading: false, error: null }),
  useMembershipsByAllocation: () => ({
    byAllocation: { "a-1": [ciraInAlpha], "a-2": [] },
    isLoading: false,
    error: null,
  }),
  useToggleAllocationAccess: () => toggle,
}));
const search = vi.hoisted((): { match: unknown } => ({ match: undefined }));
vi.mock("@/features/core/users/queries", () => ({
  useUserEmailSearch: () => ({ suggestions: [], match: search.match, emptyMessage: "" }),
}));

import { ProjectMembersTab } from "../components/ProjectMembersTab";

describe("<ProjectMembersTab />", () => {
  it("is read-only until Edit, then locks the PI and offers only Co-PI, Allocation Manager and Member", () => {
    render(<ProjectMembersTab projectId="p-1" canEditRoles canEditAccess canSearchUsers />);
    expect(screen.queryByLabelText("Role of Cira Member")).not.toBeInTheDocument();
    expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Edit" }));
    expect(screen.queryByLabelText("Role of Ada Pi")).not.toBeInTheDocument();
    const select = screen.getByLabelText("Role of Cira Member");
    const options = within(select)
      .getAllByRole("option")
      .map((o) => o.textContent);
    expect(options).toEqual(["Co-PI", "Allocation Manager", "Member"]);
  });

  it("stages edits, highlights them, and applies them only after confirmation", async () => {
    render(<ProjectMembersTab projectId="p-1" canEditRoles canEditAccess canSearchUsers />);
    fireEvent.click(screen.getByRole("button", { name: "Edit" }));
    fireEvent.click(screen.getByLabelText("Cira Member in Beta"));
    fireEvent.click(screen.getByLabelText("Cira Member in Alpha"));
    fireEvent.change(screen.getByLabelText("Role of Cira Member"), {
      target: { value: "ALLOCATION_MANAGER" },
    });
    expect(screen.getByLabelText("Cira Member in Beta").closest("td")?.className).toMatch(/amber/);
    expect(toggle.mutateAsync).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await screen.findByText("Cira Member: add to Beta");
    expect(screen.getByText("Cira Member: remove from Alpha")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Apply changes" }));

    await waitFor(() => expect(toggle.mutateAsync).toHaveBeenCalledTimes(2));
    expect(roleMutation.mutateAsync).toHaveBeenCalledWith({
      userId: "u-3",
      payload: { role: "ALLOCATION_MANAGER" },
    });
    expect(toggle.mutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ userId: "u-3", membership: ciraInAlpha }),
    );
    expect(toggle.mutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ userId: "u-3", membership: undefined }),
    );
  });

  it("enables Save as soon as a new member is added", () => {
    render(<ProjectMembersTab projectId="p-1" canEditRoles canEditAccess canSearchUsers />);
    fireEvent.click(screen.getByRole("button", { name: "Edit" }));
    expect(screen.getByRole("button", { name: "Save" })).toBeDisabled();
    search.match = {
      id: "u-9",
      email: "new@sample.example.edu",
      first_name: "New",
      last_name: "Person",
    };
    fireEvent.click(screen.getByRole("button", { name: "Add member" }));
    fireEvent.change(screen.getByLabelText("Role of new member"), { target: { value: "MEMBER" } });
    fireEvent.click(screen.getByRole("button", { name: "Confirm" }));
    expect(screen.getByText("New Person")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save" })).toBeEnabled();
  });

  it("keeps Edit visible but disabled without either write privilege", () => {
    render(
      <ProjectMembersTab
        projectId="p-1"
        canEditRoles={false}
        canEditAccess={false}
        canSearchUsers={true}
      />,
    );
    expect(screen.getByRole("button", { name: "Edit" })).toBeDisabled();
  });

  it("locks role pickers without projects write while access stays editable", () => {
    render(<ProjectMembersTab projectId="p-1" canEditRoles={false} canEditAccess canSearchUsers />);
    fireEvent.click(screen.getByRole("button", { name: "Edit" }));
    expect(screen.getByLabelText("Role of Cira Member")).toBeDisabled();
    expect(screen.getByLabelText("Cira Member in Beta")).toBeEnabled();
  });
});
