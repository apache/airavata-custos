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

import type { ProjectResponse } from "@/generated/core/types.gen";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ProjectsList } from "../components/ProjectsList";

const project: ProjectResponse = {
  id: "19000001-0000-4000-8000-000019000001",
  originated_id: "BIO130000",
  title: "Genomic Sequencing Pipeline",
  origination: "ACCESS",
  project_pi_id: "24000020-0000-4000-8000-000024000020",
  status: "ACTIVE",
  created_time: "2026-04-01T00:00:00.000Z",
};

function renderList(overrides: Partial<React.ComponentProps<typeof ProjectsList>> = {}) {
  const props = {
    rows: [project],
    isLoading: false,
    error: null,
    search: "",
    onSearchChange: vi.fn(),
    statusFilter: "all" as const,
    onStatusFilterChange: vi.fn(),
    ...overrides,
  };
  return { ...render(<ProjectsList {...props} />), props };
}

describe("<ProjectsList />", () => {
  it("renders the project title as a link to the detail page", () => {
    renderList();
    const link = screen.getByRole("link", { name: project.title });
    expect(link).toHaveAttribute("href", `/projects/${project.id}`);
  });

  it("words the empty state by whether a filter is set", () => {
    const { unmount } = renderList({ rows: [] });
    expect(screen.getByRole("heading", { name: "No projects yet" })).toBeInTheDocument();
    unmount();
    renderList({ rows: [], search: "genomics" });
    expect(screen.getByRole("heading", { name: "No matching projects" })).toBeInTheDocument();
  });

  it("describes the list as every project to Project readers", () => {
    renderList({ canReadAll: true });
    expect(screen.getByText("All projects.")).toBeInTheDocument();
  });

  it("surfaces an error state with a retry callback", () => {
    const onRetry = vi.fn();
    renderList({ error: new Error("boom"), onRetry });
    expect(screen.queryByText(/boom/)).toBeNull();
    screen.getByRole("button", { name: "Try again" }).click();
    expect(onRetry).toHaveBeenCalled();
  });
});
