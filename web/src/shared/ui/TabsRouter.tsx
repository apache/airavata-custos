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
  useCanonicalParams,
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
  // ?tab= names the active tab: the default is omitted, an unknown or unavailable one dropped.
  useCanonicalParams({ [searchParam]: active === defaultValue ? undefined : active });

  // Which ends of a sideways-scrolling tablist hide tabs; those ends fade out as the overflow cue.
  const list = React.useRef<HTMLDivElement>(null);
  const [hidden, setHidden] = React.useState({ start: false, end: false });
  const measure = React.useCallback(() => {
    const el = list.current;
    if (!el) return;
    setHidden({
      start: el.scrollLeft > 0,
      end: el.scrollLeft + el.clientWidth < el.scrollWidth - 1,
    });
  }, []);
  React.useEffect(() => {
    measure();
    window.addEventListener("resize", measure);
    return () => window.removeEventListener("resize", measure);
  }, [measure]);
  const fade = "2rem";
  const maskImage =
    hidden.start || hidden.end
      ? `linear-gradient(to right, ${hidden.start ? "transparent" : "black"}, black ${fade}, black calc(100% - ${fade}), ${hidden.end ? "transparent" : "black"})`
      : undefined;

  const handleChange = (value: string | number | null) => {
    if (typeof value !== "string") return;
    setSearchParam(searchParams, searchParam, value === defaultValue ? null : value);
  };

  return (
    <TabsPrimitive.Root value={active} onValueChange={handleChange} className={cn(className)}>
      {/* At narrow widths the tabs scroll sideways inside the list, over its bottom rule. */}
      <TabsPrimitive.List
        ref={list}
        onScroll={measure}
        style={{ maskImage }}
        className="flex gap-6 overflow-x-auto shadow-[inset_0_-1px_0_var(--border)]"
      >
        {tabs.map((tab) => (
          <TabsPrimitive.Tab
            key={tab.value}
            value={tab.value}
            className={cn(
              // Flush underline pattern per §7.4 — no boxed chrome, the tab
              // rests directly on the list's bottom border.
              "relative inline-flex shrink-0 items-center justify-center whitespace-nowrap pb-3 text-sm font-medium text-muted-foreground transition-colors",
              "hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset",
              "data-[active]:border-b-2 data-[active]:border-brand data-[active]:font-semibold data-[active]:text-foreground",
            )}
          >
            {tab.label}
          </TabsPrimitive.Tab>
        ))}
      </TabsPrimitive.List>
      {tabs.map((tab) => (
        <TabsPrimitive.Panel
          key={tab.value}
          value={tab.value}
          className={cn("pt-6", panelClassName)}
        >
          {tab.content}
        </TabsPrimitive.Panel>
      ))}
    </TabsPrimitive.Root>
  );
}
