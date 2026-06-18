import { describe, expect, it } from "vitest";
import type { Span, Trace, UISpan } from "@/features/core/audit/types";
import {
  buildTree,
  detectErrorPath,
  enrichSpan,
  flattenTree,
  getEntityRefs,
  isCodeShaped,
  rowTone,
  subtreeHasError,
  traceTone,
} from "@/features/core/audit/utils";

function span(partial: Partial<UISpan> & { span_id: string }): UISpan {
  const out: Record<string, unknown> = {
    span_id: partial.span_id,
    parent_span_id: partial.parent_span_id,
    name: "name" in partial ? partial.name : "op",
    kind: "kind" in partial ? partial.kind : 0,
    status: "status" in partial ? partial.status : 0,
    status_message: "status_message" in partial ? partial.status_message : null,
    start_time: "start_time" in partial ? partial.start_time : "2026-06-04T00:00:00.000Z",
    end_time: "end_time" in partial ? partial.end_time : "2026-06-04T00:00:00.010Z",
    attributes: "attributes" in partial ? partial.attributes : null,
    running: partial.running,
    notRun: partial.notRun,
    orphan: partial.orphan,
  };
  return out as UISpan;
}

function trace(partial: Partial<Trace> & { trace_id: string }): Trace {
  const out: Record<string, unknown> = {
    trace_id: partial.trace_id,
    root_name: "root_name" in partial ? partial.root_name : "op",
    source: "source" in partial ? partial.source : "amie",
    status: "status" in partial ? partial.status : 0,
    started_at: "started_at" in partial ? partial.started_at : "2026-06-04T00:00:00.000Z",
    ended_at: "ended_at" in partial ? partial.ended_at : "2026-06-04T00:00:00.500Z",
    span_count: "span_count" in partial ? partial.span_count : 1,
    root_event: "root_event" in partial ? partial.root_event : null,
  };
  return out as Trace;
}

describe("rowTone", () => {
  it("returns in-progress when running flag is set", () => {
    expect(rowTone(span({ span_id: "a", running: true, status: 0 }))).toBe("in-progress");
  });
  it("returns orphaned when orphan flag is set and status is null", () => {
    expect(
      rowTone(span({ span_id: "a", orphan: true, status: null as unknown as number })),
    ).toBe("orphaned");
  });
  it("returns no-status when notRun flag is set", () => {
    expect(
      rowTone(span({ span_id: "a", notRun: true, status: null as unknown as number })),
    ).toBe("no-status");
  });
  it("returns error when status code is 1", () => {
    expect(rowTone(span({ span_id: "a", status: 1 }))).toBe("error");
  });
  it("returns ok when status code is 0", () => {
    expect(rowTone(span({ span_id: "a", status: 0 }))).toBe("ok");
  });
  it("name ending with Failed yields error when status is null", () => {
    expect(
      rowTone(
        span({
          span_id: "a",
          name: "ComanageProvisioningFailed",
          status: null as unknown as number,
        }),
      ),
    ).toBe("error");
  });
  it("name containing Error yields error when status is null", () => {
    expect(
      rowTone(
        span({ span_id: "a", name: "SlurmCommandError", status: null as unknown as number }),
      ),
    ).toBe("error");
  });
  it("status_message presence yields error when status is null", () => {
    expect(
      rowTone(
        span({
          span_id: "a",
          name: "calm.name",
          status: null as unknown as number,
          status_message: "boom",
        }),
      ),
    ).toBe("error");
  });
  it("returns no-status when status is null and no heuristic matches", () => {
    expect(
      rowTone(span({ span_id: "a", name: "calm.name", status: null as unknown as number })),
    ).toBe("no-status");
  });
  it("explicit status 0 wins over heuristic-eligible names", () => {
    expect(rowTone(span({ span_id: "a", name: "noisyError", status: 0 }))).toBe("ok");
  });
});

describe("traceTone", () => {
  it("maps status 0 to ok", () => {
    expect(traceTone(trace({ trace_id: "t1", status: 0 }))).toBe("ok");
  });
  it("maps status 1 to error", () => {
    expect(traceTone(trace({ trace_id: "t1", status: 1 }))).toBe("error");
  });
  it("maps status 2 to no-status (cancelled gets muted treatment)", () => {
    expect(traceTone(trace({ trace_id: "t1", status: 2 }))).toBe("no-status");
  });
  it("maps status 3 to orphaned", () => {
    expect(traceTone(trace({ trace_id: "t1", status: 3 }))).toBe("orphaned");
  });
  it("overrides to in-progress when ended_at is null regardless of status", () => {
    expect(traceTone(trace({ trace_id: "t1", status: 0, ended_at: null }))).toBe("in-progress");
    expect(traceTone(trace({ trace_id: "t1", status: 1, ended_at: null }))).toBe("in-progress");
  });
});

