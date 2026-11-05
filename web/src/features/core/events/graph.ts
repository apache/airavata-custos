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

// Event state machine from GET /events/graph, laid out by ELK layered.
// Events are grouped in one block per entity (the part before `::`); each external trigger
// (API, a connector loop) is one box whose arrows carry the routes that start an event.
// Arrows are connector transitions (Connector.handler);
// an end card lists handlers that take an event and emit nothing.
// A trace keeps only what lies on a path from one node to another, so it lays out small and legible.
// All text is monospace, so a box is sized from its text length alone.
// Coordinates are absolute and top-left; ELK reports them relative to the enclosing block.

import ELK, { type ElkExtendedEdge, type ElkNode } from "elkjs/lib/elk.bundled.js";
import type { EventGraph, GraphEdge, GraphEvent, GraphTerm } from "./types";

export type GraphOptions = { connector: boolean; ends: boolean; entries: boolean };

export const DEFAULT_GRAPH_OPTIONS: GraphOptions = {
  connector: true,
  ends: true,
  entries: true,
};

// One arrow; handlers of one connector between the same two events share it.
type Transition = GraphEdge & { multi: GraphEdge[] };

export type Selected = { events: GraphEvent[]; edges: Transition[]; terms: GraphTerm[] };

export const entityOf = (id: string) => id.split("::")[0] ?? id;
export const verbOf = (id: string) => id.split("::")[1] ?? id;
export const triggerId = (kind: string) => `in:${kind}`;
const endId = (event: string) => `end:${event}`;

export function selectGraph(graph: EventGraph, o: GraphOptions): Selected {
  const byId = new Map(graph.events.map((e) => [e.id, e]));
  const terms = o.ends ? graph.terms : [];
  let edges = o.connector ? graph.edges : [];
  // Drop transitions into events nobody handles and that lead nowhere, until none remain.
  for (let n = -1; n !== edges.length; ) {
    n = edges.length;
    const out = new Set([...edges.map((e) => e.from), ...terms.map((t) => t.from)]);
    edges = edges.filter((e) => byId.get(e.to)?.subs || out.has(e.to));
  }
  const merged = new Map<string, Transition>();
  for (const e of edges) {
    const key = `${e.from}→${e.to}|${e.kind}|${e.conn}`;
    const m = merged.get(key);
    if (m) m.multi.push(e);
    else merged.set(key, { ...e, multi: [e] });
  }
  const transitions = [...merged.values()].map((e) =>
    e.multi.length > 1
      ? { ...e, actor: e.multi.map((a) => a.actor).join(" / "), cond: e.multi.every((a) => a.cond) }
      : e,
  );
  const used = new Set([
    ...transitions.flatMap((e) => [e.from, e.to]),
    ...terms.map((t) => t.from),
  ]);
  return {
    events: graph.events
      .filter((e) => used.has(e.id))
      .map((e) => (o.entries ? e : { ...e, entry: [] })),
    edges: transitions,
    terms,
  };
}

export type Trace = { from?: string | undefined; to?: string | undefined };

// Keeps what lies on some path from `from` and on some path to `to`; either end may be open.
export function traceGraph(sel: Selected, { from, to }: Trace): Selected {
  if (!from && !to) return sel;
  const next = new Map<string, string[]>();
  const prev = new Map<string, string[]>();
  const link = (a: string, b: string) => {
    next.set(a, [...(next.get(a) ?? []), b]);
    prev.set(b, [...(prev.get(b) ?? []), a]);
  };
  for (const e of sel.events) for (const x of e.entry) link(triggerId(x.kind), e.id);
  for (const t of sel.edges) link(t.from, t.to);
  for (const t of sel.terms) link(t.from, endId(t.from));
  const reach = (start: string | undefined, adj: Map<string, string[]>) => {
    if (!start) return null;
    const seen = new Set([start]);
    for (const queue = [start]; queue.length; ) {
      for (const n of adj.get(queue.shift() ?? "") ?? []) {
        if (!seen.has(n)) {
          seen.add(n);
          queue.push(n);
        }
      }
    }
    return seen;
  };
  const down = reach(from, next);
  const up = reach(to, prev);
  const keep = (id: string) => (!down || down.has(id)) && (!up || up.has(id));
  return {
    events: sel.events
      .filter((e) => keep(e.id))
      .map((e) => ({ ...e, entry: e.entry.filter((x) => keep(triggerId(x.kind))) })),
    edges: sel.edges.filter((t) => keep(t.from) && keep(t.to)),
    terms: sel.terms.filter((t) => keep(t.from) && keep(endId(t.from))),
  };
}

