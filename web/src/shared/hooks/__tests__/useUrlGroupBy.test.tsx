import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useUrlGroupBy } from "../useUrlGroupBy";

const replace = vi.fn();
let searchString = "";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace }),
  useSearchParams: () => new URLSearchParams(searchString),
}));

beforeEach(() => {
  replace.mockReset();
  searchString = "";
});

describe("useUrlGroupBy", () => {
  it("returns an empty list when ?gb= is absent", () => {
    const { result } = renderHook(() => useUrlGroupBy());
    expect(result.current.groupBy).toEqual([]);
  });

  it("parses comma-separated values and trims whitespace", () => {
    searchString = "gb=resource, project ,org";
    const { result } = renderHook(() => useUrlGroupBy());
    expect(result.current.groupBy).toEqual(["resource", "project", "org"]);
  });

  it("setGroupBy joins values into a single param", () => {
    const { result } = renderHook(() => useUrlGroupBy());
    act(() => {
      result.current.setGroupBy(["resource", "project"]);
    });
    const url = replace.mock.calls[0]?.[0] as string;
    expect(url).toContain("gb=resource%2Cproject");
    expect(replace.mock.calls[0]?.[1]).toEqual({ scroll: false });
  });

  it("setGroupBy with an empty list deletes the param", () => {
    searchString = "gb=resource&other=keep";
    const { result } = renderHook(() => useUrlGroupBy());
    act(() => {
      result.current.setGroupBy([]);
    });
    const url = replace.mock.calls[0]?.[0] as string;
    expect(url).not.toContain("gb=");
    expect(url).toContain("other=keep");
  });
});
