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
import { DEV_LEVEL_PRIVILEGES, DEV_LEVEL_NAMES, type DevLevel } from "../devLevels";

describe("DEV_LEVEL_PRIVILEGES", () => {
  it("viewer is read-only HPC", () => {
    expect(DEV_LEVEL_PRIVILEGES.viewer).toEqual(["hpc:read"]);
  });

  it("manager carries HPC write + AMIE read", () => {
    expect(DEV_LEVEL_PRIVILEGES.manager).toEqual(
      expect.arrayContaining(["hpc:read", "hpc:write", "amie:read"]),
    );
    expect(DEV_LEVEL_PRIVILEGES.manager).not.toContain("privileges:grant");
  });

  it("admin covers the full spec enum", () => {
    const adminSet = new Set(DEV_LEVEL_PRIVILEGES.admin);
    for (const key of zPrivilegeKey.options) {
      expect(adminSet.has(key)).toBe(true);
    }
  });

  it("every level's privileges parse against the OpenAPI enum", () => {
    for (const level of Object.keys(DEV_LEVEL_PRIVILEGES) as DevLevel[]) {
      for (const p of DEV_LEVEL_PRIVILEGES[level]) {
        expect(() => zPrivilegeKey.parse(p)).not.toThrow();
      }
    }
  });

  it("exposes a name per level", () => {
    for (const level of Object.keys(DEV_LEVEL_PRIVILEGES) as DevLevel[]) {
      expect(DEV_LEVEL_NAMES[level]).toMatch(/Dev/);
    }
  });
});
