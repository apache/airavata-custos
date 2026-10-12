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

import type { ProjectMemberResponse } from "@/generated/core/types.gen";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

function member(over: Partial<ProjectMemberResponse>): ProjectMemberResponse {
  return {
    id: "u-x",
    project_id: "p-1",
    user_id: "u-x",
    added_time: "2026-01-01T00:00:00Z",
    email: "x@sample.example.edu",
    display_name: "Member X",
    role: "MEMBER",
    status: "ACTIVE",
    allocations: [],
    ...over,
  };
}

const PI_ID = "5d3c2b1a-0000-4000-8000-0000000000a1";
const members = [
  member({ user_id: PI_ID, display_name: "Ada Pi", role: "PI" }),
  member({
    user_id: "u-2",
    display_name: "Ben Copi",
    role: "CO_PI",
    allocations: [{ id: "a-1", name: "Alloc One", role: "CO_PI" }],
  }),
  member({ user_id: "u-3", display_name: "Cira Member", role: "MEMBER" }),
  member({ user_id: "u-4", display_name: "Dee Manager", role: "ALLOCATION_MANAGER" }),
  member({
    user_id: "u-5",
    display_name: "eve@sample.example.edu",
    email: "eve@sample.example.edu",
  }),
];

const mutation = { mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false };
const toastError = vi.hoisted(() => vi.fn());
const confirm = vi.hoisted(() => vi.fn());
vi.mock("@/shared/ui/sonner", async (orig) => ({
  ...(await orig<typeof import("@/shared/ui/sonner")>()),
  confirmToast: confirm,
}));
vi.mock("sonner", async (orig) => ({
  ...(await orig<typeof import("sonner")>()),
  toast: { error: toastError, success: vi.fn(), dismiss: vi.fn() },
}));
vi.mock("../queries", () => ({
  useProjectMembers: () => ({
    data: members,
    isLoading: false,
    error: null,
    refetch: vi.fn(),
    dataUpdatedAt: Number.MAX_SAFE_INTEGER,
  }),
  useSetProjectRole: () => mutation,
}));

import { ProjectMembersTab } from "../components/ProjectMembersTab";

describe("<ProjectMembersTab />", () => {
  it("offers role actions only for Co-PI and Allocation manager tags, limited to those roles", () => {
    render(<ProjectMembersTab projectId="p-1" canManage={true} canReadAllocations={true} />);
    for (const name of ["Ada Pi", "Cira Member"]) {
      expect(screen.queryByRole("button", { name: `Edit role of ${name}` })).toBeNull();
      expect(screen.queryByRole("button", { name: `Remove role of ${name}` })).toBeNull();
    }
    for (const name of ["Ben Copi", "Dee Manager"]) {
      expect(screen.getByRole("button", { name: `Remove role of ${name}` })).toBeInTheDocument();
    }
    fireEvent.click(screen.getByRole("button", { name: "Edit role of Dee Manager" }));
    const options = within(screen.getByLabelText("Role"))
      .getAllByRole("option")
      .map((o) => o.textContent);
    expect(options).toEqual(["Co-PI", "Allocation manager"]);
  });

  // An untouched Save would write the opened role over another tab's change.
  it("closes an unchanged role edit without a request", () => {
    render(<ProjectMembersTab projectId="p-1" canManage={true} canReadAllocations={true} />);
    fireEvent.click(screen.getByRole("button", { name: "Edit role of Dee Manager" }));
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(mutation.mutateAsync).not.toHaveBeenCalled();
  });

  it("says so, with Save disabled, when another tab removed the role being edited", () => {
    const { rerender } = render(
      <ProjectMembersTab projectId="p-1" canManage={true} canReadAllocations={true} />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Edit role of Dee Manager" }));
    members[3] = { ...(members[3] as ProjectMemberResponse), role: "MEMBER" };
    rerender(<ProjectMembersTab projectId="p-1" canManage={true} canReadAllocations={true} />);
    expect(screen.getByText("Dee Manager no longer holds a role on this project")).toBeVisible();
    expect(screen.getByRole("button", { name: "Save" })).toHaveAttribute("aria-disabled", "true");
    members[3] = { ...(members[3] as ProjectMemberResponse), role: "ALLOCATION_MANAGER" };
  });

  it("confirms before removing a role", () => {
    render(<ProjectMembersTab projectId="p-1" canManage={true} canReadAllocations={true} />);
    fireEvent.click(screen.getByRole("button", { name: "Remove role of Ben Copi" }));
    expect(confirm).toHaveBeenCalledWith(
      "Remove Ben Copi's role on this project?",
      "Remove",
      expect.any(Function),
    );
    expect(mutation.mutateAsync).not.toHaveBeenCalled();
  });

  it("refuses a role for the PI, who already holds every role, without calling the API", () => {
    render(<ProjectMembersTab projectId="p-1" canManage={true} canReadAllocations={true} />);
    fireEvent.click(screen.getByRole("button", { name: "+ Assign role" }));
    fireEvent.change(screen.getByLabelText(/User ID/), { target: { value: PI_ID } });
    fireEvent.change(screen.getByLabelText("Role"), { target: { value: "CO_PI" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(toastError).toHaveBeenCalledWith("The PI already holds every project role.");
    expect(mutation.mutateAsync).not.toHaveBeenCalled();
  });

  it("assigns a role to a pasted uppercase user id as the lowercase id", () => {
    mutation.mutateAsync.mockResolvedValueOnce({});
    render(<ProjectMembersTab projectId="p-1" canManage={true} canReadAllocations={true} />);
    fireEvent.click(screen.getByRole("button", { name: "+ Assign role" }));
    fireEvent.change(screen.getByLabelText(/User ID/), {
      target: { value: " 0B9F4C1E-8D2A-4F3B-9C7D-1A2B3C4D5E6F " },
    });
    fireEvent.change(screen.getByLabelText("Role"), { target: { value: "CO_PI" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(mutation.mutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        path: { id: "p-1", userId: "0b9f4c1e-8d2a-4f3b-9c7d-1a2b3c4d5e6f" },
      }),
    );
  });

  it("assigns a role to a member from their row, without a user ID field", () => {
    mutation.mutateAsync.mockResolvedValueOnce({});
    render(<ProjectMembersTab projectId="p-1" canManage={true} canReadAllocations={true} />);
    fireEvent.click(screen.getByRole("button", { name: "Assign a role to Cira Member" }));
    expect(screen.queryByLabelText(/User ID/)).toBeNull();
    fireEvent.change(screen.getByLabelText("Role"), { target: { value: "CO_PI" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(mutation.mutateAsync).toHaveBeenCalledWith({
      path: { id: "p-1", userId: "u-3" },
      body: { role: "CO_PI" },
    });
  });

  it("shows a member's email once when their name falls back to it", () => {
    render(<ProjectMembersTab projectId="p-1" canManage={false} canReadAllocations={true} />);
    expect(screen.getAllByText("eve@sample.example.edu")).toHaveLength(1);
  });

  it("links allocation chips only for callers who can read every allocation", () => {
    const { rerender } = render(
      <ProjectMembersTab projectId="p-1" canManage={false} canReadAllocations={true} />,
    );
    expect(screen.getByRole("link", { name: /Alloc One/ })).toHaveAttribute(
      "href",
      "/allocations/a-1",
    );
    rerender(<ProjectMembersTab projectId="p-1" canManage={false} canReadAllocations={false} />);
    expect(screen.queryByRole("link", { name: /Alloc One/ })).toBeNull();
    expect(screen.getByText("Alloc One")).toBeInTheDocument();
  });
});
