import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { type ComplianceCell, ComplianceMatrix } from "../ComplianceMatrix";

type Project = { id: string; name: string };
type Resource = { id: string; name: string };

const projects: Project[] = [
  { id: "p1", name: "Acme" },
  { id: "p2", name: "Beta" },
];
const resources: Resource[] = [
  { id: "r1", name: "Anvil" },
  { id: "r2", name: "Bridges" },
  { id: "r3", name: "Cori" },
];

function score(p: Project, r: Resource): number {
  // Deterministic per-cell value; band is computed externally.
  const table: Record<string, number> = {
    "p1:r1": 0.5,
    "p1:r2": 0.78,
    "p1:r3": 0.95,
    "p2:r1": 0.0,
    "p2:r2": 0.4,
    "p2:r3": 0.82,
  };
  return table[`${p.id}:${r.id}`] ?? 0;
}

function bandOf(value: number): ComplianceCell["band"] {
  if (value < 0.01) return "empty";
  if (value >= 0.9) return "hot";
  if (value >= 0.75) return "warn";
  return "ok";
}

describe("ComplianceMatrix", () => {
  it("renders an N×M grid with row labels and column headers", () => {
    render(
      <ComplianceMatrix
        rows={projects}
        cols={resources}
        rowLabel={(p) => p.name}
        colLabel={(r) => r.name}
        cell={(p, r) => ({ value: score(p, r), band: bandOf(score(p, r)) })}
      />,
    );
    // Two row headers.
    expect(screen.getByRole("rowheader", { name: "Acme" })).toBeInTheDocument();
    expect(screen.getByRole("rowheader", { name: "Beta" })).toBeInTheDocument();
    // Three column headers (plus the corner empty <th>).
    expect(screen.getByRole("columnheader", { name: "Anvil" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "Bridges" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "Cori" })).toBeInTheDocument();
  });

  it("applies the right band per scoreFn output", () => {
    const { container } = render(
      <ComplianceMatrix
        rows={projects}
        cols={resources}
        rowLabel={(p) => p.name}
        colLabel={(r) => r.name}
        cell={(p, r) => ({ value: score(p, r), band: bandOf(score(p, r)) })}
      />,
    );
    const bands = Array.from(container.querySelectorAll("[data-band]")).map((n) =>
      n.getAttribute("data-band"),
    );
    // Order: p1×{r1,r2,r3} then p2×{r1,r2,r3}.
    expect(bands).toEqual(["ok", "warn", "hot", "empty", "ok", "warn"]);
  });

  it("fires onCellClick with the right row + col args", () => {
    const onCellClick = vi.fn();
    render(
      <ComplianceMatrix
        rows={projects}
        cols={resources}
        rowLabel={(p) => p.name}
        colLabel={(r) => r.name}
        cell={(p, r) => ({ value: score(p, r), band: bandOf(score(p, r)) })}
        onCellClick={onCellClick}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Acme — Cori: 95%/ }));
    expect(onCellClick).toHaveBeenCalledTimes(1);
    expect(onCellClick).toHaveBeenCalledWith(projects[0], resources[2]);
  });
});
