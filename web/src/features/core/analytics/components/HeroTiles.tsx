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

import { cn } from "@/lib/utils";
import {
  creditsBand,
  daysBand,
  daysUntil,
  formatCredits,
  formatDate,
  pctRemaining,
  type UrgencyBand,
} from "../lib";
import type { AnalyticsAllocation } from "../schemas";
import { BAND_METER_CLASS, BAND_PILL_CLASS, BAND_PILL_LABEL, BAND_RULE_CLASS } from "./bands";

// Number of hero tiles; the page's loading skeletons match this count.
export const HERO_TILE_COUNT = 3;

type HeroTilesProps = {
  allocation: AnalyticsAllocation;
  callerUsed: number;
  now: Date;
};

// The same three tiles for every role: the pooled balance, the project total used
// with the viewer's own share, and when the allocation ends. left + used sum to
// the award, so "your share" is a subline, not the headline.
export function HeroTiles({ allocation, callerUsed, now }: HeroTilesProps) {
  const total = allocation.initial_su_amount;
  const used = allocation.used_su_amount;
  const left = Math.max(0, total - used);
  const pctLeft = pctRemaining(total, used);
  const allocDays = daysUntil(allocation.end_time, now);

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <Tile
        title="Credits left"
        value={formatCredits(left)}
        sub={`${Math.round(pctLeft)}% of ${formatCredits(total)} total`}
        band={creditsBand(pctLeft)}
        meterPct={pctLeft}
      />
      <Tile
        title="Credits used"
        value={formatCredits(used)}
        sub={`your share: ${formatCredits(callerUsed)}`}
        band="ok"
      />
      <Tile
        title="Time remaining"
        value={String(allocDays)}
        sub={formatDate(allocation.end_time)}
        band={daysBand(allocDays)}
        unit="days"
      />
    </div>
  );
}

function Tile({
  title,
  value,
  sub,
  band,
  unit,
  meterPct,
}: {
  title: string;
  value: string;
  sub: string;
  band: UrgencyBand;
  unit?: string;
  meterPct?: number;
}) {
  return (
    <div
      className={cn(
        "rounded-lg border border-border bg-card p-4",
        band !== "ok" && BAND_RULE_CLASS[band],
      )}
    >
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
          {title}
        </span>
        {band !== "ok" ? (
          <span
            className={cn(
              "rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide",
              BAND_PILL_CLASS[band],
            )}
          >
            {BAND_PILL_LABEL[band]}
          </span>
        ) : null}
      </div>
      <div className="mt-2 flex items-baseline gap-1.5">
        <span className="font-display text-3xl font-bold text-foreground">{value}</span>
        {unit ? <span className="text-sm text-muted-foreground">{unit}</span> : null}
      </div>
      {meterPct != null ? (
        <div className="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-muted">
          <div
            className={cn("h-full rounded-full", BAND_METER_CLASS[band])}
            style={{ width: `${Math.max(0, Math.min(100, meterPct))}%` }}
          />
        </div>
      ) : null}
      <div className="mt-2 text-xs text-muted-foreground">{sub}</div>
    </div>
  );
}
