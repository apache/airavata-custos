"use client";

import { cn } from "@/lib/utils";
import { ChevronRightIcon } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import * as React from "react";

export type BreadcrumbItem = {
  label: string;
  href?: string;
};

export type BreadcrumbsProps = {
  items?: BreadcrumbItem[];
  className?: string;
};

function fromPathname(pathname: string): BreadcrumbItem[] {
  const segments = pathname.split("/").filter(Boolean);
  return segments.map((segment, i) => {
    const href = `/${segments.slice(0, i + 1).join("/")}`;
    const label = segment.replace(/-/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
    return { label, href: i === segments.length - 1 ? undefined : href };
  });
}

export function Breadcrumbs({ items, className }: BreadcrumbsProps) {
  const pathname = usePathname();
  const resolved = items ?? fromPathname(pathname ?? "/");
  if (resolved.length === 0) return null;
  return (
    <nav aria-label="Breadcrumb" className={cn("flex text-sm text-muted-foreground", className)}>
      <ol className="flex items-center gap-1">
        {resolved.map((item, i) => (
          <li key={`${item.label}-${i}`} className="flex items-center gap-1">
            {item.href ? (
              <Link
                href={item.href}
                className="rounded transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              >
                {item.label}
              </Link>
            ) : (
              <span aria-current="page" className="text-foreground">
                {item.label}
              </span>
            )}
            {i < resolved.length - 1 ? (
              <ChevronRightIcon className="size-3" aria-hidden="true" />
            ) : null}
          </li>
        ))}
      </ol>
    </nav>
  );
}
