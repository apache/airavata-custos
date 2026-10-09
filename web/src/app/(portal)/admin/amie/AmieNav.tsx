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
import { cn } from "@/lib/utils";
import type * as React from "react";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { ErrorState } from "@/shared/ui/ErrorState";
import { AMIE_TABS } from "./tabs";

// Shared AMIE console page: heading, section tabs, and a read gate on the section's subject.
export function AmiePage({
  title,
  description,
  subject,
  children,
}: {
  title: string;
  description: React.ReactNode;
  subject: (typeof AMIE_TABS)[number]["subject"];
  children: React.ReactNode;
}) {
  const canRead = useAbility().can("read", subject);
  return (
    <div className="space-y-4">
      <header className="space-y-1">
        <h1 className="font-display text-[28px] font-bold leading-tight">{title}</h1>
        <p className="text-sm text-muted-foreground">{description}</p>
      </header>
      <AmieNav />
      {canRead ? children : <ErrorState message="Not permitted." />}
    </div>
  );
}

function AmieNav() {
  const pathname = usePathname();
  const ability = useAbility();
  const tabs = AMIE_TABS.filter((tab) => ability.can("read", tab.subject));
  return (
    <nav aria-label="AMIE console sections" className="border-b border-border/60">
      <ul className="-mb-px flex flex-wrap gap-1">
        {tabs.map((tab) => {
          const active = pathname === tab.href || pathname?.startsWith(`${tab.href}/`);
          return (
            <li key={tab.href}>
              <Link
                href={tab.href}
                className={cn(
                  "inline-flex items-center px-4 py-2 text-sm font-medium transition-colors",
                  active
                    ? "border-b-2 border-brand text-foreground"
                    : "border-b-2 border-transparent text-muted-foreground hover:text-foreground",
                )}
              >
                {tab.label}
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
