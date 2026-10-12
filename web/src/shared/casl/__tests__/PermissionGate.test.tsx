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
import { defineAbilitiesFor } from "../abilities";

let privileges: string[] = [];
vi.mock("../AbilityProvider", () => ({ useAbility: () => defineAbilitiesFor(privileges) }));

import { PermissionGate } from "../PermissionGate";

describe("<PermissionGate />", () => {
  it("opens on any listed read and otherwise shows the page heading over the not-permitted state", () => {
    const gate = (
      <PermissionGate read={["Cluster", "Allocation"]} title="Resources & clusters">
        Resources
      </PermissionGate>
    );
    privileges = ["core:allocations:read"];
    const { unmount } = render(gate);
    expect(screen.getByText("Resources")).toBeInTheDocument();
    unmount();
    privileges = [];
    render(gate);
    expect(
      screen.getByRole("heading", { level: 1, name: "Resources & clusters" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveTextContent("Not permitted");
  });
});
