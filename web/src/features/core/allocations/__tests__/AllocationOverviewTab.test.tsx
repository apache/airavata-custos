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
import type { ComputeAllocation, ComputeAllocationResource } from "@/generated/core/types.gen";
import { fireEvent, render, screen } from "@testing-library/react";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";

// The resources endpoint answers with the attached resources themselves.
const attached: ComputeAllocationResource[] = [
  {
    id: "21000002-0000-4000-8000-000021000002",
    name: "ClusterA GPU",
    resource_type: "GPU",
    resource_amount: 4,
    compute_cluster_id: "04000001-0000-4000-8000-000004000001",
  },
  {
    id: "21000001-0000-4000-8000-000021000001",
    name: "ClusterA CPU",
    resource_type: "CPU",
    resource_amount: 64,
    compute_cluster_id: "04000001-0000-4000-8000-000004000001",
  },
];
const mutation = { mutate: vi.fn(), isPending: false };

vi.mock("../queries", () => ({
  useAllocationResources: () => ({ data: attached, isLoading: false, error: null }),
  useAttachResource: () => mutation,
  useDetachResource: () => mutation,
}));

vi.mock("@/features/core/resources/queries", () => ({
  useResources: () => ({
    data: [
      ...attached,
      {
        id: "res-a-big",
        name: "ClusterA Bigmem",
        compute_cluster_id: "04000001-0000-4000-8000-000004000001",
      },
      {
        id: "21000003-0000-4000-8000-000021000003",
        name: "ClusterB CPU",
        compute_cluster_id: "04000002-0000-4000-8000-000004000002",
      },
    ],
  }),
  // The GPU resource has an effective rate; the CPU resource has none.
  useEffectiveRate: (id: string) => ({
    data: id === "21000002-0000-4000-8000-000021000002" ? { rate: 0.5 } : null,
  }),
}));
vi.mock("@/features/core/projects/queries", () => ({
  useProject: () => ({ data: { title: "Genomics Project" } }),
}));
vi.mock("@/features/core/clusters/queries", () => ({
  useCluster: () => ({ data: { name: "ClusterA" } }),
}));

vi.mock("sonner", () => ({ toast: { warning: vi.fn(), dismiss: vi.fn(), success: vi.fn() } }));
vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => true }) }));

import { AllocationOverviewTab } from "../components/AllocationOverviewTab";

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

describe("<AllocationOverviewTab />", () => {
  it("names each attached resource from the resources response", () => {
    render(<AllocationOverviewTab allocation={allocation} canManage={false} canRead={false} />);
    expect(screen.getByText("ClusterA GPU")).toBeInTheDocument();
    expect(screen.getByText("ClusterA CPU")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Genomics Project" })).toHaveAttribute(
      "href",
      "/projects/19000001-0000-4000-8000-000019000001",
    );
    expect(screen.getByText("ClusterA")).toBeInTheDocument();
  });

  it("renders the effective rate, omitting it when none applies", () => {
    render(<AllocationOverviewTab allocation={allocation} canManage={false} canRead />);
    expect(screen.getByText("0.5 SU/unit")).toBeInTheDocument();
    expect(screen.getAllByText(/SU\/unit/)).toHaveLength(1);
  });

  it("offers resource writes only with manage permission", () => {
    const { unmount } = render(
      <AllocationOverviewTab allocation={allocation} canManage={false} canRead />,
    );
    expect(screen.queryByRole("button", { name: /^Detach / })).not.toBeInTheDocument();
    unmount();
    render(<AllocationOverviewTab allocation={allocation} canManage canRead />);
    expect(screen.getAllByRole("button", { name: /^Detach / })).toHaveLength(2);
  });

  it("reads no rates and offers no attach without allocations read", () => {
    render(<AllocationOverviewTab allocation={allocation} canManage canRead={false} />);
    expect(screen.queryByText(/SU\/unit/)).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "+ Attach resource" })).not.toBeInTheDocument();
  });

  it("names the resource in the detach confirmation", () => {
    render(<AllocationOverviewTab allocation={allocation} canManage canRead />);
    fireEvent.click(screen.getAllByRole("button", { name: /^Detach / })[0] as HTMLElement);
    expect(toast.warning).toHaveBeenCalledWith(
      "Detach ClusterA GPU from this allocation?",
      expect.anything(),
    );
  });

  it("offers to attach only the cluster's resources not attached yet", () => {
    render(<AllocationOverviewTab allocation={allocation} canManage canRead />);
    fireEvent.click(screen.getByRole("button", { name: "+ Attach resource" }));
    const options = screen.getAllByRole("option").map((o) => o.textContent);
    expect(options).toEqual(["ClusterA Bigmem"]);
  });
});
