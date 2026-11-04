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
import { ApiError } from "@/shared/api/client";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { Button } from "@/shared/ui/button";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Input } from "@/shared/ui/input";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/shared/ui/select";
import { Skeleton } from "@/shared/ui/skeleton";
import { ArrowLeftRightIcon, ArrowRightIcon } from "lucide-react";
import * as React from "react";
import {
  DEFAULT_GRAPH_OPTIONS,
  entityOf,
  type GraphOptions,
  type Layout,
  layoutGraph,
  neighbourhood,
  pathsThrough,
  type Selected,
  selectGraph,
  type Trace,
  traceGraph,
  triggerId,
  verbOf,
} from "../graph";
import { useEventGraph } from "../queries";
import { COLOR, EventGraphCanvas, type Lit, type Pick } from "./EventGraphCanvas";
import { EventGraphDetails } from "./EventGraphDetails";

const ANY = "*";

const SHOW: {
  key: keyof GraphOptions;
  label: string;
  swatch: React.ReactNode;
}[] = [
  { key: "core", label: "Core", swatch: <Line color={COLOR.core} /> },
  { key: "connector", label: "Connector", swatch: <Line color={COLOR.conn} /> },
  { key: "ends", label: "End steps", swatch: <span style={{ color: COLOR.end }}>⊣</span> },
  { key: "entries", label: "Triggers", swatch: <Line color={COLOR.entry} dashed /> },
];

function Line({ color, dashed = false }: { color: string; dashed?: boolean }) {
  return (
    <span
      className="inline-block h-0 w-4 border-t-2 align-middle"
      style={{ borderColor: color, borderTopStyle: dashed ? "dashed" : "solid" }}
    />
  );
}

function useTrace(): [Trace, (t: Trace) => void] {
  const params = useShallowSearchParams();
  const trace = { from: params.get("from") ?? undefined, to: params.get("to") ?? undefined };
  const setTrace = (t: Trace) => {
    const next = new URLSearchParams(params.toString());
    for (const key of ["from", "to"] as const) {
      const value = t[key];
      if (value) next.set(key, value);
      else next.delete(key);
    }
    replaceShallowSearchParams(next);
  };
  return [trace, setTrace];
}

export function EventGraphPage() {
  const graph = useEventGraph();
  const [options, setOptions] = React.useState(DEFAULT_GRAPH_OPTIONS);
  const [trace, setTrace] = useTrace();
  const [layout, setLayout] = React.useState<Layout>();
  const [pick, setPick] = React.useState<Pick>();
  const [open, setOpen] = React.useState(false);
  const [query, setQuery] = React.useState("");

  const selected = React.useMemo(
    () => (graph.data ? selectGraph(graph.data, options) : undefined),
    [graph.data, options],
  );
  const { from, to } = trace;
  const traced = React.useMemo(
    () => (selected ? traceGraph(selected, { from, to }) : undefined),
    [selected, from, to],
  );

  React.useEffect(() => {
    if (!traced) return;
    let live = true;
    layoutGraph(traced).then((l) => {
      if (live) setLayout(l);
    });
    setPick(undefined);
    return () => {
      live = false;
    };
  }, [traced]);

  const lit = React.useMemo(
    () => (layout ? highlight(layout, pick, query) : undefined),
    [layout, pick, query],
  );
  const marked = React.useMemo(
    () => new Set([from, to].filter((x): x is string => Boolean(x))),
    [from, to],
  );

  return (
    <div className="-mx-10 -mb-8 flex h-[calc(100vh-6rem)] min-h-[520px] flex-col gap-2">
      {selected ? (
        <Toolbar
          selected={selected}
          options={options}
          onOptions={setOptions}
          trace={trace}
          onTrace={setTrace}
          query={query}
          onQuery={setQuery}
        />
      ) : null}

      <div className="relative min-h-0 flex-1 overflow-hidden border-t bg-background">
        {graph.isError ? (
          graph.error instanceof ApiError && graph.error.status === 404 ? (
            <EmptyState
              className="h-full border-0"
              heading="No event graph"
              description="This server was built without one. Run `go generate ./cmd/server/` and rebuild."
            />
          ) : graph.error instanceof ApiError && graph.error.status === 403 ? (
            <EmptyState
              className="h-full border-0"
              heading="No access"
              description="The event graph needs the core:events:manage privilege."
            />
          ) : (
            <ErrorState message={graph.error.message} />
          )
        ) : traced && !traced.events.length ? (
          <EmptyState
            className="h-full border-0"
            heading="No path"
            description="Nothing connects these two with the current filters. Turn on more arrow kinds or pick other ends."
          />
        ) : layout && lit ? (
          <>
            <EventGraphCanvas
              layout={layout}
              lit={lit}
              marked={marked}
              onPick={setPick}
              // An empty click folds the open drawer first, then drops the selection.
              onClear={() => {
                if (open) return setOpen(false);
                setPick(undefined);
                setQuery("");
              }}
            />
            {pick ? (
              <EventGraphDetails
                layout={layout}
                pick={pick}
                open={open}
                onOpen={setOpen}
                onTrace={(t) => setTrace({ ...trace, ...t })}
              />
            ) : null}
          </>
        ) : (
          <Skeleton className="h-full w-full" />
        )}
      </div>
    </div>
  );
}

