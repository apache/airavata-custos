import { fireEvent, render, screen } from "@testing-library/react";
import { Cpu } from "lucide-react";
import { describe, expect, it, vi } from "vitest";
import { KpiCard } from "../KpiCard";

describe("KpiCard", () => {
  it("renders with all props (icon, title, value, delta, sparkline)", () => {
    render(
      <KpiCard
        icon={Cpu}
        title="Used SUs"
        value="124,860"
        delta={{ value: 12, unit: "% wow", direction: "up" }}
        deltaTone="positive"
        sparkline={[1, 2, 5, 7, 6, 5, 3]}
      />,
    );
    expect(screen.getByText("Used SUs")).toBeInTheDocument();
    expect(screen.getByText("124,860")).toBeInTheDocument();
    // Delta pill renders the value+unit string.
    expect(screen.getByTestId("kpi-delta")).toHaveTextContent("12% wow");
    // Sparkline footer present.
    expect(screen.getByTestId("kpi-sparkline")).toBeInTheDocument();
  });

  it("renders without sparkline when prop is omitted", () => {
    render(<KpiCard title="Days left" value="18 d" />);
    expect(screen.getByText("Days left")).toBeInTheDocument();
    expect(screen.queryByTestId("kpi-sparkline")).toBeNull();
  });

  it("labels the sparkline with the card title so each card is distinct in the a11y tree", () => {
    render(
      <KpiCard title="Burn rate" value="4,160/d" sparkline={[1, 2, 3, 4, 5]} />,
    );
    expect(
      screen.getByRole("img", { name: /Burn rate trend sparkline/i }),
    ).toBeInTheDocument();
  });

  it("fires onClick when card is clicked", () => {
    const onClick = vi.fn();
    render(<KpiCard title="Jobs run" value="312" onClick={onClick} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it("renders skeleton when loading", () => {
    render(<KpiCard title="Used SUs" value="—" loading />);
    expect(screen.getByTestId("kpi-skeleton")).toBeInTheDocument();
    // Loading state hides the real value.
    expect(screen.queryByText("Used SUs")).toBeNull();
  });
});
