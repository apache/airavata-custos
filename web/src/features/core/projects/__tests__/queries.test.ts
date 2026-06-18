import { describe, expect, it } from "vitest";
import { projectKeys } from "../queries";

describe("projectKeys", () => {
  it("namespaces under 'projects'", () => {
    expect(projectKeys.all).toEqual(["projects"]);
  });

  it("list key carries the params", () => {
    const params = { limit: 20, status: "ACTIVE" as const };
    expect(projectKeys.list(params)).toEqual(["projects", "list", params]);
  });

  it("detail key carries the id", () => {
    expect(projectKeys.detail("p-1")).toEqual(["projects", "detail", "p-1"]);
  });

  it("members key carries the project id", () => {
    expect(projectKeys.members("p-1")).toEqual(["projects", "members", "p-1"]);
  });
});
