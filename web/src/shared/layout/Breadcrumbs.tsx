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

type Segment = { label: string; href?: string };

function titleCase(slug: string): string {
  return slug.replace(/-/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

// Max 3 visible segments; collapse the middle when longer so the breadcrumb
// stays single-line at 1440 width.
function collapseSegments(segments: Segment[]): Array<Segment | { ellipsis: true }> {
  if (segments.length <= 3) return segments;
  const first = segments[0];
  const last = segments.at(-1);
  if (!first || !last) return segments;
  return [first, { ellipsis: true }, last];
}

export function Breadcrumbs() {
  const pathname = usePathname() ?? "/";
  const parts = pathname.split("/").filter(Boolean);
  if (parts.length === 0) return null;

  const segments: Segment[] = parts.map((part, i) => {
    const href = `/${parts.slice(0, i + 1).join("/")}`;
    const isLast = i === parts.length - 1;
    return { label: titleCase(part), href: isLast ? undefined : href };
  });

  const visible = collapseSegments(segments);

  return (
    <nav aria-label="Breadcrumb" className="flex items-center text-sm">
      <ol className="flex items-center gap-2">
        {visible.map((item, i) => {
          const isLast = i === visible.length - 1;
          const key = "ellipsis" in item ? "ellipsis" : (item.href ?? item.label);
          return (
            <li key={key} className="flex items-center gap-2">
              {"ellipsis" in item ? (
                <span className="text-muted-foreground">…</span>
              ) : item.href ? (
                <Link
                  href={item.href}
                  className="text-muted-foreground transition-colors hover:text-foreground"
                >
                  {item.label}
                </Link>
              ) : (
                <span aria-current="page" className="font-medium text-foreground">
                  {item.label}
                </span>
              )}
              {!isLast && (
                <span aria-hidden="true" className="text-muted-foreground">
                  /
                </span>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
