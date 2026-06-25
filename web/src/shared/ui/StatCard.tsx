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
import { UsageBar } from "@/shared/ui/UsageBar";
import type { ElementType, ReactNode } from "react";

// Card chrome shared by both variants. Bigger radius + softer shadow keeps the
// 3-up row from competing with the page title.
const cardChrome = "rounded-lg border border-border bg-card p-5 shadow-sm";

// Title row: tight uppercase muted label + optional inline lucide icon.
function StatCardTitle({ icon: Icon, title }: { icon?: ElementType; title: ReactNode }) {
  return (
    <div className="flex items-center gap-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
      {Icon ? <Icon className="h-4 w-4 stroke-[1.5]" /> : null}
      <span>{title}</span>
    </div>
  );
}

type CommonProps = {
  icon?: ElementType;
  title: ReactNode;
  value: ReactNode;
  className?: string;
};

type TextProps = CommonProps & {
  variant?: "text";
  sub?: ReactNode;
  /** @deprecated use `sub` */
  sublabel?: ReactNode;
};

type ProgressProps = CommonProps & {
  variant: "progress";
  percent: number;
};

export type StatCardProps = TextProps | ProgressProps;

export function StatCard(props: StatCardProps) {
  if (props.variant === "progress") {
    const { icon, title, value, percent, className } = props;
    const clamped = Math.max(0, Math.min(100, percent));
    return (
      <div className={cn(cardChrome, className)}>
        <StatCardTitle icon={icon} title={title} />
        <div className="mt-2 flex items-baseline justify-between gap-2">
          <span className="font-display text-3xl font-bold text-foreground tabular-nums">
            {value}
          </span>
          <span className="text-sm font-medium text-muted-foreground tabular-nums">
            {clamped.toFixed(0)}%
          </span>
        </div>
        <UsageBar value={clamped} max={100} size="md" className="mt-3" />
      </div>
    );
  }

  const { icon, title, value, sub, sublabel, className } = props;
  const subLine = sub ?? sublabel;
  return (
    <div className={cn(cardChrome, className)}>
      <StatCardTitle icon={icon} title={title} />
      <div className="mt-2 font-display text-3xl font-bold text-foreground tabular-nums">
        {value}
      </div>
      {subLine ? <div className="mt-2 text-xs text-muted-foreground">{subLine}</div> : null}
    </div>
  );
}

