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

import type { ComputeCluster, ComputeClusterUser, PrivilegeKey } from "@/generated/core/types.gen";
import { defineAbilitiesFor } from "@/shared/casl/abilities";
import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

let clusters: ComputeCluster[] = [];
let users: ComputeClusterUser[] = [];
let privileges: PrivilegeKey[] = [];
const ADA = "6a1b2c3d-0000-4000-8000-00000000ada0";

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => defineAbilitiesFor(privileges),
}));

vi.mock("@/features/core/users/queries", () => ({
  useUser: (id?: string) => ({
    data: id === ADA ? { first_name: "Ada", last_name: "Lovelace" } : undefined,
  }),
}));

vi.mock("../queries", () => ({
  useClusters: () => ({ data: clusters, isLoading: false, error: null, refetch: vi.fn() }),
  useCluster: (id?: string) => ({ data: clusters.find((c) => c.id === id) }),
  useClusterUsers: () => ({ data: users, isLoading: false, error: null, refetch: vi.fn() }),
  useCreateCluster: () => ({ mutate: vi.fn(), isPending: false }),
}));

import { ClustersTab } from "../components/ClustersTab";

beforeEach(() => {
  window.history.replaceState(null, "", "/");
  privileges = ["core:clusters:read", "core:users:read"];
  clusters = [
    { id: "04000001-0000-4000-8000-000004000001", name: "ClusterA" },
    { id: "04000003-0000-4000-8000-000004000003", name: "ClusterC" },
  ];
  users = [
    {
      id: "03000001-0000-4000-8000-000003000001",
      compute_cluster_id: "04000001-0000-4000-8000-000004000001",
      user_id: ADA,
      local_username: "alovelace",
      access_level: "USER",
      approval_status: "APPROVED",
      provisioned_at: null,
    },
  ];
});

describe("<ClustersTab />", () => {
  it("filters clusters client-side by name", () => {
    render(<ClustersTab />);
    fireEvent.change(screen.getByLabelText("Search clusters"), { target: { value: "clusterc" } });
    expect(screen.queryByText("ClusterA")).not.toBeInTheDocument();
    expect(screen.getByText("ClusterC")).toBeInTheDocument();
  });

  it("opens a drawer listing local accounts by the user's name", () => {
    render(<ClustersTab />);
    fireEvent.click(screen.getByRole("button", { name: "ClusterA" }));

    expect(screen.getByText("Cluster: ClusterA")).toBeInTheDocument();
    expect(screen.getByText("alovelace")).toBeInTheDocument();
    expect(screen.getByText("Ada Lovelace")).toBeInTheDocument();
    expect(screen.queryByText(ADA)).not.toBeInTheDocument();
  });

  it("finds an account by user ID in the loaded accounts", () => {
    render(<ClustersTab />);
    fireEvent.click(screen.getByRole("button", { name: "ClusterA" }));
    const input = screen.getByLabelText("Find account by user ID");
    const find = screen.getByRole("button", { name: "Find account" });

    fireEvent.change(input, { target: { value: "user-nobody" } });
    fireEvent.click(find);
    expect(screen.getByRole("status")).toHaveTextContent("Enter a valid user ID.");

    fireEvent.change(input, { target: { value: ADA.replace("a", "b") } });
    fireEvent.click(find);
    expect(screen.getByRole("status")).toHaveTextContent(
      "No account for that user on this cluster.",
    );

    fireEvent.change(input, { target: { value: ` ${ADA.toUpperCase()} ` } });
    fireEvent.click(find);
    expect(screen.queryByRole("status")).not.toBeInTheDocument();
    expect(screen.getAllByRole("button", { name: "alovelace" })).toHaveLength(2);
  });

  it("shows the empty state when no clusters are registered", () => {
    clusters = [];
    render(<ClustersTab />);
    expect(screen.getByRole("heading", { name: /no clusters registered/i })).toBeInTheDocument();
  });

  it("offers cluster registration only with the clusters write privilege", () => {
    render(<ClustersTab />);
    expect(screen.queryByRole("button", { name: /register cluster/i })).not.toBeInTheDocument();
    privileges = ["core:clusters:read", "core:clusters:write", "core:users:read"];
    render(<ClustersTab />);
    expect(screen.getByRole("button", { name: /register cluster/i })).toBeInTheDocument();
  });
});
