"use client";

import { Search } from "lucide-react";
import * as React from "react";
import { cn } from "@/lib/utils";
import { useDebounce } from "@/shared/hooks/useDebounce";
import { Card } from "@/shared/ui/card";
import type { RowTone } from "../types";
import type { ListFilters, StatusFilter, WindowPreset } from "./traceListUrlState";

export type TraceFilterStripProps = {
  value: ListFilters;
  onChange(next: ListFilters): void;
  sourceOptions?: ReadonlyArray<string>;
};

const STATUS_OPTIONS: { id: StatusFilter; tone: RowTone; label: string }[] = [
  { id: "error", tone: "error", label: "error" },
  { id: "ok", tone: "ok", label: "ok" },
  { id: "in-progress", tone: "in-progress", label: "in-progress" },
  { id: "orphaned", tone: "orphaned", label: "orphaned" },
];

const DEFAULT_SOURCE_OPTIONS = ["amie", "comanage", "slurm", "http", "core"] as const;
const WINDOW_OPTIONS: WindowPreset[] = ["24h", "7d", "30d"];

// Inline-style tone dot — StatusPill exposes dotOnly but the filter pill needs
// a 7px dot, not 8px.
const TONE_DOT: Record<StatusFilter, { color: string; hollow: boolean }> = {
  error: { color: "var(--custos-red-500)", hollow: false },
  ok: { color: "var(--custos-green-500)", hollow: false },
  "in-progress": { color: "var(--custos-amber-500)", hollow: false },
  orphaned: { color: "var(--muted-foreground)", hollow: true },
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
          ? "border-[color:var(--brand)] bg-[color:var(--brand-tint)] text-[color:var(--brand)]"
          : "border-[color:var(--border-strong)] bg-card text-foreground hover:bg-muted",
      )}
    >
      {children}
    </button>
  );
}

function StatusDot({ status }: { status: StatusFilter }) {
  const { color, hollow } = TONE_DOT[status];
  return (
    <span
      aria-hidden="true"
      style={{
        width: 7,
        height: 7,
        borderRadius: "50%",
        background: hollow ? "transparent" : color,
        boxShadow: hollow ? `inset 0 0 0 1.4px ${color}` : "none",
      }}
    />
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

export function TraceFilterStrip({ value, onChange, sourceOptions }: TraceFilterStripProps) {
  // Local search debounced before pushing to the URL — avoids navigating on
  // every keystroke and resets to page 1 once the user pauses.
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
    if (debounced === valueRef.current.q) return;
    lastRemote.current = debounced;
    onChangeRef.current({ ...valueRef.current, q: debounced, page: 1 });
  }, [debounced]);

  const toggleStatus = (id: StatusFilter) => {
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

  const sources = sourceOptions ?? DEFAULT_SOURCE_OPTIONS;

  return (
    <Card className="rounded-xl px-4 py-3 shadow-sm" data-testid="trace-filter-strip">
      <div className="flex flex-wrap items-start gap-4">
        <fieldset className="flex items-center gap-2">
          <GroupLegend>STATUS</GroupLegend>
          <div className="flex flex-wrap gap-1.5">
            {STATUS_OPTIONS.map((opt) => (
              <FilterPill
                key={opt.id}
                active={value.status.includes(opt.id)}
                onClick={() => toggleStatus(opt.id)}
              >
                <StatusDot status={opt.id} />
                {opt.label}
              </FilterPill>
            ))}
          </div>
        </fieldset>

        <Divider />

        <fieldset className="flex items-center gap-2">
          <GroupLegend>SOURCE</GroupLegend>
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

      <div className="mt-3 flex flex-wrap items-center gap-2">
        <div className="relative min-w-[260px] flex-1">
          <Search
            className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
            aria-hidden="true"
          />
          <input
            type="search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search trace_id / span_id / entity / action…"
            aria-label="Search traces"
            className={cn(
              "h-[38px] w-full rounded-md border bg-card pl-9 pr-3 text-[13.5px] text-foreground outline-none",
              "border-[color:var(--border-strong)] focus-visible:ring-2 focus-visible:ring-ring",
            )}
          />
        </div>
        {value.failingOver24h && (
          <button
            type="button"
            data-testid="failing-over-24h-chip"
            aria-label="Clear Failing >24h filter"
            onClick={() => onChange({ ...value, failingOver24h: false, page: 1 })}
            className={cn(
              "inline-flex h-7 items-center gap-1.5 rounded-md px-2.5 text-[12.5px] font-semibold transition-colors",
              "border focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
              "border-[color:var(--brand)] bg-[color:var(--brand-tint)] text-[color:var(--brand)]",
            )}
          >
            <span>Failing &gt;24h</span>
            <span aria-hidden="true">×</span>
          </button>
        )}
      </div>
    </Card>
  );
}
