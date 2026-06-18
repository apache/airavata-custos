import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { BurnDownBar } from "../BurnDownBar";
import { Sparkline } from "../Sparkline";
import { StackedAreaUsage } from "../StackedAreaUsage";

describe("chart wrappers", () => {
  it("StackedAreaUsage renders without throwing", () => {
    const data = [
      { date: "2026-05-01", cpu: 12, gpu: 4 },
      { date: "2026-05-02", cpu: 18, gpu: 6 },
    ];
    const { container } = render(
      <StackedAreaUsage data={data} seriesKeys={["cpu", "gpu"]} ariaLabel="usage" />,
    );
    expect(container.querySelector("[aria-label=usage]")).toBeTruthy();
  });

  it("BurnDownBar renders without throwing", () => {
    const { container } = render(<BurnDownBar used={20} projected={50} capacity={100} />);
    expect(container.querySelector("[role=img]")).toBeTruthy();
  });

  it("Sparkline renders without throwing", () => {
    const data = [{ value: 1 }, { value: 4 }, { value: 2 }];
    const { container } = render(<Sparkline data={data} ariaLabel="sparkline" />);
    expect(container.querySelector("[aria-label=sparkline]")).toBeTruthy();
  });
});
