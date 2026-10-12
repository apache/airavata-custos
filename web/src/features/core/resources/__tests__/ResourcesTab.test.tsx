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

import type { ComputeAllocationResourceSummary, ComputeCluster } from "@/generated/core/types.gen";
import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

let resources: ComputeAllocationResourceSummary[] = [];
const clusters: ComputeCluster[] = [
  { id: "04000001-0000-4000-8000-000004000001", name: "ClusterA" },
  { id: "04000002-0000-4000-8000-000004000002", name: "ClusterB" },
];

vi.mock("../queries", () => ({
  useResourceSummaries: () => ({
    data: resources,
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  }),
  useCreateResource: () => ({ mutate: vi.fn(), isPending: false }),
}));

const useClusters = vi.fn((enabled: boolean) => ({ data: enabled ? clusters : undefined }));
vi.mock("@/features/core/clusters/queries", () => ({
  useClusters: (enabled: boolean) => useClusters(enabled),
}));

// The rates drawer owns its own queries; stub it so this suite stays on the
// table, its filters, and the cluster resolution.
vi.mock("../components/ResourcesRatesDrawer", () => ({
  ResourcesRatesDrawer: ({ resourceName }: { resourceName: string | null }) => (
    <div>{resourceName ? `rates-drawer:${resourceName}` : null}</div>
  ),
}));

vi.mock("../components/ResourceDetailDrawer", () => ({ ResourceDetailDrawer: () => null }));

let granted: string[] = [];
vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({
    can: (action: string, subject: string) => granted.includes(`${action} ${subject}`),
  }),
}));

import { ResourcesTab } from "../components/ResourcesTab";

beforeEach(() => {
  window.history.replaceState(null, "", "/");
  granted = [];
  useClusters.mockClear();
  resources = [
    {
      id: "21000001-0000-4000-8000-000021000001",
      name: "ClusterA CPU",
      resource_type: "CPU",
      resource_amount: 1000,
      compute_cluster_id: "04000001-0000-4000-8000-000004000001",
      allocation_count: 3,
      total_allocated: 50000,
      total_used_su: 12500,
      rate_count: 2,
    },
    {
      id: "21000004-0000-4000-8000-000021000004",
      name: "ClusterB GPU",
      resource_type: "GPU",
      resource_amount: 200,
      compute_cluster_id: "04000002-0000-4000-8000-000004000002",
      allocation_count: 0,
      total_allocated: 0,
      total_used_su: 0,
      rate_count: 0,
    },
  ];
});

describe("<ResourcesTab />", () => {
  it("names clusters only with Cluster read, and takes a Cluster ID without it", () => {
    granted = ["write Allocation"];
    const { unmount } = render(<ResourcesTab />);
    expect(useClusters).toHaveBeenLastCalledWith(false);
    expect(screen.queryByLabelText("Filter by cluster")).not.toBeInTheDocument();
    expect(screen.queryByText(/04000002-/)).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "+ New resource" }));
    expect(screen.getByLabelText("Cluster ID")).toBeInTheDocument();
    unmount();

    granted = ["write Allocation", "read Cluster"];
    render(<ResourcesTab />);
    expect(useClusters).toHaveBeenLastCalledWith(true);
    expect(screen.getByRole("option", { name: "ClusterB" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "+ New resource" })).toBeInTheDocument();
  });

  it("filters by search text", () => {
    render(<ResourcesTab />);
    fireEvent.change(screen.getByLabelText("Search resources"), { target: { value: "gpu" } });
    expect(screen.queryByText("ClusterA CPU")).not.toBeInTheDocument();
    expect(screen.getByText("ClusterB GPU")).toBeInTheDocument();
  });

  it("filters by the selected cluster", () => {
    granted = ["read Cluster"];
    render(<ResourcesTab />);
    fireEvent.change(screen.getByLabelText("Filter by cluster"), {
      target: { value: "04000002-0000-4000-8000-000004000002" },
    });
    expect(screen.queryByText("ClusterA CPU")).not.toBeInTheDocument();
    expect(screen.getByText("ClusterB GPU")).toBeInTheDocument();
  });

  it("opens the rates drawer for the chosen resource", () => {
    render(<ResourcesTab />);
    fireEvent.click(screen.getByRole("button", { name: "Rates (2) of ClusterA CPU" }));
    expect(screen.getByText("rates-drawer:ClusterA CPU")).toBeInTheDocument();
  });

  it("shows the empty state when no resources are registered", () => {
    resources = [];
    render(<ResourcesTab />);
    expect(screen.getByRole("heading", { name: /no resources registered/i })).toBeInTheDocument();
  });
});
