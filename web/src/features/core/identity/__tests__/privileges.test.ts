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

import { describe, expect, it } from "vitest";
import { buildPrivilegeGroups, domainLabel, resourceLabel } from "../privileges";
import type { MyAccess } from "../queries";

const access: MyAccess = {
  roles: [
    {
      role: { id: "role-admin", name: "Administrator", is_system: true },
      privileges: ["core:users:read", "core:users:write", "core:roles:manage"],
      grant: { role_id: "role-admin" },
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
    "core:roles:manage",
    "core:traces:read",
    "amie:packets:read",
    "amie:packets:write",
    "future:widgets:read",
  ],
};

function must<T>(value: T | undefined): T {
  if (value === undefined) throw new Error("expected a value");
  return value;
}

describe("buildPrivilegeGroups", () => {
  const groups = buildPrivilegeGroups(access);
  const core = must(groups.find((g) => g.domain === "core"));
  const users = must(core.rows.find((r) => r.resource === "users"));

  it("groups by domain with title-cased labels, unknown domains generic", () => {
    expect(domainLabel("core")).toBe("Core");
    expect(domainLabel("amie")).toBe("AMIE");
    expect(domainLabel("future")).toBe("Future");
    expect(resourceLabel("packets")).toBe("Packets");
    expect(groups.map((g) => g.label).sort()).toEqual(["AMIE", "Core", "Future"]);
  });

  it("one row per resource with one chip per action", () => {
    expect(users.actions.map((a) => a.action)).toEqual(["read", "write"]);
    expect(users.rawKeys).toEqual(["core:users:read", "core:users:write"]);
  });

  it("attributes provenance: role name for role-derived, Direct grant otherwise", () => {
    expect(users.provenance).toBe("role");
    expect(users.roleId).toBe("role-admin");
    expect(users.provenanceLabel).toBe("Administrator");

    const traces = must(core.rows.find((r) => r.resource === "traces"));
    expect(traces.provenance).toBe("direct");
    expect(traces.provenanceLabel).toBe("Direct grant");
  });

  it("renders unknown-domain keys generically without a hardcoded list", () => {
    const future = must(groups.find((g) => g.domain === "future"));
    const row = must(future.rows.at(0));
    expect(row.resourceLabel).toBe("Widgets");
    expect(row.provenanceLabel).toBe("Direct grant");
  });
});
