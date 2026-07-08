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
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Privilege } from "@/features/core/identity/types";
import { defineAbilitiesFor } from "@/shared/casl/abilities";

let currentPrivileges: Privilege[] = [];

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => defineAbilitiesFor(currentPrivileges),
}));

// Render only the tab labels/content so the gate logic is tested without
// pulling in each tab's own data queries.
vi.mock("@/shared/ui/TabsRouter", () => ({
  TabsRouter: ({ tabs }: { tabs: Array<{ value: string; label: string }> }) => (
    <div>
      {tabs.map((t) => (
        <div key={t.value}>tab:{t.label}</div>
      ))}
    </div>
  ),
}));
vi.mock("@/features/core/clusters/components/ClustersTab", () => ({
  ClustersTab: () => <div>clusters-tab</div>,
}));
vi.mock("@/features/core/resources/components/ResourcesTab", () => ({
  ResourcesTab: () => <div>resources-tab</div>,
}));

import { ResourcesClustersTabs } from "../ResourcesClustersTabs";

beforeEach(() => {
  currentPrivileges = [];
});

describe("<ResourcesClustersTabs />", () => {
  it("shows both tabs when the ability can read Cluster and Allocation", () => {
    currentPrivileges = ["core:clusters:read", "core:allocations:read"];
    render(<ResourcesClustersTabs />);
    expect(screen.getByText("tab:Clusters")).toBeInTheDocument();
    expect(screen.getByText("tab:Resources")).toBeInTheDocument();
  });

  it("hides the Resources tab without allocation read", () => {
    currentPrivileges = ["core:clusters:read"];
    render(<ResourcesClustersTabs />);
    expect(screen.getByText("tab:Clusters")).toBeInTheDocument();
    expect(screen.queryByText("tab:Resources")).not.toBeInTheDocument();
  });

  it("hides the Clusters tab without cluster read", () => {
    currentPrivileges = ["core:allocations:read"];
    render(<ResourcesClustersTabs />);
    expect(screen.queryByText("tab:Clusters")).not.toBeInTheDocument();
    expect(screen.getByText("tab:Resources")).toBeInTheDocument();
  });
});
