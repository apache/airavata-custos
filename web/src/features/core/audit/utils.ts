// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

import type {
  TraceDetailResponse,
  TraceEvent,
  TraceNode,
  TraceSummary,
} from "@/generated/core/types.gen";

export type RowTone = "ok" | "error" | "in-progress";

// The backend ships a recursive tree; the tree view joins a flat list by
// parent_span_id, and the header summarises the root like a list row.
export function traceView({ trace_id, status, tree, deliveries, truncated }: TraceDetailResponse): {
  trace: TraceSummary | undefined;
  spans: TraceEvent[];
  truncated: boolean;
} {
  const spans: TraceEvent[] = [];
  const walk = ({ children, ...event }: TraceNode) => {
    spans.push(event);
    children.forEach(walk);
  };
  tree.forEach(walk);
  const root = tree[0];
  const trace = root && {
    trace_id,
    root_operation: root.event_type,
    source: root.source,
    status,
    started_at: root.created_at,
    ended_at: new Date(Math.max(...spans.map((s) => Date.parse(s.created_at)))).toISOString(),
    event_count: spans.length,
    deliveries: {
      pending: deliveries.filter((d) => d.status === "PENDING").length,
      succeeded: deliveries.filter((d) => d.status === "SUCCEEDED").length,
      failed: deliveries.filter((d) => d.status === "FAILED").length,
      attempts: Math.max(0, ...deliveries.filter((d) => d.status === "PENDING").map((d) => d.attempts)),
    },
  };
  return { trace, spans, truncated };
}

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

export function traceTone({ status }: TraceSummary): RowTone {
  return status === "in_progress" ? "in-progress" : status;
}

// A code-shaped action has no whitespace and at least one `.` or `:` separator
// (so `comanage.create_person` is mono; `http.POST /users` falls back to sans).
export function isCodeShaped(action: string): boolean {
  if (/\s/.test(action)) return false;
  return /[.:]/.test(action);
}

// Walk every error span up to the root, collecting the path. An "error leaf"
// is an error span with no error descendant — the precise failing row.
export function detectErrorPath(spans: TraceEvent[]): {
  pathSet: Set<string>;
  errorLeafIds: string[];
} {
  const byId = new Map<string, TraceEvent>();
  for (const s of spans) byId.set(s.span_id, s);

  const childrenOf = new Map<string, TraceEvent[]>();
  for (const s of spans) {
    if (!s.parent_span_id) continue;
    const list = childrenOf.get(s.parent_span_id) ?? [];
    list.push(s);
    childrenOf.set(s.parent_span_id, list);
  }

  const errors = spans.filter((s) => s.status === "error");
  const pathSet = new Set<string>();
  for (const e of errors) {
    let cursor: TraceEvent | undefined = e;
    while (cursor && !pathSet.has(cursor.span_id)) {
      pathSet.add(cursor.span_id);
      cursor = cursor.parent_span_id ? byId.get(cursor.parent_span_id) : undefined;
    }
  }

  const errorLeafIds: string[] = [];
  for (const e of errors) {
    const kids = childrenOf.get(e.span_id) ?? [];
    const hasErrorDescendant = kids.some((k) => k.status === "error");
    if (!hasErrorDescendant) errorLeafIds.push(e.span_id);
  }

  return { pathSet, errorLeafIds };
}

export type TreeNode = {
  span: TraceEvent;
  depth: number;
  children: TreeNode[];
  parent: TreeNode | null;
};

// Join spans into a tree by parent_span_id. A span whose parent wrote no row is
// a root, as the backend treats it, so retry siblings keep their place.
export function buildTree(spans: TraceEvent[]): {
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
  if (node.span.status === "error") return true;
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

// One event per referenced entity, deduped across spans.
export function getEntityRefs(spans: TraceEvent[]): TraceEvent[] {
  const seen = new Set<string>();
  return spans.filter(({ entity_type, entity_id }) => {
    const key = `${entity_type}::${entity_id}`;
    if (!entity_id || seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}
