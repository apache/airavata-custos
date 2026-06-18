import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { StatCardRow } from "../StatCardRow";

describe("StatCardRow", () => {
  it("defaults to a 3-col responsive grid (backward-compat with allocation header)", () => {
    const { container } = render(
      <StatCardRow>
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </StatCardRow>,
    );
    const row = container.firstElementChild;
    expect(row?.className).toContain("md:grid-cols-3");
  });

  it("renders 5-col KPI strip when cols=5", () => {
    const { container } = render(
      <StatCardRow cols={5}>
        <div>A</div>
      </StatCardRow>,
    );
    const row = container.firstElementChild;
    expect(row?.className).toContain("sm:grid-cols-2");
    expect(row?.className).toContain("lg:grid-cols-5");
  });

  it("renders 4-col strip when cols=4", () => {
    const { container } = render(
      <StatCardRow cols={4}>
        <div>A</div>
      </StatCardRow>,
    );
    expect(container.firstElementChild?.className).toContain("lg:grid-cols-4");
  });

  it("passes through className", () => {
    const { container } = render(
      <StatCardRow cols={5} className="custom-marker">
        <div>A</div>
      </StatCardRow>,
    );
    expect(container.firstElementChild?.className).toContain("custom-marker");
  });
});
