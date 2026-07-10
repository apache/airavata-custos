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

import { fireEvent, render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

function member(over: Record<string, unknown>) {
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
  member({ id: "pm-1", display_name: "Ada Pi", role: "PI" }),
  member({ id: "pm-2", display_name: "Ben Copi", role: "CO_PI" }),
  member({ id: "pm-3", display_name: "Cira Member", role: "MEMBER" }),
];

const mutation = { mutate: vi.fn(), isPending: false };
vi.mock("../queries", () => ({
  useProjectMembers: () => ({ data: members, isLoading: false, error: null, refetch: vi.fn() }),
  useUpdateProjectMember: () => mutation,
  useRemoveProjectMember: () => mutation,
}));

import { ProjectMembersTab } from "../components/ProjectMembersTab";

describe("<ProjectMembersTab />", () => {
  it("offers no Edit or Remove for PI or Co-PI members", () => {
    render(<ProjectMembersTab projectId="p-1" canManage={true} />);
    for (const name of ["Ada Pi", "Ben Copi"]) {
      expect(screen.queryByRole("button", { name: `Edit ${name}` })).not.toBeInTheDocument();
      expect(screen.queryByRole("button", { name: `Remove ${name}` })).not.toBeInTheDocument();
    }
    expect(screen.getByRole("button", { name: "Edit Cira Member" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Remove Cira Member" })).toBeInTheDocument();
  });

  it("limits the edit role picker to Member and Allocation Manager", () => {
    render(<ProjectMembersTab projectId="p-1" canManage={true} />);
    fireEvent.click(screen.getByRole("button", { name: "Edit Cira Member" }));
    const select = screen.getByLabelText("Role");
    const options = within(select).getAllByRole("option").map((o) => o.textContent);
    expect(options).toEqual(["Allocation Manager", "Member"]);
  });
});
