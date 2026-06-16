import { describe, expect, it } from "vitest";
import { allocationKeys } from "../queries";

describe("allocationKeys", () => {
  it("namespaces under 'allocations'", () => {
    expect(allocationKeys.all).toEqual(["allocations"]);
  });

  it("list key carries the params", () => {
    const params = { limit: 20, status: "ACTIVE" as const };
    expect(allocationKeys.list(params)).toEqual(["allocations", "list", params]);
  });

  it("detail key carries the id", () => {
    expect(allocationKeys.detail("alloc-1")).toEqual(["allocations", "detail", "alloc-1"]);
  });

  it("resources key extends detail", () => {
    expect(allocationKeys.resources("alloc-1")).toEqual([
      "allocations",
      "detail",
      "alloc-1",
      "resources",
    ]);
  });

  it("members key extends detail", () => {
    expect(allocationKeys.members("alloc-1")).toEqual([
      "allocations",
      "detail",
      "alloc-1",
      "members",
    ]);
  });

  it("change-requests list key carries params", () => {
    expect(allocationKeys.changeRequests({ status: "PENDING" })).toEqual([
      "allocations",
      "change-requests",
      "list",
      { status: "PENDING" },
    ]);
  });

  it("change-requests detail key carries the id", () => {
    expect(allocationKeys.changeRequestDetail("cr-1")).toEqual([
      "allocations",
      "change-requests",
      "detail",
      "cr-1",
    ]);
  });

  it("change-requests events key carries the id", () => {
    expect(allocationKeys.changeRequestEvents("cr-1")).toEqual([
      "allocations",
      "change-requests",
      "events",
      "cr-1",
    ]);
  });
});
