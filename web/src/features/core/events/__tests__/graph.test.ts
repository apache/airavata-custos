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

import { describe, expect, it } from "vitest";
import graphFixture from "../__fixtures__/graph.json";
import {
  DEFAULT_GRAPH_OPTIONS,
  entityOf,
  layoutGraph,
  neighbourhood,
  selectGraph,
  splitActor,
  traceGraph,
} from "../graph";
import { eventGraphSchema } from "../schemas";
import type { EventGraph, GraphEdge } from "../types";
import type { LaidEdge, Layout } from "../graph";

const edge = (
  from: string,
  to: string,
  actor: string,
  extra: Partial<GraphEdge> = {},
): GraphEdge => ({
  from,
  to,
  kind: "conn",
  conn: "c",
  actor,
  sites: [],
  chain: [],
  cond: false,
  ...extra,
});

const ev = (id: string, subs = 1) => ({ id, entry: [], subs });

describe("selectGraph", () => {
  it("drops arrows into dead ends", () => {
    const graph: EventGraph = {
      events: [ev("a::x"), ev("a::y"), ev("a::dead", 0), ev("a::z")],
      edges: [edge("a::x", "a::y", "c.H.f"), edge("a::x", "a::dead", "c.H.g")],
      terms: [],
    };
    const sel = selectGraph(graph, DEFAULT_GRAPH_OPTIONS);
    expect(sel.edges.map((e) => e.to)).toEqual(["a::y"]);
    expect(sel.events.map((e) => e.id)).toEqual(["a::x", "a::y"]);
  });

  it("merges handlers of one connector between the same two events", () => {
    const graph: EventGraph = {
      events: [ev("a::x"), ev("a::y")],
      edges: [edge("a::x", "a::y", "c.H.f", { cond: true }), edge("a::x", "a::y", "c.H.g")],
      terms: [],
    };
    const [merged] = selectGraph(graph, DEFAULT_GRAPH_OPTIONS).edges;
    expect(merged?.actor).toBe("c.H.f / c.H.g");
    expect(merged?.cond).toBe(false);
  });
});

describe("traceGraph", () => {
  const graph: EventGraph = {
    events: [
      { id: "a::x", entry: [{ kind: "API", label: "POST /a" }], subs: 1 },
      { id: "a::y", entry: [{ kind: "amie", label: "loop" }], subs: 1 },
      ev("a::z"),
      ev("b::w"),
    ],
    edges: [
      edge("a::x", "a::y", "c.H.f"),
      edge("a::y", "a::z", "c.H.g"),
      edge("b::w", "a::z", "c.H.h"),
    ],
    terms: [{ from: "a::z", conn: "c", actor: "c.H.end", site: "", effects: [], cond: false }],
  };
  const sel = selectGraph(graph, DEFAULT_GRAPH_OPTIONS);
  const ids = (s: ReturnType<typeof traceGraph>) => s.events.map((e) => e.id);

  it("keeps everything downstream of a trigger and only that trigger's arrows", () => {
    const down = traceGraph(sel, { from: "in:API" });
    expect(ids(down)).toEqual(["a::x", "a::y", "a::z"]);
    expect(down.events.flatMap((e) => e.entry.map((x) => x.kind))).toEqual(["API"]);
    expect(down.terms).toHaveLength(1);
  });

  it("keeps only the paths between two events and drops end steps past the end", () => {
    const path = traceGraph(sel, { from: "a::x", to: "a::z" });
    expect(ids(path)).toEqual(["a::x", "a::y", "a::z"]);
    expect(path.edges.map((e) => e.actor)).toEqual(["c.H.f", "c.H.g"]);
    expect(path.terms).toHaveLength(0);
    expect(path.events.every((e) => e.entry.length === 0)).toBe(true);
  });

  it("keeps everything upstream of an event, triggers included", () => {
    const up = traceGraph(sel, { to: "a::y" });
    expect(ids(up)).toEqual(["a::x", "a::y"]);
    expect(up.events.flatMap((e) => e.entry.map((x) => x.kind)).sort()).toEqual(["API", "amie"]);
  });

  it("is empty when nothing connects the two ends", () => {
    expect(ids(traceGraph(sel, { from: "a::z", to: "a::x" }))).toEqual([]);
  });
});

describe("splitActor", () => {
  it("puts the owner on one line and the function on the next", () => {
    expect(splitActor("vast-storage-provisioner.Subscriber.ensureUser")).toEqual([
      "vast-storage-provisioner.Subscriber",
      "ensureUser",
    ]);
    expect(splitActor("f")).toEqual(["", "f"]);
  });
});

describe("layoutGraph", () => {
  it("places every node inside its entity block and routes every arrow", async () => {
    const graph = eventGraphSchema.parse(graphFixture);
    const layout = await layoutGraph(selectGraph(graph, DEFAULT_GRAPH_OPTIONS));
    const block = new Map(layout.groups.map((g) => [g.label, g]));
    for (const n of layout.nodes) {
      if (n.kind === "entry") continue;
      const g = block.get(entityOf(n.kind === "end" ? n.from : n.id));
      expect(g).toBeDefined();
      if (!g) continue;
      expect(n.x).toBeGreaterThanOrEqual(g.x);
      expect(n.y).toBeGreaterThanOrEqual(g.y);
      expect(n.x + n.width).toBeLessThanOrEqual(g.x + g.width);
      expect(n.y + n.height).toBeLessThanOrEqual(g.y + g.height);
    }
    expect(layout.nodes.filter((n) => n.kind === "entry").map((n) => n.id)).toContain("in:API");
    for (const e of layout.edges) expect(e.points.length).toBeGreaterThanOrEqual(2);
    const lit = neighbourhood(layout, "compute_allocation::create");
    expect(lit.nodes).toContain("in:API");
    expect(lit.nodes).toContain("end:compute_allocation::create");
  });
});

describe("neighbourhood", () => {
  it("lights every path into and out of a node", () => {
    const edge = (from: string, to: string, kind: "entry" | "end" = "entry") =>
      ({ id: `${from}>${to}`, from, to, kind, routes: [], points: [], label: null }) as LaidEdge;
    const layout: Layout = {
      width: 0,
      height: 0,
      groups: [],
      nodes: [],
      edges: [
        edge("in:API", "a"),
        edge("a", "b"),
        edge("x", "b"),
        edge("b", "c"),
        edge("b", "end:b", "end"),
      ],
    };
    expect([...neighbourhood(layout, "b").nodes].sort()).toEqual([
      "a",
      "b",
      "c",
      "end:b",
      "in:API",
      "x",
    ]);
    expect([...neighbourhood(layout, "a").nodes].sort()).toEqual([
      "a",
      "b",
      "c",
      "end:b",
      "in:API",
    ]);
  });
});
