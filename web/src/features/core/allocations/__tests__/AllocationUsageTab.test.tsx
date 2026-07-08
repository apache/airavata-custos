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

import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { AllocationUsage, ComputeAllocation } from "../schemas";

let usage: AllocationUsage[] = [];

vi.mock("../queries", () => ({
  useAllocationUsage: () => ({ data: usage, isLoading: false, error: null, refetch: vi.fn() }),
}));

vi.mock("@/features/core/resources/queries", () => ({
  useResourceSummaries: () => ({
    data: [
      { id: "res-a-gpu", name: "ClusterA GPU" },
      { id: "res-c-cpu", name: "ClusterC RM" },
    ],
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  }),
}));

import { AllocationUsageTab } from "../components/AllocationUsageTab";

const allocation: ComputeAllocation = {
  id: "alloc-003",
  project_id: "project-002",
  name: "Plasma GPU",
  status: "ACTIVE",
  compute_cluster_id: "cluster-a",
  initial_su_amount: 250000,
  start_time: "2026-04-12T00:00:00.000Z",
  end_time: "2027-04-11T00:00:00.000Z",
};

function usageRow(overrides: Partial<AllocationUsage>): AllocationUsage {
  return {
    id: "use-x",
    compute_allocation_id: "alloc-003",
    used_raw_amount: 0,
    used_su_amount: 0,
    last_updated: "2026-05-18T11:00:00.000Z",
    user_id: "user-pi-002",
    job_id: "job-1",
    compute_allocation_resource_id: "res-a-gpu",
    ...overrides,
  };
}

beforeEach(() => {
  usage = [
    usageRow({ id: "use-001", used_su_amount: 60000, compute_allocation_resource_id: "res-a-gpu" }),
    usageRow({ id: "use-002", used_su_amount: 20000, compute_allocation_resource_id: "res-a-gpu" }),
    usageRow({ id: "use-003", used_su_amount: 24000, compute_allocation_resource_id: "res-c-cpu" }),
  ];
});

describe("<AllocationUsageTab />", () => {
  it("renders a per-resource bar with summed SU, one per resource", () => {
    render(<AllocationUsageTab allocation={allocation} />);
    expect(screen.getByText("ClusterA GPU")).toBeInTheDocument();
    expect(screen.getByText("80,000 SU")).toBeInTheDocument();
    expect(screen.getByText("ClusterC RM")).toBeInTheDocument();
    expect(screen.getByText("24,000 SU")).toBeInTheDocument();
  });

  it("keeps the total usage bar", () => {
    render(<AllocationUsageTab allocation={allocation} />);
    expect(screen.getByText("104,000 / 250,000 SUs")).toBeInTheDocument();
  });

  it("shows the empty state when there is no usage", () => {
    usage = [];
    render(<AllocationUsageTab allocation={allocation} />);
    expect(screen.getByRole("heading", { name: /no usage recorded/i })).toBeInTheDocument();
  });
});
