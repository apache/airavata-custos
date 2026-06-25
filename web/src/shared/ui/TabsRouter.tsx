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
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { Tabs as TabsPrimitive } from "@base-ui/react/tabs";
import * as React from "react";

export type TabsRouterTab = {
  value: string;
  label: React.ReactNode;
  content: React.ReactNode;
};

export type TabsRouterProps = {
  tabs: TabsRouterTab[];
  defaultValue: string;
  searchParam?: string;
  className?: string;
  // Per-panel class override. Use to make a specific tab fill remaining
  // height (`min-h-0 flex-1 flex flex-col`) instead of the default natural
  // height — needed for tabs with their own internal scroll container.
  panelClassName?: string | Record<string, string | undefined>;
  /**
   * Action node rendered on the right of the tab strip, sharing the same
   * bottom border. When set as a record keyed by tab value, the right-slot
   * content swaps to match the active tab; tabs without an entry render nothing.
   */
  rightSlot?: React.ReactNode | Record<string, React.ReactNode | undefined>;
};

export function TabsRouter({
  tabs,
  defaultValue,
  searchParam = "tab",
  className,
  panelClassName,
  rightSlot,
}: TabsRouterProps) {
  const searchParams = useShallowSearchParams();
  const activeRaw = searchParams.get(searchParam);
  const active = tabs.some((t) => t.value === activeRaw) ? (activeRaw as string) : defaultValue;

  const handleChange = (value: string | number | null) => {
    if (typeof value !== "string") return;
    const params = new URLSearchParams(searchParams.toString());
    if (value === defaultValue) params.delete(searchParam);
    else params.set(searchParam, value);
    replaceShallowSearchParams(params);
  };

  // Plain ReactNode (React elements include `$$typeof`) vs the per-tab record
  // share the same union — narrow by looking for that React-element marker.
  const isReactElement = (node: unknown): node is React.ReactNode =>
    node === null ||
    typeof node === "string" ||
    typeof node === "number" ||
    typeof node === "boolean" ||
    (typeof node === "object" && node !== null && "$$typeof" in node) ||
    Array.isArray(node);

  const resolvedRightSlot: React.ReactNode =
    rightSlot && typeof rightSlot === "object" && !isReactElement(rightSlot)
      ? ((rightSlot as Record<string, React.ReactNode | undefined>)[active] ?? null)
      : (rightSlot ?? null);

  return (
    <TabsPrimitive.Root value={active} onValueChange={handleChange} className={cn(className)}>
      <div className="flex items-center justify-between border-b border-border">
        <TabsPrimitive.List className="flex gap-6">
          {tabs.map((tab) => (
            <TabsPrimitive.Tab
              key={tab.value}
              value={tab.value}
              className={cn(
                // Flush underline pattern per §7.4 — no boxed chrome, the tab
                // rests directly on the list's bottom border.
                "relative -mb-px inline-flex items-center justify-center pb-3 text-sm font-medium text-muted-foreground transition-colors",
                "hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                "data-[active]:border-b-2 data-[active]:border-brand data-[active]:font-semibold data-[active]:text-brand",
              )}
            >
              {tab.label}
            </TabsPrimitive.Tab>
          ))}
        </TabsPrimitive.List>
        {resolvedRightSlot ? <div className="pb-3">{resolvedRightSlot}</div> : null}
      </div>
      {tabs.map((tab) => {
        const perPanel =
          panelClassName && typeof panelClassName === "object"
            ? panelClassName[tab.value]
            : panelClassName;
        return (
          <TabsPrimitive.Panel
            key={tab.value}
            value={tab.value}
            className={cn("pt-6", perPanel)}
          >
            {tab.content}
          </TabsPrimitive.Panel>
        );
      })}
    </TabsPrimitive.Root>
  );
}
