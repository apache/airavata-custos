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
import { describe, expect, it, vi } from "vitest";
import type { ComputeAllocation } from "../schemas";

const members = [
  {
    id: "m-1",
    compute_allocation_id: "alloc-001",
    user_id: "u-1",
    display_name: "Paula Ivey",
    email: "paula@sample.example.edu",
    role: "PI",
    membership_status: "ACTIVE",
  },
  {
    id: "m-2",
    compute_allocation_id: "alloc-001",
    user_id: "u-2",
    display_name: "Carl Osei",
    email: "carl@sample.example.edu",
    role: "CO_PI",
    membership_status: "ACTIVE",
  },
  {
    id: "m-3",
    compute_allocation_id: "alloc-001",
    user_id: "u-3",
    display_name: "Mina Frey",
    email: "mina@sample.example.edu",
    role: "MEMBER",
    membership_status: "ACTIVE",
  },
];

const mutation = { mutate: vi.fn(), isPending: false };
const roleMutation = { mutate: vi.fn(), isPending: false };
vi.mock("../queries", () => ({
  useAllocationMembers: () => ({ data: members, isLoading: false, error: null, refetch: vi.fn() }),
  useAddMember: () => mutation,
  useSetMemberProjectRole: () => roleMutation,
  useRemoveMember: () => mutation,
}));

import { AllocationMembersTab } from "../components/AllocationMembersTab";

const allocation = {
  id: "alloc-001",
  name: "Test Allocation",
  project_id: "proj-1",
} as ComputeAllocation;

describe("<AllocationMembersTab />", () => {
  it("offers no Edit or Remove for PI or Co-PI members", () => {
    render(<AllocationMembersTab allocation={allocation} canManage={true} />);
    for (const name of ["Paula Ivey", "Carl Osei"]) {
      expect(screen.queryByRole("button", { name: `Edit ${name}` })).not.toBeInTheDocument();
      expect(screen.queryByRole("button", { name: `Remove ${name}` })).not.toBeInTheDocument();
    }
    expect(screen.getByRole("button", { name: "Edit Mina Frey" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Remove Mina Frey" })).toBeInTheDocument();
  });

  it("offers no actions without manage permission", () => {
    render(<AllocationMembersTab allocation={allocation} canManage={false} />);
    expect(screen.queryByRole("button", { name: /^Remove / })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /^Edit / })).not.toBeInTheDocument();
  });

  it("writes the role to the parent project keyed on user id", () => {
    render(<AllocationMembersTab allocation={allocation} canManage={true} />);
    fireEvent.click(screen.getByRole("button", { name: "Edit Mina Frey" }));
    fireEvent.change(screen.getByLabelText("Role"), { target: { value: "ALLOCATION_MANAGER" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(roleMutation.mutate).toHaveBeenCalledWith(
      { projectId: "proj-1", userId: "u-3", role: "ALLOCATION_MANAGER" },
      expect.anything(),
    );
  });
});
