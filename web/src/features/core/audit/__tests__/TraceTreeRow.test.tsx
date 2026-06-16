import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { TraceTreeRow } from "@/features/core/audit/components/TraceTreeRow";
import type { UISpan } from "@/features/core/audit/types";
import { buildTree, type VisibleRow } from "@/features/core/audit/utils";

function makeSpan(p: Partial<UISpan> & { span_id: string }): UISpan {
  return {
    span_id: p.span_id,
    parent_span_id: p.parent_span_id,
    name: p.name ?? "op",
    kind: p.kind ?? 0,
    status: p.status ?? 0,
    status_message: p.status_message ?? null,
    start_time: p.start_time ?? "2026-06-04T00:00:00.000Z",
    end_time: "end_time" in p ? (p.end_time as string | null) : "2026-06-04T00:00:00.100Z",
    attributes: p.attributes ?? null,
    running: p.running,
    notRun: p.notRun,
    orphan: p.orphan,
  } as UISpan;
}

function asVisibleRow(spans: UISpan[]): { row: VisibleRow } {
  const { roots } = buildTree(spans);
  const r0 = roots[0];
  if (!r0) throw new Error("expected a root");
  return { row: { node: r0, depth: r0.depth, hasChildren: r0.children.length > 0 } };
}

describe("TraceTreeRow", () => {
  function commonProps() {
    return {
      tone: "ok" as const,
      source: "amie",
      isSelected: false,
      isOnErrorPath: false,
      isPreciseFailure: false,
      isExpanded: false,
      hasHiddenError: false,
      onSelect: vi.fn(),
      onToggle: vi.fn(),
      rowRef: vi.fn(),
    };
  }

  it("indents per depth and prefixes with …/ when depth exceeds the cap", () => {
    const root = makeSpan({ span_id: "r" });
    const d1 = makeSpan({ span_id: "d1", parent_span_id: "r" });
    const d2 = makeSpan({ span_id: "d2", parent_span_id: "d1" });
    const d3 = makeSpan({ span_id: "d3", parent_span_id: "d2" });
    const d4 = makeSpan({ span_id: "d4", parent_span_id: "d3" });
    const d5 = makeSpan({ span_id: "d5", parent_span_id: "d4" });
    const d6 = makeSpan({ span_id: "d6", parent_span_id: "d5", name: "deep.thing" });
    const { roots } = buildTree([root, d1, d2, d3, d4, d5, d6]);
    let n = roots[0];
    while (n?.children[0]) n = n.children[0];
    if (!n || n.span.span_id !== "d6") throw new Error("expected to reach d6");
    render(
      <TraceTreeRow
        {...commonProps()}
        row={{ node: n, depth: n.depth, hasChildren: false }}
      />,
    );
    expect(screen.getByText("…/deep.thing")).toBeInTheDocument();
  });

  it("renders the status dot per tone (precise failure shows red wash + leading bar)", () => {
    const root = makeSpan({ span_id: "r", status: 1 });
    const { row } = asVisibleRow([root]);
    const { container } = render(
      <TraceTreeRow {...commonProps()} row={row} tone="error" isPreciseFailure />,
    );
    const node = container.querySelector('[data-precise="true"]');
    expect(node).not.toBeNull();
    const bar = node?.querySelector('span[aria-hidden="true"]');
    expect(bar).toBeTruthy();
    const style = (node as HTMLElement)?.getAttribute("style") ?? "";
    expect(style).toMatch(/var\(--tone-error-bg\)/);
  });

  it("ancestor error row keeps calm — no precise treatment", () => {
    const root = makeSpan({ span_id: "r", status: 1 });
    const child = makeSpan({ span_id: "c", parent_span_id: "r", status: 1 });
    const { roots } = buildTree([root, child]);
    const ancestor = roots[0];
    if (!ancestor) throw new Error("missing");
    const { container } = render(
      <TraceTreeRow
        {...commonProps()}
        row={{ node: ancestor, depth: 0, hasChildren: true }}
        tone="error"
        isOnErrorPath
        isPreciseFailure={false}
      />,
    );
    const node = container.querySelector('[data-on-error-path="true"]');
    expect(node).not.toBeNull();
    expect(node?.getAttribute("data-precise")).toBeNull();
    const style = (node as HTMLElement)?.getAttribute("style") ?? "";
    expect(style).not.toMatch(/tone-error-bg/);
  });

  it("shows the (orphan) tag when span.orphan is true", () => {
    const orphan = makeSpan({ span_id: "o", parent_span_id: "missing", orphan: true });
    const { row } = asVisibleRow([orphan]);
    render(<TraceTreeRow {...commonProps()} row={row} />);
    expect(screen.getByText("(orphan)")).toBeInTheDocument();
  });

  it("renders the hidden-error badge on a collapsed parent with an errored descendant", () => {
    const root = makeSpan({ span_id: "r" });
    const { row } = asVisibleRow([root]);
    render(<TraceTreeRow {...commonProps()} row={row} hasHiddenError />);
    expect(screen.getByTitle("Contains a failed span")).toBeInTheDocument();
  });
});
