import { describe, expect, it } from "vitest";
import { identityKeys } from "../queries";

describe("identityKeys", () => {
  it("namespaces queries under identity", () => {
    expect(identityKeys.all).toEqual(["identity"]);
    expect(identityKeys.current()).toEqual(["identity", "current"]);
    expect(identityKeys.privileges()).toEqual(["identity", "privileges"]);
  });
});
