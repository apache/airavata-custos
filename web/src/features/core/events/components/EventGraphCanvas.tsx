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

"use client";

import { cn } from "@/lib/utils";
import * as React from "react";
import {
  END_GAP,
  endLines,
  entityOf,
  type LaidEdge,
  type LaidNode,
  type Layout,
  LINE,
  SIZE,
  verbOf,
} from "../graph";
import type { GraphTerm } from "../types";

export const COLOR = {
  core: "var(--tone-accent-fg)",
  conn: "var(--tone-warn-fg)",
  entry: "var(--tone-info-fg)",
  end: "var(--tone-warn-fg)",
} as const;

export type Pick =
  | { type: "node"; id: string }
  | { type: "edge"; id: string }
  | { type: "term"; from: string; term: GraphTerm };

export type Lit = { edges: Set<string> | null; nodes: Set<string> | null };

// The size the layout measured with; the canvas renders in the mono font throughout.
const font = (key: keyof typeof SIZE): React.CSSProperties => ({
  fontSize: SIZE[key],
  fontWeight: key === "label" ? 400 : 600,
});

type View = { k: number; x: number; y: number };

// Below this scale the in-block titles are too small to read.
const ZOOMED_OUT = 0.7;

// Fills the canvas height and starts at the left; the rest pans into view.
function fitView(layout: Layout, el: SVGSVGElement): View {
  const { clientWidth: w, clientHeight: h } = el;
  const k = Math.min(1.5, (h - 8) / layout.height);
  const x = k * layout.width > w - 8 ? 4 : (w - k * layout.width) / 2;
  return { k, x, y: (h - k * layout.height) / 2 };
}

const groupOf = (n: LaidNode) => entityOf(n.kind === "end" ? n.from : n.id);

// One hue per block, so neighbouring entities read apart in both themes.
const HUES = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--tone-error-fg)",
  "var(--chart-5)",
];
const tint = (hue: string, pct: number) => `color-mix(in oklab, ${hue} ${pct}%, var(--background))`;

