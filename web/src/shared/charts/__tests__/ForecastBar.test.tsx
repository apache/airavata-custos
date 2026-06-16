import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ForecastBar } from "../ForecastBar";

describe("ForecastBar", () => {
  it("positions the marker at (projected/capacity)% across the bar", () => {
    render(<ForecastBar used={20} projected={75} capacity={100} />);
    const marker = screen.getByTestId("forecast-marker");
    expect(marker.style.left).toBe("75%");
  });

  it("renders red caption tone when exhaust < endDate", () => {
    render(
      <ForecastBar
        used={50}
        projected={90}
        capacity={100}
        exhaustDate="2026-11-14"
        endDate="2026-12-07"
      />,
    );
    const caption = screen.getByTestId("forecast-caption");
    expect(caption.className).toContain("var(--custos-red-700)");
    expect(caption.textContent).toMatch(/days before award end/);
  });

  it("uses muted caption tone when exhaust >= endDate", () => {
    render(
      <ForecastBar
        used={50}
        projected={90}
        capacity={100}
        exhaustDate="2027-01-01"
        endDate="2026-12-07"
      />,
    );
    const caption = screen.getByTestId("forecast-caption");
    expect(caption.className).toContain("text-muted-foreground");
    expect(caption.textContent).toMatch(/days after award end/);
  });

  it("omits caption when exhaustDate is missing", () => {
    render(<ForecastBar used={20} projected={40} capacity={100} endDate="2026-12-07" />);
    expect(screen.queryByTestId("forecast-caption")).toBeNull();
  });

  it("renders an exhaust-only caption when endDate is missing", () => {
    render(<ForecastBar used={20} projected={40} capacity={100} exhaustDate="2026-11-14" />);
    const caption = screen.getByTestId("forecast-caption");
    expect(caption.textContent).toMatch(/Projected exhaust/);
  });
});
