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

import { brand } from "@/shared/brand";
import type { AppAbility } from "@/shared/casl/abilities";
import {
  Activity,
  Building2,
  ChartColumn,
  ClipboardCheck,
  ClipboardList,
  FolderKanban,
  HardDrive,
  KeyRound,
  type LucideIcon,
  Send,
  Server,
  UserCog,
} from "lucide-react";
import type { Metadata } from "next";

// The one name of each route, shared by the sidebar, tabs, page headings, breadcrumbs and tab titles.
export const ROUTE_LABELS = {
  "/allocations": "Allocations",
  "/projects": "Projects",
  "/analytics": "Analytics",
  "/settings": "Settings",
  "/change-requests": "Change requests",
  "/admin/users": "Users & permissions",
  "/admin/users/management": "User management",
  "/admin/users/cluster-accounts": "Cluster accounts",
  "/admin/users/roles": "Role management",
  "/admin/traces": "Traces",
  "/admin/events": "Events",
  "/admin/amie": "AMIE",
  "/admin/amie/packets": "Inbox",
  "/admin/amie/failed": "Failed",
  "/admin/temp-accounts": "Temporary accounts",
  "/admin/organizations": "Organizations",
  "/admin/resources": "Resources & clusters",
} as const;

type RoutePath = keyof typeof ROUTE_LABELS;

export function routeLabel(href: string): string | undefined {
  return ROUTE_LABELS[href as RoutePath];
}

// Next hands dynamic segments over still percent-encoded; a malformed escape stays as typed.
export function decodeSegment(segment: string): string {
  try {
    return decodeURIComponent(segment);
  } catch {
    return segment;
  }
}

// A UUID or a 32-hex trace id.
const HEX_ID = /^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|[0-9a-f]{32})$/i;

// A ?param id as pasted (URLSearchParams has already decoded it): trimmed, hex ids lowercased;
// blank becomes undefined.
export function normalizeQueryId(value: string | null | undefined): string | undefined {
  const id = (value ?? "").trim();
  if (!id) return undefined;
  return HEX_ID.test(id) ? id.toLowerCase() : id;
}

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

// A ?param that must name a record by UUID: normalised, else undefined, so no request is made for it.
export function queryUuid(value: string | null | undefined): string | undefined {
  const id = normalizeQueryId(value);
  return id && UUID.test(id) ? id : undefined;
}

const TRACE_ID = /^[0-9a-f]{32}$/;

// A ?param that must name a trace by its 32-hex id: normalised, else undefined.
export function queryTraceId(value: string | null | undefined): string | undefined {
  const id = normalizeQueryId(value);
  return id && TRACE_ID.test(id) ? id : undefined;
}

// A raw route segment id: decoded once, then normalised like a ?param id.
export const normalizeRouteId = (raw: string | null | undefined) =>
  normalizeQueryId(decodeSegment(raw ?? ""));

type AbilityCheck = { action: string; subject: string };

export type NavGroup = "allocations" | "admin";

export type NavItem = {
  href: string;
  label: string;
  icon: LucideIcon;
  group: NavGroup;
  // Shown when the caller holds any of the listed abilities.
  ability?: AbilityCheck | AbilityCheck[];
};

export const NAV_GROUP_LABELS: Record<NavGroup, string> = {
  allocations: "My allocations",
  admin: "Site administration",
};

// The Users & permissions tabs, each shown to a caller who can open it: write-only callers get the
// create controls there, and grant-only callers the roles page's direct privilege holders card.
export const USERS_TABS = [
  {
    href: "/admin/users/management",
    ability: [
      { action: "read", subject: "User" },
      { action: "write", subject: "User" },
    ],
  },
  {
    href: "/admin/users/cluster-accounts",
    ability: [
      { action: "read", subject: "Cluster" },
      { action: "write", subject: "Cluster" },
    ],
  },
  {
    href: "/admin/users/roles",
    ability: [
      { action: "write", subject: "Role" },
      { action: "write", subject: "PrivilegeGrant" },
    ],
  },
] as const satisfies Array<{ href: RoutePath; ability: AbilityCheck[] }>;

