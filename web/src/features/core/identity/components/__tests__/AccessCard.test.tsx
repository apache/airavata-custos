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
import { AccessCard } from "../AccessCard";
import type { MyAccess } from "../../queries";

const access: MyAccess = {
  provenance: true,
  roles: [
    {
      role: {
        id: "role-admin",
        name: "Administrator",
        description: "Full management.",
        is_system: true,
      },
      privileges: ["core:users:read", "core:users:write"],
      grant: { role_id: "role-admin", granted_by: "portal-admin", granted_at: "2026-03-12T09:00:00Z" },
    },
    {
      role: { id: "role-amie", name: "AMIE Operator" },
      privileges: ["amie:packets:read", "amie:packets:write"],
      grant: { role_id: "role-amie" },
    },
  ],
  direct: [{ privilege: "core:traces:read" }],
  privileges: [
    "core:users:read",
    "core:users:write",
    "core:traces:read",
    "amie:packets:read",
    "amie:packets:write",
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
    expect(roleCard("role-admin")).toHaveTextContent("Administrator");
    expect(screen.getByText("SYSTEM")).toBeInTheDocument();
    const users = rowFor("core:users");
    expect(users).toHaveTextContent("read");
    expect(users).toHaveTextContent("write");
    expect(screen.getByText("amie:packets")).toBeInTheDocument();
    expect(screen.queryByText("Users")).not.toBeInTheDocument();
  });

  it("shows provenance per row: role name or Direct grant", () => {
    render(<AccessCard access={access} />);
    expect(rowFor("core:users")).toHaveTextContent("Administrator");
    expect(rowFor("core:traces")).toHaveTextContent("Direct grant");
  });

  it("highlights a role's privileges on hover", () => {
    render(<AccessCard access={access} />);
    const users = rowFor("core:users");
    const packets = rowFor("amie:packets");
    expect(users.className).not.toContain("brand-tint");

    fireEvent.mouseEnter(roleCard("role-admin"));
    expect(users.className).toContain("brand-tint");
    // Only the hovered role's rows highlight, not another role's.
    expect(packets.className).not.toContain("brand-tint");
  });

  it("renders empty states when the caller has no roles", () => {
    render(<AccessCard access={{ provenance: true, roles: [], direct: [], privileges: [] }} />);
    expect(screen.getByText("No roles assigned.")).toBeInTheDocument();
  });

  it("shows effective privileges without provenance when the reads are gated", () => {
    render(
      <AccessCard
        access={{
          provenance: false,
          roles: [],
          direct: [],
          privileges: ["core:users:read", "amie:packets:write"],
        }}
      />,
    );
    expect(screen.getByText(/role details are not visible/i)).toBeInTheDocument();
    expect(screen.getByText("core:users")).toBeInTheDocument();
    expect(screen.getByText("amie:packets")).toBeInTheDocument();
    // Never claim "Direct grant" without the provenance data.
    expect(screen.queryByText("Direct grant")).not.toBeInTheDocument();
  });
});
