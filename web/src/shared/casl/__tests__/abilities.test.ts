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
import { PRIVILEGE_ABILITY_MAP, defineAbilitiesFor } from "../abilities";

// Update alongside each connector's privileges.go registry.
const KNOWN_CONNECTOR_KEYS = [
  "amie:packets:read",
  "amie:packets:write",
  "amie:replies:read",
  "amie:replies:write",
  "amie:unmapped:read",
  "amie:unmapped:write",
  "temp-account:accounts:read",
  "temp-account:accounts:write",
] as const;

describe("PRIVILEGE_ABILITY_MAP", () => {
  it.each([...zPrivilegeKey.options, ...KNOWN_CONNECTOR_KEYS])("maps %s to a rule", (key) => {
    expect(PRIVILEGE_ABILITY_MAP[key]?.length).toBeGreaterThan(0);
  });
});

describe("defineAbilitiesFor", () => {
  it("returns no rules for unknown keys and does not throw", () => {
    const ability = defineAbilitiesFor(["not:a:real:key"]);
    expect(ability.can("read", "Allocation")).toBe(false);
  });

  it("composes rules across multiple privileges without granting write from read", () => {
    const ability = defineAbilitiesFor(["core:allocations:read", "amie:packets:write"]);
    expect(ability.can("read", "Allocation")).toBe(true);
    expect(ability.can("write", "Allocation")).toBe(false);
    expect(ability.can("write", "AmiePacket")).toBe(true);
    expect(ability.can("write", "Cluster")).toBe(false);
  });
});
