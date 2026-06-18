import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { SourcePill } from "@/features/core/audit/components/primitives/SourcePill";

describe("SourcePill", () => {
  it.each(["amie", "comanage", "slurm", "http", "core"])("renders the %s source", (src) => {
    render(<SourcePill source={src} />);
    expect(screen.getByText(src)).toBeInTheDocument();
  });

  it("falls back to muted styling for an unknown source without crashing", () => {
    render(<SourcePill source="future-connector" />);
    const node = screen.getByText("future-connector");
    expect(node).toBeInTheDocument();
    expect(node.className).toContain("bg-muted");
  });

  it("lowercases the displayed source name", () => {
    render(<SourcePill source="AMIE" />);
    expect(screen.getByText("amie")).toBeInTheDocument();
  });
});
