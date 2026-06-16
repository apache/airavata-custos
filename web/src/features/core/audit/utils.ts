import type { EntityRef, RowTone, Span, Trace, UISpan } from "./types";

// Truncate a hex id to a leading prefix for compact display in tables/badges.
export function shortHex(value: string | undefined | null, length = 8): string {
  if (!value) return "";
  return value.length > length ? value.slice(0, length) : value;
}

export function formatDurationMs(ms: number): string {
  if (!Number.isFinite(ms) || ms < 0) return "—";
  if (ms < 1) return "<1ms";
  if (ms < 1_000) return `${Math.round(ms)}ms`;
  const seconds = ms / 1_000;
  if (seconds < 60) return `${seconds.toFixed(seconds < 10 ? 2 : 1)}s`;
  const minutes = Math.floor(seconds / 60);
  const remaining = Math.round(seconds - minutes * 60);
  return `${minutes}m ${remaining}s`;
}

export function durationBetween(startIso: string, endIso?: string | null): number | null {
  if (!endIso) return null;
  const start = Date.parse(startIso);
  const end = Date.parse(endIso);
  if (Number.isNaN(start) || Number.isNaN(end)) return null;
  return Math.max(0, end - start);
}

export function formatAbsoluteUtc(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return `${d.toISOString().replace("T", " ").replace("Z", "")} UTC`;
}

const MIN_MS = 60 * 1000;
const HOUR_MS = 60 * MIN_MS;
const DAY_MS = 24 * HOUR_MS;

export function formatRelative(iso: string, now: number = Date.now()): string {
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return iso;
  const ageMs = Math.max(0, now - t);
  if (ageMs < MIN_MS) return "just now";
  if (ageMs < HOUR_MS) return `${Math.floor(ageMs / MIN_MS)} min ago`;
  if (ageMs < DAY_MS) return `${Math.floor(ageMs / HOUR_MS)}h ago`;
  return `${Math.floor(ageMs / DAY_MS)}d ago`;
}

// Sonner is imported lazily so tests don't have to stub it when loading utils.
export async function copyTraceId(traceId: string): Promise<void> {
  const { toast } = await import("sonner");
  try {
    await navigator.clipboard.writeText(traceId);
    toast.success("Copied", { description: shortHex(traceId, 12) });
  } catch {
    toast.error("Could not copy to clipboard");
  }
}

// Run-state flags win over wire status, which wins over name heuristics.
export function rowTone(row: UISpan): RowTone {
  if (row.running) return "in-progress";
  if (row.orphan && row.status == null) return "orphaned";
  if (row.notRun) return "no-status";
  if (row.status === 1) return "error";
  if (row.status === 0) return "ok";
  const a = row.name || "";
  if (/Failed$/.test(a) || /Error/.test(a) || row.status_message) return "error";
  if (row.status == null) return "no-status";
  return "ok";
}

// Trace-level tone: 0->ok, 1->error, 2->no-status (cancelled rendered muted),
// 3->orphaned. `ended_at == null` always overrides to in-progress so a live
// trace shows the amber pulse even before its terminal status is recorded.
export function traceTone(trace: Trace): RowTone {
  if (trace.ended_at == null) return "in-progress";
  switch (trace.status) {
    case 0:
      return "ok";
    case 1:
      return "error";
    case 2:
      return "no-status";
    case 3:
      return "orphaned";
    default:
      return "no-status";
  }
}

// Augment a wire span with the run-state flags rowTone needs. `notRun` is
// "skipped because parent failed": child sat null with no end_time under an
// errored ancestor.
export function enrichSpan(span: Span, byId: Map<string, Span>, parentIsError: boolean): UISpan {
  const running = span.end_time == null && !parentIsError;
  const orphan = span.parent_span_id != null && !byId.has(span.parent_span_id);
  const notRun = parentIsError && span.status == null && span.end_time == null;
  return { ...span, running, orphan, notRun };
}

// A code-shaped action has no whitespace and at least one `.` or `:` separator
// (so `comanage.create_person` is mono; `http.POST /users` falls back to sans).
export function isCodeShaped(action: string): boolean {
  if (/\s/.test(action)) return false;
  return /[.:]/.test(action);
}

