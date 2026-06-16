import { describe, expect, it } from "vitest";
import {
  callerPrivilegesSchema,
  privilegeKeySchema,
  privilegesResponseSchema,
} from "../schemas";

describe("identity schemas", () => {
  it("re-exports the generated PrivilegeKey enum", () => {
    expect(() => privilegeKeySchema.parse("hpc:read")).not.toThrow();
    expect(() => privilegeKeySchema.parse("not-a-privilege")).toThrow();
  });

  it("accepts the OpenAPI CallerPrivileges shape", () => {
    const parsed = callerPrivilegesSchema.parse({ privileges: ["amie:read"] });
    expect(parsed.privileges).toEqual(["amie:read"]);
  });

  it("unwraps the response transform to a Privilege[] (default [])", () => {
    expect(privilegesResponseSchema.parse({ privileges: ["hpc:read"] })).toEqual(["hpc:read"]);
    expect(privilegesResponseSchema.parse({})).toEqual([]);
  });
});
