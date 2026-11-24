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

import { Button } from "@/shared/ui/button";
import { ChevronDownIcon, ChevronUpIcon } from "lucide-react";
import * as React from "react";
import { type LaidEdge, type Layout, splitActor, type Trace } from "../graph";
import type { GraphTerm } from "../types";
import { COLOR, type Pick } from "./EventGraphCanvas";

export function EventGraphDetails({
  layout,
  pick,
  open,
  onOpen,
  onTrace,
}: {
  layout: Layout;
  pick: Pick;
  open: boolean;
  onOpen: (open: boolean) => void;
  onTrace: (trace: Trace) => void;
}) {
  // Share of the canvas height; the top edge drags or steps with the arrow keys.
  const [share, setShare] = React.useState(0.5);
  const resize = (next: number) => setShare(Math.min(0.9, Math.max(0.15, next)));
  const title =
    pick.type === "term"
      ? `${pick.from} ⊣ ${splitActor(pick.term.actor)[1]}`
      : pick.type === "edge"
        ? pick.id.split("|")[0]
        : pick.id.replace(/^in:/, "");
  if (!open) {
    return (
      <Button
        variant="outline"
        size="sm"
        className="absolute bottom-3 left-1/2 z-10 max-w-[60%] -translate-x-1/2 shadow-sm"
        onClick={() => onOpen(true)}
      >
        <ChevronUpIcon />
        <span className="truncate font-mono text-xs">{title}</span>
      </Button>
    );
  }
  return (
    <aside
      className="absolute inset-x-0 bottom-0 z-10 flex animate-in flex-col rounded-t-lg border-t bg-card text-sm shadow-lg duration-200 slide-in-from-bottom"
      style={{ height: `${share * 100}%` }}
      aria-label="Details"
    >
      {/* biome-ignore lint/a11y/useSemanticElements: a focusable, draggable separator is a widget; hr is static. */}
      <div
        role="separator"
        aria-orientation="horizontal"
        aria-label="Resize details"
        aria-valuenow={Math.round(share * 100)}
        tabIndex={0}
        className="flex h-3 shrink-0 cursor-row-resize touch-none items-center justify-center outline-none focus-visible:bg-muted"
        onPointerDown={(e) => e.currentTarget.setPointerCapture(e.pointerId)}
        onPointerMove={(e) => {
          const frame = e.currentTarget.parentElement?.parentElement;
          if (!frame || !e.currentTarget.hasPointerCapture(e.pointerId)) return;
          const r = frame.getBoundingClientRect();
          resize((r.bottom - e.clientY) / r.height);
        }}
        onKeyDown={(e) => {
          if (e.key === "ArrowUp" || e.key === "ArrowDown") {
            e.preventDefault();
            resize(share + (e.key === "ArrowUp" ? 0.05 : -0.05));
          }
        }}
      >
        <span className="h-1 w-10 rounded-full bg-border" />
      </div>
      <Button
        variant="ghost"
        size="icon-sm"
        aria-label="Close details"
        className="absolute top-1 right-2"
        onClick={() => onOpen(false)}
      >
        <ChevronDownIcon />
      </Button>
      <div className="min-h-0 flex-1 overflow-auto px-3 pb-3">
        {pick.type === "term" ? (
          <TermDetails from={pick.from} term={pick.term} />
        ) : pick.type === "edge" ? (
          <EdgeDetails edge={layout.edges.find((e) => e.id === pick.id)} />
        ) : (
          <NodeDetails layout={layout} id={pick.id} onTrace={onTrace} />
        )}
      </div>
    </aside>
  );
}

const Key = ({ children }: { children: React.ReactNode }) => (
  <div className="mt-2 text-xs text-muted-foreground">{children}</div>
);
const Mono = ({ children }: { children: React.ReactNode }) => (
  <code className="font-mono text-xs break-all">{children}</code>
);
const Dot = ({ kind }: { kind: keyof typeof COLOR }) => (
  <span style={{ color: COLOR[kind] }}>● </span>
);

function TermDetails({ from, term }: { from: string; term: GraphTerm }) {
  return (
    <>
      <h2 className="font-mono text-sm font-semibold break-all">
        {from} ⊣ {splitActor(term.actor)[1]}
      </h2>
      <div className="text-xs text-muted-foreground">
        connector end step{term.cond ? " · conditional subscription" : ""}
      </div>
      <Key>Handled by</Key>
      <div>
        <Dot kind="end" />
        <Mono>{term.actor}</Mono>
      </div>
      <div className="font-mono text-[11.5px] break-all text-muted-foreground">{term.site}</div>
      <Key>External effects</Key>
      {term.effects.length ? (
        <ul className="list-disc pl-5">
          {term.effects.map((x) => (
            <li key={x}>
              <Mono>{x}</Mono>
            </li>
          ))}
        </ul>
      ) : (
        <div className="text-xs text-muted-foreground">none recorded</div>
      )}
    </>
  );
}

