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
import type { ResourceSummary } from "../schemas";

let resources: ResourceSummary[] = [];
const clusters: ComputeCluster[] = [
  { id: "cluster-a", name: "ClusterA" },
  { id: "cluster-b", name: "ClusterB" },
];

vi.mock("../queries", () => ({
  useResourceSummaries: () => ({ data: resources, isLoading: false, error: null, refetch: vi.fn() }),
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
      id: "res-a-cpu",
      name: "ClusterA CPU",
      resource_type: "CPU",
      resource_amount: 1000,
      compute_cluster_id: "cluster-a",
      allocation_count: 3,
      total_allocated: 50000,
      total_used_su: 12500,
      rate_count: 2,
    },
    {
      id: "res-b-gpu",
      name: "ClusterB GPU",
      resource_type: "GPU",
      resource_amount: 200,
      compute_cluster_id: "cluster-b",
      allocation_count: 0,
      total_allocated: 0,
      total_used_su: 0,
      rate_count: 0,
    },
  ];
});

describe("<ResourcesTab />", () => {
  it("resolves the cluster name as plain text, not a link", () => {
    render(<ResourcesTab />);
    expect(screen.getByText("ClusterA CPU")).toBeInTheDocument();
    expect(screen.getAllByText("ClusterA", { exact: true }).length).toBeGreaterThan(0);
    expect(screen.queryByRole("link", { name: "ClusterA" })).not.toBeInTheDocument();
  });

  it("filters by search text", () => {
    render(<ResourcesTab />);
    fireEvent.change(screen.getByLabelText("Search resources"), { target: { value: "gpu" } });
    expect(screen.queryByText("ClusterA CPU")).not.toBeInTheDocument();
    expect(screen.getByText("ClusterB GPU")).toBeInTheDocument();
  });

  it("filters by the selected cluster", () => {
    render(<ResourcesTab />);
    fireEvent.change(screen.getByLabelText("Filter by cluster"), {
      target: { value: "cluster-b" },
    });
    expect(screen.queryByText("ClusterA CPU")).not.toBeInTheDocument();
    expect(screen.getByText("ClusterB GPU")).toBeInTheDocument();
  });

  it("opens the rates drawer for the chosen resource", () => {
    render(<ResourcesTab />);
    fireEvent.click(screen.getByRole("button", { name: "Rates (2)" }));
    expect(screen.getByText("rates-drawer:ClusterA CPU")).toBeInTheDocument();
  });

  it("renders aggregate columns with a usage bar", () => {
    render(<ResourcesTab />);
    expect(screen.getByText("50,000")).toBeInTheDocument();
    expect(screen.getByText("12,500")).toBeInTheDocument();
    expect(screen.getAllByText("25.0%").length).toBeGreaterThan(0);
  });

  it("shows the empty state when no resources are registered", () => {
    resources = [];
    render(<ResourcesTab />);
    expect(screen.getByRole("heading", { name: /no resources registered/i })).toBeInTheDocument();
  });
});
