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
import { act, fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const { createProject, createAllocation, toastError, push } = vi.hoisted(() => ({
  createProject: vi.fn(async () => ({ id: "project-new", title: "Climate" })),
  createAllocation: vi.fn(),
  toastError: vi.fn(),
  push: vi.fn(),
}));

vi.mock("next/navigation", () => ({ useRouter: () => ({ push }) }));

vi.mock("../queries", () => ({
  useCreateProject: () => ({ mutateAsync: createProject }),
}));
vi.mock("@/features/core/allocations/queries", () => ({
  useCreateAllocation: () => ({ mutateAsync: createAllocation }),
}));
vi.mock("@/features/core/users/queries", () => ({
  useUserEmailSearch: () => ({ match: { id: "user-pi" }, suggestions: [], emptyMessage: "" }),
}));
vi.mock("@/features/core/clusters/queries", () => ({
  useClusters: () => ({ data: [{ id: "cluster-a", name: "ClusterA" }] }),
  useCluster: () => ({ data: { name: "ClusterA" } }),
}));
vi.mock("@/features/core/resources/queries", () => ({
  useResourceSummaries: () => ({ data: [] }),
}));
vi.mock("sonner", () => ({ toast: { error: toastError, success: vi.fn() } }));

import { CreateProjectDialog } from "../components/CreateProjectDialog";

const click = (name: string) => fireEvent.click(screen.getByRole("button", { name }));

function addDraft(name: string) {
  fireEvent.change(screen.getByLabelText("Name"), { target: { value: name } });
  fireEvent.change(screen.getByLabelText("Initial SUs"), { target: { value: "1000" } });
  click("Add allocation");
}

describe("<CreateProjectDialog />", () => {
  it("creates nothing until confirmed, then the project and each allocation under it", async () => {
    createAllocation.mockResolvedValueOnce({}).mockRejectedValueOnce(new Error("quota"));
    render(
      <QueryClientProvider client={new QueryClient()}>
        <CreateProjectDialog onOpenChange={vi.fn()} />
      </QueryClientProvider>,
    );

    fireEvent.change(screen.getByLabelText("Title"), { target: { value: "Climate" } });
    fireEvent.change(screen.getByLabelText("PI"), { target: { value: "pi@example.edu" } });
    click("Next");
    addDraft("GPU");
    addDraft("CPU");
    click("Next");
    expect(createProject).not.toHaveBeenCalled();

    await act(async () => click("Confirm and create"));
    expect(createProject).toHaveBeenCalledWith({
      title: "Climate",
      origination: "internal",
      project_pi_id: "user-pi",
    });
    expect(createAllocation.mock.calls.map(([p]) => [p.project_id, p.name])).toEqual([
      ["project-new", "GPU"],
      ["project-new", "CPU"],
    ]);
    expect(toastError).toHaveBeenCalledWith("Project Climate created; 1 allocation(s) failed.");
    expect(push).toHaveBeenCalledWith("/projects/project-new");
  });
});
