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
import type { RowTone } from "../../types";

export type StatusPillProps = {
  tone: RowTone;
  label?: string;
  size?: "sm" | "md";
  dotOnly?: boolean;
  className?: string;
};

const DEFAULT_LABEL: Record<RowTone, string> = {
  ok: "ok",
  error: "error",
  "in-progress": "in progress",
  orphaned: "orphaned",
  "no-status": "no status",
};

// Solid dot for ok/error/in-progress, hollow ring for orphaned/no-status —
// matches the "calm vs called-out" hierarchy in the tree.
const TONE_STYLES: Record<
  RowTone,
  { dot: string; pill: string; hollow: boolean; pulse: boolean }
> = {
  ok: {
    dot: "bg-[color:var(--custos-green-500)]",
    pill: "bg-[color:var(--tone-ok-bg)] text-[color:var(--tone-ok-fg)]",
    hollow: false,
    pulse: false,
  },
  error: {
    dot: "bg-[color:var(--custos-red-500)]",
    pill: "bg-[color:var(--tone-error-bg)] text-[color:var(--tone-error-fg)]",
    hollow: false,
    pulse: false,
  },
  "in-progress": {
    dot: "bg-[color:var(--custos-amber-500)]",
    pill: "bg-[color:var(--tone-warn-bg)] text-[color:var(--tone-warn-fg)]",
    hollow: false,
    pulse: true,
  },
  orphaned: {
    dot: "ring-[1.5px] ring-inset ring-[color:var(--muted-foreground)]",
    pill: "bg-muted text-muted-foreground",
    hollow: true,
    pulse: false,
  },
  "no-status": {
    dot: "ring-[1.5px] ring-inset ring-[color:var(--muted-foreground)]",
    pill: "bg-muted text-muted-foreground",
    hollow: true,
    pulse: false,
  },
};

export function StatusPill({
  tone,
  label,
  size = "md",
  dotOnly = false,
  className,
}: StatusPillProps) {
  const styles = TONE_STYLES[tone];
  const displayLabel = label ?? DEFAULT_LABEL[tone];
  const dot = (
    <span
      aria-hidden="true"
      className={cn(
        "h-2 w-2 shrink-0 rounded-full",
        styles.hollow ? "bg-transparent" : styles.dot,
        styles.hollow ? styles.dot : null,
        styles.pulse ? "custos-pulse-dot" : null,
      )}
    />
  );

  if (dotOnly) {
    return (
      <output
        aria-label={`Status: ${displayLabel}`}
        className={cn("inline-flex items-center", className)}
      >
        {dot}
      </output>
    );
  }

  return (
    <output
      aria-label={`Status: ${displayLabel}`}
      className={cn(
        "inline-flex items-center gap-1.5 rounded-md font-semibold whitespace-nowrap leading-none",
        size === "sm" ? "h-5 px-2 text-[11px]" : "h-[22px] px-2 text-xs",
        styles.pill,
        className,
      )}
    >
      {dot}
      {displayLabel}
    </output>
  );
}