// A handler A.B.fn splits into where it lives (A.B) and the function (fn).
export function splitActor(actor: string): [string, string] {
  const i = actor.lastIndexOf(".");
  return i < 0 ? ["", actor] : [actor.slice(0, i), actor.slice(i + 1)];
}

// Long names break at camelCase and separator boundaries, so a label grows taller, not wider.
export function wrap(name: string, max = 16): string[] {
  const lines: string[] = [];
  for (const token of name.split(/(?=[A-Z])|(?<=[._-])/)) {
    const last = lines.length - 1;
    if (last >= 0 && (lines[last] ?? "").length + token.length <= max) lines[last] += token;
    else lines.push(token);
  }
  return lines;
}

export const SIZE = { event: 12.5, entry: 12, group: 13, label: 11 } as const;
const textWidth = (lines: string[], size: number) =>
  Math.max(0, ...lines.map((l) => l.length)) * size * 0.61;

export const LINE = 14;
// Gap between handlers on an end card.
export const END_GAP = 4;
const PAD = 8;

type Point = { x: number; y: number };
type Box = { x: number; y: number; width: number; height: number };

export type LaidNode = Box &
  (
    | { kind: "event"; id: string; event: GraphEvent }
    | { kind: "entry"; id: string; label: string }
    | { kind: "end"; id: string; from: string; terms: GraphTerm[] }
  );

export type LaidEdge = {
  id: string;
  from: string;
  to: string;
  points: Point[];
  label: (Point & { lines: string[]; width: number; height: number }) | null;
} & (
  | { kind: "conn"; transition: Transition }
  | { kind: "entry"; routes: string[] }
  | { kind: "end" }
);

export type Layout = {
  width: number;
  height: number;
  groups: (Box & { id: string; label: string })[];
  nodes: LaidNode[];
  edges: LaidEdge[];
};

type Spec = (LaidEdge extends infer E
  ? E extends LaidEdge
    ? Omit<E, "points" | "label">
    : never
  : never) & {
  lines: string[];
};

const elk = new ELK();

