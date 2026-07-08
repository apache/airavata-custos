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
import type { ComputeCluster, ComputeClusterUser } from "../schemas";

let clusters: ComputeCluster[] = [];
let users: ComputeClusterUser[] = [];

vi.mock("../queries", () => ({
  useClusters: () => ({ data: clusters, isLoading: false, error: null, refetch: vi.fn() }),
  useClusterUsers: () => ({ data: users, isLoading: false, error: null, refetch: vi.fn() }),
}));

import { ClustersTab } from "../components/ClustersTab";

beforeEach(() => {
  clusters = [
    { id: "cluster-anvil", name: "Anvil" },
    { id: "cluster-bridges2", name: "Bridges-2" },
  ];
  users = [
    {
      id: "ccu-1",
      compute_cluster_id: "cluster-anvil",
      user_id: "user-ada-lovelace",
      local_username: "alovelace",
    },
  ];
});

describe("<ClustersTab />", () => {
  it("filters clusters client-side by name", () => {
    render(<ClustersTab />);
    fireEvent.change(screen.getByLabelText("Search clusters"), { target: { value: "bridges" } });
    expect(screen.queryByText("Anvil")).not.toBeInTheDocument();
    expect(screen.getByText("Bridges-2")).toBeInTheDocument();
  });

  it("opens a drawer with local users, rendering the user id as plain text", () => {
    render(<ClustersTab />);
    fireEvent.click(screen.getByRole("button", { name: "Anvil" }));

    expect(screen.getByText("Cluster: Anvil")).toBeInTheDocument();
    expect(screen.getByText("alovelace")).toBeInTheDocument();
    expect(screen.getByText("user-ada-lovelace")).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "user-ada-lovelace" })).not.toBeInTheDocument();
  });

  it("shows the empty state when no clusters are registered", () => {
    clusters = [];
    render(<ClustersTab />);
    expect(screen.getByRole("heading", { name: /no clusters registered/i })).toBeInTheDocument();
  });
});
