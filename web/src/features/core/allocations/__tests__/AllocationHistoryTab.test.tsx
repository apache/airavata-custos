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

import type { ComputeAllocation, ComputeAllocationDiff } from "@/generated/core/types.gen";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

let diffs: ComputeAllocationDiff[] = [];

vi.mock("../queries", () => ({
  useDiff: () => ({}),
  useDeleteDiff: () => ({}),
  useRecordDiff: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useAllocationDiffs: () => ({ data: diffs, isLoading: false, error: null, refetch: vi.fn() }),
}));

vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => true }) }));

import { AllocationHistoryTab } from "../components/AllocationHistoryTab";

const allocation: ComputeAllocation = {
  id: "01000001-0000-4000-8000-000001000001",
  project_id: "19000001-0000-4000-8000-000019000001",
  name: "Genomic GPU Pool",
  status: "ACTIVE",
  compute_cluster_id: "04000001-0000-4000-8000-000004000001",
  initial_su_amount: 250000,
  start_time: "2026-04-01T00:00:00.000Z",
  end_time: "2027-03-31T00:00:00.000Z",
  posix_group: null,
};

beforeEach(() => {
  diffs = [
    {
      id: "diff-a",
      compute_allocation_id: "01000001-0000-4000-8000-000001000001",
      diff_type: "ALLOCATION_STATUS_CHANGE",
      new_su_amount: 250000,
      status: "ACTIVE",
      timestamp: "2026-04-01T00:00:00.000Z",
      description: "Activated for the funding period.",
    },
    {
      id: "diff-b",
      compute_allocation_id: "01000001-0000-4000-8000-000001000001",
      diff_type: "USAGE_UPDATE",
      new_su_amount: 300000,
      status: "ACTIVE",
      timestamp: "2026-06-20T14:05:00.000Z",
    },
    {
      id: "diff-c",
      compute_allocation_id: "01000001-0000-4000-8000-000001000001",
      diff_type: "RESOURCE_ATTACHED",
      new_su_amount: 300000,
      status: "INACTIVE",
      timestamp: "2026-05-15T08:30:00.000Z",
    },
  ];
});

describe("<AllocationHistoryTab />", () => {
  it("lists diffs newest first by label", () => {
    render(<AllocationHistoryTab allocation={allocation} canManage={false} />);
    const rows = screen.getAllByRole("row").slice(1); // drop the header row
    expect(within(rows[0] as HTMLElement).getByText("Usage updated")).toBeInTheDocument();
    expect(within(rows[0] as HTMLElement).getByText("300,000")).toBeInTheDocument();
    expect(within(rows[1] as HTMLElement).getByText("Resource attached")).toBeInTheDocument();
    expect(within(rows[2] as HTMLElement).getByText("Status changed")).toBeInTheDocument();
  });

  it("starts a new diff from the latest recorded SU total", () => {
    render(<AllocationHistoryTab allocation={allocation} canManage />);
    fireEvent.click(screen.getByRole("button", { name: "+ Record diff" }));
    expect(screen.getByLabelText("New SU total")).toHaveValue(300000);
  });
});
