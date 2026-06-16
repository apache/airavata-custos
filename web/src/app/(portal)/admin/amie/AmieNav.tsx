"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

const TABS = [
  { href: "/admin/amie/packets", label: "Inbox" },
  { href: "/admin/amie/failed", label: "Failed" },
  { href: "/admin/amie/replies", label: "Replies" },
  { href: "/admin/amie/reconcile", label: "Reconcile" },
] as const;

export function AmieNav() {
  const pathname = usePathname();
  return (
    <nav aria-label="AMIE console sections" className="border-b border-border/60">
      <ul className="-mb-px flex flex-wrap gap-1">
        {TABS.map((tab) => {
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
