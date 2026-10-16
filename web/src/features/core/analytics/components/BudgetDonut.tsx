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

import { formatSU, pluralize } from "@/shared/format";
import * as React from "react";
import { CHART_OTHER_COLOR, formatPercent } from "../lib";

// folded items are listed individually in the legend but share one ring arc,
// so fifty hair-thin sectors never happen.
export type BudgetSlice = {
  key: string;
  label: string;
  value: number;
  color: string;
  folded?: boolean;
  // The same usage in the resource's own unit, e.g. "49,742 GPU-hours".
  native?: string;
};

// The track doubles as the unused remainder, so it stays a light neutral,
// distinct from the darker chart-5 that marks the folded-tail arc.
const REMAINDER_COLOR = "var(--brand-tint)";
const OTHERS_KEY = "__others";

// A donut against the allocation's full SU budget: colored sectors are
// consumed SU, the track that shows through is what remains. Hovering
// (or tapping) a sector or legend row swaps the center to that slice's
// numbers. When consumption exceeds the budget, sectors span the whole ring
// and the center flips to how far over the budget the allocation is.
export function BudgetDonut({
  items,
  total,
  ariaContext,
}: {
  items: BudgetSlice[];
  total: number;
  ariaContext: string;
}) {
  const [active, setActive] = React.useState<string | null>(null);

  const used = items.reduce((a, s) => a + s.value, 0);
  const over = used > total;
  const denominator = Math.max(total, used);
  const remaining = Math.max(0, total - used);
  const pctOf = (v: number) => (denominator > 0 ? (v / denominator) * 100 : 0);

  const shown = items.filter((s) => !s.folded);
  const folded = items.filter((s) => s.folded);
  const arcs: BudgetSlice[] = [...shown];
  if (folded.length > 0) {
    arcs.push({
      key: OTHERS_KEY,
      label: `+${folded.length} ${pluralize("other", folded.length)}`,
      value: folded.reduce((a, s) => a + s.value, 0),
      color: CHART_OTHER_COLOR,
    });
  }

  // A folded legend row highlights the shared tail arc.
  const arcKeyFor = (s: BudgetSlice) => (s.folded ? OTHERS_KEY : s.key);
  const activeItem =
    items.find((s) => s.key === active) ?? arcs.find((s) => s.key === active) ?? null;
  const toggle = (key: string) => setActive((cur) => (cur === key ? null : key));

  return (
    // Side by side only when the card itself is wide, not the viewport.
    <div className="@container w-full">
      <div className="flex flex-col items-center gap-6 @2xl:flex-row @2xl:gap-8">
        <svg
          viewBox="0 0 42 42"
          className="h-[220px] w-[220px] shrink-0 @2xl:h-[260px] @2xl:w-[260px]"
          role="img"
          aria-label={`${ariaContext}: ${formatSU(used)} of ${formatSU(total)} SU used`}
        >
          <circle cx="21" cy="21" r="15.915" fill="none" stroke={REMAINDER_COLOR} strokeWidth="7" />
          {(() => {
            let cumulative = 0;
            return arcs.map((s) => {
              // r = 15.915 gives a circumference of ~100, so a percentage maps
              // straight to a dash length. Offset 25 starts the ring at 12 o'clock.
              const pct = pctOf(s.value);
              const isActive = active === s.key;
              const dimmed = active !== null && !isActive;
              const arc = (
                <circle
                  key={s.key}
                  cx="21"
                  cy="21"
                  r="15.915"
                  fill="none"
                  stroke={s.color}
                  strokeWidth={isActive ? 8.2 : 7}
                  strokeDasharray={`${pct} ${100 - pct}`}
                  strokeDashoffset={25 - cumulative}
                  opacity={dimmed ? 0.4 : 1}
                  className="cursor-default motion-safe:transition-[opacity,stroke-width]"
                  data-slice={s.key}
                  onMouseEnter={() => setActive(s.key)}
                  onMouseLeave={() => setActive(null)}
                >
                  <title>{`${s.label}: ${formatSU(s.value)} SU, ${formatPercent(pct)}`}</title>
                </circle>
              );
              cumulative += pct;
              return arc;
            });
          })()}
          <CenterText
            activeItem={activeItem}
            over={over}
            used={used}
            total={total}
            remaining={remaining}
            pctOf={pctOf}
          />
        </svg>
        <div aria-live="polite" className="sr-only">
          {activeItem
            ? `${activeItem.label}: ${formatSU(activeItem.value)} SU, ${formatPercent(pctOf(activeItem.value))}`
            : null}
        </div>
        <div className="w-full min-w-0 max-w-[420px] flex-1">
          <ul className="max-h-none space-y-0.5 overflow-y-auto pr-2 @2xl:max-h-[260px]">
            {items.map((s) => (
              <LegendRow
                key={s.key}
                label={s.label}
                value={s.value}
                native={s.native}
                pct={pctOf(s.value)}
                color={s.color}
                active={active === s.key || active === arcKeyFor(s)}
                dimmed={active !== null && active !== s.key && active !== arcKeyFor(s)}
                onEnter={() => setActive(s.key)}
                onLeave={() => setActive(null)}
                onToggle={() => toggle(s.key)}
              />
            ))}
          </ul>
          {!over && (
            <div className="mt-2 border-t border-border pt-2">
              <ul>
                <LegendRow
                  label="Available"
                  value={remaining}
                  pct={pctOf(remaining)}
                  color={REMAINDER_COLOR}
                  emphasis
                />
              </ul>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function CenterText({
  activeItem,
  over,
  used,
  total,
  remaining,
  pctOf,
}: {
  activeItem: BudgetSlice | null;
  over: boolean;
  used: number;
  total: number;
  remaining: number;
  pctOf: (v: number) => number;
}) {
  if (activeItem) {
    return (
      <>
        <text x="21" y="17.5" textAnchor="middle" className="fill-muted-foreground text-[2.4px]">
          {activeItem.label.length > 24 ? `${activeItem.label.slice(0, 23)}…` : activeItem.label}
        </text>
        <text
          x="21"
          y="22.5"
          textAnchor="middle"
          className="fill-foreground text-[4.4px] font-semibold"
        >
          {formatSU(activeItem.value)}
        </text>
        <text x="21" y="26.5" textAnchor="middle" className="fill-muted-foreground text-[2.2px]">
          {formatPercent(pctOf(activeItem.value))} · of {formatSU(total)}
        </text>
      </>
    );
  }
  if (over) {
    return (
      <>
        <text
          x="21"
          y="20.5"
          textAnchor="middle"
          className="fill-[color:var(--tone-error-fg)] text-[5px] font-semibold"
        >
          -{formatSU(used - total)}
        </text>
        <text x="21" y="25" textAnchor="middle" className="fill-muted-foreground text-[2.4px]">
          over the {formatSU(total)} budget
        </text>
      </>
    );
  }
  return (
    <>
      <text
        x="21"
        y="20.5"
        textAnchor="middle"
        className="fill-foreground text-[5px] font-semibold"
      >
        {formatSU(remaining)}
      </text>
      <text x="21" y="25" textAnchor="middle" className="fill-muted-foreground text-[2.4px]">
        of {formatSU(total)} available
      </text>
    </>
  );
}

function LegendRow({
  label,
  value,
  native,
  pct,
  color,
  active,
  dimmed,
  emphasis,
  onEnter,
  onLeave,
  onToggle,
}: {
  label: string;
  value: number;
  native?: string;
  pct: number;
  color: string;
  active?: boolean;
  dimmed?: boolean;
  emphasis?: boolean;
  onEnter?: () => void;
  onLeave?: () => void;
  onToggle?: () => void;
}) {
  const amount = `${formatPercent(pct)} · ${formatSU(value)}${native ? ` · ${native}` : ""}`;
  return (
    <li>
      <button
        type="button"
        className={`flex w-full items-start gap-2 rounded px-1 py-1 text-left text-sm outline-none motion-safe:transition-opacity ${
          active ? "bg-[color:var(--brand-tint)]" : ""
        } ${dimmed ? "opacity-50" : ""} ${onToggle ? "" : "cursor-default"}`}
        onMouseEnter={onEnter}
        onMouseLeave={onLeave}
        onFocus={onEnter}
        onBlur={onLeave}
        onClick={onToggle}
      >
        <span className="mt-1.5 h-2.5 w-2.5 shrink-0 rounded-full" style={{ background: color }} />
        <span className="flex min-w-0 flex-1 flex-col">
          <span className={`truncate ${emphasis ? "font-medium" : ""}`} title={label}>
            {label}
          </span>
          <span className="truncate text-xs tabular-nums text-muted-foreground" title={amount}>
            {amount}
          </span>
        </span>
      </button>
    </li>
  );
}
