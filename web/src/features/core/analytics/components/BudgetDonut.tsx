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

import * as React from "react";
import { CHART_OTHER_COLOR, formatCredits, formatCreditsFull, formatPercent } from "../lib";

// folded items are listed individually in the legend but share one ring arc,
// so fifty hair-thin sectors never happen.
export type BudgetSlice = {
  key: string;
  label: string;
  value: number;
  color: string;
  folded?: boolean;
};

// The track doubles as the unused remainder, so it stays a light neutral,
// distinct from the darker chart-5 that marks the folded-tail arc.
const REMAINDER_COLOR = "var(--brand-tint)";
const OTHERS_KEY = "__others";

// A donut against the allocation's full credit budget: colored sectors are
// consumed credits, the track that shows through is what remains. Hovering
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
      label: `+${folded.length} ${folded.length === 1 ? "other" : "others"}`,
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
    <div className="flex flex-col items-center gap-6 md:flex-row md:items-center md:gap-8">
      <svg
        viewBox="0 0 42 42"
        className="h-[220px] w-[220px] shrink-0 md:h-[260px] md:w-[260px]"
        role="img"
        aria-label={`${ariaContext}: ${formatCreditsFull(used)} of ${formatCreditsFull(total)} credits used`}
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
                <title>{`${s.label}: ${formatCreditsFull(s.value)} credits, ${formatPercent(pct)}`}</title>
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
          ? `${activeItem.label}: ${formatCreditsFull(activeItem.value)} credits, ${formatPercent(pctOf(activeItem.value))}`
          : null}
      </div>
      <div className="w-full min-w-0 max-w-[420px] flex-1">
        <ul className="max-h-none space-y-0.5 overflow-y-auto pr-2 md:max-h-[260px]">
          {items.map((s) => (
            <LegendRow
              key={s.key}
              label={s.label}
              value={s.value}
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
          {formatCredits(activeItem.value)}
        </text>
        <text x="21" y="26.5" textAnchor="middle" className="fill-muted-foreground text-[2.2px]">
          {formatPercent(pctOf(activeItem.value))} · of {formatCredits(total)}
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
          -{formatCredits(used - total)}
        </text>
        <text x="21" y="25" textAnchor="middle" className="fill-muted-foreground text-[2.4px]">
          over the {formatCredits(total)} budget
        </text>
      </>
    );
  }
  return (
    <>
      <text x="21" y="20.5" textAnchor="middle" className="fill-foreground text-[5px] font-semibold">
        {formatCredits(remaining)}
      </text>
      <text x="21" y="25" textAnchor="middle" className="fill-muted-foreground text-[2.4px]">
        of {formatCredits(total)} available
      </text>
    </>
  );
}

function LegendRow({
  label,
  value,
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
  pct: number;
  color: string;
  active?: boolean;
  dimmed?: boolean;
  emphasis?: boolean;
  onEnter?: () => void;
  onLeave?: () => void;
  onToggle?: () => void;
}) {
  return (
    <li>
      <button
        type="button"
        className={`flex w-full max-w-[380px] items-center gap-2 rounded px-1 py-1 text-left text-sm outline-none motion-safe:transition-opacity ${
          active ? "bg-[color:var(--brand-tint)]" : ""
        } ${dimmed ? "opacity-50" : ""} ${onToggle ? "" : "cursor-default"}`}
        onMouseEnter={onEnter}
        onMouseLeave={onLeave}
        onFocus={onEnter}
        onBlur={onLeave}
        onClick={onToggle}
      >
        <span className="h-2.5 w-2.5 shrink-0 rounded-full" style={{ background: color }} />
        <span
          className={`min-w-0 flex-1 truncate ${emphasis ? "font-medium" : ""}`}
          title={label}
        >
          {label}
        </span>
        <span className="shrink-0 tabular-nums text-muted-foreground">
          {formatPercent(pct)} · {formatCredits(value)}
        </span>
      </button>
    </li>
  );
}
