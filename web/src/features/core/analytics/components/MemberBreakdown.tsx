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

import { formatCredits } from "../lib";
import type { UsageMember } from "../schemas";

const TOP_MEMBERS = 5;

function initials(name: string): string {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((p) => p[0]?.toUpperCase() ?? "")
    .join("");
}

// Ranked per-member consumption. Rendered only for project managers; the caller
// passes the (non-null) member list.
export function MemberBreakdown({ members }: { members: UsageMember[] }) {
  const ranked = [...members].sort((a, b) => b.used - a.used);
  const top = ranked.slice(0, TOP_MEMBERS);
  const rest = ranked.slice(TOP_MEMBERS);
  const restTotal = rest.reduce((a, m) => a + m.used, 0);
  const max = Math.max(1, ranked[0]?.used ?? 0);

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <h2 className="text-sm font-semibold">Usage by member</h2>
      <p className="mb-3 text-xs text-muted-foreground">last 30 days</p>
      {ranked.length === 0 ? (
        <p className="text-sm text-muted-foreground">No member usage yet.</p>
      ) : (
        <ul className="space-y-2.5">
          {top.map((m) => (
            <Row
              key={m.user_id}
              label={m.name}
              used={m.used}
              max={max}
              initials={initials(m.name)}
            />
          ))}
          {rest.length > 0 ? (
            <Row
              label={`+${rest.length} ${rest.length === 1 ? "other" : "others"}`}
              used={restTotal}
              max={max}
              muted
            />
          ) : null}
        </ul>
      )}
    </div>
  );
}

function Row({
  label,
  used,
  max,
  initials: init,
  muted,
}: {
  label: string;
  used: number;
  max: number;
  initials?: string;
  muted?: boolean;
}) {
  const pct = (used / max) * 100;
  return (
    <li className="flex items-center gap-3">
      <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-muted text-[10px] font-semibold text-muted-foreground">
        {init ?? "···"}
      </span>
      <div className="min-w-0 flex-1">
        <div className="mb-0.5 flex items-baseline justify-between gap-2 text-sm">
          <span className="truncate">{label}</span>
          <span className="tabular-nums text-muted-foreground">{formatCredits(used)}</span>
        </div>
        <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
          <div
            className="h-full rounded-full"
            style={{ width: `${pct}%`, background: muted ? "var(--chart-5)" : "var(--chart-1)" }}
          />
        </div>
      </div>
    </li>
  );
}
