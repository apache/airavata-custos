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
import { BurnDownBar } from "@/shared/charts/BurnDownBar";

export type ForecastBarProps = {
  used: number;
  projected: number;
  capacity: number;
  exhaustDate?: Date | string | null;
  endDate?: Date | string | null;
  label?: string;
  onClick?: () => void;
  className?: string;
};

function toDate(input: Date | string | null | undefined): Date | null {
  if (!input) return null;
  if (input instanceof Date) return Number.isNaN(input.getTime()) ? null : input;
  const ms = Date.parse(input);
  return Number.isNaN(ms) ? null : new Date(ms);
}

function formatDate(d: Date): string {
  return d.toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" });
}

function daysBetween(a: Date, b: Date): number {
  // Math.abs + round so "0.99 days" reads as "1 day" not "0 days".
  return Math.round(Math.abs(b.getTime() - a.getTime()) / (24 * 60 * 60 * 1000));
}

export function ForecastBar({
  used,
  projected,
  capacity,
  exhaustDate,
  endDate,
  label,
  onClick,
  className,
}: ForecastBarProps) {
  const exhaust = toDate(exhaustDate);
  const end = toDate(endDate);
  const safeCapacity = capacity <= 0 ? 1 : capacity;
  const markerPct = Math.min(100, Math.max(0, (projected / safeCapacity) * 100));

  const exhaustBeforeEnd = exhaust && end ? exhaust.getTime() < end.getTime() : null;
  const markerTone =
    exhaustBeforeEnd === true
      ? "bg-[color:var(--custos-red-500)]"
      : "bg-[color:var(--custos-gray-500)]";
  const captionTone =
    exhaustBeforeEnd === true
      ? "text-[color:var(--custos-red-700)]"
      : "text-muted-foreground";

  let caption: string | null = null;
  if (exhaust) {
    if (end) {
      const days = daysBetween(exhaust, end);
      const direction = exhaustBeforeEnd === true ? "before" : "after";
      caption = `Projected exhaust ${formatDate(exhaust)} — ${days} days ${direction} award end`;
    } else {
      caption = `Projected exhaust ${formatDate(exhaust)}`;
    }
  }

  const content = (
    <div className={cn("w-full space-y-1", className)}>
      {label ? <div className="text-xs font-medium text-foreground">{label}</div> : null}
      <div className="relative">
        <BurnDownBar
          used={used}
          projected={projected}
          capacity={capacity}
          ariaLabel={
            label
              ? `${label}: ${used} used of ${capacity}, projected ${projected}`
              : undefined
          }
        />
        <span
          aria-hidden="true"
          data-testid="forecast-marker"
          className={cn(
            "pointer-events-none absolute top-0 h-3 w-0.5 -translate-x-1/2",
            markerTone,
          )}
          style={{ left: `${markerPct}%` }}
        />
      </div>
      {caption ? (
        <div className={cn("text-xs", captionTone)} data-testid="forecast-caption">
          {caption}
        </div>
      ) : null}
    </div>
  );

  if (onClick) {
    return (
      <button
        type="button"
        onClick={onClick}
        className="w-full text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 rounded-md"
      >
        {content}
      </button>
    );
  }

  return content;
}
