import { describe, expect, it } from "vitest";
import {
  roleDetailResponseSchema,
  userListResponseSchema,
  userRolesResponseSchema,
} from "../schemas";

const user = {
  id: "user-1",
  email: "user@example.org",
  first_name: "Example",
  status: "ACTIVE",
};

describe("user schemas", () => {
  it("parses a complete paginated user response", () => {
    expect(userListResponseSchema.parse({ items: [user], total: 1 })).toEqual({
      items: [user],
      total: 1,
    });
  });

  it("rejects missing pagination totals", () => {
    expect(() => userListResponseSchema.parse({ items: [user] })).toThrow();
  });

  it("normalizes null role lists to an empty array", () => {
    expect(userRolesResponseSchema.parse(null)).toEqual([]);
  });

  it("preserves arbitrary privilege keys returned by role detail", () => {
    const result = roleDetailResponseSchema.parse({
      role: { id: "role-1", name: "Admin" },
      privileges: ["core:roles:manage", "connector:custom:inspect"],
    });
    expect(result.privileges).toEqual(["core:roles:manage", "connector:custom:inspect"]);
  });
});
