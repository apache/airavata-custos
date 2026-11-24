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

const role = (id: string, name: string) => ({
  id,
  name,
  description: null,
  is_system: false,
  created_at: "2026-01-01T00:00:00Z",
});
const grant = (role_id: string) => ({
  user_id: "u-1",
  role_id,
  granted_at: "",
  granted_by: null,
  reason: null,
});

const access: MyAccess = {
  provenance: true,
  roles: [
    {
      role: {
        ...role("role-admin", "Administrator"),
        description: "Full management.",
        is_system: true,
      },
      privileges: ["core:users:read", "core:users:write"],
      grant: {
        ...grant("role-admin"),
        granted_by: "portal-admin",
        granted_at: "2026-03-12T09:00:00Z",
      },
    },
    {
      role: role("role-amie", "AMIE Operator"),
      privileges: ["amie:packets:read", "amie:packets:write"],
      grant: grant("role-amie"),
    },
  ],
  direct: [
    {
      id: "grant-1",
      user_id: "u-1",
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
    "amie:packets:read",
    "amie:packets:write",
  ],
};

function rowFor(prefix: string): HTMLElement {
  const row = screen.getByText(prefix).closest("div");
  if (!row) throw new Error(`no row for ${prefix}`);
  return row;
}

function roleCard(id: string): HTMLElement {
  const card = document.querySelector(`[data-role-id="${id}"]`);
  if (!(card instanceof HTMLElement)) throw new Error(`no card for ${id}`);
  return card;
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
    expect(screen.getByText(/no roles assigned/i)).toBeInTheDocument();
    expect(screen.getByText("core:users")).toBeInTheDocument();
    expect(screen.getByText("amie:packets")).toBeInTheDocument();
    // Never claim "Direct grant" without the provenance data.
    expect(screen.queryByText("Direct grant")).not.toBeInTheDocument();
  });

  it("lists held roles with attribution even without provenance", () => {
    render(
      <AccessCard
        access={{
          provenance: false,
          roles: [
            {
              role: { ...role("role-reviewer", "Reviewer"), description: "Reads reports." },
              privileges: ["core:users:read"],
              grant: { ...grant("role-reviewer"), granted_at: "2026-07-01T09:00:00Z" },
            },
          ],
          direct: [],
          privileges: ["core:users:read", "core:traces:read"],
        }}
      />,
    );
    expect(roleCard("role-reviewer")).toHaveTextContent("Reviewer");
    expect(roleCard("role-reviewer")).toHaveTextContent("Granted");
    expect(rowFor("core:users")).toHaveTextContent("Reviewer");
    // A key not carried by any visible role must not claim "Direct grant".
    expect(rowFor("core:traces")).not.toHaveTextContent("Direct grant");
  });
});
