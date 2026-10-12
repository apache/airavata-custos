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

let privileges: string[] = [];
vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({
    can: (action: string, subject: string) => privileges.includes(`${action}:${subject}`),
  }),
}));

import type { ComputeAllocation } from "@/generated/core/types.gen";
import { AllocationsList } from "../components/AllocationsList";

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
  it("describes the list by the caller's scope", () => {
    privileges = [];
    const { unmount } = renderList();
    expect(
      screen.getByText("Compute resources granted to you across projects."),
    ).toBeInTheDocument();
    unmount();
    privileges = ["read:Allocation"];
    renderList();
    expect(screen.getByText("All compute allocations.")).toBeInTheDocument();
  });

  it("renders the allocation name as a link to the detail page", () => {
    renderList();
    const link = screen.getByRole("link", { name: allocation.name });
    expect(link).toHaveAttribute("href", `/allocations/${allocation.id}`);
  });

  it("words the empty state by whether a filter is set", () => {
    const { unmount } = renderList({ rows: [] });
    expect(screen.getByRole("heading", { name: "No allocations yet" })).toBeInTheDocument();
    unmount();
    renderList({ rows: [], statusFilter: "INACTIVE" });
    expect(screen.getByRole("heading", { name: "No matching allocations" })).toBeInTheDocument();
  });

  it("surfaces an error state with a retry callback", () => {
    const onRetry = vi.fn();
    renderList({ error: new Error("boom"), onRetry });
    expect(screen.queryByText(/boom/)).not.toBeInTheDocument();
    screen.getByRole("button", { name: "Try again" }).click();
    expect(onRetry).toHaveBeenCalled();
  });
});
