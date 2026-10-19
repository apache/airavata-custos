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

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";
import projects from "@/features/core/projects/__fixtures__/projects.json";
import { AllocationsList } from "../components/AllocationsList";
import type { ComputeAllocation } from "@/generated/core/zod.gen";

vi.mock("next/navigation", () => ({ useRouter: () => ({ push: vi.fn() }) }));

const allocation: ComputeAllocation = {
  id: "alloc-001",
  project_id: "project-001",
  name: "Genomic GPU Pool",
  status: "ACTIVE",
  compute_cluster_id: "cluster-a",
  initial_su_amount: 250000,
  start_time: "2026-04-01T00:00:00.000Z",
  end_time: "2027-03-31T00:00:00.000Z",
  posix_group: null,
  group_provisioned_at: null,
};

function renderList(overrides: Partial<React.ComponentProps<typeof AllocationsList>> = {}) {
  const props = {
    rows: [allocation],
    isLoading: false,
    error: null,
    search: "",
    onSearchChange: vi.fn(),
    statusFilter: "all" as const,
    onStatusFilterChange: vi.fn(),
    canEditResources: false,
    ...overrides,
  };
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return {
    ...render(
      <QueryClientProvider client={client}>
        <AllocationsList {...props} />
      </QueryClientProvider>,
    ),
    props,
  };
}

const server = setupServer(
  http.get<{ id: string }>("*/api/v1/projects/:id", ({ params }) =>
    HttpResponse.json(projects.find((p) => p.id === params.id)),
  ),
  http.get("*/api/v1/compute-allocations/:id/usages/total", () =>
    HttpResponse.json({ compute_allocation_id: "alloc-001", total_su_amount: 50000 }),
  ),
  http.get("*/api/v1/compute-allocations/:id/resources", () =>
    HttpResponse.json([
      {
        id: "res-a",
        name: "gpu-a100",
        resource_type: "gres/gpu",
        resource_amount: 8,
        compute_cluster_id: "cluster-a",
      },
    ]),
  ),
);

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterAll(() => server.close());
afterEach(() => server.resetHandlers());

describe("<AllocationsList />", () => {
  it("renders the allocation name as a link to the detail page", () => {
    renderList();
    const link = screen.getByRole("link", { name: allocation.name });
    expect(link).toHaveAttribute("href", `/allocations/${allocation.id}`);
  });

  it("groups rows under a project header with the allocation's resources and remaining SUs", async () => {
    renderList();
    expect(await screen.findByRole("link", { name: /Go to project/ })).toHaveAttribute(
      "href",
      `/projects/${allocation.project_id}`,
    );
    expect(await screen.findByText("gpu-a100")).toBeInTheDocument();
    expect(await screen.findByText("200,000")).toBeInTheDocument();
  });

  it("shows the empty state when no rows match the filters", () => {
    renderList({ rows: [] });
    expect(screen.getByText("No allocations match the current filters.")).toBeInTheDocument();
  });

  it("filters by status", () => {
    const inactive: ComputeAllocation = {
      ...allocation,
      id: "alloc-002",
      name: "Legacy Pool",
      status: "INACTIVE",
    };
    renderList({ rows: [allocation, inactive], statusFilter: "INACTIVE" });
    expect(screen.queryByRole("link", { name: allocation.name })).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: inactive.name })).toBeInTheDocument();
  });

  it("surfaces an error state with a retry callback", () => {
    const onRetry = vi.fn();
    renderList({ error: new Error("boom"), onRetry });
    expect(screen.getByText(/boom/)).toBeInTheDocument();
  });
});
