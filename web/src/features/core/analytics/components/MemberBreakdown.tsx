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

import { formatCredits, formatPercent } from "../lib";
import type { UsageMember } from "../schemas";

const TOP_MEMBERS = 5;

// Distinct hues for the named members; the tail rolls into the neutral "others".
const SLICE_COLORS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-1-soft)",
];
const OTHERS_COLOR = "var(--chart-5)";

type Slice = { key: string; label: string; used: number; color: string };

// Ranked per-member consumption as a share donut: who used the allocation and by
// how much. Rendered only for project managers; the caller passes the member list.
export function MemberBreakdown({ members }: { members: UsageMember[] }) {
  const ranked = [...members].sort((a, b) => b.used - a.used);
  const top = ranked.slice(0, TOP_MEMBERS);
  const rest = ranked.slice(TOP_MEMBERS);
  const restTotal = rest.reduce((a, m) => a + m.used, 0);
  const total = ranked.reduce((a, m) => a + m.used, 0);

  const slices: Slice[] = top.map((m, i) => ({
    key: m.user_id,
    label: m.name,
    used: m.used,
    color: SLICE_COLORS[i % SLICE_COLORS.length] ?? OTHERS_COLOR,
  }));
  if (rest.length > 0) {
    slices.push({
      key: "others",
      label: `+${rest.length} ${rest.length === 1 ? "other" : "others"}`,
      used: restTotal,
      color: OTHERS_COLOR,
    });
  }

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <h2 className="text-sm font-semibold">Usage by member</h2>
      <p className="mb-3 text-xs text-muted-foreground">share of credits used · last 30 days</p>
      {ranked.length === 0 ? (
        <p className="text-sm text-muted-foreground">No member usage yet.</p>
      ) : (
        <div className="flex flex-col items-center gap-4 sm:flex-row">
          <Donut slices={slices} total={total} />
          <ul className="w-full min-w-0 flex-1 space-y-1.5">
            {slices.map((s) => (
              <li key={s.key} className="flex items-center gap-2 text-sm">
                <span
                  className="h-2.5 w-2.5 shrink-0 rounded-full"
                  style={{ background: s.color }}
                />
                <span className="min-w-0 flex-1 truncate">{s.label}</span>
                <span className="shrink-0 tabular-nums text-muted-foreground">
                  {formatPercent(total > 0 ? (s.used / total) * 100 : 0)} · {formatCredits(s.used)}
                </span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

// r = 15.915 gives a circumference of ~100, so each slice's percentage maps
// straight to a dash length. Offset 25 starts the ring at 12 o'clock.
function Donut({ slices, total }: { slices: Slice[]; total: number }) {
  let cumulative = 0;
  return (
    <svg viewBox="0 0 42 42" className="h-32 w-32 shrink-0" aria-hidden="true">
      <circle cx="21" cy="21" r="15.915" fill="none" stroke="var(--brand-tint)" strokeWidth="5" />
      {slices.map((s) => {
        const pct = total > 0 ? (s.used / total) * 100 : 0;
        const dash = (
          <circle
            key={s.key}
            cx="21"
            cy="21"
            r="15.915"
            fill="none"
            stroke={s.color}
            strokeWidth="5"
            strokeDasharray={`${pct} ${100 - pct}`}
            strokeDashoffset={25 - cumulative}
          />
        );
        cumulative += pct;
        return dash;
      })}
      <text
        x="21"
        y="20.5"
        textAnchor="middle"
        className="fill-foreground text-[4px] font-semibold"
      >
        {formatCredits(total)}
      </text>
      <text x="21" y="25" textAnchor="middle" className="fill-muted-foreground text-[2.4px]">
        total
      </text>
    </svg>
  );
}
