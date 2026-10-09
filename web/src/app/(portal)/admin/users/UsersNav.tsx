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

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useClusterAccounts } from "@/features/core/cluster-accounts/queries";
import { cn } from "@/lib/utils";
import { useAbility } from "@/shared/casl/AbilityProvider";

const TABS = [
  { href: "/admin/users/management", label: "User Management" },
  {
    href: "/admin/users/cluster-accounts",
    label: "Cluster Accounts",
    ability: { action: "read", subject: "Cluster" } as const,
    badge: "pending-cluster-accounts",
  },
  {
    href: "/admin/users/roles",
    label: "Role Management",
    ability: { action: "write", subject: "Role" } as const,
  },
] as const;

// Shows how many accounts wait for approval, so the admin sees it from the
// other tabs too.
function PendingClusterAccountsBadge() {
  const query = useClusterAccounts({ approval_status: "PENDING", limit: 1 });
  const count = query.data?.total ?? 0;
  if (count === 0) return null;
  return (
    <span
      aria-label={`${count} pending approval`}
      className="rounded-full bg-[color:var(--tone-warn-bg)] px-2 py-0.5 text-[11px] font-semibold tabular-nums text-[color:var(--tone-warn-fg)]"
    >
      {count}
    </span>
  );
}

export function UsersNav({ rightSlot }: { rightSlot?: React.ReactNode }) {
  const pathname = usePathname();
  const ability = useAbility();
  const tabs = TABS.filter(
    (tab) => !("ability" in tab) || ability.can(tab.ability.action, tab.ability.subject),
  );

  return (
    <nav
      aria-label="Users & permissions sections"
      className="flex min-h-9 items-end justify-between border-b border-border/60"
    >
      <ul className="-mb-px flex flex-wrap gap-1">
        {tabs.map((tab) => {
          const active = pathname === tab.href || pathname?.startsWith(`${tab.href}/`);
          return (
            <li key={tab.href}>
              <Link
                href={tab.href}
                className={cn(
                  "inline-flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors",
                  active
                    ? "border-b-2 border-brand text-foreground"
                    : "border-b-2 border-transparent text-muted-foreground hover:text-foreground",
                )}
              >
                {tab.label}
                {"badge" in tab ? <PendingClusterAccountsBadge /> : null}
              </Link>
            </li>
          );
        })}
      </ul>
      {rightSlot ?? null}
    </nav>
  );
}
