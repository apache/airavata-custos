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

import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ComputeCluster } from "@/features/core/clusters/schemas";
import type { ComputeAllocationResource } from "../schemas";

let resources: ComputeAllocationResource[] = [];
const clusters: ComputeCluster[] = [
  { id: "cluster-anvil", name: "Anvil" },
  { id: "cluster-delta", name: "Delta" },
];

vi.mock("../queries", () => ({
  useResources: () => ({ data: resources, isLoading: false, error: null, refetch: vi.fn() }),
}));

vi.mock("@/features/core/clusters/queries", () => ({
  useClusters: () => ({ data: clusters, isLoading: false, error: null, refetch: vi.fn() }),
}));

// The rates drawer owns its own queries; stub it so this suite stays on the
// table, its filters, and the cluster resolution.
vi.mock("../components/ResourcesRatesDrawer", () => ({
  ResourcesRatesDrawer: ({ resourceName }: { resourceName: string | null }) => (
    <div>{resourceName ? `rates-drawer:${resourceName}` : null}</div>
  ),
}));

import { ResourcesTab } from "../components/ResourcesTab";

beforeEach(() => {
  resources = [
    {
      id: "res-anvil-cpu",
      name: "Anvil CPU",
      resource_type: "CPU",
      resource_amount: 1000,
      compute_cluster_id: "cluster-anvil",
    },
    {
      id: "res-delta-gpu",
      name: "Delta GPU",
      resource_type: "GPU",
      resource_amount: 200,
      compute_cluster_id: "cluster-delta",
    },
  ];
});

describe("<ResourcesTab />", () => {
  it("resolves the cluster name as plain text, not a link", () => {
    render(<ResourcesTab />);
    expect(screen.getByText("Anvil CPU")).toBeInTheDocument();
    expect(screen.getAllByText("Anvil", { exact: true }).length).toBeGreaterThan(0);
    expect(screen.queryByRole("link", { name: "Anvil" })).not.toBeInTheDocument();
  });

  it("filters by search text", () => {
    render(<ResourcesTab />);
    fireEvent.change(screen.getByLabelText("Search resources"), { target: { value: "gpu" } });
    expect(screen.queryByText("Anvil CPU")).not.toBeInTheDocument();
    expect(screen.getByText("Delta GPU")).toBeInTheDocument();
  });

  it("filters by the selected cluster", () => {
    render(<ResourcesTab />);
    fireEvent.change(screen.getByLabelText("Filter by cluster"), {
      target: { value: "cluster-delta" },
    });
    expect(screen.queryByText("Anvil CPU")).not.toBeInTheDocument();
    expect(screen.getByText("Delta GPU")).toBeInTheDocument();
  });

  it("opens the rates drawer for the chosen resource", () => {
    render(<ResourcesTab />);
    fireEvent.click(screen.getAllByRole("button", { name: "Rates" })[0] as HTMLElement);
    expect(screen.getByText("rates-drawer:Anvil CPU")).toBeInTheDocument();
  });

  it("shows the empty state when no resources are registered", () => {
    resources = [];
    render(<ResourcesTab />);
    expect(screen.getByRole("heading", { name: /no resources registered/i })).toBeInTheDocument();
  });
});
