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
import type { ComputeAllocation, ComputeAllocationResource } from "@/generated/core/types.gen";

let attached: ComputeAllocationResource[] = [];
const mutation = { mutate: vi.fn(), isPending: false };

vi.mock("../queries", () => ({
  useAllocationResourceMapping: () => ({ data: { resource_amount: 4, resource_time: 600 }, isLoading: false }),
  useAllocationResources: () => ({
    data: attached,
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  }),
  useAttachResource: () => mutation,
  useUpdateResourceMapping: () => mutation,
  useDetachResource: () => mutation,
}));

vi.mock("@/features/core/resources/queries", () => ({
  useResources: () => ({ data: [] }),
  // res-a-gpu has an effective rate; res-b-gpu 404s (query error, no data).
  useEffectiveRate: (id: string) =>
    id === "res-a-gpu"
      ? { data: { rate: 0.5 }, isError: false }
      : { data: undefined, isError: true },
}));

vi.mock("@/features/core/projects/queries", () => ({ useProject: () => ({ data: undefined }) }));
vi.mock("@/features/core/clusters/queries", () => ({ useCluster: () => ({ data: undefined }) }));

import { AllocationOverviewTab } from "../components/AllocationOverviewTab";

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
  attached = [
    {
      id: "res-a-gpu",
      name: "ClusterA GPU",
      resource_type: "GPU",
      resource_amount: 50000,
      compute_cluster_id: "cluster-a",
    },
    {
      id: "res-b-gpu",
      name: "ClusterB GPU A100",
      resource_type: "GPU",
      resource_amount: 16000,
      compute_cluster_id: "cluster-b",
    },
  ];
});

describe("<AllocationOverviewTab />", () => {
  it("renders attached resource names, types, and the allocation's grant", () => {
    render(<AllocationOverviewTab allocation={allocation} canManage={false} />);
    expect(screen.getByText("ClusterA GPU")).toBeInTheDocument();
    expect(screen.getByText("ClusterB GPU A100")).toBeInTheDocument();
    expect(screen.getAllByText("4 · 600 min")).toHaveLength(2);
  });

  it("renders the effective rate, omitting it when the lookup 404s", () => {
    render(<AllocationOverviewTab allocation={allocation} canManage={false} />);
    expect(screen.getByText("0.5 SU/unit")).toBeInTheDocument();
    // res-b-gpu has no effective rate, so only one rate line shows.
    expect(screen.getAllByText(/SU\/unit/)).toHaveLength(1);
  });

  it("offers resource writes only with manage permission", () => {
    const { unmount } = render(<AllocationOverviewTab allocation={allocation} canManage={false} />);
    expect(screen.queryByRole("button", { name: "Detach" })).not.toBeInTheDocument();
    unmount();
    render(<AllocationOverviewTab allocation={allocation} canManage />);
    expect(screen.getAllByRole("button", { name: "Detach" })).toHaveLength(2);
  });
});