function Toolbar({
  selected,
  options,
  onOptions,
  trace,
  onTrace,
  query,
  onQuery,
}: {
  selected: Selected;
  options: GraphOptions;
  onOptions: (o: GraphOptions) => void;
  trace: Trace;
  onTrace: (t: Trace) => void;
  query: string;
  onQuery: (q: string) => void;
}) {
  // Trace ends come from what the filters show, so a pick always has something to trace.
  const triggers = [...new Set(selected.events.flatMap((e) => e.entry.map((x) => x.kind)))].sort();
  const byEntity = [...Map.groupBy(selected.events, (e) => entityOf(e.id))]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([entity, events]) => ({ entity, events: events.map((e) => e.id) }));

  return (
    <div className="flex flex-wrap items-center gap-2 px-10 xl:flex-nowrap">
      <h1 className="mr-2 shrink-0 text-lg font-semibold tracking-tight whitespace-nowrap">
        Event graph
      </h1>
      <EndPicker
        label="From"
        value={trace.from}
        onChange={(from) => onTrace({ ...trace, from })}
        triggers={triggers}
        groups={byEntity}
      />
      <ArrowRightIcon aria-hidden="true" className="size-4 text-muted-foreground" />
      <EndPicker
        label="To"
        value={trace.to}
        onChange={(to) => onTrace({ ...trace, to })}
        groups={byEntity}
      />
      <Button
        variant="ghost"
        size="icon-sm"
        aria-label="Swap start and end"
        disabled={Boolean(trace.from?.startsWith("in:"))}
        onClick={() => onTrace({ from: trace.to, to: trace.from })}
      >
        <ArrowLeftRightIcon />
      </Button>
      {trace.from || trace.to ? (
        <Button variant="ghost" size="sm" onClick={() => onTrace({})}>
          Clear
        </Button>
      ) : null}
      {SHOW.map(({ key, label, swatch }) => (
        <Button
          key={key}
          variant="outline"
          size="sm"
          aria-pressed={options[key]}
          onClick={() => onOptions({ ...options, [key]: !options[key] })}
          className={cn(!options[key] && "text-muted-foreground opacity-60")}
        >
          {swatch}
          {label}
        </Button>
      ))}
      <Input
        type="search"
        value={query}
        onChange={(e) => onQuery(e.target.value)}
        placeholder="Highlight…"
        aria-label="Highlight event or handler"
        className="ml-auto h-8 w-60 min-w-28 shrink"
      />
    </div>
  );
}

function EndPicker({
  label,
  value,
  onChange,
  triggers = [],
  groups,
}: {
  label: string;
  value: string | undefined;
  onChange: (value: string | undefined) => void;
  triggers?: string[];
  groups: { entity: string; events: string[] }[];
}) {
  const name = (id: string) => (id.startsWith("in:") ? id.slice(3) : id);
  return (
    <Select value={value ?? ANY} onValueChange={(v) => onChange(!v || v === ANY ? undefined : v)}>
      <SelectTrigger aria-label={label} className="h-8 w-56 min-w-32 shrink px-3">
        <SelectValue>
          {(v: string) => (
            <span className="truncate">
              <span className="text-muted-foreground">{label}: </span>
              <span className="font-mono text-xs">{v === ANY ? "any" : name(v)}</span>
            </span>
          )}
        </SelectValue>
      </SelectTrigger>
      <SelectContent className="max-h-96">
        <SelectItem value={ANY}>any</SelectItem>
        {triggers.length ? (
          <SelectGroup>
            <SelectLabel>Triggers</SelectLabel>
            {triggers.map((t) => (
              <SelectItem key={t} value={triggerId(t)}>
                {t}
              </SelectItem>
            ))}
          </SelectGroup>
        ) : null}
        {groups.map(({ entity, events }) => (
          <SelectGroup key={entity}>
            <SelectLabel className="font-mono">{entity}</SelectLabel>
            {events.map((id) => (
              <SelectItem key={id} value={id} className="pl-5 font-mono text-xs">
                {verbOf(id)}
              </SelectItem>
            ))}
          </SelectGroup>
        ))}
      </SelectContent>
    </Select>
  );
}

function highlight(layout: Layout, pick: Pick | undefined, query: string): Lit {
  if (pick?.type === "node") return neighbourhood(layout, pick.id);
  if (pick?.type === "term") {
    const e = layout.edges.find((x) => x.kind === "end" && x.from === pick.from);
    return pathsThrough(layout, pick.from, `end:${pick.from}`, e?.id);
  }
  if (pick?.type === "edge") {
    const e = layout.edges.find((x) => x.id === pick.id);
    return e ? pathsThrough(layout, e.from, e.to, e.id) : { edges: null, nodes: null };
  }
  const s = query.trim().toLowerCase();
  if (!s) return { edges: null, nodes: null };
  const hit = (x: string) => x.toLowerCase().includes(s);
  return {
    edges: new Set(
      layout.edges
        .filter((e) =>
          [
            e.from,
            e.to,
            ...(e.kind === "entry" ? e.routes : []),
            ...(e.kind === "core" || e.kind === "conn" ? [e.transition.actor] : []),
          ].some(hit),
        )
        .map((e) => e.id),
    ),
    nodes: new Set(
      layout.nodes
        .filter((n) => hit(n.id) || (n.kind === "end" && n.terms.some((t) => hit(t.actor))))
        .map((n) => n.id),
    ),
  };
}
