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

import type { ComputeAllocation, ComputeAllocationUsage } from "@/generated/core/types.gen";
import { render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

let usage: ComputeAllocationUsage[] = [];
let members: Array<{ user_id: string; display_name?: string }> = [];

vi.mock("@/features/core/resources/queries", () => ({
  useResources: () => ({
    data: [
      { id: "21000002-0000-4000-8000-000021000002", name: "ClusterA GPU" },
      { id: "21000001-0000-4000-8000-000021000001", name: "ClusterA CPU" },
    ],
  }),
}));
vi.mock("@/features/core/analytics/queries", () => ({ useUsageSummary: () => ({}) }));
vi.mock("../queries", () => ({
  useUsageRecord: () => ({}),
  useDeleteUsage: () => ({}),
  useAllocationUsage: () => ({ data: usage, isLoading: false, error: null, refetch: vi.fn() }),
  useUsageTotal: () => ({ data: usage.reduce((acc, r) => acc + r.used_su_amount, 0) }),
  useAllocationMembers: () => ({ data: members }),
  useAllocationResources: () => ({
    data: [{ compute_allocation_resource_id: "21000002-0000-4000-8000-000021000002" }],
  }),
  useRecordUsage: () => ({ isPending: false }),
  useUserUsageTotal: () => ({ data: 0 }),
}));

vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => false }) }));

import { AllocationUsageTab } from "../components/AllocationUsageTab";

const allocation: ComputeAllocation = {
  id: "01000003-0000-4000-8000-000001000003",
  project_id: "19000002-0000-4000-8000-000019000002",
  name: "Plasma GPU",
  status: "ACTIVE",
  compute_cluster_id: "04000001-0000-4000-8000-000004000001",
  initial_su_amount: 250000,
  start_time: "2026-04-12T00:00:00.000Z",
  end_time: "2027-04-11T00:00:00.000Z",
  posix_group: null,
};

function usageRow(overrides: Partial<ComputeAllocationUsage>): ComputeAllocationUsage {
  return {
    id: "23000000-0000-4000-8000-000023000000",
    compute_allocation_id: "01000003-0000-4000-8000-000001000003",
    used_raw_amount: 0,
    used_su_amount: 0,
    last_updated: "2026-05-18T11:00:00.000Z",
    user_id: "24000021-0000-4000-8000-000024000021",
    job_id: "11000002-0000-4000-8000-000011000002",
    compute_allocation_resource_id: "21000002-0000-4000-8000-000021000002",
    ...overrides,
  };
}

beforeEach(() => {
  members = [];
  usage = [
    usageRow({
      id: "23000001-0000-4000-8000-000023000001",
      used_su_amount: 60000,
      compute_allocation_resource_id: "21000002-0000-4000-8000-000021000002",
    }),
    usageRow({
      id: "23000002-0000-4000-8000-000023000002",
      used_su_amount: 20000,
      compute_allocation_resource_id: "21000002-0000-4000-8000-000021000002",
    }),
    usageRow({
      id: "23000003-0000-4000-8000-000023000003",
      used_su_amount: 24000,
      compute_allocation_resource_id: "21000001-0000-4000-8000-000021000001",
    }),
  ];
});

describe("<AllocationUsageTab />", () => {
  it("renders a per-resource bar with summed SU, one per resource, named from the catalog", () => {
    render(<AllocationUsageTab allocation={allocation} canRead canManage={false} />);
    expect(screen.getByLabelText("ClusterA GPU share of usage")).toBeInTheDocument();
    expect(screen.getByText("80,000 SU")).toBeInTheDocument();
    expect(screen.getByLabelText("ClusterA CPU share of usage")).toBeInTheDocument();
    expect(screen.getByText("24,000 SU")).toBeInTheDocument();
  });

  it("measures total usage against the initial SU amount", () => {
    render(<AllocationUsageTab allocation={allocation} canRead canManage={false} />);
    expect(screen.getByText("104,000 / 250,000 initial SUs")).toBeInTheDocument();
  });

  it("names each record's user from the member list", () => {
    members = [
      { user_id: "24000021-0000-4000-8000-000024000021", display_name: "Pat Investigator" },
    ];
    render(<AllocationUsageTab allocation={allocation} canRead canManage={false} />);
    const records = screen.getByRole("table", { name: "Usage records" });
    expect(within(records).getAllByText("Pat Investigator")).toHaveLength(3);
  });

  it("disables recording usage, with the reason, on an allocation without members", () => {
    render(<AllocationUsageTab allocation={allocation} canRead canManage />);
    expect(screen.getByRole("button", { name: "+ Record usage" })).toHaveAttribute(
      "aria-disabled",
      "true",
    );
    expect(
      screen.getByText("Add a member to this allocation before recording usage."),
    ).toBeInTheDocument();
  });

  it("sums usage by users who are not members into an Other row", () => {
    members = [{ user_id: "24000021-0000-4000-8000-000024000021" }];
    usage = [
      ...usage,
      usageRow({
        id: "23000004-0000-4000-8000-000023000004",
        used_su_amount: 1500,
        user_id: "24000099-0000-4000-8000-000024000099",
      }),
    ];
    render(<AllocationUsageTab allocation={allocation} canRead canManage={false} />);
    expect(screen.getByText("Other (not a member)").nextSibling).toHaveTextContent("1,500 SU");
  });

  it("shows the empty state when there is no usage", () => {
    usage = [];
    render(<AllocationUsageTab allocation={allocation} canRead canManage={false} />);
    expect(screen.getByRole("heading", { name: /no usage recorded/i })).toBeInTheDocument();
  });
});
