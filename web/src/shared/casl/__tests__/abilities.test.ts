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
import { zPrivilegeKey } from "@/generated/core/zod.gen";
import type { Privilege } from "@/features/core/identity/types";
import { PRIVILEGE_ABILITY_MAP, defineAbilitiesFor } from "../abilities";

// Update alongside each connector's privileges.go registry.
const KNOWN_CONNECTOR_KEYS = [
  "amie:packets:read",
  "amie:packets:write",
  "amie:replies:read",
  "amie:replies:write",
  "amie:unmapped:read",
  "amie:unmapped:write",
] as const;

describe("PRIVILEGE_ABILITY_MAP", () => {
  it("covers every core PrivilegeKey the OpenAPI enum declares", () => {
    const declared = zPrivilegeKey.options;
    for (const key of declared) {
      expect(PRIVILEGE_ABILITY_MAP[key], `${key} unmapped`).toBeDefined();
    }
  });

  it("covers every known connector privilege key", () => {
    for (const key of KNOWN_CONNECTOR_KEYS) {
      expect(PRIVILEGE_ABILITY_MAP[key], `${key} unmapped`).toBeDefined();
    }
  });

  it("declares at least one rule per privilege", () => {
    for (const [key, rules] of Object.entries(PRIVILEGE_ABILITY_MAP)) {
      expect(rules.length, `${key} should have >=1 rule`).toBeGreaterThan(0);
    }
  });
});

const cases: Array<[Privilege, Array<[string, string, boolean]>]> = [
  [
    "core:clusters:read",
    [
      ["read", "Cluster", true],
      ["manage", "Cluster", false],
    ],
  ],
  [
    "core:clusters:write",
    [
      ["read", "Cluster", true],
      ["manage", "Cluster", true],
    ],
  ],
  [
    "core:allocations:read",
    [
      ["read", "Allocation", true],
      ["manage", "Allocation", false],
    ],
  ],
  [
    "core:allocations:write",
    [
      ["read", "Allocation", true],
      ["manage", "Allocation", true],
    ],
  ],
  [
    "core:projects:read",
    [
      ["read", "Project", true],
      ["manage", "Project", false],
    ],
  ],
  [
    "core:projects:write",
    [
      ["read", "Project", true],
      ["manage", "Project", true],
    ],
  ],
  [
    "core:users:read",
    [
      ["read", "User", true],
      ["manage", "User", false],
    ],
  ],
  [
    "core:users:write",
    [
      ["read", "User", true],
      ["manage", "User", true],
    ],
  ],
  [
    "core:organizations:read",
    [
      ["read", "Organization", true],
      ["manage", "Organization", false],
    ],
  ],
  [
    "core:organizations:write",
    [
      ["read", "Organization", true],
      ["manage", "Organization", true],
    ],
  ],
  [
    "core:traces:read",
    [
      ["read", "Trace", true],
      ["read", "AuditEvent", true],
      ["manage", "Trace", false],
    ],
  ],
  [
    "core:privileges:grant",
    [
      ["manage", "PrivilegeGrant", true],
      ["read", "Allocation", false],
    ],
  ],
  [
    "core:roles:manage",
    [
      ["manage", "Role", true],
      ["read", "AMIE", false],
    ],
  ],
  [
    "amie:packets:read",
    [
      ["read", "AMIE", true],
      ["manage", "AMIE", false],
    ],
  ],
  [
    "amie:packets:write",
    [
      ["read", "AMIE", true],
      ["manage", "AMIE", true],
    ],
  ],
  [
    "amie:replies:read",
    [
      ["read", "AMIE", true],
      ["manage", "AMIE", false],
    ],
  ],
  [
    "amie:unmapped:write",
    [
      ["read", "AMIE", true],
      ["manage", "AMIE", true],
    ],
  ],
];

describe("defineAbilitiesFor (table-driven)", () => {
  for (const [privilege, expectations] of cases) {
    describe(`with [${privilege}]`, () => {
      const ability = defineAbilitiesFor([privilege]);
      for (const [action, subject, expected] of expectations) {
        it(`${expected ? "grants" : "denies"} ${action} ${subject}`, () => {
          expect(ability.can(action, subject)).toBe(expected);
        });
      }
    });
  }

  it("yields an empty ability when no privileges are present", () => {
    const ability = defineAbilitiesFor([]);
    expect(ability.can("read", "Allocation")).toBe(false);
    expect(ability.can("manage", "AMIE")).toBe(false);
  });

  it("returns no rules for unknown keys and does not throw", () => {
    const ability = defineAbilitiesFor(["not:a:real:key"]);
    expect(ability.can("read", "Allocation")).toBe(false);
  });

  it("composes rules across multiple privileges", () => {
    const ability = defineAbilitiesFor([
      "core:allocations:read",
      "amie:packets:write",
    ]);
    expect(ability.can("read", "Allocation")).toBe(true);
    expect(ability.can("manage", "AMIE")).toBe(true);
    expect(ability.can("manage", "Cluster")).toBe(false);
  });
});
