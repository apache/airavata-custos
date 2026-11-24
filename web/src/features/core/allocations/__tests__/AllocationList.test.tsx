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

import type { ComputeAllocation } from "@/generated/core/zod.gen";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

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

vi.mock("../queries", () => ({
  useAllocationsByProject: () => ({ data: { items: [allocation], total: 1 } }),
  useAllocationUsedSU: () => ({ data: 50000 }),
}));
vi.mock("@/shared/casl/permissions", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/shared/casl/permissions")>()),
  usePermissions: () => ({}),
}));

import { AllocationList } from "../components/AllocationList";

const project = {
  id: "project-001",
  title: "Genomic Sequencing Pipeline",
  ref: "ACCESS-BIO130000",
};

describe("<AllocationList />", () => {
  it("links each allocation under its project with the SUs left", () => {
    render(<AllocationList project={project} />);
    expect(screen.getByRole("link", { name: allocation.name })).toHaveAttribute(
      "href",
      `/projects/${allocation.project_id}/${allocation.id}`,
    );
    expect(screen.getByText("200,000")).toBeInTheDocument();
  });
});