// Walk every error span up to the root, collecting the path. An "error leaf"
// is an error span with no error descendant — the precise failing row.
export function detectErrorPath(spans: UISpan[]): {
  pathSet: Set<string>;
  errorLeafIds: string[];
} {
  const byId = new Map<string, UISpan>();
  for (const s of spans) byId.set(s.span_id, s);

  const childrenOf = new Map<string, UISpan[]>();
  for (const s of spans) {
    if (!s.parent_span_id) continue;
    const list = childrenOf.get(s.parent_span_id) ?? [];
    list.push(s);
    childrenOf.set(s.parent_span_id, list);
  }

  const errors = spans.filter((s) => rowTone(s) === "error");
  const pathSet = new Set<string>();
  for (const e of errors) {
    let cursor: UISpan | undefined = e;
    while (cursor && !pathSet.has(cursor.span_id)) {
      pathSet.add(cursor.span_id);
      cursor = cursor.parent_span_id ? byId.get(cursor.parent_span_id) : undefined;
    }
  }

  const errorLeafIds: string[] = [];
  for (const e of errors) {
    const kids = childrenOf.get(e.span_id) ?? [];
    const hasErrorDescendant = kids.some((k) => rowTone(k) === "error");
    if (!hasErrorDescendant) errorLeafIds.push(e.span_id);
  }

  return { pathSet, errorLeafIds };
}

export type TreeNode = {
  span: UISpan;
  depth: number;
  children: TreeNode[];
  parent: TreeNode | null;
};

// Join spans into a tree by parent_span_id. Orphans (parent set but
// unresolved) and true roots both surface as roots so retry siblings keep
// their place.
export function buildTree(spans: UISpan[]): {
  roots: TreeNode[];
  byId: Map<string, TreeNode>;
} {
  const byId = new Map<string, TreeNode>();
  for (const span of spans) {
    byId.set(span.span_id, { span, depth: 0, children: [], parent: null });
  }
  const roots: TreeNode[] = [];
  for (const span of spans) {
    const node = byId.get(span.span_id);
    if (!node) continue;
    const parentId = span.parent_span_id;
    const parent = parentId ? byId.get(parentId) : undefined;
    if (parent) {
      node.parent = parent;
      node.depth = parent.depth + 1;
      parent.children.push(node);
    } else {
      roots.push(node);
    }
  }
  // Settle depths for cases where a child was visited before its parent.
  const walk = (n: TreeNode, depth: number) => {
    n.depth = depth;
    for (const c of n.children) walk(c, depth + 1);
  };
  for (const r of roots) walk(r, 0);
  return { roots, byId };
}

export function subtreeHasError(node: TreeNode): boolean {
  if (rowTone(node.span) === "error") return true;
  for (const c of node.children) if (subtreeHasError(c)) return true;
  return false;
}

export type VisibleRow = { node: TreeNode; depth: number; hasChildren: boolean };

// Flatten honoring `expanded`. In errorsOnly mode, only rows whose span is on
// the error path are emitted — collapsed subtrees still elide their children.
export function flattenTree(
  roots: TreeNode[],
  expanded: Set<string>,
  errorsOnly: boolean,
  errorPathSet: Set<string>,
): VisibleRow[] {
  const out: VisibleRow[] = [];
  const walk = (node: TreeNode, depth: number) => {
    const hasChildren = node.children.length > 0;
    const onPath = errorPathSet.has(node.span.span_id);
    if (!errorsOnly || onPath) {
      out.push({ node, depth, hasChildren });
    }
    if (hasChildren && expanded.has(node.span.span_id)) {
      for (const c of node.children) walk(c, depth + 1);
    }
  };
  for (const r of roots) walk(r, 0);
  return out;
}

const ENTITY_ATTR_MAP: ReadonlyArray<{ attr: string; kind: string }> = [
  { attr: "amie.packet_id", kind: "AMIE packet" },
  { attr: "entity.user_id", kind: "User" },
  { attr: "entity.project_id", kind: "Project" },
  { attr: "project.id", kind: "Project" },
  { attr: "comanage.co_person_id", kind: "CO person" },
  { attr: "allocation.id", kind: "Allocation" },
  { attr: "slurm.account", kind: "Cluster account" },
];

function readAttr(span: Span, key: string): string | undefined {
  const attrs = span.attributes;
  if (!attrs || typeof attrs !== "object") return undefined;
  const raw = (attrs as Record<string, unknown>)[key];
  if (raw == null) return undefined;
  if (typeof raw === "string") return raw;
  if (typeof raw === "number" || typeof raw === "boolean") return String(raw);
  return undefined;
}

// Scan span attributes for the known entity keys and dedupe by
// `${kind}::${primaryId}` so the same packet/user surfacing across spans
// collapses to one card.
export function getEntityRefs(spans: Span[]): EntityRef[] {
  const out: EntityRef[] = [];
  const seen = new Set<string>();
  for (const span of spans) {
    for (const { attr, kind } of ENTITY_ATTR_MAP) {
      const value = readAttr(span, attr);
      if (!value) continue;
      const key = `${kind}::${value}`;
      if (seen.has(key)) continue;
      seen.add(key);
      out.push({ kind, primaryId: value, attrs: { [attr]: value } });
    }
  }
  return out;
}
