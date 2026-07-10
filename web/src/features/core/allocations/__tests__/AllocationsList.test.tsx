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
import { describe, expect, it, vi } from "vitest";
import { AllocationsList } from "../components/AllocationsList";
import type { ComputeAllocation } from "../schemas";

const allocation: ComputeAllocation = {
  id: "alloc-001",
  project_id: "project-001",
  name: "Genomic GPU Pool",
  status: "ACTIVE",
  compute_cluster_id: "cluster-001",
  initial_su_amount: 250000,
  start_time: "2026-04-01T00:00:00.000Z",
  end_time: "2027-03-31T00:00:00.000Z",
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
    ...overrides,
  };
  return { ...render(<AllocationsList {...props} />), props };
}

describe("<AllocationsList />", () => {
  it("renders the allocation name as a link to the detail page", () => {
    renderList();
    const link = screen.getByRole("link", { name: allocation.name });
    expect(link).toHaveAttribute("href", `/allocations/${allocation.id}`);
  });

  it("shows the empty state when no rows match the filters", () => {
    renderList({ rows: [] });
    expect(screen.getByRole("heading", { name: /no allocations yet/i })).toBeInTheDocument();
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
