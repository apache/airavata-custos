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
import { useAbility } from "@/shared/casl/AbilityProvider";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { NAV_GROUP_LABELS, type NavGroup, type NavItem, portalNav } from "./nav";

const GROUP_ORDER: NavGroup[] = ["allocations", "admin"];

export function Sidebar() {
  const pathname = usePathname();
  const ability = useAbility();

  const visible = portalNav.filter((item) => {
    if (!item.ability) return true;
    return ability.can(item.ability.action, item.ability.subject);
  });

  const groups = GROUP_ORDER.map((group) => ({
    group,
    items: visible.filter((item) => item.group === group),
  })).filter((g) => g.items.length > 0);

  return (
    <aside className="flex w-[240px] shrink-0 flex-col border-r border-border bg-sidebar text-sidebar-foreground">
      <div className="px-6 pt-4 pb-6">
        <Link href="/" className="block">
          <img src="/custos-logo.svg" alt="Custos" className="h-8 w-auto" />
        </Link>
      </div>

      <nav className="flex flex-col">
        {groups.map(({ group, items }, idx) => (
          <div key={group} className={cn("flex flex-col", idx > 0 && "mt-4")}>
            <div className="px-6 pt-2 pb-1 text-[11px] font-semibold uppercase tracking-wider text-custos-gray-400">
              {NAV_GROUP_LABELS[group]}
            </div>
            {items.map((item) => (
              <SidebarLink key={item.href} item={item} active={isActive(pathname, item.href)} />
            ))}
          </div>
        ))}
      </nav>
    </aside>
  );
}

function isActive(pathname: string, href: string): boolean {
  return pathname === href || pathname.startsWith(`${href}/`);
}

function SidebarLink({ item, active }: { item: NavItem; active: boolean }) {
  const Icon = item.icon;
  return (
    <Link
      href={item.href}
      aria-current={active ? "page" : undefined}
      className={cn(
        "relative flex h-11 items-center gap-3 px-6 text-sm font-medium transition",
        active
          ? "bg-[var(--sidebar-active)] font-semibold text-foreground"
          : "text-muted-foreground hover:bg-[var(--sidebar-hover)] hover:text-foreground",
      )}
    >
      <Icon className={cn("h-5 w-5 stroke-[1.75]", active && "text-brand")} />
      <span className="truncate">{item.label}</span>
      {active && <span className="absolute top-2 right-0 bottom-2 w-1 rounded-l-full bg-brand" />}
    </Link>
  );
}
