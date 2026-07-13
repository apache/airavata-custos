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
import type { AttachedResource, ComputeAllocation } from "../schemas";

let attached: AttachedResource[] = [];

vi.mock("../queries", () => ({
  useAllocationResources: () => ({
    data: attached,
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  }),
}));

vi.mock("@/features/core/resources/queries", () => ({
  // res-a-gpu has an effective rate; res-b-gpu 404s (query error, no data).
  useEffectiveRate: (id: string) =>
    id === "res-a-gpu"
      ? { data: { rate: 0.5 }, isError: false }
      : { data: undefined, isError: true },
}));

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
  it("renders attached resource names, types, and amounts", () => {
    render(<AllocationOverviewTab allocation={allocation} />);
    expect(screen.getByText("ClusterA GPU")).toBeInTheDocument();
    expect(screen.getByText("ClusterB GPU A100")).toBeInTheDocument();
    expect(screen.getByText("50,000")).toBeInTheDocument();
  });

  it("renders the effective rate, omitting it when the lookup 404s", () => {
    render(<AllocationOverviewTab allocation={allocation} />);
    expect(screen.getByText("0.5 SU/unit")).toBeInTheDocument();
    // res-b-gpu has no effective rate, so only one rate line shows.
    expect(screen.getAllByText(/SU\/unit/)).toHaveLength(1);
  });

  it("shows the empty state when nothing is attached", () => {
    attached = [];
    render(<AllocationOverviewTab allocation={allocation} />);
    expect(screen.getByText(/no resources attached/i)).toBeInTheDocument();
  });
});
