import { describe, expect, it } from "vitest";
import { zPrivilegeKey } from "@/generated/core/zod.gen";
import type { Privilege } from "@/features/core/identity/types";
import { PRIVILEGE_ABILITY_MAP, defineAbilitiesFor } from "../abilities";

describe("PRIVILEGE_ABILITY_MAP", () => {
  it("covers every PrivilegeKey the OpenAPI enum declares", () => {
    const declared = zPrivilegeKey.options;
    const mapped = Object.keys(PRIVILEGE_ABILITY_MAP).sort();
    expect(mapped).toEqual([...declared].sort());
  });

  it("declares at least one rule per privilege", () => {
    for (const [key, rules] of Object.entries(PRIVILEGE_ABILITY_MAP)) {
      expect(rules.length, `${key} should have >=1 rule`).toBeGreaterThan(0);
    }
  });
});

const cases: Array<[Privilege, Array<[string, string, boolean]>]> = [
  [
    "amie:read",
    [
      ["read", "AMIE", true],
      ["manage", "AMIE", false],
    ],
  ],
  [
    "amie:write",
    [
      ["read", "AMIE", true],
      ["manage", "AMIE", true],
    ],
  ],
  [
    "hpc:read",
    [
      ["read", "Allocation", true],
      ["read", "Project", true],
      ["read", "Trace", true],
      ["read", "AuditEvent", true],
      ["manage", "Allocation", false],
    ],
  ],
  [
    "hpc:write",
    [
      ["manage", "Allocation", true],
      ["manage", "Project", true],
      ["read", "Allocation", true],
    ],
  ],
  [
    "signer:read",
    [
      ["read", "Signer", true],
      ["manage", "Signer", false],
    ],
  ],
  [
    "signer:write",
    [
      ["read", "Signer", true],
      ["manage", "Signer", true],
    ],
  ],
  [
    "privileges:grant",
    [
      ["manage", "PrivilegeGrant", true],
      ["read", "Allocation", false],
    ],
  ],
  [
    "roles:manage",
    [
      ["manage", "Role", true],
      ["read", "AMIE", false],
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

  it("composes rules across multiple privileges", () => {
    const ability = defineAbilitiesFor(["hpc:read", "amie:write"]);
    expect(ability.can("read", "Allocation")).toBe(true);
    expect(ability.can("manage", "AMIE")).toBe(true);
    expect(ability.can("manage", "Signer")).toBe(false);
  });
});
