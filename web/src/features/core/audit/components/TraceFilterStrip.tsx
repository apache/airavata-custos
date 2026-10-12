/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

"use client";

import { cn } from "@/lib/utils";
import { useDebounce } from "@/shared/hooks/useDebounce";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { keepFocus } from "@/shared/ui/focus";
import { Search } from "lucide-react";
import * as React from "react";
import type { TraceStatus } from "../types";
import { TRACE_STATUS_WORDS } from "../utils";
import { type ListFilters, STATUS_FILTERS, type WindowPreset } from "./traceListUrlState";

export type TraceFilterStripProps = {
  value: ListFilters;
  onChange(next: ListFilters): void;
  sourceOptions?: ReadonlyArray<string>;
  sourcesError?: Error | null;
  onRetrySources?: () => void;
};

const WINDOW_OPTIONS: WindowPreset[] = ["24h", "7d", "30d"];

const STATUS_DOT: Record<TraceStatus, string> = {
  error: "bg-[color:var(--tone-error-fg)]",
  in_progress: "bg-[color:var(--tone-warn-fg)]",
  ok: "bg-[color:var(--tone-ok-fg)]",
};

function FilterPill({
  active,
  onClick,
  children,
  radio,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
  radio?: boolean;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={cn(
        "inline-flex h-7 items-center gap-1.5 px-2.5 text-[12.5px] font-semibold transition-colors",
        "border focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        radio ? "rounded-[14px]" : "rounded-md",
        active
          ? "border-[color:var(--brand)] bg-[color:var(--brand-tint)] text-foreground"
          : "border-[color:var(--border-strong)] bg-card text-foreground hover:bg-muted",
      )}
    >
      {children}
    </button>
  );
}

function GroupLegend({ children }: { children: React.ReactNode }) {
  return (
    <legend className="text-[11.5px] font-semibold uppercase tracking-[0.04em] text-muted-foreground">
      {children}
    </legend>
  );
}

function Divider() {
  return (
    <div
      aria-hidden="true"
      className="mx-1 hidden h-7 w-px self-center bg-[color:var(--border)] md:block"
    />
  );
}

export function TraceFilterStrip({
  value,
  onChange,
  sourceOptions,
  sourcesError,
  onRetrySources,
}: TraceFilterStripProps) {
  // The search is debounced before it reaches the URL, so the list does not
  // refetch on every keystroke.
  const [search, setSearch] = React.useState(value.q);
  const lastRemote = React.useRef(value.q);
  const valueRef = React.useRef(value);
  const onChangeRef = React.useRef(onChange);
  valueRef.current = value;
  onChangeRef.current = onChange;
  React.useEffect(() => {
    if (value.q !== lastRemote.current) {
      lastRemote.current = value.q;
      setSearch(value.q);
    }
  }, [value.q]);
  const debounced = useDebounce(search, 300);
  React.useEffect(() => {
    const q = debounced.trim();
    if (q === valueRef.current.q) return;
    lastRemote.current = q;
    onChangeRef.current({ ...valueRef.current, q, page: 1 });
  }, [debounced]);

  const toggleStatus = (id: TraceStatus) => {
    const has = value.status.includes(id);
    const next = has ? value.status.filter((s) => s !== id) : [...value.status, id];
    onChange({ ...value, status: next, page: 1 });
  };

  const toggleSource = (id: string) => {
    const has = value.source.includes(id);
    const next = has ? value.source.filter((s) => s !== id) : [...value.source, id];
    onChange({ ...value, source: next, page: 1 });
  };

  const pickWindow = (w: WindowPreset) => {
    if (w === value.window) return;
    onChange({ ...value, window: w, page: 1 });
  };

  const sources = sourceOptions ?? value.source;
  const sourceGroup = React.useRef<HTMLFieldSetElement>(null);
  const retrySources = () => {
    keepFocus(() => sourceGroup.current?.querySelector("button") ?? null);
    onRetrySources?.();
  };

  return (
    <Card className="rounded-xl px-4 py-3 shadow-sm" data-testid="trace-filter-strip">
      <div className="flex flex-wrap items-start gap-4">
        <fieldset className="flex items-center gap-2">
          <GroupLegend>STATUS</GroupLegend>
          <div className="flex flex-wrap gap-1.5">
            {STATUS_FILTERS.map((id) => (
              <FilterPill
                key={id}
                active={value.status.includes(id)}
                onClick={() => toggleStatus(id)}
              >
                <span
                  aria-hidden="true"
                  className={cn("size-[7px] rounded-full", STATUS_DOT[id])}
                />
                {TRACE_STATUS_WORDS[id]}
              </FilterPill>
            ))}
          </div>
        </fieldset>

        <Divider />

        <fieldset ref={sourceGroup} className="flex items-center gap-2">
          <GroupLegend>SOURCE</GroupLegend>
          {sourcesError ? (
            <span role="alert" className="inline-flex items-center gap-2 text-[12.5px]">
              Couldn't load sources.
              <Button size="sm" variant="outline" onClick={retrySources}>
                Try again
              </Button>
            </span>
          ) : (
            <div className="flex flex-wrap gap-1.5">
              {sources.map((s) => (
                <FilterPill
                  key={s}
                  active={value.source.includes(s)}
                  onClick={() => toggleSource(s)}
                >
                  {s}
                </FilterPill>
              ))}
            </div>
          )}
        </fieldset>

        <Divider />

        <fieldset className="flex items-center gap-2">
          <GroupLegend>WINDOW</GroupLegend>
          <div className="flex flex-wrap gap-1.5">
            {WINDOW_OPTIONS.map((w) => (
              <FilterPill key={w} active={value.window === w} onClick={() => pickWindow(w)} radio>
                {w}
              </FilterPill>
            ))}
          </div>
        </fieldset>
      </div>

      <div className="relative mt-3 max-w-[520px]">
        <Search
          className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
          aria-hidden="true"
        />
        <input
          type="search"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search traces"
          aria-label="Search traces"
          className={cn(
            "h-[38px] w-full rounded-md border bg-card pl-9 pr-3 text-[13.5px] text-foreground outline-none",
            "border-[color:var(--border-strong)] focus-visible:ring-2 focus-visible:ring-ring",
          )}
        />
      </div>
    </Card>
  );
}