// Left to right, so each step of a chain is a column.
export async function layoutGraph(sel: Selected): Promise<Layout> {
  const base = {
    "elk.algorithm": "layered",
    "elk.direction": "RIGHT",
    "elk.edgeRouting": "ORTHOGONAL",
    "elk.hierarchyHandling": "INCLUDE_CHILDREN",
    "elk.layered.spacing.nodeNodeBetweenLayers": "12",
    "elk.layered.spacing.edgeNodeBetweenLayers": "6",
    "elk.spacing.nodeNode": "8",
    "elk.spacing.edgeNode": "8",
    "elk.spacing.edgeEdge": "4",
    "elk.spacing.edgeLabel": "3",
    // Labels sit above their arrow, so a long name adds height instead of stretching the arrow.
    "elk.edgeLabels.inline": "false",
    "elk.layered.edgeLabels.sideSelection": "ALWAYS_UP",
    "elk.layered.mergeEdges": "false",
    "elk.layered.nodePlacement.strategy": "NETWORK_SIMPLEX",
    // Pulls nodes along their layer to shorten long arrows; about 3% narrower at no cost.
    "elk.layered.compaction.postCompaction.strategy": "EDGE_LENGTH",
  };
  const meta = new Map<string, LaidNode>();
  const groups = new Map<string, ElkNode>();
  const parent = new Map<string, string>();
  const root: ElkNode = { id: "root", layoutOptions: base, children: [], edges: [] };
  const containers = new Map<string, ElkNode>([["root", root]]);

  const place = (node: LaidNode, entity?: string) => {
    meta.set(node.id, node);
    // One port in, one port out: arrows into a box meet at a single arrowhead, and arrows out leave as a trunk.
    const port = (side: string, dir: string) => ({
      id: `${node.id}:${dir}`,
      width: 0,
      height: 0,
      layoutOptions: { "elk.port.side": side },
    });
    const elkNode: ElkNode = {
      id: node.id,
      width: node.width,
      height: node.height,
      layoutOptions: { "elk.portConstraints": "FIXED_SIDE" },
      ports: [port("WEST", "in"), port("EAST", "out")],
    };
    if (!entity) {
      root.children?.push(elkNode);
      return;
    }
    const gid = `grp:${entity}`;
    let g = groups.get(gid);
    if (!g) {
      g = {
        id: gid,
        layoutOptions: {
          ...base,
          "elk.padding": "[top=32,left=10,bottom=10,right=10]",
          // Wide enough for the block's title.
          "elk.nodeSize.constraints": "MINIMUM_SIZE",
          "elk.nodeSize.minimum": `(${Math.ceil(textWidth([entity], SIZE.group) + 24)},0)`,
        },
        children: [],
        edges: [],
      };
      groups.set(gid, g);
      containers.set(gid, g);
      root.children?.push(g);
    }
    g.children?.push(elkNode);
    parent.set(node.id, gid);
  };

  const box = { x: 0, y: 0 };
  for (const event of sel.events) {
    const width = textWidth([verbOf(event.id)], SIZE.event) + 2 * PAD + 6;
    place({ ...box, kind: "event", id: event.id, event, width, height: 28 }, entityOf(event.id));
  }

  const specs: Spec[] = [];
  for (const event of sel.events) {
    const byKind = new Map<string, string[]>();
    for (const x of event.entry) byKind.set(x.kind, [...(byKind.get(x.kind) ?? []), x.label]);
    for (const [kind, routes] of byKind) {
      const id = triggerId(kind);
      if (!meta.has(id)) {
        place({
          ...box,
          kind: "entry",
          id,
          label: kind,
          width: textWidth([kind], SIZE.entry) + 2 * PAD + 8,
          height: 28,
        });
      }
      // A connector loop's root is a function, the same for every event it starts; only routes are worth a label.
      // Routes are "METHOD /path"; the method alone keeps the label short, the path is in the details.
      const methods = new Set(
        routes.filter((r) => r.includes(" /")).map((r) => r.split(" ")[0] ?? r),
      );
      const lines = methods.size ? [[...methods].join(" · ")] : [];
      specs.push({ id: `${id}→${event.id}`, kind: "entry", from: id, to: event.id, routes, lines });
    }
  }
  for (const t of sel.edges) {
    specs.push({
      id: `${t.from}→${t.to}|${t.actor}`,
      kind: t.kind,
      from: t.from,
      to: t.to,
      transition: t,
      lines: t.multi.flatMap((a) => wrap(splitActor(a.actor)[1])),
    });
  }
  const termsByEvent = new Map<string, GraphTerm[]>();
  for (const t of sel.terms) termsByEvent.set(t.from, [...(termsByEvent.get(t.from) ?? []), t]);
  for (const [from, terms] of termsByEvent) {
    if (!meta.has(from)) continue;
    const id = endId(from);
    const lines = terms.map(endLines);
    const width = textWidth(lines.flat(), SIZE.label) + 2 * PAD;
    const height = lines.flat().length * LINE + terms.length * END_GAP + 4;
    place({ ...box, kind: "end", id, from, terms, width, height }, entityOf(from));
    specs.push({ id: `${from}⊣`, kind: "end", from, to: id, lines: [] });
  }

  const unit = (id: string) => parent.get(id) ?? "root";
  const labelBox = (lines: string[]) => ({
    width: textWidth(lines, SIZE.label) + 6,
    height: lines.length * LINE + 2,
  });
  specs.forEach((s, i) => {
    const a = unit(s.from);
    const edge: ElkExtendedEdge = {
      id: `e${i}`,
      sources: [`${s.from}:out`],
      targets: [`${s.to}:in`],
      // ELK ignores a label whose text is empty, so it carries the real lines.
      labels: s.lines.length
        ? [{ id: `l${i}`, text: s.lines.join("\n"), ...labelBox(s.lines) }]
        : [],
    };
    containers.get(a === unit(s.to) ? a : "root")?.edges?.push(edge);
  });

  const out = await elk.layout(root);
  const absolute = new Map<string, Box>();
  const routed = new Map<string, ElkExtendedEdge>();
  const walk = (n: ElkNode, ox: number, oy: number) => {
    for (const e of n.edges ?? []) {
      for (const s of e.sections ?? []) {
        s.startPoint = { x: s.startPoint.x + ox, y: s.startPoint.y + oy };
        s.endPoint = { x: s.endPoint.x + ox, y: s.endPoint.y + oy };
        s.bendPoints = (s.bendPoints ?? []).map((p) => ({ x: p.x + ox, y: p.y + oy }));
      }
      for (const l of e.labels ?? []) {
        l.x = (l.x ?? 0) + ox;
        l.y = (l.y ?? 0) + oy;
      }
      routed.set(e.id, e);
    }
    for (const c of n.children ?? []) {
      const x = (c.x ?? 0) + ox;
      const y = (c.y ?? 0) + oy;
      absolute.set(c.id, { x, y, width: c.width ?? 0, height: c.height ?? 0 });
      walk(c, x, y);
    }
  };
  walk(out, 0, 0);

  const nodes = [...meta.values()].map((n) => ({ ...n, ...absolute.get(n.id) }) as LaidNode);
  const edges = specs.map((s, i) => {
    const { lines, ...rest } = s;
    const e = routed.get(`e${i}`);
    const sec = e?.sections?.[0];
    const points = sec ? [sec.startPoint, ...(sec.bendPoints ?? []), sec.endPoint] : [];
    const l = e?.labels?.[0];
    const at = { x: (l?.x ?? 0) + (l?.width ?? 0) / 2, y: (l?.y ?? 0) + (l?.height ?? 0) / 2 };
    return {
      ...rest,
      points,
      label: lines.length ? { ...at, ...labelBox(lines), lines } : null,
    } as LaidEdge;
  });

  return {
    width: out.width ?? 0,
    height: out.height ?? 0,
    groups: [...groups.keys()].map((id) => ({
      id,
      label: id.slice(4),
      ...(absolute.get(id) as Box),
    })),
    nodes,
    edges,
  };
}

