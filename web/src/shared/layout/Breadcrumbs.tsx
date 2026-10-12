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

import { Link } from "@/shared/ui/link";
import { useParams, usePathname } from "next/navigation";
import { Fragment } from "react";
import { useBreadcrumbLabels } from "./BreadcrumbLabelsProvider";
import { decodeSegment, normalizeRouteId, routeLabel } from "./nav";

type Segment = { label: string; href?: string };

// Path prefixes with no page of their own.
const UNLINKED = new Set(["/admin"]);

function titleCase(slug: string): string {
  return slug.replace(/-/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

// A detail page's crumb until it registers the record's own label.
const RECORD_LABELS: Record<string, string> = {
  "/allocations": "Allocation",
  "/projects": "Project",
  "/change-requests": "Change request",
  "/admin/amie/packets": "Packet",
  "/admin/traces": "Trace",
};

export function Breadcrumbs() {
  const pathname = usePathname() ?? "/";
  const labels = useBreadcrumbLabels();
  const ids = new Set(
    Object.values(useParams() ?? {})
      .flat()
      .flatMap((id) => (id ? [decodeSegment(id)] : [])),
  );
  const parts = pathname.split("/").filter(Boolean);
  if (parts.length === 0) return null;

  const segments: Segment[] = parts.map((raw, i) => {
    const href = `/${parts.slice(0, i + 1).join("/")}`;
    const part = decodeSegment(raw);
    const isLast = i === parts.length - 1;
    return {
      // Detail pages register their label under the normalized route id.
      label:
        labels[normalizeRouteId(raw) ?? part] ??
        routeLabel(href) ??
        (ids.has(part) ? RECORD_LABELS[`/${parts.slice(0, i).join("/")}`] : undefined) ??
        titleCase(part),
      href: isLast || UNLINKED.has(href) ? undefined : href,
    };
  });

  const n = segments.length;
  // Below lg only the parent and the current label stay ("… / Parent / Current"). Ancestors keep up
  // to 16ch and give way down to 6ch before the current label drops below 8ch.
  return (
    <nav aria-label="Breadcrumb" className="flex min-w-0 flex-1 items-center text-sm">
      <ol className="flex min-w-0 flex-1 items-center gap-2">
        {segments.map((item, i) => {
          const isLast = i === n - 1;
          const ancestor = `min-w-0 truncate text-muted-foreground ${item.href ? "transition-colors hover:text-foreground" : ""}`;
          const shown = i < n - 2 ? "hidden lg:flex" : "flex";
          return (
            <Fragment key={item.href ?? item.label}>
              {i === 0 && n >= 3 ? (
                <li
                  aria-hidden="true"
                  className="flex shrink-0 gap-2 text-muted-foreground lg:hidden"
                >
                  … <span>/</span>
                </li>
              ) : null}
              <li
                className={
                  isLast
                    ? "flex min-w-[8ch] flex-1 basis-0"
                    : `${shown} min-w-[6ch] max-w-[16ch] flex-initial`
                }
              >
                {item.href ? (
                  <Link href={item.href} title={item.label} className={ancestor}>
                    {item.label}
                  </Link>
                ) : isLast ? (
                  <span
                    aria-current="page"
                    title={item.label}
                    className="min-w-0 truncate font-medium text-foreground"
                  >
                    {item.label}
                  </span>
                ) : (
                  <span title={item.label} className={ancestor}>
                    {item.label}
                  </span>
                )}
              </li>
              {!isLast && (
                <li aria-hidden="true" className={`${shown} shrink-0 text-muted-foreground`}>
                  /
                </li>
              )}
            </Fragment>
          );
        })}
      </ol>
    </nav>
  );
}
