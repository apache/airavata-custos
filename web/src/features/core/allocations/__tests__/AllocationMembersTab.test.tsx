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

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { setupServer } from "msw/node";
import type * as React from "react";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";
import type { ComputeAllocation } from "@/generated/core/zod.gen";
import { slurmConflicts, slurmHandlers } from "@/mocks/handlers/slurm";

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

vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => true }) }));

const server = setupServer(...slurmHandlers);
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const renderTab = (ui: React.ReactElement) =>
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      {ui}
    </QueryClientProvider>,
  );

const mutation = { mutate: vi.fn(), isPending: false };
vi.mock("@/features/core/users/components/UserDetailsDrawer", () => ({
  UserDetailsDrawer: () => null,
}));
vi.mock("../queries", () => ({
  useAllocationMembers: () => ({ data: members, isLoading: false, error: null, refetch: vi.fn() }),
  useAddMember: () => mutation,
  useSetMemberStatus: () => mutation,
}));
vi.mock("@/features/core/projects/queries", () => ({
  useUpdateProjectMember: () => mutation,
}));
vi.mock("@/features/core/users/queries", () => ({
  useUserEmailSearch: () => ({ suggestions: [], match: undefined, emptyMessage: "" }),
}));

import { AllocationMembersTab } from "../components/AllocationMembersTab";

const allocation: ComputeAllocation = {
  id: "alloc-001",
  project_id: "project-001",
  name: "Test Allocation",
  status: "ACTIVE",
  compute_cluster_id: "cluster-001",
  initial_su_amount: 100000,
  start_time: "2026-04-01T00:00:00Z",
  end_time: "2027-03-31T00:00:00Z",
  posix_group: null,
  group_provisioned_at: null,
};

describe("<AllocationMembersTab />", () => {
  it("offers no Edit or Deactivate for PI or Co-PI members", () => {
    renderTab(
      <AllocationMembersTab allocation={allocation} canEditAccess canEditRoles canSearchUsers />,
    );
    for (const name of ["Paula Ivey", "Carl Osei"]) {
      expect(screen.queryByRole("button", { name: `Edit ${name}` })).not.toBeInTheDocument();
      expect(screen.queryByRole("button", { name: `Deactivate ${name}` })).not.toBeInTheDocument();
    }
    expect(screen.getByRole("button", { name: "Edit Mina Frey" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Deactivate Mina Frey" })).toBeInTheDocument();
  });

  it("disables each action whose privilege is missing", () => {
    renderTab(
      <AllocationMembersTab
        allocation={allocation}
        canEditAccess={false}
        canEditRoles
        canSearchUsers
      />,
    );
    expect(screen.getByRole("button", { name: "Deactivate Mina Frey" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Edit Mina Frey" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "+ Add member" })).toBeDisabled();
  });

  it("lists this allocation's conflicts, marking those without a membership", async () => {
    server.use(
      slurmConflicts([
        {
          key: "member:bio1:ghost:access",
          name: "access",
          label: "Access",
          user: "ghost",
          kind: "access",
          allocation_id: "alloc-001",
          custos: null,
          slurm: "access",
          choices: ["custos"],
        },
        {
          key: "member:bio9:ghost2:access",
          name: "access",
          label: "Access",
          user: "ghost2",
          kind: "access",
          allocation_id: "other",
          custos: null,
          slurm: "access",
          choices: ["custos"],
        },
      ]),
    );
    renderTab(
      <AllocationMembersTab allocation={allocation} canEditAccess canEditRoles canSearchUsers />,
    );
    expect(await screen.findByText("ghost (not in Custos) · Access")).toBeInTheDocument();
    expect(screen.queryByText(/ghost2/)).not.toBeInTheDocument();
  });
});
