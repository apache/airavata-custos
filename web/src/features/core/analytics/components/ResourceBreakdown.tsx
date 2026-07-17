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

import { formatCredits, formatNative, formatPercent } from "../lib";
import type { UsageResource, UsageSummary } from "../schemas";

// Approximate the caller's own native usage from their share of the credits;
// the summary carries native totals but not a per-caller native breakdown.
function callerNative(r: UsageResource): number {
  if (r.used <= 0) return 0;
  return Math.round(r.used_native * (r.used_by_caller / r.used));
}

// One card for every role. Each resource shows the viewer's own slice against
// the rest of the team; when there is no team (solo) or the viewer hasn't run
// anything, it collapses to a plain consumption bar rather than showing
// "team 0" / "you 0".
export function ResourceBreakdown({ summary }: { summary: UsageSummary }) {
  const rows = summary.by_resource;
  // Each bar is the resource's share of total credits consumed, so the bars
  // sum to the period total; resources carry no per-resource cap in v1.
  const totalUsed = rows.reduce((sum, r) => sum + r.used, 0);
  const anyCallerUsage = rows.some((r) => r.used_by_caller > 0);

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <h2 className="text-sm font-semibold">Usage by resource</h2>
      <p className="mb-3 text-xs text-muted-foreground">
        {anyCallerUsage
          ? "your share vs the rest of the team · last 30 days"
          : "consumed · last 30 days"}
      </p>
      {rows.length === 0 ? (
        <p className="text-sm text-muted-foreground">No usage recorded yet.</p>
      ) : (
        <ul className="space-y-3">
          {rows.map((r) => (
            <ResourceRow key={r.resource_id} resource={r} total={totalUsed} />
          ))}
        </ul>
      )}
    </div>
  );
}

function ResourceRow({ resource, total }: { resource: UsageResource; total: number }) {
  const mine = resource.used_by_caller;
  const team = Math.max(0, resource.used - mine);
  const share = total > 0 ? (resource.used / total) * 100 : 0;
  const minePct = (mine / total) * 100;
  const teamPct = (team / total) * 100;

  const labelParts: string[] = [];
  if (mine > 0) labelParts.push(`you ${formatCredits(mine)}`);
  if (team > 0)
    labelParts.push(mine > 0 ? `team ${formatCredits(team)}` : `${formatCredits(team)} used`);
  const detail = labelParts.length
    ? labelParts.join(" · ")
    : `${formatCredits(resource.used)} used`;
  const label = `${formatPercent(share)} · ${detail}`;

  const nativeLine =
    mine > 0
      ? `${formatNative(callerNative(resource))} ${resource.native_unit} (you)`
      : `${formatNative(resource.used_native)} ${resource.native_unit}`;

  return (
    <li>
      <div className="mb-1 flex items-baseline justify-between text-sm">
        <span className="font-medium">{resource.name}</span>
        <span className="tabular-nums text-muted-foreground">{label}</span>
      </div>
      <div className="flex h-2.5 w-full overflow-hidden rounded-full bg-[color:var(--brand-tint)]">
        {mine > 0 ? <div style={{ width: `${minePct}%`, background: "var(--chart-1)" }} /> : null}
        {team > 0 ? (
          <div
            style={{
              width: `${teamPct}%`,
              background: mine > 0 ? "var(--chart-1-soft)" : "var(--chart-1)",
            }}
          />
        ) : null}
      </div>
      <p className="mt-1 text-xs text-muted-foreground">{nativeLine}</p>
    </li>
  );
}
