"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

const TABS = [
  { href: "/admin/users/management", label: "User Management" },
  { href: "/admin/users/roles", label: "Role Management" },
] as const;

export function UsersNav({ rightSlot }: { rightSlot?: React.ReactNode }) {
  const pathname = usePathname();
  return (
    <nav
      aria-label="Users & permissions sections"
      className="flex min-h-9 items-end justify-between border-b border-border/60"
    >
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
      {rightSlot ?? null}
    </nav>
  );
}
