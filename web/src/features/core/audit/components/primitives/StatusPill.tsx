/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { cn } from "@/lib/utils";
import {
  CircleCheckIcon,
  CircleDashedIcon,
  CircleXIcon,
  ClockIcon,
  type LucideIcon,
  RotateCwIcon,
} from "lucide-react";
import type { ListStatus } from "../../types";
import { STATUS_WORDS } from "../../utils";

export type PillStatus = ListStatus | "ok" | "error" | "muted";

export type StatusPillProps = {
  status: PillStatus;
  label?: string;
  size?: "sm" | "md";
  className?: string;
};

const STYLES: Record<PillStatus, { icon: LucideIcon; className: string; word: string }> = {
  failed: {
    icon: CircleXIcon,
    className: "bg-[color:var(--tone-error-bg)] text-[color:var(--tone-error-fg)]",
    word: STATUS_WORDS.failed,
  },
  error: {
    icon: CircleXIcon,
    className: "bg-[color:var(--tone-error-bg)] text-[color:var(--tone-error-fg)]",
    word: STATUS_WORDS.failed,
  },
  retrying: {
    icon: RotateCwIcon,
    className: "bg-[color:var(--tone-warn-bg)] text-[color:var(--tone-warn-fg)]",
    word: STATUS_WORDS.retrying,
  },
  waiting: {
    icon: ClockIcon,
    className: "bg-[color:var(--tone-info-bg)] text-[color:var(--tone-info-fg)]",
    word: STATUS_WORDS.waiting,
  },
  done: {
    icon: CircleCheckIcon,
    className: "bg-[color:var(--tone-ok-bg)] text-[color:var(--tone-ok-fg)]",
    word: STATUS_WORDS.done,
  },
  ok: {
    icon: CircleCheckIcon,
    className: "bg-[color:var(--tone-ok-bg)] text-[color:var(--tone-ok-fg)]",
    word: STATUS_WORDS.done,
  },
  muted: {
    icon: CircleDashedIcon,
    className: "border border-dashed border-border bg-muted text-muted-foreground",
    word: "No status",
  },
};

export function StatusPill({ status, label, size = "md", className }: StatusPillProps) {
  const { icon: Icon, className: tone, word } = STYLES[status];
  const text = label ?? word;
  return (
    <output
      aria-label={`Status: ${text}`}
      className={cn(
        "inline-flex items-center gap-1 rounded-md font-semibold whitespace-nowrap leading-none tabular-nums",
        size === "sm" ? "h-5 px-1.5 text-[11px]" : "h-[22px] px-2 text-xs",
        tone,
        className,
      )}
    >
      <Icon className={size === "sm" ? "size-3" : "size-3.5"} aria-hidden="true" />
      {text}
    </output>
  );
}
