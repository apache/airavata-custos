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

import { RefreshCw } from "lucide-react";

import { cn } from "@/lib/utils";

export type LastSyncedBadgeProps = {
  syncedAt: Date | string;
  onRefetch?: () => void;
  /** Override current time for deterministic rendering / tests. */
  now?: Date;
  className?: string;
};

const MIN_MS = 60 * 1000;
const HOUR_MS = 60 * MIN_MS;
const DAY_MS = 24 * HOUR_MS;

type Tint = "fresh" | "stale" | "expired";

function tintFor(ageMs: number): Tint {
  if (ageMs < HOUR_MS) return "fresh";
  if (ageMs < DAY_MS) return "stale";
  return "expired";
}

function formatAbbreviated(ageMs: number): string {
  if (ageMs < MIN_MS) return "just now";
  if (ageMs < HOUR_MS) return `${Math.floor(ageMs / MIN_MS)}m ago`;
  if (ageMs < DAY_MS) return `${Math.floor(ageMs / HOUR_MS)}h ago`;
  return `${Math.floor(ageMs / DAY_MS)}d ago`;
}

const tintStyles: Record<Tint, string> = {
  fresh: "bg-[color:var(--custos-green-50)] text-[color:var(--custos-green-700)]",
  stale: "bg-[color:var(--custos-amber-50)] text-[color:var(--custos-amber-700)]",
  expired: "bg-[color:var(--custos-red-50)] text-[color:var(--custos-red-700)]",
};

export function LastSyncedBadge({
  syncedAt,
  onRefetch,
  now,
  className,
}: LastSyncedBadgeProps) {
  const syncedDate = syncedAt instanceof Date ? syncedAt : new Date(syncedAt);
  const reference = now ?? new Date();
  const ageMs = Math.max(0, reference.getTime() - syncedDate.getTime());
  const tint = tintFor(ageMs);

  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium",
        tintStyles[tint],
        className,
      )}
      data-tint={tint}
    >
      <span aria-hidden="true" className="h-2 w-2 rounded-full bg-current" />
      <span>Synced {formatAbbreviated(ageMs)}</span>
      {onRefetch ? (
        <button
          type="button"
          onClick={onRefetch}
          aria-label="Refetch data"
          className="ml-0.5 inline-flex items-center justify-center rounded-full p-0.5 text-current/80 transition-colors hover:bg-current/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-current/50"
        >
          <RefreshCw className="h-3 w-3 stroke-[1.5]" />
        </button>
      ) : null}
    </span>
  );
}
