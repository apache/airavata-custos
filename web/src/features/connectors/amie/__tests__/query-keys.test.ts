import { describe, expect, it } from "vitest";
import { amieKeys } from "../queries";

describe("amieKeys factory", () => {
  it("all is stable", () => {
    expect(amieKeys.all).toEqual(["amie"]);
  });

  it("packets is parameterized by filter object", () => {
    expect(amieKeys.packets({ status: "FAILED" })).toEqual([
      "amie",
      "packets",
      "list",
      { status: "FAILED" },
    ]);
  });

  it("packet detail keys are namespaced by id", () => {
    expect(amieKeys.packet("pkt-1")).toEqual(["amie", "packets", "detail", "pkt-1"]);
  });

  it("events keys are namespaced by packet id", () => {
    expect(amieKeys.events("pkt-1")).toEqual(["amie", "packets", "events", "pkt-1"]);
  });

  it("stats keys carry the window param", () => {
    expect(amieKeys.stats({ window: "30d" })).toEqual(["amie", "stats", { window: "30d" }]);
  });
});
