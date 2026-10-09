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

import failedFixture from "@/features/core/audit/__fixtures__/trace.amie.failed.json";
import {
  buildTree,
  detectErrorPath,
  flattenTree,
  getEntityRefs,
  subtreeHasError,
  traceView,
} from "@/features/core/audit/utils";
import type { TraceEvent } from "@/generated/core/types.gen";
import { zGetAuditTracesByTraceIdResponse } from "@/generated/core/zod.gen";
import { describe, expect, it } from "vitest";

function span(partial: Partial<TraceEvent> & { span_id: string }): TraceEvent {
  return { id: partial.span_id, event_type: "op", status: "ok", source: "core", created_at: "2026-06-08T14:00:00Z", ...partial };
}

describe("traceView", () => {
  it("flattens the tree depth-first and summarises the root", () => {
    const { trace, spans } = traceView(zGetAuditTracesByTraceIdResponse.parse(failedFixture));
    expect(spans.map((s) => s.span_id)).toEqual([
      "0123456789abcdef",
      "11111111aaaaaaaa",
      "22222222bbbbbbbb",
    ]);
    expect(trace).toMatchObject({
      root_operation: "amie.process_event:request_account_create",
      status: "error",
      event_count: 3,
      ended_at: "2026-06-08T14:21:33.412Z",
    });
  });

  it("ends the trace at the latest span, not the last visited", () => {
    const { trace } = traceView({
      trace_id: "t",
      status: "ok",
      truncated: false,
      deliveries: [],
      tree: [
        {
          ...span({ span_id: "a" }),
          children: [
            { ...span({ span_id: "b", created_at: "2026-06-08T14:00:05Z" }), children: [] },
            { ...span({ span_id: "c", created_at: "2026-06-08T14:00:03Z" }), children: [] },
          ],
        },
      ],
    });
    expect(trace?.ended_at).toBe("2026-06-08T14:00:05.000Z");
  });

  it("counts deliveries like the trace list, attempts from the pending ones", () => {
    const delivery = (status: "PENDING" | "SUCCEEDED" | "FAILED", attempts: number) => ({
      id: `${status}-${attempts}`,
      event_type: "e",
      subscriber: "s",
      status,
      attempts,
      next_run_at: "2026-06-08T14:00:00Z",
      span_id: "a",
    });
    const { trace } = traceView({
      trace_id: "t",
      status: "in_progress",
      truncated: false,
      deliveries: [delivery("PENDING", 3), delivery("PENDING", 1), delivery("SUCCEEDED", 1), delivery("FAILED", 5)],
      tree: [{ ...span({ span_id: "a" }), children: [] }],
    });
    expect(trace?.deliveries).toEqual({ pending: 2, succeeded: 1, failed: 1, attempts: 3 });
  });
});

describe("detectErrorPath", () => {
  it("walks an error leaf up to the root, collecting every ancestor", () => {
    const root = span({ span_id: "r", status: "error" });
    const mid = span({ span_id: "m", parent_span_id: "r", status: "error" });
    const leaf = span({ span_id: "l", parent_span_id: "m", status: "error" });
    const { pathSet, errorLeafIds } = detectErrorPath([root, mid, leaf]);
    expect(pathSet.has("r")).toBe(true);
    expect(pathSet.has("m")).toBe(true);
    expect(pathSet.has("l")).toBe(true);
    expect(errorLeafIds).toEqual(["l"]);
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
    const retry = span({ span_id: "rt", parent_span_id: "r", event_type: "retry:something" });
    const { roots } = buildTree([root, a, retry]);
    expect(roots).toHaveLength(1);
    const r0 = roots[0];
    if (!r0) throw new Error("expected a root");
    expect(r0.children.map((c) => c.span.span_id).sort()).toEqual(["a", "rt"]);
  });

  it("treats a span whose parent wrote no row as a root", () => {
    const unaudited = span({ span_id: "o", parent_span_id: "missing" });
    const real = span({ span_id: "r" });
    const { roots } = buildTree([unaudited, real]);
    expect(roots).toHaveLength(2);
    expect(roots.map((n) => n.span.span_id).sort()).toEqual(["o", "r"]);
  });
});

describe("subtreeHasError", () => {
  it("returns true when a descendant is errored", () => {
    const root = span({ span_id: "r", status: "ok" });
    const errChild = span({ span_id: "c", parent_span_id: "r", status: "error" });
    const { byId } = buildTree([root, errChild]);
    const node = byId.get("r");
    if (!node) throw new Error("missing");
    expect(subtreeHasError(node)).toBe(true);
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
    const r = span({ span_id: "r", status: "error" });
    const ok = span({ span_id: "ok", parent_span_id: "r", status: "ok" });
    const err = span({ span_id: "err", parent_span_id: "r", status: "error" });
    const { roots } = buildTree([r, ok, err]);
    const { pathSet } = detectErrorPath([r, ok, err]);
    const rows = flattenTree(roots, new Set(["r"]), true, pathSet);
    expect(rows.map((v) => v.node.span.span_id).sort()).toEqual(["err", "r"]);
  });

  it("errorsOnly still respects collapsed subtrees", () => {
    const r = span({ span_id: "r", status: "error" });
    const err = span({ span_id: "err", parent_span_id: "r", status: "error" });
    const { roots } = buildTree([r, err]);
    const { pathSet } = detectErrorPath([r, err]);
    const rows = flattenTree(roots, new Set(), true, pathSet);
    expect(rows.map((v) => v.node.span.span_id)).toEqual(["r"]);
  });
});

describe("getEntityRefs", () => {
  it("keeps one event per entity and skips spans without one", () => {
    const refs = getEntityRefs([
      span({ span_id: "a", entity_type: "packet", entity_id: "pkt_1" }),
      span({ span_id: "b", entity_type: "packet", entity_id: "pkt_1" }),
      span({ span_id: "c" }),
    ]);
    expect(refs.map((r) => r.span_id)).toEqual(["a"]);
  });
});
