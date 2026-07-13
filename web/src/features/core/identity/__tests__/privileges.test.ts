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
import { buildPrivilegeRows } from "../privileges";
import type { MyAccess } from "../queries";

const access: MyAccess = {
  provenance: true,
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

describe("buildPrivilegeRows", () => {
  const rows = buildPrivilegeRows(access);

  it("renders one row per prefix with the action segments, sorted", () => {
    expect(rows.map((r) => r.prefix)).toEqual([
      "amie:packets",
      "core:roles",
      "core:traces",
      "core:users",
      "future:widgets",
    ]);
    const users = must(rows.find((r) => r.prefix === "core:users"));
    expect(users.actions).toEqual(["read", "write"]);
    const packets = must(rows.find((r) => r.prefix === "amie:packets"));
    expect(packets.actions).toEqual(["read", "write"]);
  });

  it("attributes provenance: role name for role-derived, Direct grant otherwise", () => {
    const users = must(rows.find((r) => r.prefix === "core:users"));
    expect(users.provenance).toBe("role");
    expect(users.roleId).toBe("role-admin");
    expect(users.provenanceLabel).toBe("Administrator");

    const traces = must(rows.find((r) => r.prefix === "core:traces"));
    expect(traces.provenance).toBe("direct");
    expect(traces.provenanceLabel).toBe("Direct grant");

    // Unknown keys fall back to Direct grant only when provenance was readable.
    const future = must(rows.find((r) => r.prefix === "future:widgets"));
    expect(future.provenanceLabel).toBe("Direct grant");
  });

  it("never claims Direct grant without the provenance reads", () => {
    const gated = buildPrivilegeRows({
      provenance: false,
      roles: [],
      direct: [],
      privileges: ["core:users:read"],
    });
    expect(must(gated.at(0)).provenance).toBe("unknown");
    expect(must(gated.at(0)).provenanceLabel).toBe("");
  });
});
