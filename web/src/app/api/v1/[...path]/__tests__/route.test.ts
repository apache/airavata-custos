import { describe, expect, it } from "vitest";

import { responseBodyForStatus } from "../proxy-response";

describe("responseBodyForStatus", () => {
  it.each([204, 205, 304])("returns null for bodyless status %s", (status) => {
    expect(responseBodyForStatus(status, "")).toBeNull();
  });

  it("preserves an ordinary response body", () => {
    expect(responseBodyForStatus(200, '{"ok":true}')).toBe('{"ok":true}');
  });
});
