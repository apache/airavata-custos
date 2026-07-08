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

import { render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { AllocationDiff, ComputeAllocation } from "../schemas";

let diffs: AllocationDiff[] = [];

vi.mock("../queries", () => ({
  useAllocationDiffs: () => ({ data: diffs, isLoading: false, error: null, refetch: vi.fn() }),
}));

import { AllocationHistoryTab } from "../components/AllocationHistoryTab";

const allocation: ComputeAllocation = {
  id: "alloc-001",
  project_id: "project-001",
  name: "Genomic GPU Pool",
  status: "ACTIVE",
  compute_cluster_id: "cluster-a",
  initial_su_amount: 250000,
  start_time: "2026-04-01T00:00:00.000Z",
  end_time: "2027-03-31T00:00:00.000Z",
};

beforeEach(() => {
  diffs = [
    {
      id: "diff-a",
      compute_allocation_id: "alloc-001",
      diff_type: "ALLOCATION_STATUS_CHANGE",
      new_su_amount: 250000,
      status: "ACTIVE",
      timestamp: "2026-04-01T00:00:00.000Z",
      description: "Activated for the funding period.",
    },
    {
      id: "diff-b",
      compute_allocation_id: "alloc-001",
      diff_type: "USAGE_UPDATE",
      new_su_amount: 250000,
      status: "ACTIVE",
      timestamp: "2026-06-20T14:05:00.000Z",
    },
    {
      id: "diff-c",
      compute_allocation_id: "alloc-001",
      diff_type: "RESOURCE_ATTACHED",
      status: "INACTIVE",
      timestamp: "2026-05-15T08:30:00.000Z",
    },
  ];
});

describe("<AllocationHistoryTab />", () => {
  it("lists diffs newest first", () => {
    render(<AllocationHistoryTab allocation={allocation} />);
    const rows = screen.getAllByRole("row").slice(1); // drop the header row
    expect(within(rows[0] as HTMLElement).getByText("USAGE_UPDATE")).toBeInTheDocument();
    expect(within(rows[1] as HTMLElement).getByText("RESOURCE_ATTACHED")).toBeInTheDocument();
    expect(within(rows[2] as HTMLElement).getByText("ALLOCATION_STATUS_CHANGE")).toBeInTheDocument();
  });

  it("renders the status badge and a placeholder for a missing SU amount", () => {
    render(<AllocationHistoryTab allocation={allocation} />);
    expect(screen.getByText("INACTIVE")).toBeInTheDocument();
    const resourceRow = screen.getByText("RESOURCE_ATTACHED").closest("tr");
    expect(within(resourceRow as HTMLElement).getByText("-")).toBeInTheDocument();
  });

  it("shows the empty state when there is no history", () => {
    diffs = [];
    render(<AllocationHistoryTab allocation={allocation} />);
    expect(
      screen.getByRole("heading", { name: /no history recorded for this allocation/i }),
    ).toBeInTheDocument();
  });
});