export const portalNav: NavItem[] = [
  // No ability gates on this group: the backend scopes these reads to the
  // caller's own memberships.
  {
    href: "/allocations",
    label: ROUTE_LABELS["/allocations"],
    icon: Server,
    group: "allocations",
  },
  {
    href: "/projects",
    label: ROUTE_LABELS["/projects"],
    icon: FolderKanban,
    group: "allocations",
  },
  {
    href: "/analytics",
    label: ROUTE_LABELS["/analytics"],
    icon: ChartColumn,
    group: "allocations",
  },
  {
    href: "/admin/users",
    label: ROUTE_LABELS["/admin/users"],
    icon: UserCog,
    group: "admin",
    ability: USERS_TABS.flatMap<AbilityCheck>((tab) => tab.ability),
  },
  {
    href: "/admin/traces",
    label: ROUTE_LABELS["/admin/traces"],
    icon: Activity,
    group: "admin",
    ability: { action: "read", subject: "Trace" },
  },
  {
    href: "/admin/events",
    label: ROUTE_LABELS["/admin/events"],
    icon: Send,
    group: "admin",
    ability: { action: "read", subject: "Trace" },
  },
  {
    href: "/admin/amie",
    label: ROUTE_LABELS["/admin/amie"],
    icon: ClipboardList,
    group: "admin",
    ability: { action: "read", subject: "AmiePacket" },
  },
  {
    href: "/admin/temp-accounts",
    label: ROUTE_LABELS["/admin/temp-accounts"],
    icon: KeyRound,
    group: "admin",
    ability: [
      { action: "read", subject: "TempAccount" },
      { action: "write", subject: "TempAccount" },
    ],
  },
  {
    href: "/admin/organizations",
    label: ROUTE_LABELS["/admin/organizations"],
    icon: Building2,
    group: "admin",
    ability: [
      { action: "read", subject: "Organization" },
      { action: "write", subject: "Organization" },
    ],
  },
  {
    href: "/change-requests",
    label: ROUTE_LABELS["/change-requests"],
    icon: ClipboardCheck,
    group: "admin",
    ability: { action: "read", subject: "Allocation" },
  },
  {
    href: "/admin/resources",
    label: ROUTE_LABELS["/admin/resources"],
    icon: HardDrive,
    group: "admin",
    ability: [
      { action: "read", subject: "Cluster" },
      { action: "write", subject: "Cluster" },
      { action: "read", subject: "Allocation" },
      { action: "write", subject: "Allocation" },
    ],
  },
];

export function isNavItemVisible(
  item: { ability?: AbilityCheck | readonly AbilityCheck[] },
  ability: AppAbility,
): boolean {
  return !item.ability || [item.ability].flat().some((a) => ability.can(a.action, a.subject));
}

// "<Page> · Admin" for the site administration pages, "<Page> · <brand>" for the caller's own.
export function pageMetadata(path: RoutePath): Metadata {
  const item = portalNav.find((i) => path === i.href || path.startsWith(`${i.href}/`));
  const area = item?.group === "admin" ? "Admin" : brand.name;
  return { title: `${ROUTE_LABELS[path]} · ${area}` };
}

const ROUTE_IDS: Array<[string, (id: string) => string | undefined]> = [
  ["/allocations", queryUuid],
  ["/projects", queryUuid],
  ["/change-requests", queryUuid],
  ["/admin/amie/packets", queryUuid],
  ["/admin/traces", queryTraceId],
];

// A path with its detail id canonicalised (trimmed, lowercased); null when the id cannot name a
// record. Middleware redirects to the first and serves the second as not found, on the server.
export function canonicalPath(path: string): string | null {
  for (const [route, parse] of ROUTE_IDS) {
    const raw = path.startsWith(`${route}/`) ? path.slice(route.length + 1) : undefined;
    if (raw === undefined || raw.includes("/")) continue;
    const id = parse(decodeSegment(raw));
    return id === undefined ? null : `${route}/${id}`;
  }
  return path;
}
