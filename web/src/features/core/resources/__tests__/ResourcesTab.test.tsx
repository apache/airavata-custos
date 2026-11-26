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

import type { ComputeAllocationResourceSummary } from "@/generated/core/zod.gen";
import type { Partition } from "@/generated/slurm/zod.gen";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen } from "@testing-library/react";
import { setupServer } from "msw/node";
import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, it, vi } from "vitest";
import { slurmHandlers, slurmView } from "@/mocks/handlers/slurm";

let resources: ComputeAllocationResourceSummary[] = [];
let partitions: Partition[] = [];

vi.mock("../queries", () => ({
  useResourceSummaries: () => ({
    data: resources,
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  }),
  useEffectiveRate: (id: string) => ({
    data: id === "res-a-cpu" ? { rate: 2.5 } : undefined,
    isLoading: false,
  }),
  useDeleteResource: () => ({ mutate: vi.fn(), isPending: false, error: null }),
}));

vi.mock("@/shared/casl/permissions", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/shared/casl/permissions")>()),
  usePermissions: () => ({ allocationsWrite: false }),
}));

// The rates drawer owns its own queries; stub it so this suite stays on the
// table and its filters.
vi.mock("../components/ResourcesRatesDrawer", () => ({
  ResourcesRatesDrawer: ({ resourceName }: { resourceName: string | null }) => (
    <div>{resourceName ? `rates-drawer:${resourceName}` : null}</div>
  ),
}));

import { ResourcesTab } from "../components/ResourcesTab";

const server = setupServer(...slurmHandlers);
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

function renderTab() {
  server.use(slurmView({ partitions }));
  render(
    <QueryClientProvider client={new QueryClient()}>
      <ResourcesTab clusterId="cluster-a" />
    </QueryClientProvider>,
  );
}

beforeEach(() => {
  partitions = [
    { name: "cpu", nodes: "c[1-4]", state: "UP", max_time_minutes: 60, default_time_minutes: 30 },
    { name: "debug", nodes: "d1", state: "UP", max_time_minutes: null, default_time_minutes: null },
  ];
  resources = [
    {
      id: "res-a-cpu",
      name: "cpu",
      resource_type: "CPU",
      resource_amount: 1000,
      compute_cluster_id: "cluster-a",
      allocation_count: 3,
      total_allocated: 50000,
      total_used_su: 12500,
      rate_count: 2,
    },
    {
      id: "res-a-old",
      name: "old",
      resource_type: "CPU",
      resource_amount: 200,
      compute_cluster_id: "cluster-a",
      allocation_count: 0,
      total_allocated: 0,
      total_used_su: 0,
      rate_count: 0,
    },
    {
      id: "res-b-gpu",
      name: "gpu",
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
  it("joins the cluster's SLURM partitions with their billing", async () => {
    renderTab();
    expect(await screen.findByText("debug")).toBeInTheDocument();
    expect(screen.getByText("Not billed")).toBeInTheDocument();
    expect(screen.getByText("Not in SLURM")).toBeInTheDocument();
    expect(screen.queryByText("gpu")).not.toBeInTheDocument();
  });

  it("opens the rates drawer for the chosen resource", async () => {
    renderTab();
    fireEvent.click(await screen.findByRole("button", { name: "Rates (2)" }));
    expect(screen.getByText("rates-drawer:cpu")).toBeInTheDocument();
  });

  it("shows each resource's effective rate", async () => {
    renderTab();
    expect(await screen.findByText("2.5 SU/unit·h")).toBeInTheDocument();
    expect(screen.getByText("No rate")).toBeInTheDocument();
  });

  it("shows the empty state when the cluster has no partitions", async () => {
    resources = [];
    partitions = [];
    renderTab();
    expect(await screen.findByText("No resources registered.")).toBeInTheDocument();
  });
});