describe("enrichSpan", () => {
  it("marks orphan when parent_span_id is set but unresolved", () => {
    const child = span({ span_id: "c", parent_span_id: "missing" }) as Span;
    const byId = new Map<string, Span>([[child.span_id, child]]);
    expect(enrichSpan(child, byId, false).orphan).toBe(true);
  });
  it("does not mark orphan when parent resolves", () => {
    const parent = span({ span_id: "p" }) as Span;
    const child = span({ span_id: "c", parent_span_id: "p" }) as Span;
    const byId = new Map<string, Span>([
      [parent.span_id, parent],
      [child.span_id, child],
    ]);
    expect(enrichSpan(child, byId, false).orphan).toBe(false);
  });
  it("marks running when end_time is null and parent isn't errored", () => {
    const s = span({ span_id: "a", end_time: null }) as Span;
    const byId = new Map<string, Span>([[s.span_id, s]]);
    expect(enrichSpan(s, byId, false).running).toBe(true);
  });
  it("marks notRun when parent errored and the span has no status or end_time", () => {
    const s = span({
      span_id: "a",
      end_time: null,
      status: null as unknown as number,
    }) as Span;
    const byId = new Map<string, Span>([[s.span_id, s]]);
    const enriched = enrichSpan(s, byId, true);
    expect(enriched.notRun).toBe(true);
    expect(enriched.running).toBe(false);
  });
});

describe("isCodeShaped", () => {
  it("returns true for dotted identifiers", () => {
    expect(isCodeShaped("comanage.create_person")).toBe(true);
  });
  it("returns true for colon-separated identifiers", () => {
    expect(isCodeShaped("amie.process:request_account_create")).toBe(true);
  });
  it("returns false for phrases that contain whitespace", () => {
    expect(isCodeShaped("http.POST /users")).toBe(false);
    expect(isCodeShaped("phrase with spaces")).toBe(false);
  });
  it("returns false for plain bareword identifiers without . or :", () => {
    expect(isCodeShaped("create")).toBe(false);
  });
});

describe("detectErrorPath", () => {
  it("walks an error leaf up to the root, collecting every ancestor", () => {
    const root = span({ span_id: "r", parent_span_id: undefined, status: 1 });
    const mid = span({ span_id: "m", parent_span_id: "r", status: 1 });
    const leaf = span({ span_id: "l", parent_span_id: "m", status: 1 });
    const { pathSet, errorLeafIds } = detectErrorPath([root, mid, leaf]);
    expect(pathSet.has("r")).toBe(true);
    expect(pathSet.has("m")).toBe(true);
    expect(pathSet.has("l")).toBe(true);
    expect(errorLeafIds).toEqual(["l"]);
  });
  it("returns empty results when no errors are present", () => {
    const { pathSet, errorLeafIds } = detectErrorPath([
      span({ span_id: "r", status: 0 }),
      span({ span_id: "c", parent_span_id: "r", status: 0 }),
    ]);
    expect(pathSet.size).toBe(0);
    expect(errorLeafIds).toEqual([]);
  });
});

describe("buildTree", () => {
  it("joins children to parents via parent_span_id and computes depths", () => {
    const root = span({ span_id: "r" });
    const mid = span({ span_id: "m", parent_span_id: "r" });
    const leaf = span({ span_id: "l", parent_span_id: "m" });
    const { roots, byId } = buildTree([root, mid, leaf]);
    expect(roots).toHaveLength(1);
    const r0 = roots[0];
    if (!r0) throw new Error("expected a root");
    expect(r0.span.span_id).toBe("r");
    expect(r0.depth).toBe(0);
    expect(r0.children).toHaveLength(1);
    const midNode = byId.get("m");
    expect(midNode?.depth).toBe(1);
    expect(midNode?.parent?.span.span_id).toBe("r");
    expect(byId.get("l")?.depth).toBe(2);
  });

  it("keeps retry siblings as separate children — does not re-parent", () => {
    const root = span({ span_id: "r" });
    const a = span({ span_id: "a", parent_span_id: "r" });
    const retry = span({ span_id: "rt", parent_span_id: "r", name: "retry:something" });
    const { roots } = buildTree([root, a, retry]);
    expect(roots).toHaveLength(1);
    const r0 = roots[0];
    if (!r0) throw new Error("expected a root");
    expect(r0.children.map((c) => c.span.span_id).sort()).toEqual(["a", "rt"]);
  });

  it("treats spans with unresolved parent_span_id as roots (orphan-style)", () => {
    const orphan = span({ span_id: "o", parent_span_id: "missing" });
    const real = span({ span_id: "r" });
    const { roots } = buildTree([orphan, real]);
    expect(roots).toHaveLength(2);
    expect(roots.map((n) => n.span.span_id).sort()).toEqual(["o", "r"]);
  });
});

