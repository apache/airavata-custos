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

import {
  CircleCheckIcon,
  CircleXIcon,
  ClockIcon,
  type LucideIcon,
  RotateCwIcon,
  UnplugIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";
import type { DeliveryState } from "../utils";

const STATE_STYLE: Record<DeliveryState, { label: string; icon: LucideIcon; className: string }> = {
  succeeded: {
    label: "Succeeded",
    icon: CircleCheckIcon,
    className: "bg-[color:var(--tone-ok-bg)] text-[color:var(--tone-ok-fg)]",
  },
  failed: {
    label: "Failed",
    icon: CircleXIcon,
    className: "bg-[color:var(--tone-error-bg)] text-[color:var(--tone-error-fg)]",
  },
  retrying: {
    label: "Retrying",
    icon: RotateCwIcon,
    className: "bg-[color:var(--tone-warn-bg)] text-[color:var(--tone-warn-fg)]",
  },
  waiting: {
    label: "Waiting",
    icon: ClockIcon,
    className: "bg-[color:var(--tone-info-bg)] text-[color:var(--tone-info-fg)]",
  },
  "not-running": {
    label: "Not running",
    icon: UnplugIcon,
    className: "border border-dashed border-border bg-muted text-muted-foreground",
  },
};

export function DeliveryStatusBadge({ state }: { state: DeliveryState }) {
  const { label, icon: Icon, className } = STATE_STYLE[state];
  return (
    <span
      className={cn(
        "inline-flex h-[22px] items-center gap-1 rounded-md px-2 text-xs font-semibold whitespace-nowrap",
        className,
      )}
    >
      <Icon className="size-3.5" aria-hidden />
      {label}
    </span>
  );
}
