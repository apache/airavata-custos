import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { StatusPill } from "@/features/core/audit/components/primitives/StatusPill";
import type { RowTone } from "@/features/core/audit/types";

const TONES: RowTone[] = ["ok", "error", "in-progress", "orphaned", "no-status"];

describe("StatusPill", () => {
  it("renders each tone with a default label", () => {
    for (const tone of TONES) {
      const { unmount } = render(<StatusPill tone={tone} />);
      expect(screen.getByRole("status")).toBeInTheDocument();
      unmount();
    }
  });

  it("dotOnly variant renders only the dot (no label text)", () => {
    render(<StatusPill tone="ok" dotOnly />);
    const node = screen.getByRole("status");
    expect(node).toHaveAttribute("aria-label", "Status: ok");
    expect(node.textContent ?? "").toBe("");
  });

  it("applies the pulsing class for in-progress", () => {
    const { container } = render(<StatusPill tone="in-progress" />);
    expect(container.querySelector(".custos-pulse-dot")).not.toBeNull();
  });

  it("does not apply the pulsing class for non-in-progress tones", () => {
    const { container } = render(<StatusPill tone="ok" />);
    expect(container.querySelector(".custos-pulse-dot")).toBeNull();
  });

  it("uses a custom label when provided", () => {
    render(<StatusPill tone="error" label="failed" />);
    expect(screen.getByText("failed")).toBeInTheDocument();
  });
});