describe("subtreeHasError", () => {
  it("returns true when a descendant is errored", () => {
    const root = span({ span_id: "r", status: 0 });
    const errChild = span({ span_id: "c", parent_span_id: "r", status: 1 });
    const { byId } = buildTree([root, errChild]);
    const node = byId.get("r");
    if (!node) throw new Error("missing");
    expect(subtreeHasError(node)).toBe(true);
  });
  it("returns false for an all-ok subtree", () => {
    const root = span({ span_id: "r", status: 0 });
    const child = span({ span_id: "c", parent_span_id: "r", status: 0 });
    const { byId } = buildTree([root, child]);
    const rNode = byId.get("r");
    if (!rNode) throw new Error("missing");
    expect(subtreeHasError(rNode)).toBe(false);
  });
});

describe("flattenTree", () => {
  it("emits only roots when nothing is expanded", () => {
    const r = span({ span_id: "r" });
    const c = span({ span_id: "c", parent_span_id: "r" });
    const { roots } = buildTree([r, c]);
    const rows = flattenTree(roots, new Set(), false, new Set());
    expect(rows.map((v) => v.node.span.span_id)).toEqual(["r"]);
    const r0 = rows[0];
    if (!r0) throw new Error("expected a row");
    expect(r0.hasChildren).toBe(true);
    expect(r0.depth).toBe(0);
  });

  it("walks into expanded subtrees with depth incremented", () => {
    const r = span({ span_id: "r" });
    const c = span({ span_id: "c", parent_span_id: "r" });
    const { roots } = buildTree([r, c]);
    const rows = flattenTree(roots, new Set(["r"]), false, new Set());
    expect(rows.map((v) => `${v.node.span.span_id}@${v.depth}`)).toEqual(["r@0", "c@1"]);
  });

  it("errorsOnly filters out rows not on the error path", () => {
    const r = span({ span_id: "r", status: 1 });
    const ok = span({ span_id: "ok", parent_span_id: "r", status: 0 });
    const err = span({ span_id: "err", parent_span_id: "r", status: 1 });
    const { roots } = buildTree([r, ok, err]);
    const { pathSet } = detectErrorPath([r, ok, err] as UISpan[]);
    const rows = flattenTree(roots, new Set(["r"]), true, pathSet);
    expect(rows.map((v) => v.node.span.span_id).sort()).toEqual(["err", "r"]);
  });

  it("errorsOnly still respects collapsed subtrees", () => {
    const r = span({ span_id: "r", status: 1 });
    const err = span({ span_id: "err", parent_span_id: "r", status: 1 });
    const { roots } = buildTree([r, err]);
    const { pathSet } = detectErrorPath([r, err] as UISpan[]);
    const rows = flattenTree(roots, new Set(), true, pathSet);
    expect(rows.map((v) => v.node.span.span_id)).toEqual(["r"]);
  });
});

describe("getEntityRefs", () => {
  it("dedupes the same entity surfacing on multiple spans", () => {
    const a = span({ span_id: "a", attributes: { "amie.packet_id": "pkt_1" } });
    const b = span({ span_id: "b", attributes: { "amie.packet_id": "pkt_1" } });
    const refs = getEntityRefs([a, b]);
    expect(refs).toHaveLength(1);
    expect(refs[0]).toMatchObject({ kind: "AMIE packet", primaryId: "pkt_1" });
  });
  it("handles spans with null attributes without crashing", () => {
    const a = span({ span_id: "a", attributes: null });
    expect(getEntityRefs([a])).toEqual([]);
  });
  it("extracts each known entity key", () => {
    const spans = [
      span({ span_id: "1", attributes: { "amie.packet_id": "pkt" } }),
      span({ span_id: "2", attributes: { "entity.user_id": "usr" } }),
      span({ span_id: "3", attributes: { "entity.project_id": "prj" } }),
      span({ span_id: "4", attributes: { "project.id": "prj2" } }),
      span({ span_id: "5", attributes: { "comanage.co_person_id": "co" } }),
      span({ span_id: "6", attributes: { "allocation.id": "alloc" } }),
      span({ span_id: "7", attributes: { "slurm.account": "TG-CCR" } }),
    ];
    const kinds = getEntityRefs(spans).map((r) => r.kind);
    expect(kinds).toEqual([
      "AMIE packet",
      "User",
      "Project",
      "Project",
      "CO person",
      "Allocation",
      "Cluster account",
    ]);
  });
});