export const endLines = (t: GraphTerm) =>
  wrap(splitActor(t.actor)[1]).map((l, i) => `${i ? "  " : "⊣ "}${l}`);

type Lit = { nodes: Set<string>; edges: Set<string> };

// Everything on a path into `id` ("up") or out of it ("down"), with the arrows between.
function reach(layout: Layout, id: string, dir: "up" | "down", lit: Lit) {
  lit.nodes.add(id);
  for (const queue = [id]; queue.length; ) {
    const at = queue.shift();
    for (const e of layout.edges) {
      const [near, far] = dir === "up" ? [e.to, e.from] : [e.from, e.to];
      if (near !== at) continue;
      lit.edges.add(e.id);
      if (!lit.nodes.has(far)) queue.push(lit.nodes.add(far) && far);
    }
  }
  return lit;
}

// What stays lit when a node is picked: every path through it, from its triggers to its end steps.
export function neighbourhood(layout: Layout, nodeId: string) {
  return reach(
    layout,
    nodeId,
    "down",
    reach(layout, nodeId, "up", { nodes: new Set(), edges: new Set() }),
  );
}

// A handler or arrow lights the paths into where it starts, itself, and the paths out of where it lands.
export function pathsThrough(layout: Layout, from: string, to: string, edgeId?: string) {
  const lit = reach(layout, from, "up", { nodes: new Set(), edges: new Set() });
  if (edgeId) lit.edges.add(edgeId);
  return reach(layout, to, "down", lit);
}
