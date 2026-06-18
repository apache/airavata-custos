import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ageHoursOf, formatDate, pluralize } from "../utils";

describe("amie utils", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-06-08T12:00:00Z"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("ageHoursOf returns 0 for undefined", () => {
    expect(ageHoursOf(undefined)).toBe(0);
  });

  it("ageHoursOf returns 0 for an unparseable iso", () => {
    expect(ageHoursOf("not-a-date")).toBe(0);
  });

  it("ageHoursOf computes hours since the timestamp", () => {
    expect(ageHoursOf("2026-06-08T10:00:00Z")).toBeCloseTo(2, 5);
  });

  it("formatDate returns em dash for undefined", () => {
    expect(formatDate(undefined)).toBe("—");
  });

  it("formatDate returns em dash for an unparseable iso", () => {
    expect(formatDate("not-a-date")).toBe("—");
  });

  it("formatDate returns a non-empty locale string for a valid iso", () => {
    expect(formatDate("2026-06-08T12:00:00Z")).not.toBe("—");
  });

  it("pluralize returns singular when count is 1", () => {
    expect(pluralize("retry", 1)).toBe("retry");
  });

  it("pluralize uses default 's' suffix", () => {
    expect(pluralize("packet", 3)).toBe("packets");
  });

  it("pluralize prefers an explicit plural form", () => {
    expect(pluralize("retry", 3, "retries")).toBe("retries");
  });
});
