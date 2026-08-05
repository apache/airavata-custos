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
  StatusBadge,
  type StatusBadgeProps,
  type StatusBadgeVariant,
} from "@/shared/ui/StatusBadge";
import type { ElementType, ReactNode } from "react";

export function MetaRow({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  // Not a <dl> — items aren't true term/definition pairs (status pill has no
  // term, value can be a node, etc.). A labeled group reads correctly for AT.
  return (
    <div
      role="group"
      aria-label="Metadata"
      className={cn(
        "inline-flex flex-wrap items-center gap-y-2 rounded-full bg-brand-tint/60 px-4 py-1 text-sm text-foreground [&>*:not(:first-child)]:ml-3 [&>*:not(:first-child)]:border-l [&>*:not(:first-child)]:border-border [&>*:not(:first-child)]:pl-3",
        className,
      )}
    >
      {children}
    </div>
  );
}

type MetaTone = "success" | "warning" | "danger";

type MetaItemBase = {
  className?: string;
};

type DefaultMetaItem = MetaItemBase & {
  variant?: "default";
  icon?: ElementType;
  label?: string;
  value: ReactNode;
  tone?: never;
};

type StatusMetaItem = MetaItemBase & {
  variant: "status";
  icon?: never;
  label?: never;
  value: ReactNode;
  tone?: MetaTone;
};

export type MetaItemProps = DefaultMetaItem | StatusMetaItem;

const toneToVariant: Record<MetaTone, StatusBadgeVariant> = {
  success: "active",
  warning: "warning",
  danger: "rejected",
};

export function MetaItem(props: MetaItemProps) {
  if (props.variant === "status") {
    const variant: StatusBadgeProps["variant"] = props.tone
      ? toneToVariant[props.tone]
      : "active";
    return <StatusBadge variant={variant} label={String(props.value)} className={props.className} />;
  }

  const { icon: Icon, label, value, className } = props;
  return (
    <div className={cn("flex items-center gap-2", className)}>
      {Icon ? <Icon className="h-4 w-4 stroke-[1.5] text-muted-foreground" /> : null}
      {label ? <span className="text-muted-foreground">{label} :</span> : null}
      <span className="font-medium text-foreground">{value}</span>
    </div>
  );
}
