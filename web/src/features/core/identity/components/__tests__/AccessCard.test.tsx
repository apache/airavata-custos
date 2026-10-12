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
import { describe, expect, it } from "vitest";
import type { MyAccess } from "../../queries";
import { AccessCard } from "../AccessCard";

const access: MyAccess = {
  provenance: true,
  roles: [
    {
      role: {
        id: "22000001-0000-4000-8000-000022000001",
        name: "Administrator",
        description: "Full management.",
        is_system: true,
        created_at: "2026-01-01T00:00:00Z",
      },
      privileges: ["core:users:read", "core:users:write"],
      granted_at: "2026-03-12T09:00:00Z",
    },
    {
      role: {
        id: "role-ops",
        name: "Cluster Operator",
        description: null,
        is_system: false,
        created_at: "2026-01-01T00:00:00Z",
      },
      privileges: ["core:clusters:read", "core:clusters:write"],
      granted_at: "2026-01-01T00:00:00Z",
    },
  ],
  direct: [
    {
      id: "g1",
      user_id: "u1",
      privilege: "core:traces:read",
      granted_at: "2026-01-01T00:00:00Z",
      granted_by: null,
      reason: null,
    },
  ],
  privileges: [
    "core:users:read",
    "core:users:write",
    "core:traces:read",
    "core:clusters:read",
    "core:clusters:write",
  ],
};

function rowFor(prefix: string): HTMLElement {
  return screen.getByText(prefix).closest("div") as HTMLElement;
}

function roleCard(id: string): HTMLElement {
  return document.querySelector(`[data-role-id="${id}"]`) as HTMLElement;
}

describe("AccessCard", () => {
  it("renders prefix rows with raw action chips", () => {
    render(<AccessCard access={access} />);
    expect(screen.getByRole("heading", { name: "Roles" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Effective privileges" })).toBeInTheDocument();
    expect(roleCard("22000001-0000-4000-8000-000022000001")).toHaveTextContent("Administrator");
    expect(roleCard("22000001-0000-4000-8000-000022000001")).toHaveTextContent("Granted");
    expect(screen.getByText("SYSTEM")).toBeInTheDocument();
    const users = rowFor("core:users");
    expect(users).toHaveTextContent("read");
    expect(users).toHaveTextContent("write");
    expect(screen.getByText("core:clusters")).toBeInTheDocument();
    expect(screen.queryByText("Users")).not.toBeInTheDocument();
  });

  it("highlights a role's privileges on hover", () => {
    render(<AccessCard access={access} />);
    const users = rowFor("core:users");
    const clusters = rowFor("core:clusters");
    expect(users.className).not.toContain("brand-tint");

    fireEvent.mouseEnter(roleCard("22000001-0000-4000-8000-000022000001"));
    expect(users.className).toContain("brand-tint");
    // Only the hovered role's rows highlight, not another role's.
    expect(clusters.className).not.toContain("brand-tint");
  });

  it("lists held roles with their grant date even without provenance", () => {
    render(
      <AccessCard
        access={{
          provenance: false,
          roles: [
            {
              role: {
                id: "role-reviewer",
                name: "Reviewer",
                description: "Reads reports.",
                is_system: false,
                created_at: "2026-01-01T00:00:00Z",
              },
              privileges: ["core:users:read"],
              granted_at: "2026-07-01T09:00:00Z",
            },
          ],
          direct: [],
          privileges: ["core:users:read"],
        }}
      />,
    );
    expect(roleCard("role-reviewer")).toHaveTextContent("Reviewer");
    expect(roleCard("role-reviewer")).toHaveTextContent("Granted");
    expect(rowFor("core:users")).toHaveTextContent("Reviewer");
  });
});
