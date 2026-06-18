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
