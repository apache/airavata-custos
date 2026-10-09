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

const members = [
  member({ user_id: "u-1", display_name: "Ada Pi", role: "PI" }),
  member({ user_id: "u-2", display_name: "Ben Copi", role: "CO_PI" }),
  member({ user_id: "u-3", display_name: "Cira Member", role: "MEMBER" }),
  member({ user_id: "u-4", display_name: "Dee Manager", role: "ALLOCATION_MANAGER" }),
];

const mutation = { mutate: vi.fn(), isPending: false };
vi.mock("../queries", () => ({
  useProjectMembers: () => ({ data: members, isLoading: false, error: null, refetch: vi.fn() }),
  useSetProjectRole: () => mutation,
}));

import { ProjectMembersTab } from "../components/ProjectMembersTab";

describe("<ProjectMembersTab />", () => {
  it("offers role actions only for Co-PI and Allocation Manager tags, limited to those roles", () => {
    render(<ProjectMembersTab projectId="p-1" canManage={true} />);
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
    expect(options).toEqual(["Co-PI", "Allocation Manager"]);
  });
});