export function EventGraphCanvas({
  layout,
  lit,
  marked,
  onPick,
  onClear,
}: {
  layout: Layout;
  lit: Lit;
  marked: Set<string>;
  onPick: (p: Pick) => void;
  onClear: () => void;
}) {
  const svg = React.useRef<SVGSVGElement>(null);
  const [view, setView] = React.useState<View>({ k: 1, x: 0, y: 0 });
  const drag = React.useRef<{ x: number; y: number; moved: boolean } | null>(null);

  // Pans stop where the content's edge meets the canvas margin; content that fits stays centred.
  const pan = React.useCallback(
    (dx: number, dy: number) => {
      const el = svg.current;
      if (!el) return;
      const axis = (at: number, content: number, size: number) =>
        content + 8 <= size ? (size - content) / 2 : Math.min(4, Math.max(size - 4 - content, at));
      setView((v) => ({
        k: v.k,
        x: axis(v.x + dx, v.k * layout.width, el.clientWidth),
        y: axis(v.y + dy, v.k * layout.height, el.clientHeight),
      }));
    },
    [layout],
  );
  const refit = React.useCallback(() => {
    if (svg.current) setView(fitView(layout, svg.current));
  }, [layout]);

  React.useLayoutEffect(() => {
    const el = svg.current;
    if (!el) return;
    refit();
    const observer = new ResizeObserver(refit);
    observer.observe(el);
    return () => observer.disconnect();
  }, [refit]);

  // React registers wheel listeners as passive, so panning binds natively to stop page scroll.
  React.useEffect(() => {
    const el = svg.current;
    if (!el) return;
    const onWheel = (e: WheelEvent) => {
      e.preventDefault();
      pan(-e.deltaX, -e.deltaY);
    };
    el.addEventListener("wheel", onWheel, { passive: false });
    return () => el.removeEventListener("wheel", onWheel);
  }, [pan]);

  const order = layout.groups.map((g) => g.label).sort();
  const hue = (label: string) => HUES[order.indexOf(label) % HUES.length] ?? "var(--chart-5)";
  const litGroups = lit.nodes
    ? new Set(layout.nodes.filter((n) => lit.nodes?.has(n.id) && n.kind !== "entry").map(groupOf))
    : null;
  const dim = (set: Set<string> | null, id: string) =>
    set && !set.has(id) ? "opacity-[.12]" : undefined;
  const picked = (p: Pick) => (e: React.MouseEvent) => {
    e.stopPropagation();
    onPick(p);
  };

  return (
    <svg
      ref={svg}
      role="img"
      aria-label="Event state machine"
      className="block h-full w-full cursor-grab touch-none select-none font-mono"
      onPointerDown={(e) => {
        drag.current = { x: e.clientX, y: e.clientY, moved: false };
      }}
      onPointerMove={(e) => {
        const d = drag.current;
        if (!d || (!d.moved && Math.hypot(e.clientX - d.x, e.clientY - d.y) < 3)) return;
        // Capture only once moving, so a plain click still reaches the box under it.
        if (!d.moved) e.currentTarget.setPointerCapture(e.pointerId);
        pan(e.clientX - d.x, e.clientY - d.y);
        drag.current = { x: e.clientX, y: e.clientY, moved: true };
      }}
      onClickCapture={(e) => {
        if (drag.current?.moved) e.stopPropagation();
        drag.current = null;
      }}
      onClick={onClear}
    >
      <defs>
        <style>{"@keyframes eg-flow { to { stroke-dashoffset: var(--eg-period); } }"}</style>
        {(["core", "conn", "entry"] as const).map((k) => (
          <marker
            key={k}
            id={`eg-${k}`}
            viewBox="0 -4 8 8"
            refX={8}
            refY={0}
            markerWidth={7}
            markerHeight={7}
            orient="auto"
          >
            <path d="M0,-4L8,0L0,4" style={{ fill: COLOR[k] }} />
          </marker>
        ))}
      </defs>
      <g transform={`translate(${view.x},${view.y}) scale(${view.k})`}>
        {layout.groups.map((g) => (
          <g key={g.id} className={dim(litGroups, g.label)}>
            <rect
              x={g.x}
              y={g.y}
              width={g.width}
              height={g.height}
              rx={10}
              style={{ fill: tint(hue(g.label), 9), stroke: tint(hue(g.label), 55) }}
            />
            <text
              x={g.x + 12}
              y={g.y + 21}
              visibility={view.k < ZOOMED_OUT ? "hidden" : undefined}
              style={{ ...font("group"), fill: tint(hue(g.label), 85) }}
            >
              {g.label}
            </text>
          </g>
        ))}
        {layout.edges.map((e) => (
          <EdgeView
            key={e.id}
            edge={e}
            className={dim(lit.edges, e.id)}
            flowing={lit.edges?.has(e.id) ?? true}
            onClick={picked(
              e.kind === "entry" ? { type: "node", id: e.to } : { type: "edge", id: e.id },
            )}
          />
        ))}
        {layout.nodes.map((n) => (
          <NodeView
            key={n.id}
            node={n}
            marked={marked.has(n.id)}
            className={dim(lit.nodes, n.id)}
            picked={picked}
          />
        ))}
        {layout.edges.map((e) =>
          e.label ? (
            <EdgeLabel key={`${e.id}:label`} edge={e} className={dim(lit.edges, e.id)} />
          ) : null,
        )}
        {view.k < ZOOMED_OUT
          ? // Zoomed out, block names keep a readable on-screen size above their blocks.
            layout.groups.map((g) => (
              <text
                key={`${g.id}:title`}
                x={g.x}
                y={g.y - 4 / view.k}
                className={cn("pointer-events-none stroke-background", dim(litGroups, g.label))}
                style={{
                  ...font("group"),
                  fill: tint(hue(g.label), 85),
                  fontSize: 12 / view.k,
                  strokeWidth: 4 / view.k,
                  paintOrder: "stroke",
                }}
              >
                {g.label}
              </text>
            ))
          : null}
      </g>
    </svg>
  );
}

const pathOf = (e: LaidEdge) => `M${e.points.map((p) => `${p.x},${p.y}`).join("L")}`;

