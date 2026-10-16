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

import type { ComputeAllocation } from "@/generated/core/types.gen";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const members = [
  {
    id: "m-1",
    compute_allocation_id: "01000001-0000-4000-8000-000001000001",
    user_id: "u-1",
    display_name: "Paula Ivey",
    email: "paula@sample.example.edu",
    role: "PI",
    membership_status: "ACTIVE",
  },
  {
    id: "m-2",
    compute_allocation_id: "01000001-0000-4000-8000-000001000001",
    user_id: "u-2",
    display_name: "Carl Osei",
    email: "carl@sample.example.edu",
    role: "CO_PI",
    membership_status: "ACTIVE",
    start_time: "2001-01-01T00:00:00Z",
    end_time: "2001-06-30T00:00:00Z",
  },
  {
    id: "m-3",
    compute_allocation_id: "01000001-0000-4000-8000-000001000001",
    user_id: "u-3",
    display_name: "Mina Frey",
    email: "mina@sample.example.edu",
    role: "MEMBER",
    membership_status: "ACTIVE",
    start_time: "2099-01-01T00:00:00Z",
    end_time: "2100-01-01T00:00:00Z",
  },
];

const mutation = { mutate: vi.fn(), mutateAsync: vi.fn(async () => ({})), isPending: false };
vi.mock("../queries", () => ({
  useMembership: () => ({}),
  useRefreshAllocationMembers: () => vi.fn(),
  useUpdateMembership: () => mutation,
  useSetMembershipStatus: () => mutation,
  useAllocationMembers: () => ({ data: members, isLoading: false, error: null, refetch: vi.fn() }),
  useAddMember: () => mutation,
  useRemoveMember: () => mutation,
}));

vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => false }) }));

import { AllocationMembersTab } from "../components/AllocationMembersTab";

const allocation: ComputeAllocation = {
  id: "01000001-0000-4000-8000-000001000001",
  project_id: "19000001-0000-4000-8000-000019000001",
  name: "Test Allocation",
  status: "ACTIVE",
  compute_cluster_id: "04000001-0000-4000-8000-000004000001",
  initial_su_amount: 250000,
  start_time: "2026-04-01T00:00:00.000Z",
  end_time: "2027-03-31T00:00:00.000Z",
  posix_group: null,
};

describe("<AllocationMembersTab />", () => {
  it("offers Remove for every member, whatever their project role", () => {
    render(<AllocationMembersTab allocation={allocation} canManage={true} canRead />);
    for (const name of ["Paula Ivey", "Carl Osei", "Mina Frey"]) {
      expect(screen.getByRole("button", { name: `Remove ${name}` })).toBeInTheDocument();
    }
  });

  it("dates each membership and notes one that has ended or not yet started", () => {
    render(<AllocationMembersTab allocation={allocation} canManage={false} canRead />);
    const row = (name: string) => screen.getByText(name).closest("tr") as HTMLElement;
    expect(within(row("Carl Osei")).getByText("Jun 30, 2001")).toBeInTheDocument();
    expect(within(row("Carl Osei")).getByText("ended Jun 30, 2001")).toBeInTheDocument();
    expect(within(row("Mina Frey")).getByText("starts Jan 1, 2099")).toBeInTheDocument();
    expect(within(row("Paula Ivey")).queryByText(/^(ended|starts) /)).not.toBeInTheDocument();
  });

  it("offers no actions without manage permission", () => {
    render(<AllocationMembersTab allocation={allocation} canManage={false} canRead />);
    expect(screen.queryByRole("button", { name: /^Remove / })).not.toBeInTheDocument();
  });

  it("starts a new member of an ended allocation on its last day", async () => {
    const ended = { ...allocation, end_time: "2001-06-30T00:00:00Z" };
    render(<AllocationMembersTab allocation={ended} canManage canRead />);
    fireEvent.click(screen.getByRole("button", { name: "+ Add member" }));
    fireEvent.change(screen.getByLabelText(/User ID/), {
      target: { value: "0B7C2F1E-4A6D-4E8B-9C3A-5F2D1E0A9B7C" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Add member" }));
    await waitFor(() =>
      expect(mutation.mutateAsync).toHaveBeenCalledWith({
        body: {
          compute_allocation_id: "01000001-0000-4000-8000-000001000001",
          user_id: "0b7c2f1e-4a6d-4e8b-9c3a-5f2d1e0a9b7c",
          start_time: "2001-06-30T00:00:00Z",
          end_time: "2001-06-30T00:00:00Z",
        },
      }),
    );
  });
});
