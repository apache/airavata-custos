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

import { Menu as MenuPrimitive } from "@base-ui/react/menu";
import { Check, ChevronDown } from "lucide-react";
import type { ReactNode } from "react";

import { cn } from "@/lib/utils";
import { Button } from "@/shared/ui/button";

export type GroupByOption = {
  value: string;
  label: string;
};

export type GroupByChipProps = {
  label: string;
  value: string;
  options: GroupByOption[];
  onChange: (value: string) => void;
  className?: string;
};

export function GroupByChip({ label, value, options, onChange, className }: GroupByChipProps) {
  const current = options.find((o) => o.value === value);
  return (
    <MenuPrimitive.Root>
      <MenuPrimitive.Trigger
        render={
          <Button variant="outline" size="sm" className={className}>
            <span className="text-muted-foreground">{label}:</span>
            <span className="font-medium">{current?.label ?? value}</span>
            <ChevronDown aria-hidden="true" className="h-3 w-3 stroke-[1.5]" />
          </Button>
        }
      />
      <MenuPrimitive.Portal>
        <MenuPrimitive.Positioner sideOffset={4} className="z-50 outline-none">
          <MenuPrimitive.Popup
            className={cn(
              "z-50 min-w-40 rounded-lg bg-popover p-1 text-popover-foreground shadow-md ring-1 ring-foreground/10",
              "data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0",
            )}
          >
            {options.map((opt) => {
              const isCurrent = opt.value === value;
              return (
                <MenuPrimitive.Item
                  key={opt.value}
                  onClick={() => onChange(opt.value)}
                  className="relative flex cursor-default items-center gap-2 rounded-md py-1 pr-2 pl-7 text-sm outline-hidden select-none focus:bg-accent focus:text-accent-foreground"
                  data-current={isCurrent ? "" : undefined}
                >
                  {isCurrent ? (
                    <Check aria-hidden="true" className="absolute left-2 h-3.5 w-3.5 stroke-[1.5]" />
                  ) : null}
                  <span>{opt.label}</span>
                </MenuPrimitive.Item>
              );
            })}
          </MenuPrimitive.Popup>
        </MenuPrimitive.Positioner>
      </MenuPrimitive.Portal>
    </MenuPrimitive.Root>
  );
}

export function GroupByChipGroup({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <div
      role="group"
      aria-label="Group by"
      className={cn("flex flex-wrap items-center gap-2", className)}
    >
      {children}
    </div>
  );
}
