import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { PacketStatusBadge, ReplyStatusBadge } from "../components/PacketStatusBadge";

describe("PacketStatusBadge", () => {
  it("renders the status text", () => {
    render(<PacketStatusBadge status="PROCESSED" />);
    expect(screen.getByText("PROCESSED")).toBeInTheDocument();
  });

  it("adds the loud '!' marker on aged FAILED packets", () => {
    const { container } = render(<PacketStatusBadge status="FAILED" ageHours={36} />);
    expect(container.textContent).toContain("FAILED");
    expect(container.textContent).toContain("!");
  });

  it("does not add the '!' marker on fresh FAILED packets", () => {
    const { container } = render(<PacketStatusBadge status="FAILED" ageHours={1} />);
    expect(container.textContent).toBe("FAILED");
  });
});

describe("ReplyStatusBadge", () => {
  it("renders the reply status text", () => {
    render(<ReplyStatusBadge status="ACKED" />);
    expect(screen.getByText("ACKED")).toBeInTheDocument();
  });
});