function EdgeView({
  edge,
  className,
  flowing,
  onClick,
}: {
  edge: LaidEdge;
  className?: string | undefined;
  flowing: boolean;
  onClick: (e: React.MouseEvent) => void;
}) {
  const d = pathOf(edge);
  const cond = (edge.kind === "core" || edge.kind === "conn") && edge.transition.cond;
  // The line itself flows source to target: its dashes march; a solid arrow dashes only while flowing.
  const [dash, period] =
    edge.kind === "entry" ? ["4 3", 7] : cond ? ["2 3", 5] : flowing ? ["9 3", 12] : [undefined, 0];
  return (
    <g className={className}>
      <path
        d={d}
        fill="none"
        style={{ stroke: COLOR[edge.kind], ["--eg-period" as string]: `-${period}px` }}
        strokeWidth={edge.kind === "entry" ? 1.2 : 1.6}
        strokeDasharray={dash}
        markerEnd={edge.kind === "end" ? undefined : `url(#eg-${edge.kind})`}
        className={flowing ? "motion-safe:animate-[eg-flow_0.8s_linear_infinite]" : undefined}
      />
      {edge.kind === "end" ? null : (
        <path
          d={d}
          fill="none"
          stroke="transparent"
          strokeWidth={12}
          className="cursor-pointer"
          onClick={onClick}
        >
          <title>
            {edge.kind === "entry"
              ? edge.routes.join("\n")
              : edge.transition.multi.map((a) => a.actor).join("\n")}
          </title>
        </path>
      )}
    </g>
  );
}

function EdgeLabel({ edge, className }: { edge: LaidEdge; className?: string | undefined }) {
  if (!edge.label) return null;
  const { x, y, lines, width, height } = edge.label;
  return (
    <g transform={`translate(${x},${y})`} className={cn("pointer-events-none", className)}>
      <rect
        x={-width / 2}
        y={-height / 2}
        width={width}
        height={height}
        rx={3}
        className="fill-background/90"
      />
      <text textAnchor="middle" style={{ ...font("label"), fill: COLOR[edge.kind] }}>
        {lines.map((l, i) => (
          <tspan key={`${i}:${l}`} x={0} y={-height / 2 + 12 + i * LINE}>
            {l}
          </tspan>
        ))}
      </text>
    </g>
  );
}

function NodeView({
  node,
  marked,
  className,
  picked,
}: {
  node: LaidNode;
  marked: boolean;
  className?: string | undefined;
  picked: (p: Pick) => (e: React.MouseEvent) => void;
}) {
  const ring = marked ? (
    <rect
      x={-4}
      y={-4}
      width={node.width + 8}
      height={node.height + 8}
      rx={9}
      fill="none"
      className="stroke-brand"
      strokeWidth={2}
    />
  ) : null;
  if (node.kind === "event") {
    const { event } = node;
    return (
      <g
        transform={`translate(${node.x},${node.y})`}
        className={cn("cursor-pointer", className)}
        onClick={picked({ type: "node", id: node.id })}
      >
        <title>{event.id}</title>
        {ring}
        <rect
          width={node.width}
          height={node.height}
          rx={6}
          className="fill-card"
          style={{ stroke: event.entry.length ? COLOR.entry : "var(--color-muted-foreground)" }}
          strokeWidth={event.entry.length ? 2 : 1.2}
          strokeDasharray={event.subs ? undefined : "3 2"}
        />
        <text
          x={node.width / 2}
          y={node.height / 2 + 4.5}
          textAnchor="middle"
          className="fill-foreground"
          style={font("event")}
        >
          {verbOf(event.id)}
        </text>
      </g>
    );
  }
  if (node.kind === "entry") {
    return (
      <g
        transform={`translate(${node.x},${node.y})`}
        className={cn("cursor-pointer", className)}
        onClick={picked({ type: "node", id: node.id })}
      >
        {ring}
        <rect
          width={node.width}
          height={node.height}
          rx={14}
          className="fill-card"
          style={{ stroke: COLOR.entry }}
        />
        <text
          x={node.width / 2}
          y={node.height / 2 + 4.5}
          textAnchor="middle"
          style={{ ...font("entry"), fill: COLOR.entry }}
        >
          {node.label}
        </text>
      </g>
    );
  }
  return (
    <g transform={`translate(${node.x},${node.y})`} className={className}>
      <rect
        width={node.width}
        height={node.height}
        rx={4}
        className="fill-card"
        style={{ stroke: COLOR.end }}
        strokeDasharray="3 2"
      />
      {node.terms.map((term, i) => {
        const before = node.terms.slice(0, i).flatMap(endLines).length;
        const top = 2 + before * LINE + i * END_GAP;
        return (
          <text
            key={term.actor}
            className="cursor-pointer"
            style={{ ...font("label"), fill: COLOR.end }}
            onClick={picked({ type: "term", from: node.from, term })}
          >
            <title>{`${node.from} ⊣ ${term.actor}`}</title>
            {endLines(term).map((l, j) => (
              <tspan key={`${j}:${l}`} x={8} y={top + 13 + j * LINE} xmlSpace="preserve">
                {l}
              </tspan>
            ))}
          </text>
        );
      })}
    </g>
  );
}
