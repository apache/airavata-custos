import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { PacketsTrendChart } from "../components/PacketsTrendChart";

describe("PacketsTrendChart", () => {
  it("renders an empty-state when there are no buckets", () => {
    render(<PacketsTrendChart buckets={[]} />);
    expect(screen.getByText(/No packet activity/i)).toBeInTheDocument();
  });

  it("renders the chart container and legend when buckets exist", () => {
    render(
      <PacketsTrendChart
        buckets={[
          { date: "2026-05-20", status: "PROCESSED", type: "request_project_create", count: 5 },
          { date: "2026-05-20", status: "FAILED", type: "request_account_create", count: 1 },
          { date: "2026-05-21", status: "PROCESSED", type: "request_project_create", count: 7 },
        ]}
      />,
    );
    expect(screen.getByLabelText(/AMIE packets per day/i)).toBeInTheDocument();
    expect(screen.getByText("PROCESSED")).toBeInTheDocument();
    expect(screen.getByText("FAILED")).toBeInTheDocument();
  });
});
