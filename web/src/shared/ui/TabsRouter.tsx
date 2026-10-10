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
  setSearchParam,
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
  // Per-panel class, e.g. to let a panel with its own scroll container fill the remaining height.
  panelClassName?: string;
};

export function TabsRouter({
  tabs,
  defaultValue,
  searchParam = "tab",
  className,
  panelClassName,
}: TabsRouterProps) {
  const searchParams = useShallowSearchParams();
  const activeRaw = searchParams.get(searchParam);
  const active = tabs.find((t) => t.value === activeRaw)?.value ?? defaultValue;

  const handleChange = (value: string | number | null) => {
    if (typeof value !== "string") return;
    setSearchParam(searchParams, searchParam, value === defaultValue ? null : value);
  };

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
                "data-[active]:border-b-2 data-[active]:border-brand data-[active]:font-semibold data-[active]:text-foreground",
              )}
            >
              {tab.label}
            </TabsPrimitive.Tab>
          ))}
        </TabsPrimitive.List>
      </div>
      {tabs.map((tab) => (
        <TabsPrimitive.Panel key={tab.value} value={tab.value} className={cn("pt-6", panelClassName)}>
          {tab.content}
        </TabsPrimitive.Panel>
      ))}
    </TabsPrimitive.Root>
  );
}