function EdgeDetails({ edge }: { edge: LaidEdge | undefined }) {
  if (edge?.kind !== "conn") return null;
  const t = edge.transition;
  return (
    <>
      <h2 className="font-mono text-sm font-semibold break-all">
        {t.from} → {t.to}
      </h2>
      <div className="text-xs text-muted-foreground">
        connector transition · {t.conn}
        {t.cond ? " · conditional" : ""}
      </div>
      <Key>Performed by</Key>
      {t.multi.map((a) => (
        <div key={a.actor} className="mt-1.5">
          <Dot kind={a.kind} />
          <Mono>{a.actor}</Mono>
          {a.cond ? <span className="text-xs text-muted-foreground"> conditional</span> : null}
          <ul className="list-disc pl-5 font-mono text-[11.5px] break-all text-muted-foreground">
            {a.sites.map((s) => (
              <li key={s}>{s}</li>
            ))}
          </ul>
          <div className="font-mono text-[11.5px] break-all text-muted-foreground">
            {a.chain.join(" → ")}
          </div>
        </div>
      ))}
    </>
  );
}

function TraceButtons({
  id,
  onTrace,
  to = true,
}: {
  id: string;
  onTrace: (t: Trace) => void;
  to?: boolean;
}) {
  return (
    <div className="mt-2 flex gap-2">
      <Button variant="outline" size="xs" onClick={() => onTrace({ from: id })}>
        Trace from here
      </Button>
      {to ? (
        <Button variant="outline" size="xs" onClick={() => onTrace({ to: id })}>
          Trace to here
        </Button>
      ) : null}
    </div>
  );
}

function NodeDetails({
  layout,
  id,
  onTrace,
}: {
  layout: Layout;
  id: string;
  onTrace: (t: Trace) => void;
}) {
  const node = layout.nodes.find((n) => n.id === id);
  if (!node) return null;
  const into = layout.edges.filter((e) => e.to === id);
  const outOf = layout.edges.filter((e) => e.from === id);
  if (node.kind === "entry") {
    return (
      <>
        <h2 className="text-sm font-semibold">{node.label}</h2>
        <div className="text-xs text-muted-foreground">external trigger</div>
        <TraceButtons id={id} onTrace={onTrace} to={false} />
        <Key>Starts</Key>
        <ul className="list-disc pl-5">
          {outOf.map((e) => (
            <li key={e.id}>
              <Mono>{e.to}</Mono>
              {e.kind === "entry" ? (
                <div className="font-mono text-[11.5px] text-muted-foreground">
                  {e.routes.join(", ")}
                </div>
              ) : null}
            </li>
          ))}
        </ul>
      </>
    );
  }
  if (node.kind !== "event") return null;
  const { event } = node;
  const transitions = (edges: LaidEdge[], side: "from" | "to") => {
    const ts = edges.flatMap((e) => (e.kind === "conn" ? [e.transition] : []));
    return ts.length ? (
      <ul className="list-disc pl-5">
        {ts.map((t) => (
          <li key={`${t[side]}|${t.actor}`}>
            <Dot kind={t.kind} />
            <Mono>{t[side]}</Mono> <span className="text-xs text-muted-foreground">via</span>{" "}
            <Mono>{t.actor}</Mono>
            {t.cond ? <span className="text-xs text-muted-foreground"> conditional</span> : null}
          </li>
        ))}
      </ul>
    ) : (
      <div className="text-xs text-muted-foreground">none</div>
    );
  };
  return (
    <>
      <h2 className="font-mono text-sm font-semibold break-all">{event.id}</h2>
      <div className="text-xs text-muted-foreground">
        {event.subs ? `${event.subs} subscription${event.subs > 1 ? "s" : ""}` : "no subscriber"}
      </div>
      <TraceButtons id={id} onTrace={onTrace} />
      <Key>Started by</Key>
      {event.entry.length ? (
        <ul className="list-disc pl-5">
          {event.entry.map((x) => (
            <li key={`${x.kind}|${x.label}`}>
              <span className="text-xs text-muted-foreground">{x.kind}</span> <Mono>{x.label}</Mono>
            </li>
          ))}
        </ul>
      ) : (
        <div className="text-xs text-muted-foreground">only by other events</div>
      )}
      <Key>Follows from</Key>
      {transitions(into, "from")}
      <Key>Leads to</Key>
      {transitions(outOf, "to")}
    </>
  );
}
